package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golangProject/config"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

// Define the response structure for the token
type KeycloakResponse struct {
	AccessToken  string `json:"access_token"` //keep storing the token until it expires
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

var currentToken *KeycloakResponse //store a pointer to a KeycloakResponse struct instance

// This function will handle the /keycloak route
func KeycloakHandler(c echo.Context) error {
	// Assuming you're fetching some data from Keycloak here
	accessToken, err := GetAccessToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not get access token"})
	}
	return c.JSON(http.StatusOK, map[string]string{"access_token": accessToken})
}

func GetAccessToken() (string, error) {
	url := "http://localhost:8080/realms/{realm}/protocol/openid-connect/token"
	// Create the request body for the token request
	reqBody := map[string]string{
		"client_id":     "admin-cli",
		"client_secret": os.Getenv("CLIENT_SECRET"),
		"grant_type":    "client_credentials",
	}
	jsonData, _ := json.Marshal(reqBody)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	var keycloakResp KeycloakResponse
	if err := json.NewDecoder(resp.Body).Decode(&keycloakResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return keycloakResp.AccessToken, nil
}

func TokenExchangeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse Google ID token from the request
	idToken := r.FormValue("id_token")
	if idToken == "" {
		http.Error(w, "ID token is required", http.StatusBadRequest)
		return
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
		http.Error(w, fmt.Sprintf("Error making request to Keycloak: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading Keycloak response: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Keycloak returned an error: %s", string(body)), http.StatusInternalServerError)
		return
	}

	var keycloakResponse KeycloakResponse
	if err := json.Unmarshal(body, &keycloakResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing Keycloak response: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the tokens
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": keycloakResponse.AccessToken,
	})
}

func CheckAccessToken() (string, error) {
	// If token is expired, refresh it
	if currentToken != nil && isTokenExpired(currentToken) {
		// Refresh the token (by calling Keycloak's token endpoint)
		newToken, err := refreshToken()
		if err != nil {
			return "", err
		}
		currentToken = newToken
	}

	if currentToken == nil {
		// Get the token for the first time
		token, err := GetAccessToken()
		if err != nil {
			return "", err
		}
		currentToken = token
	}

	return currentToken.AccessToken, nil
}

func isTokenExpired(token *KeycloakResponse) bool {
	// Simple check based on expiration time (current time vs. expiration)
	expirationTime := time.Now().Unix() - int64(token.ExpiresIn)
	return expirationTime > 0
}

func refreshToken() (*KeycloakResponse, error) {
	// Logic to refresh the token if a refresh token is available
}
