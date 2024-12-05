package keycloak

import (
	"encoding/json"
	"fmt"
	"golangProject/config"
	"io"
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

//var keycloakResponse KeycloakResponse

func TokenExchangeHandler(c echo.Context) error {
	// Parse Google ID token from the request
	idToken := c.FormValue("id_token")
	if idToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID token is required",
		})
	}

	cfg := config.GetConfig()

	keycloakTokenUrl := fmt.Sprintf("%s:8080/realms/%s/protocol/openid-connect/token", cfg.Server, cfg.Realm)

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
			"error": fmt.Sprintf("Error decoding response: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  clientKeycloak.AccessToken,
		"refresh_token": clientKeycloak.RefreshToken,
	})
}
