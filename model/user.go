package model

import (
	"errors"
	"strings"
	"web-server/database"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" gorm:"unique"`
}

func AddUser(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	//CREATE USER
	if err := database.DB.Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Conflict",
				"error":   "Email already exists.",
			})
		}

		// Handle other database errors
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// Check If Email Exists
func isDuplicateKeyError(err error) bool {
	// Assuming you're using GORM with a PostgreSQL database
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func GetAllUsers(c *fiber.Ctx) error {
	users := new([]User)

	//GET ALL USERS
	if err := database.DB.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})

	}

	return c.Status(200).JSON(&users)
}

func UpdateUser(c *fiber.Ctx) error {
	user := new(User)

	// First, find the user by ID
	if err := database.DB.First(user, c.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Handle the case where the user with the given ID is not found
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}
		// Handle other database errors
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	// Parse the request body to update user fields
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   err.Error(),
		})
	}

	// Save the updated user to the database
	if err := database.DB.Save(user).Error; err != nil {
		// Handle database save error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)

}
