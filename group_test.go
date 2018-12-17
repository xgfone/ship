package ship

import (
	"testing"
)

func TestGroup(t *testing.T) {
	s := New(Config{Prefix: "/v1"})
	group := s.Group("/group")
	group.Route("/route1", NothingHandler()).GET()
	group.Route("/route2", NothingHandler()).POST()

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
