package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golangProject/config"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"time"

	"github.com/labstack/echo/v4"
)

// // Define the response structure for the token
// type KeycloakResponse struct {
// 	AccessToken  string `json:"access_token"` //keep storing the token until it expires
// 	RefreshToken string `json:"refresh_token"`
// 	ExpiresIn    int    `json:"expires_in"`
// 	TokenType    string `json:"token_type"`
// }

// Struct to hold the response from Keycloak for user search
type KeycloakUserData struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Struct for linking external identity provider (Google)
type IdentityProviderLink struct {
	IdentityProvider string `json:"identityProvider"` // Name of the identity provider (e.g., "google")
	UserID           string `json:"userId"`           // Keycloak user ID
	ExternalUserID   string `json:"userName"`         // Google user ID
}

//var keycloakResponse KeycloakResponse

func TokenExchangeHandler(c echo.Context) error {
	// Parse Google ID token from the request
	idToken := c.FormValue("id_token")
	googleEmail := c.FormValue("google_email")
	googleID := c.FormValue("google_id")
	fmt.Println("Google Access Token is" + idToken)
	fmt.Println("Google ID is" + googleID)
	if idToken == "" && googleEmail == "" && googleID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID token is required",
		})
	}

	cfg := config.GetConfig()
	accessToken, err := GetKeycloakToken()
	if accessToken == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error getting access token to Keycloak: %v"),
		})
	}
	user, err := getUserByEmail(accessToken, googleEmail)
	if user != nil {
		linking, err := getFederatedIdentities(accessToken, user.ID)
		if err != nil {
			fmt.Println("no linking")
			log.Fatalf("Error fetching federated identities: %v", err)
		}
		if linking == nil || len(linking) == 0 {
			// Step 2: If the user exists, link Google account to the existing user
			err := linkGoogleAccountToUser(accessToken, user.ID, googleID, user.Username)
			if err != nil {
				log.Fatalf("Error linking Google account to Keycloak user: %v", err)
			}
			fmt.Printf("User already exists and Google account linked with ID: %s\n", user.ID)
		}
	}

	//get access token
	keycloakTokenUrl := fmt.Sprintf("%s:8080/realms/%s/protocol/openid-connect/token", cfg.Server, cfg.Realm)
	fmt.Println("fml")
	// Prepare the request to Keycloak
	form := url.Values{}
	form.Add("client_id", cfg.ClientID)
	form.Add("client_secret", cfg.ClientSecret)
	form.Add("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	form.Add("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	form.Add("subject_token", idToken)
	form.Add("subject_issuer", "google")
	form.Add("redirect_uri", cfg.RedirectURI)

	// Make the request to Keycloak
	resp, err := http.PostForm(keycloakTokenUrl, form)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error making request to Keycloak: %v", err),
		})
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error reading Keycloak response: %v", err),
		})
	}

	if resp.StatusCode != http.StatusOK {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Keycloak returned an error: %s", string(body)),
		})
	}

	if err := json.Unmarshal(body, &keycloakResponse); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error parsing Keycloak response: %v", err),
		})
	}

	// Explicitly set the content-type if needed (though c.JSON sets it by default to application/json)
	c.Response().Header().Set("Content-Type", "application/x-www-form-urlencoded")

	// Respond with the access token
	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  keycloakResponse.AccessToken,
		"refresh_token": keycloakResponse.RefreshToken,
	})
}

func CheckAccessToken() error { //check access token before executing any action
	// If token is expired, refresh it
	if keycloakResponse.AccessToken != "" && isTokenExpired() {
		// Refresh the token (by calling Keycloak's token endpoint)
		err := refreshToken()
		if err != nil {
			return fmt.Errorf("failed to refresh token: %v", err)
		}
	}

	if keycloakResponse.AccessToken == "" {
		return fmt.Errorf("no access token found, user needs to log in again")
	}

	return nil
}

// Function to get user from Keycloak by email
func getUserByEmail(keycloakAdminToken, email string) (*KeycloakUserData, error) {

	cfg := config.GetConfig()
	url := fmt.Sprintf("%s:8080/admin/realms/%s/users?email=%s", cfg.Server, cfg.Realm, email)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get email: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+keycloakAdminToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user from Keycloak, status: %s", resp.Status)
	}

	var users []KeycloakUserData
	body, _ := io.ReadAll(resp.Body)

	fmt.Println("Response Body:", string(body))

	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Parsed Users: %+v\n", users)
	if len(users) > 0 {
		return &users[0], nil
	}
	fmt.Println("No users found in the response")
	return nil, nil
}

func isTokenExpired() bool {

	// Store the token issuance time
	tokenIssuedAt := time.Now()

	// Calculate expiration time by adding the expires_in duration to the current time
	tokenExpiration := tokenIssuedAt.Add(time.Second * time.Duration(keycloakResponse.ExpiresIn))

	// Check if the token is still valid
	if time.Now().Before(tokenExpiration) {
		fmt.Println("Access token is still valid.")
		return false
	} else {
		fmt.Println("Access token has expired.")
		return true
	}
}

