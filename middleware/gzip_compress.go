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
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/xgfone/ship/v4"
)

// GZipConfig is used to configure the GZIP middleware.
type GZipConfig struct {
	// Level is the compression level.
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

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) error {
			if strings.Contains(ctx.GetHeader(ship.HeaderAcceptEncoding), "gzip") {
				ctx.AddHeader(ship.HeaderVary, ship.HeaderAcceptEncoding)
				ctx.SetHeader(ship.HeaderContentEncoding, "gzip")

				resp := ctx.ResponseWriter()
				writer := ship.GetResponseFromPool(resp)
				newWriter, err := gzip.NewWriterLevel(writer, conf.Level)
				if err != nil {
					return err
				}

				defer func() {
					if writer.Size == 0 {
						ctx.DelHeader(ship.HeaderContentEncoding)
						ctx.SetResponse(resp)
						newWriter.Reset(ioutil.Discard)
					}
					newWriter.Close()
					ship.PutResponseIntoPool(writer)
				}()

				gzipWriter := &gzipResponseWriter{Writer: newWriter, ResponseWriter: resp}
				ctx.SetResponse(gzipWriter)
			}

			return next(ctx)
		}
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusNoContent {
		g.Header().Del(ship.HeaderContentEncoding)
	}
	g.Header().Del(ship.HeaderContentLength)
	g.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Header().Get(ship.HeaderContentType) == "" {
		g.Header().Set(ship.HeaderContentType, http.DetectContentType(b))
	}
	return g.Writer.Write(b)
}

func (g *gzipResponseWriter) Flush() {
	g.Writer.(*gzip.Writer).Flush()
}
