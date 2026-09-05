package main

import (
	"context"
	"net/http"
)

type standartReq struct {
	Name      string `json:"name"`
	Age       int    `json:"age"`
	UserID    string `header:"user_id"`
	SessionID string `cookie:"session_id"`
}

type standartResp struct {
	Name      string `json:"name"`
	Age       int    `json:"age"`
	UserID    string `header:"User-ID" json:"-"`
	SessionID string `cookie:"session_id,2h45m" json:"-"`
}

func standartHandler(ctx context.Context, req *standartReq) (*standartResp, int, error) {
	return &standartResp{
		Name:      req.Name,
		Age:       req.Age,
		UserID:    req.UserID,
		SessionID: req.SessionID,
	}, http.StatusOK, nil
}
