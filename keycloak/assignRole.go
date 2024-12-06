package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golangProject/config"
	"net/http"
)

// get subgroup ID
func GetRoleAndAssignToGroup(subGroupName, groupID string) error {
	cfg := config.GetConfig()

	// Map subgroup names to role names
	roleName := map[string]string{
		"Owner":       "owner",
		"Admin":       "admin",
		"Account":     "account",
		"Packer":      "packer",
		"DeliveryMan": "deliveryman",
	}[subGroupName]

	if roleName == "" {
		return fmt.Errorf("role not found for subgroup: %s", subGroupName)
	}

	// Get role details
	roleFetchUrl := fmt.Sprintf(
		"%s:8080/admin/realms/%s/clients/%s/roles/%s", cfg.Server, cfg.Realm, cfg.ClientUrlID, roleName,
	)

	req, err := http.NewRequest("GET", roleFetchUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create role fetch request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+clientKeycloak.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute role fetch request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch role: %v", resp.Status)
	}

	var roleData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&roleData); err != nil {
		return fmt.Errorf("failed to decode role response: %v", err)
	}

	roleID, ok := roleData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid role ID format")
	}

	// Assign role to group
	assignRoleUrl := fmt.Sprintf(
		"%s:8080/admin/realms/%s/groups/%s/role-mappings/clients/%s",
		cfg.Server, cfg.Realm, groupID, cfg.ClientUrlID,
	)

	rolePayload := []map[string]interface{}{
		{
			"id":   roleID,
			"name": roleName,
		},
	}
	payload, err := json.Marshal(rolePayload)
	if err != nil {
		return fmt.Errorf("failed to create role assign request: %v", err)
	}
	assignReq, err := http.NewRequest("POST", assignRoleUrl, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create role assign request: %v", err)
	}
	assignReq.Header.Set("Authorization", "Bearer "+clientKeycloak.AccessToken)
	assignReq.Header.Set("Content-Type", "application/json")

	assignResp, err := client.Do(assignReq)
	if err != nil {
		return fmt.Errorf("failed to execute role assign request: %v", err)
	}
	defer assignResp.Body.Close()

	if assignResp.StatusCode != http.StatusNoContent {
		//respBody, _ := io.ReadAll(assignResp.Body)
		return fmt.Errorf("failed to assign role: %v", resp.Status)
	}

	// Optional: Add user to the group if role is "admin"
	if subGroupName == "Owner" {
		if err := AddUserToGroup(groupID); err != nil {
			return fmt.Errorf("Failed to add user to group: %v", err)
		}
	}

	fmt.Printf("Role \"%s\" assigned to subgroup \"%s\" successfully.\n", roleName, subGroupName)
	return nil
}
