package model

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

func verifyJWT(tokenString string) (jwt.MapClaims, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	JWTSecretKey := os.Getenv("JWT_SECRET_KEY")

	key := []byte(JWTSecretKey)

	// Parse and verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the token using the key
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if token.Valid {
		// Access token claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			return claims, nil
		}
		return nil, fmt.Errorf("token claims are not valid")
	} else {
		return nil, fmt.Errorf("token is not valid")
	}
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

func Protected(c *fiber.Ctx) error {
	var token string
	startsWith := "Bearer"
	authHeader := c.Get("Authorization")
	user := new(User)

	if authHeader != "" && strings.HasPrefix(authHeader, startsWith) {
		// Split the Authorization Into Array
		token = strings.Fields(authHeader)[1]

	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing Token,Login To Continue",
			"Error":   "Unauthorized",
		})
	}

	claims, err := verifyJWT(token)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Wrong Token",
			"Error":   "Unauthorized",
		})
	}

	if err := DB.First(user, claims["sub"]).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Handle the case where the user with the given ID is not found
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "User not found with that token",
				"error":   "Unauthorized",
			})
		}
		// Handle other database errors
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	c.Locals("userRole", claims["role"])

	return c.Next()

}

func Restricted(allowedRoles ...string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("userRole")

		// Check if the user's role is in the allowedRoles slice.
		allowed := false

		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		// If the user's role is not in the allowedRoles slice, return a Forbidden response.
		if !allowed {
			return c.Status(fiber.StatusForbidden).SendString("Forbidden, Login with proper rights")
		}

		return c.Next()
	}
}
