package main

import (
	"fmt"
	"golangProject/keycloak"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	// Initialize Echo instance
	e := echo.New()

	// Enable CORS with middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8000"}, // Replace with the URL of your frontend
		AllowMethods: []string{echo.GET, echo.POST},     // Allow only GET and POST methods
	}))

	// Load the .env file
	envFilePath := "C:/Users/User/Documents/.env" // Use absolute path
	err := godotenv.Load(envFilePath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Set up the POST route for token exchange
	e.POST("/token", keycloak.TokenExchangeHandler)

	// Set up the POST route to get user info
	e.GET("/userinfo", keycloak.GetUserInfoHandler)

	// Define the /logout route and bind it to the LogoutHandler
	e.POST("/logout", keycloak.LogoutHandler)

	//get client access token
	e.POST("/client", keycloak.GetAccessToken)

	//create group
	e.POST("/create-group", keycloak.CreateGroup)

	//get groupid
	e.POST("/create-subgroup", keycloak.AddSubGroups)

	// Define your routes
	e.POST("/encrypt", keycloak.EncryptHandler)

	e.POST("/decrypt", keycloak.DecryptHandler)

	// Start the server on port 8010
	fmt.Println("Server is running on port 8010")
	if err := e.Start(":8010"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
