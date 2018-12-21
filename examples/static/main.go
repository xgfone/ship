package main

import (
	"flag"
	"net/http"

	"github.com/xgfone/ship"
)

var (
	localDir string
	showDir  bool
)

func setupRouter(rpath, localpath string, showDir bool) *ship.Ship {
	router := ship.New()
	if showDir {
		router.Route(rpath).StaticFS(http.Dir(localDir))
	} else {
		router.Route(rpath).Static(localpath)
	}

	return router
}

func main() {
	flag.BoolVar(&showDir, "showdir", false, "If true, show the directory")
	flag.StringVar(&localDir, "dir", ".", "The local directory to be served")
	flag.Parse()

	router := setupRouter("/static", localDir, showDir)
	http.ListenAndServe(":8080", router)
}
