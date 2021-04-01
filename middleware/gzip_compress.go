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

package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/xgfone/ship/v4"
)

// GZipConfig is used to configure the GZIP middleware.
type GZipConfig struct {
	// Level is the compression level, range [-1, 9].
	//
	// Default: -1 (default compression level)
	Level int
}

// Gzip returns a middleware to compress the response body by GZIP.
func Gzip(config *GZipConfig) Middleware {
	var conf GZipConfig
	if config != nil {
		conf = *config
	}

	if conf.Level < gzip.HuffmanOnly || conf.Level > gzip.BestCompression {
		panic(fmt.Errorf("gzip: invalid compression level '%d'", conf.Level))
	}

	gpool := sync.Pool{New: func() interface{} {
		w, err := gzip.NewWriterLevel(nil, conf.Level)
		if err != nil {
			panic(err)
		}
		return &gzipResponse{w: w}
	}}

	releaseGzipResponse := func(r *gzipResponse) { r.w.Close(); gpool.Put(r) }
	acquireGzipResponse := func(w http.ResponseWriter) (r *gzipResponse) {
		r = gpool.Get().(*gzipResponse)
		r.ResponseWriter = w
		r.w.Reset(w)
		return
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) error {
			if strings.Contains(ctx.GetHeader(ship.HeaderAcceptEncoding), "gzip") {
				ctx.AddHeader(ship.HeaderVary, ship.HeaderAcceptEncoding)
				ctx.SetHeader(ship.HeaderContentEncoding, "gzip")

				resp := ctx.ResponseWriter()
				gresp := acquireGzipResponse(resp)
				defer releaseGzipResponse(gresp)
				ctx.SetResponse(gresp)
			}

			return next(ctx)
		}
	}
}

type gzipResponse struct {
	http.ResponseWriter
	w *gzip.Writer
}

func (g *gzipResponse) Write(b []byte) (int, error) { return g.w.Write(b) }
func (g *gzipResponse) Flush()                      { g.w.Flush() }
