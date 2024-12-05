package keycloak

import (
	"encoding/json"
	"fmt"
	"golangProject/config"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// UserInfo represents the structure of the user info response
type UserInfo struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Username   string `json:"preferred_username"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	// Add other fields as necessary
}

var keycloakData KeycloakResponse

func GetUserInfoHandler(c echo.Context) error {
	// Call the FetchUserInfo function
	userInfo, err := FetchUserInfo()
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	// Return user info to the frontend
	return c.JSON(http.StatusOK, userInfo)
}

// FetchUserInfo fetches user information using the access token
func FetchUserInfo() (*UserInfo, error) {
	cfg := config.GetConfig()

	// Construct the user info endpoint URL
	userInfoUrl := fmt.Sprintf("%s:8080/realms/%s/protocol/openid-connect/userinfo", cfg.Server, cfg.Realm)

	// Make the HTTP GET request
	req, err := http.NewRequest("GET", userInfoUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+keycloakData.AccessToken)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user info, status: %d, response: %s", resp.StatusCode, body)
	}

	// Decode the user info response
	var userInfo UserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &userInfo, nil
}
