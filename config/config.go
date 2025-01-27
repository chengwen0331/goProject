package config

type MyConfig struct {
	Server       string
	Realm        string
	ClientID     string
	ClientSecret string
	ClientUrlID  string
	RedirectURI  string
}

// GetConfig initializes and returns a MyConfig instance
func GetConfig() *MyConfig {
	return &MyConfig{
		Server:       "http://192.168.11.27",
		Realm:        "fml",
		ClientID:     "fml_client",
		ClientSecret: "8V2hrYbQoNS7y0eAjDv8X9kCzvGaMWu8",
		ClientUrlID:  "ed8241ee-7bd3-45eb-b7e8-d9cb09f79236",
		RedirectURI:  "http://192.168.11.27:9080/",
	}
}
