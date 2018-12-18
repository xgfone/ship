package main

import (
	"net/http"

	"github.com/xgfone/ship"
)

func setupRouter() *ship.Ship {
	router := ship.New()
	router.Route("/ping").GET(func(ctx ship.Context) error {
		return ctx.JSON(200, map[string]interface{}{"message": "pong"})
	})

	return router
}

func main() {
	router := setupRouter()
	http.ListenAndServe(":8080", router)
}
