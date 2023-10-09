package main

import (
	"web-server/database"
	"web-server/model"

	"github.com/gofiber/fiber/v2"
)

func routers(app *fiber.App) {
	app.Get("/users", model.GetAllUsers)
	app.Post("/user", model.AddUser)
	app.Patch("/user/:id", model.UpdateUser)
	// app.Delete("/user/:id", DeleteUser)
}

func main() {
	database.InitMigration()
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome To My First Go Web Backend!!!")
	})

	routers(app)

	app.Listen(":3000")
}
