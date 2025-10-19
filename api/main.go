package main

import (
	"fmt"
	"log"
	"os"
	"url_shortner_service/routes" // importing the routes

	"github.com/gofiber/fiber/v2"                   // importing fiber
	"github.com/gofiber/fiber/v2/middleware/logger" // importing logger , logger is used to log the requests in the terminal
	"github.com/joho/godotenv"                      // importing godotenv
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL) // v1 is used so that app using the old url dont break suddenly when new url is formed
}

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	app := fiber.New() // new fiber app just like express

	app.Use(logger.New())

	setupRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))

}
