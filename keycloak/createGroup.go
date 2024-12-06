package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golangProject/config"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CreateGroupRequest represents the request structure for creating a group
type CreateGroupRequest struct {
	Name       string              `json:"name"`
	Attributes map[string][]string `json:"attributes"`
}

var createGroup CreateGroupRequest

// CreateGroup creates a group in Keycloak
func CreateGroup(c echo.Context) error {
	cfg := config.GetConfig()
	// Fetch the access token from the frontend or other secure source
	//accessToken := clientKeycloak.AccessToken
	if clientKeycloak.AccessToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing access token",
		})
	}

	//fmt.Println("Keycloak access token retrieved:", clientKeycloak.AccessToken)

	// Read the group data from the request body
	if err := c.Bind(&createGroup); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Failed to bind request: %v", err),
		})
	}

	// Construct the Keycloak URL to create a group
	url := fmt.Sprintf("%s:8080/admin/realms/%s/groups", cfg.Server, cfg.Realm)

	// Prepare the request body
	groupData, err := json.Marshal(createGroup)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to marshal request data: %v", err),
		})
	}

	// Send the POST request to Keycloak to create the group
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(groupData))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to create request: %v", err),
		})
	}

	// Set headers for authentication and content type
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+clientKeycloak.AccessToken)

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to execute request: %v", err),
		})
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode == http.StatusCreated {
		// Return success response without parsing the body
		groupResp, err := FetchGroupAndAddSubGroups(c)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to add subgroups: %v", err),
			})
		}

		// Return success response without parsing the body
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":  "Group created successfully",
			"group_id": groupResp,
		})
	} else {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to create group: %v", resp.Status),
		})
	}
}

func FetchGroupAndAddSubGroups(c echo.Context) (string, error) {
	cfg := config.GetConfig()

	// Access token from your configuration or secure source
	//accessToken := clientKeycloak.AccessToken
	if clientKeycloak.AccessToken == "" {
		return "", fmt.Errorf("Missing access token")
	}

	fmt.Println("Create Group name:", createGroup.Name)
	// Fetch the group by name
	groupFetchUrl := fmt.Sprintf("%s:8080/admin/realms/%s/groups?search=%s", cfg.Server, cfg.Realm, createGroup.Name)
	req, err := http.NewRequest("GET", groupFetchUrl, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+clientKeycloak.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to fetch group: %v", resp.Status)
	}

	// Decode the response to get the group ID
	var groups []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return "", fmt.Errorf("Failed to decode response: %v", err)
	}

	var groupID string
	for _, group := range groups {
		if group["name"] == createGroup.Name {
			groupID, _ = group["id"].(string)
			break
		}
	}
	if groupID == "" {
		return "", fmt.Errorf("Group with the name %s not found", createGroup.Name)
	}

	// Return the group ID
	return groupID, nil
}

// AddSubGroups function to create subgroups
func AddSubGroups(c echo.Context) error {
	cfg := config.GetConfig()

	if clientKeycloak.AccessToken == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authorization token is missing",
		})
	}

	var data map[string]interface{}
	if err := c.Bind(&data); err != nil {
		fmt.Printf("Error binding data: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	fmt.Printf("Request Data: %+v\n", data)

	// Extract groupID and access_token from the data
	groupID, ok := data["group_id"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "group_id is required",
		})
	}

	subGroups := []string{"Owner", "Admin", "Account", "Delivery Man", "Packer"}

	for _, subgroup := range subGroups {
		url := fmt.Sprintf("%s:8080/admin/realms/%s/groups/%s/children", cfg.Server, cfg.Realm, groupID)
		subgroupData := map[string]interface{}{
			"name": subgroup,
		}
		payload, err := json.Marshal(subgroupData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to marshal subgroup data: %v", err),
			})
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to create request: %v", err),
			})
		}
		req.Header.Set("Authorization", "Bearer "+clientKeycloak.AccessToken)
		req.Header.Set("Content-Type", "application/json") //ensures proper formatting of the response data sent to the client.

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to execute request: %v", err)
		}

		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Request to Keycloak: %s\nPayload: %s\nResponse: %s\n", url, string(payload), string(respBody))

		if resp.StatusCode != http.StatusCreated {
			fmt.Printf("Failed to create subgroup %s: %s\n", subgroup, string(respBody))
			continue
			// return c.JSON(http.StatusInternalServerError, map[string]string{
			// 	"error": fmt.Sprintf("Failed to create subgroup %s: %s - %s", subgroup, resp.Status, string(respBody)),
			// })
		}
		fmt.Printf("Subgroup \"%s\" created successfully.\n", subgroup)

		// Extract subgroup ID from response and assign roles
		var createdSubgroup map[string]interface{}
		if err := json.Unmarshal(respBody, &createdSubgroup); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to parse subgroup creation response: %v", err),
			})
		}

		subgroupID, ok := createdSubgroup["id"].(string)
		if !ok || subgroupID == "" {
			fmt.Println("Failed to retrieve subgroup ID.")
			continue
		}

		// Assign role to the created subgroup
		err = GetRoleAndAssignToGroup(subgroup, subgroupID)
		if err != nil {
			fmt.Printf("Failed to assign role to subgroup \"%s\": %v\n", subgroup, err)
		}
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Subgroups created successfully",
	})
}
