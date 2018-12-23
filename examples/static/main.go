// Copyright 2018 xgfone <xgfone@126.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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