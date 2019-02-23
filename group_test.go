// Copyright 2018 xgfone
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

package ship

import (
	"testing"
)

func TestGroup(t *testing.T) {
	s := New(SetPrefix("/v1"))
	group := s.Group("/group")
	group.Route("/route1").GET(NothingHandler())
	group.Route("/route2").POST(NothingHandler())

	i := 0
	s.Traverse(func(name, method, path string) {
		switch i {
		case 0:
			if name != "" || method != "GET" || path != "/v1/group/route1" {
				t.Fail()
			}
			i++
		case 1:
			if name != "" || method != "POST" || path != "/v1/group/route2" {
				t.Fail()
			}
			i++
		default:
			t.Fail()
		}
	})
}
