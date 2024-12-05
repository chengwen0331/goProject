package keycloak

// Define the response structure for the token
type KeycloakResponse struct {
	AccessToken  string `json:"access_token"` //keep storing the token until it expires
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
