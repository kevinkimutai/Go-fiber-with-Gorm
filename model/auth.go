package model

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
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

func createJWT(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"name": user.LastName,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	JWTSecretKey := os.Getenv("JWT_SECRET_KEY")

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

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

	if !matched {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
			"error":   "Wrong Email or Password",
		})
	}

	//GENERATE JWT
	jwt, err := createJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Error",
			"error":   "Something went wrong when generating JWT token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   jwt,
	})

}
