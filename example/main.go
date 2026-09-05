package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/ksckaan1/fibergh"
)

func main() {
	router := fiber.New()

	router.Post("/example", fibergh.GH(ExampleHandler))

	err := router.Listen(":3030")
	if err != nil {
		log.Fatal(err)
	}
}

type exampleReq struct {
	Name      string `json:"name"`
	Age       int    `json:"age"`
	UserID    string `header:"user_id"`
	SessionID string `cookie:"session_id"`
}

type exampleResp struct {
	Name      string `json:"name"`
	Age       int    `json:"age"`
	UserID    string `header:"User-ID" json:"-"`
	SessionID string `cookie:"session_id,2h45m" json:"-"`
}

func ExampleHandler(ctx context.Context, req *exampleReq) (*exampleResp, int, error) {
	return &exampleResp{
		Name:      req.Name,
		Age:       req.Age,
		UserID:    req.UserID,
		SessionID: req.SessionID,
	}, http.StatusOK, nil
}
