package keycloak

// Define the response structure for the token
type SwitchCompany struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	//Path       string              `json:"path"`
	Attributes map[string][]string `json:"attributes"`
}

var switchCompany SwitchCompany
