package keycloak

import (
	"fmt"
	"golangProject/config"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// LogoutHandler is the handler function to log out the user
func LogoutHandler(c echo.Context) error {
	// // Retrieve the access token and refresh token from the request headers or body
	// accessToken := c.Request().Header.Get("Authorization")
	// refreshToken := c.FormValue("refresh_token") // Or fetch from the request body

	// // Remove 'Bearer ' from the access token if present
	// if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
	// 	accessToken = accessToken[7:]
	// }

	// Call the Logout function with the access token and refresh token
	err := Logout()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("Logout failed: %v", err)})
	}

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{"message": "Successfully logged out"})
}

func Logout() error {
	cfg := config.GetConfig()

	fmt.Println("Keycloak refresh token retrieved:", keycloakResponse.RefreshToken)

	// Construct the logout URL
	logoutUrl := fmt.Sprintf("%s:8080/realms/%s/protocol/openid-connect/logout", cfg.Server, cfg.Realm)

	// Prepare the request to Keycloak
	data := url.Values{}
	data.Add("client_id", cfg.ClientID)
	data.Add("client_secret", cfg.ClientSecret)
	data.Add("refresh_token", keycloakResponse.RefreshToken)

	// Create the HTTP request
	req, err := http.NewRequest("POST", logoutUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the authorization header with the access token
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+keycloakResponse.AccessToken)

	// Create a client and make the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if the logout was successful
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Successfully logged out.")
		return nil
	} else {
		return fmt.Errorf("logout failed with status: %d, response: %s", resp.StatusCode, string(body))
	}
}
