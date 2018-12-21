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

package middleware

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/utils"
)

// Gzip returns a middleware to compress the response body by GZIP.
func Gzip(level ...int) Middleware {
	glevel := -1
	if len(level) > 0 {
		glevel = level[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {
			if strings.Contains(ctx.Request().Header.Get(ship.HeaderAcceptEncoding), "gzip") {
				resp := ctx.Response()
				resp.Header().Add(ship.HeaderVary, ship.HeaderAcceptEncoding)
				resp.Header().Set(ship.HeaderContentEncoding, "gzip")

				writer := utils.GetResponseFromPool(resp)
				newWriter, err := gzip.NewWriterLevel(writer, glevel)
				if err != nil {
					return err
				}

				defer func() {
					if writer.Size == 0 {
						resp.Header().Del(ship.HeaderContentEncoding)
						ctx.SetResponse(resp)
						newWriter.Reset(ioutil.Discard)
					}
					newWriter.Close()
					utils.PutResponseIntoPool(writer)
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
		g.ResponseWriter.Header().Del(ship.HeaderContentEncoding)
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

func (g *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return g.ResponseWriter.(http.Hijacker).Hijack()
}

func (g *gzipResponseWriter) CloseNotify() <-chan bool {
	return g.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
