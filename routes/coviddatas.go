package routes

import (
	"example.com/go-demo/controllers" // replace
	"github.com/gofiber/fiber/v2"
)

func CoviddatasRoute(route fiber.Router) {
	route.Get("/", controllers.FetchAndStoreCoviddata)
}

func GPSLocationRoute(route fiber.Router) {
	route.Get("/", controllers.FetchCoviddataAtLocation)
}
