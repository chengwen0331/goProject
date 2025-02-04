package handler

import (
	"encoding/json"
	"golangProject/database"
	"golangProject/modal"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SaveUser(c echo.Context) error {
	var user modal.User
	if err := json.NewDecoder(c.Request().Body).Decode(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO User (company_id, permission, email, address, first_name, last_name, nationality, role, city, gender, phone, status, zipcode) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (email) DO NOTHING",
		2,
		"{}",
		user.Email,
		"Address 11",
		user.FirstName,
		user.LastName,
		"Nationality 11",
		"Role 0",
		"City 11",
		"Female",
		"123456",
		true,
		"ZIP1011",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database insert error"})
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction commit failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Products saved successfully"})
}
