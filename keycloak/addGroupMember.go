package keycloak

import (
	"fmt"
	"golangProject/config"
	"io"
	"net/http"
)

func AddUserToGroup(groupID string) error {
	cfg := config.GetConfig()
	fmt.Printf("UserInfo: ", userInfo.Sub)

	// Construct the URL to add a user to a group in Keycloak
	url := fmt.Sprintf("%s:8080/admin/realms/%s/users/%s/groups/%s",
		cfg.Server, cfg.Realm, userInfo.Sub, groupID)

	// Create the PUT request to add the user to the group
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the required headers, including the authorization token
	req.Header.Set("Authorization", "Bearer "+clientKeycloak.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code for successful addition (204 No Content)
	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("User added to group successfully.")
		return nil
	}

	// If the request fails, print the error
	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("failed to add user to group: %s - %s", resp.Status, string(respBody))
}
