package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ksckaan1/fibergh"
)

func main() {
	router := fiber.New()

	router.Post("/standart", fibergh.GH(standartHandler))
	router.Post("/sse", fibergh.GHforSSE(5*time.Second, sseHandler))

	err := router.Listen(":3030")
	if err != nil {
		log.Fatal(err)
	}
}
