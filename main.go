package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

	// Retrieve the secret key from .env
	secretKey := os.Getenv("SECRET_KEY")
	fmt.Printf("SECRET_KEY: %s\n", secretKey)

	// Define your routes
	e.POST("/encrypt", encryptHandler)

	// Start the server on port 8010
	fmt.Println("Server is running on port 8010")
	if err := e.Start(":8010"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func encryptHandler(c echo.Context) error {
	var requestData map[string]string
	if err := json.NewDecoder(c.Request().Body).Decode(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	invitationCode := requestData["invitation_code"]
	//fmt.Println("Received invitation code:", invitationCode)

	// Encrypt the invitation code
	encryptedCode, err := encryptCode(invitationCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	fmt.Println("Received encrypted code:", encryptedCode)

	// Perform your encryption or other logic here
	//return c.JSON(http.StatusOK, map[string]string{"message": "Invitation code processed", "encrypted_code": encryptedCode})
	return c.JSON(http.StatusOK, map[string]string{"encrypted_code": encryptedCode})
}

func encryptCode(code string) (string, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if len(secretKey) > 32 {
		secretKey = secretKey[:32] // Truncate to 32 bytes for AES-256
	}

	// Ensure secretKey length matches AES key size (16, 24, or 32 bytes)
	key := []byte(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Encrypt the code
	plaintext := []byte(code)                                // The input code that needs to be encrypted, converted into a byte slice.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext)) // Allocate space for the ciphertext. It needs enough space to store the IV (Initialization Vector) + the encrypted data.
	iv := ciphertext[:aes.BlockSize]                         // Randomly generate the IV. This ensures that each encryption is unique even with the same plaintext.
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)                // Create a new CFB encrypter. It uses the AES block cipher and the generated IV.
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext) // Encrypt the plaintext by applying the XORKeyStream to the plaintext bytes. The result is written to ciphertext, starting from the position after the IV.

	// Return the encrypted code as a base64-encoded string
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}
