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

// Predefine some variables
const (
	CharsetUTF8 = "charset=UTF-8"
	PROPFIND    = "PROPFIND"
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + CharsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + CharsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + CharsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + CharsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + CharsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + CharsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

// MIME slice types
var (
	MIMEApplicationJSONs                  = []string{MIMEApplicationJSON}
	MIMEApplicationJSONCharsetUTF8s       = []string{MIMEApplicationJSONCharsetUTF8}
	MIMEApplicationJavaScripts            = []string{MIMEApplicationJavaScript}
	MIMEApplicationJavaScriptCharsetUTF8s = []string{MIMEApplicationJavaScriptCharsetUTF8}
	MIMEApplicationXMLs                   = []string{MIMEApplicationXML}
	MIMEApplicationXMLCharsetUTF8s        = []string{MIMEApplicationXMLCharsetUTF8}
	MIMETextXMLs                          = []string{MIMETextXML}
	MIMETextXMLCharsetUTF8s               = []string{MIMETextXMLCharsetUTF8}
	MIMEApplicationForms                  = []string{MIMEApplicationForm}
	MIMEApplicationProtobufs              = []string{MIMEApplicationProtobuf}
	MIMEApplicationMsgpacks               = []string{MIMEApplicationMsgpack}
	MIMETextHTMLs                         = []string{MIMETextHTML}
	MIMETextHTMLCharsetUTF8s              = []string{MIMETextHTMLCharsetUTF8}
	MIMETextPlains                        = []string{MIMETextPlain}
	MIMETextPlainCharsetUTF8s             = []string{MIMETextPlainCharsetUTF8}
	MIMEMultipartForms                    = []string{MIMEMultipartForm}
	MIMEOctetStreams                      = []string{MIMEOctetStream}
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptedLanguage    = "Accept-Language"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderConnection          = "Connection"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderEtag                = "Etag"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-Ip"
	HeaderXServerID           = "X-Server-Id"
	HeaderXRequestID          = "X-Request-Id"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"
	HeaderReferer             = "Referer"
	HeaderUserAgent           = "User-Agent"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)
