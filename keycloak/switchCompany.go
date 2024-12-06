package keycloak

// Define the response structure for the token
type SwitchCompany struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var switchCompany SwitchCompany