func refreshToken() error {

	cfg := config.GetConfig()

	if keycloakResponse.RefreshToken == "" {
		return fmt.Errorf("refresh token is required but missing")
	}

	// Set up the request to Keycloak for refreshing the token
	keycloakTokenUrl := fmt.Sprintf("%s:8080/realms/%s/protocol/openid-connect/token", cfg.Server, cfg.Realm)

	// Prepare the request to Keycloak
	form := url.Values{}
	form.Add("client_id", cfg.ClientID)
	form.Add("client_secret", cfg.ClientSecret)
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", keycloakResponse.RefreshToken)

	// Make the request to Keycloak
	req, err := http.NewRequest("POST", keycloakTokenUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to refresh token, response: %s", string(body))
	}

	err = json.Unmarshal(body, &keycloakResponse)

	if err != nil {
		return fmt.Errorf("failed to parse response body: %v", err)
	}

	// Tokens are now updated in keycloakResponse
	fmt.Println("Access token and refresh token have been updated.")
	return nil

}

// get client access token
func GetAccessToken(c echo.Context) error {
	cfg := config.GetConfig()
	// Construct the Keycloak token URL
	keycloakTokenUrl := fmt.Sprintf("%s:8080/realms/%s/protocol/openid-connect/token", cfg.Server, cfg.Realm)

	// Prepare the data for the POST request
	data := url.Values{}
	data.Add("client_id", cfg.ClientID)
	data.Add("client_secret", cfg.ClientSecret)
	data.Add("grant_type", "client_credentials")

	// Make the POST request to Keycloak
	resp, err := http.PostForm(keycloakTokenUrl, data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error making request to Keycloak: %v", err),
		})
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error reading Keycloak response: %v", err),
		})
	}

	if resp.StatusCode != http.StatusOK {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Keycloak returned an error: %s", string(body)),
		})
	}

	// Decode the response body into a map
	if err := json.Unmarshal(body, &clientKeycloak); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error decoding response: %s", string(body)),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  clientKeycloak.AccessToken,
		"refresh_token": clientKeycloak.RefreshToken,
	})
	//return clientKeycloak.AccessToken, nil
}

func GetKeycloakToken() (string, error) {
	cfg := config.GetConfig()
	// Construct the Keycloak token URL
	clientTokenUrl := fmt.Sprintf("%s:8080/realms/%s/protocol/openid-connect/token", cfg.Server, cfg.Realm)

	// Prepare the data for the POST request
	data := url.Values{}
	data.Add("client_id", cfg.ClientID)
	data.Add("client_secret", cfg.ClientSecret)
	data.Add("grant_type", "client_credentials")

	// Make the POST request to Keycloak
	resp, err := http.PostForm(clientTokenUrl, data)
	if err != nil {
		fmt.Printf("HTTP Request Error: %v\n", err)
		return "", fmt.Errorf("Error making request to Keycloak: %v", err)
	}

	resp.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading Keycloak response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Keycloak returned an error: %s", string(body))
	}

	if err := json.Unmarshal(body, &clientKeycloak); err != nil {
		return "", fmt.Errorf("Error decoding response: %s", string(body))
	}

	return clientKeycloak.AccessToken, nil
}

// Function to link the Google account to the Keycloak user
func linkGoogleAccountToUser(keycloakAdminToken, keycloakUserID, googleUserID, keycloakUsername string) error {
	fmt.Println("keycloak id" + keycloakUserID)
	fmt.Println("google id" + googleUserID)
	cfg := config.GetConfig()
	url := fmt.Sprintf("%s:8080/admin/realms/%s/users/%s/federated-identity/google", cfg.Server, cfg.Realm, keycloakUserID)
	link := IdentityProviderLink{
		IdentityProvider: "google", // Google is the identity provider
		UserID:           googleUserID,
		ExternalUserID:   keycloakUsername,
	}

	linkData, err := json.Marshal(link)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(linkData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+keycloakAdminToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// if resp.StatusCode != http.StatusCreated {
	// 	return fmt.Errorf("failed to link Google account to user, status: %s", resp.Status)
	// }

	fmt.Println("Google account linked successfully to Keycloak user.")
	return err
}

func getFederatedIdentities(keycloakAdminToken, keycloakUserID string) ([]IdentityProviderLink, error) {
	cfg := config.GetConfig() // Replace with your configuration getter function

	// Construct the URL for fetching federated identities
	url := fmt.Sprintf("%s:8080/admin/realms/%s/users/%s/federated-identity", cfg.Server, cfg.Realm, keycloakUserID)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}

	// Set the Authorization header with the admin token
	req.Header.Set("Authorization", "Bearer "+keycloakAdminToken)

	// Create an HTTP client and set the timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request to Keycloak: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to retrieve federated identities, status: %s", resp.Status)
	}

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %v", err)
	}

	// Parse the response body into a slice of FederatedIdentity
	var identities []IdentityProviderLink
	err = json.Unmarshal(body, &identities)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling response: %v", err)
	}

	// Return the list of federated identities
	return identities, nil
}
