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
