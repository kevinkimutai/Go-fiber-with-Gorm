package model

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func comparePasswords(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Check If Email Exists
func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func createJWT() {}

func SignUp(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	//Hash Password
	hashedPwd, err := hashPassword(user.Password)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	user.Password = hashedPwd

	//CREATE USER
	if err := DB.Create(user).Error; err != nil {
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

func Login(c *fiber.Ctx) error {
	loginUser := new(loginRequest)
	user := new(User)

	if err := c.BodyParser(loginUser); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	if loginUser.Email == "" || loginUser.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   "Missing Required Fields",
		})
	}

	if err := DB.Where("email = ?", &loginUser.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	//Compare passwords
	matched := comparePasswords(loginUser.Password, user.Password)

	fmt.Println("Matched", matched)

	if !matched {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
			"error":   "Wrong Email or Password",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
	})

}
