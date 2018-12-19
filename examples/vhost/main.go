package main

import (
	"net/http"

	"github.com/xgfone/ship"
)

func setupRouter() *ship.Ship {
	router := ship.New()
	router.Route("/router").GET(func(c ship.Context) error { return c.String(200, "default") })

	vhost1 := router.VHost("host1.example.com")
	vhost1.Route("/router").GET(func(c ship.Context) error { return c.String(200, "vhost1") })

	vhost2 := router.VHost("host2.example.com")
	vhost2.Route("/router").GET(func(c ship.Context) error { return c.String(200, "vhost2") })

	return router
}

func main() {
	router := setupRouter()
	http.ListenAndServe(":8080", router)
}
