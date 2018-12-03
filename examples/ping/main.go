package main

import (
	"net/http"

	"github.com/xgfone/ship"
)

func setupRouter() ship.Router {
	router := ship.NewRouter()
	router.Get("/ping", func(ctx ship.Context) error {
		return ctx.JSON(200, map[string]interface{}{"message": "pong"})
	})

	return router
}

func main() {
	router := setupRouter()
	http.ListenAndServe(":8080", router)
}
