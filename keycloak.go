package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

// Define the response structure for the token
type KeycloakResponse struct {
	AccessToken string `json:"access_token"`
}

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
