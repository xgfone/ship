package ship

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

const middlewareoutput = `
pre m1 start
pre m2 start
pre m2 end
pre m1 end
use m1 start
use m2 start
group m1 start
group m2 start
route m1 start
route m2 start
route m2 end
route m1 end
group m2 end
group m1 end
use m2 end
use m1 end
`

func TestMiddleware(t *testing.T) {
	bs := bytes.NewBufferString("\n")
	s := New()
	s.Pre(func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("pre m1 start\n")
			err := next(ctx)
			bs.WriteString("pre m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("pre m2 start\n")
			err := next(ctx)
			bs.WriteString("pre m2 end\n")
			return err
		}
	})

	s.Use(func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("use m1 start\n")
			err := next(ctx)
			bs.WriteString("use m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("use m2 start\n")
			err := next(ctx)
			bs.WriteString("use m2 end\n")
			return err
		}
	})

	group := s.Group("/v1")
	group.Use(func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("group m1 start\n")
			err := next(ctx)
			bs.WriteString("group m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("group m2 start\n")
			err := next(ctx)
			bs.WriteString("group m2 end\n")
			return err
		}
	})

	group.R("/route").Use(func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("route m1 start\n")
			err := next(ctx)
			bs.WriteString("route m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx Context) error {
			bs.WriteString("route m2 start\n")
			err := next(ctx)
			bs.WriteString("route m2 end\n")
			return err
		}
	}).GET(OkHandler())

	req := httptest.NewRequest(http.MethodGet, "/v1/route", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if bs.String() != middlewareoutput {
		t.Error(bs.String())
		t.Fail()
	}
}
