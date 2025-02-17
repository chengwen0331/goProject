package keycloak

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func EncryptHandler(c echo.Context) error {
	var requestData map[string]string
	if err := json.NewDecoder(c.Request().Body).Decode(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	groupID := requestData["groupID"]
	subgroupID := requestData["subgroupID"]
	//fmt.Println("Received invitation code:", invitationCode)

	validity := time.Duration(5 * time.Minute)
	expiry := time.Now().Add(validity).UnixMilli()
	// Convert to milliseconds since epoch if needed

	// Concatenate into invitation code
	invitationCode := fmt.Sprintf("%s/%s/%d", groupID, subgroupID, expiry)

	// Output the result
	fmt.Println("Invitation Code:", invitationCode)

	// Encrypt the invitation code
	encryptedCode, err := encryptCode(invitationCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	fmt.Println("Received ccode:", encryptedCode)

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
	fmt.Println("Received cccode:", ciphertext[aes.BlockSize:])
	// Return the encrypted code as a base64-encoded string
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptHandler(c echo.Context) error {
	secretKey := os.Getenv("SECRET_KEY")
	fmt.Println("Secret Key:", secretKey)
	// fmt.Println(len([]byte("b'h\x94\xd6\xcag\x99\xf0_\xe0_j\xf4\xd8\x07\xf2\xbf\r\xf8\xc1\xfb\xa1kb@\xbb\x7f\xfd\x88\xe4'")))

	if len(secretKey) > 32 {
		secretKey = secretKey[:32] // Truncate to 32 bytes for AES-256
	}

	// Read the encrypted code from the request body
	var requestData map[string]string
	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error1": err.Error()})
	}

	encryptedCode := requestData["invitation_code"]
	decodedCode, err := decryptCode(encryptedCode, secretKey)
	if err != nil {
		fmt.Println("Decryption error:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error2": "invalid code"})
	}
	fmt.Println("Decryption code:", decodedCode)

	valid, err := validateCode(decodedCode)
	if err != nil || !valid {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "code invalid or expired"})
	}

	// Extract parts from the decoded code
	parts := strings.Split(decodedCode, "/")

	fmt.Println("group id:", parts[0])
	fmt.Println("role id:", parts[1])

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Code is valid!",
		"part_0":  parts[0],
		"part_1":  parts[1]})
}

func decryptCode(encryptedCode, secretKey string) (string, error) {
	fmt.Printf("Received encryptedCode using secret key: %s\n", encryptedCode)

	key := []byte(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.URLEncoding.DecodeString(encryptedCode)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func validateCode(decodedCode string) (bool, error) {
	parts := strings.Split(decodedCode, "/")
	if len(parts) != 3 {
		return false, errors.New("invalid code format")
	}

	// expiryTimestamp, err := time.Parse("20060102150405", parts[2])

	// Parse the Unix timestamp (milliseconds) and convert it to time
	timestampMillis, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return false, fmt.Errorf("error parsing timestamp: %v", err)
	}
	expiryTimestamp := time.Unix(0, timestampMillis*int64(time.Millisecond))

	fmt.Println("Decoded code parts:", parts)
	fmt.Println("Expiry timestamp part:", parts[2])

	if err != nil {
		return false, err
	}

	// Format the expiry timestamp into a human-readable string
	expiryFormatted := expiryTimestamp.Format("2006-01-02 15:04:05") // Change the format as needed

	// Log the expiry and current time
	fmt.Println("Code expiry time:", expiryFormatted)

	fmt.Println("Current time:", time.Now())

	if time.Now().After(expiryTimestamp) {
		return false, errors.New("code expired")
	}

	return true, nil
}
