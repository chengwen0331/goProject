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
		Server:       "http://192.168.0.230",
		Realm:        "G-SSO-Connect",
		ClientID:     "frontend-login",
		ClientSecret: "intkmDx6ZK7QdCtQShUudO0q6z5mQBmb",
		//ClientUrlID:  "ed8241ee-7bd3-45eb-b7e8-d9cb09f79236",
		RedirectURI: "http://192.168.11.27:9080/",
	}
}
