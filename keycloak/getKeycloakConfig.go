package keycloak

import (
	"golangProject/config"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetKeycloakInfoHandler returns Keycloak config data
func GetKeycloakInfoHandler(c echo.Context) error {
	cfg := config.GetConfig()

	return c.JSON(http.StatusOK, map[string]string{
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
	})
}
