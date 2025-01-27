package keycloak

import (
	//"encoding/json"
	"fmt"
	"golangProject/config"
	"net/http"

	"github.com/labstack/echo/v4"
)

func FetchGroupInfoHandler(c echo.Context) error {
	cfg := config.GetConfig()

	// Get the group ID from the request
	groupID := c.Param("groupId")
	if groupID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Group ID is required",
		})
	}

	// Access token retrieval (replace with your actual implementation)
	accessToken := clientKeycloak.AccessToken
	if accessToken == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Access token is missing",
		})
	}

	// Construct the Keycloak group info URL
	groupInfoURL := fmt.Sprintf("%s:8080/admin/realms/%s/groups/%s", cfg.Server, cfg.Realm, groupID)

	// Create the HTTP request
	req, err := http.NewRequest("GET", groupInfoURL, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to create request: %v", err),
		})
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to send request: %v", err),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.JSON(resp.StatusCode, map[string]string{
			"error": fmt.Sprintf("Failed to fetch group info: %s", resp.Status),
		})
	}

	// Parse the response body
	// var groupInfo GroupInfo
	// if err := json.NewDecoder(resp.Body).Decode(&groupInfo); err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{
	// 		"error": fmt.Sprintf("Failed to parse response: %v", err),
	// 	})
	// }

	// Return group info as a JSON response
	return c.JSON(http.StatusOK, "groupInfo")
}
