package main

import (
	"log"
	"os"

	"example.com/go-demo/config"
	"example.com/go-demo/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success":     true,
			"message":     "** You are at the root endpoint **",
			"github_repo": "https://github.com/R-D-G/Covid19API-s",
		})
	})

	api := app.Group("/api")

	routes.CoviddatasRoute(api.Group("/Coviddatas"))
	routes.GPSLocationRoute(api.Group("/GPSbasedcoviddata"))
}

func main() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	config.ConnectDB()

	setupRoutes(app)

	port := os.Getenv("PORT")
	err := app.Listen(":" + port)

	if err != nil {
		log.Fatal("Error app failed to start")
		panic(err)
	}
}
