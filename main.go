package main

import (
	"web-server/model"

	"github.com/gofiber/fiber/v2"
	//"gorm.io/gorm/logger"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func routers(app *fiber.App) {

	//Auth Routes
	app.Post("/auth/login", model.Login)
	app.Post("/auth/signup", model.SignUp)

	//User Routes
	app.Get("/user", model.Protected, model.Restricted("admin"), model.GetAllUsers)
	app.Patch("/user/:id", model.UpdateUser)
	// app.Delete("/user/:id", DeleteUser)
}

func main() {
	model.InitMigration()
	app := fiber.New()

	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome To My First Go Web Backend!!!")
	})

	routers(app)

	app.Listen(":3000")
}
