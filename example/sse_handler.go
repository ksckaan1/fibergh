package main

import (
	"context"
	"fmt"
	"time"
)

type sseReq struct {
	Email  string `json:"email"`
	UserID string `header:"user-id"`
}

type sseMsg struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func sseHandler(ctx context.Context, req *sseReq, send func(name string, msg sseMsg) error) error {
	fmt.Println("req received", req)

	for range 10 {
		err := send("message", sseMsg{
			Name: "Kaan",
			Age:  29,
		})
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}
