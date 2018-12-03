package main

import (
	"net/http"

	"github.com/xgfone/ship"
)

func main() {
	router := ship.NewRouter()

	router.Get("/ping", func(ctx ship.Context) error {
		return ctx.JSON(200, map[string]interface{}{"message": "pong"})
	})

	http.ListenAndServe(":8080", router)
}
