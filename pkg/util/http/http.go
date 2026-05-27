// Copyright 2023 The frp Authors
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

package http

import (
	"encoding/base64"
	"net"
	"net/http"
	"strings"
)

func OkResponse() *http.Response {
	header := make(http.Header)

	res := &http.Response{
		Status:     "OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
	}
	return res
}

func ProxyUnauthorizedResponse() *http.Response {
	header := make(http.Header)
	header.Set("Proxy-Authenticate", `Basic realm="Restricted"`)
	res := &http.Response{
		Status:     "Proxy Authentication Required",
		StatusCode: 407,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
	}
	return res
}

// canonicalHost strips port from host if present and returns the canonicalized
// host name.
func CanonicalHost(host string) (string, error) {
	var err error
	host = strings.ToLower(host)
	if hasPort(host) {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return "", err
		}
	}
	// Strip trailing dot from fully qualified domain names.
	host = strings.TrimSuffix(host, ".")
	return host, nil
}

// hasPort reports whether host contains a port number. host may be a host
// name, an IPv4 or an IPv6 address.
func hasPort(host string) bool {
	colons := strings.Count(host, ":")
	if colons == 0 {
		return false
	}
	if colons == 1 {
		return true
	}
	return host[0] == '[' && strings.Contains(host, "]:")
}

func ParseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func BasicAuth(username, passwd string) string {
	auth := username + ":" + passwd
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// websocketHeaderCases maps the Go canonical form of RFC 6455 WebSocket
// headers (as produced by textproto.CanonicalMIMEHeaderKey) back to the
// mixed-case form spelled in the RFC. Some strict embedded HTTP parsers
// (notably several IP-camera firmwares) do byte-wise header-name matching
// and silently ignore upgrade requests when Sec-WebSocket-* arrives as
// Sec-Websocket-*, causing the upgrade to hang until the proxy times out.
var websocketHeaderCases = map[string]string{
	"Sec-Websocket-Key":        "Sec-WebSocket-Key",
	"Sec-Websocket-Version":    "Sec-WebSocket-Version",
	"Sec-Websocket-Protocol":   "Sec-WebSocket-Protocol",
	"Sec-Websocket-Extensions": "Sec-WebSocket-Extensions",
	"Sec-Websocket-Accept":     "Sec-WebSocket-Accept",
}

// PreserveWebSocketHeaderCase rewrites any Sec-Websocket-* keys in h to
// their RFC 6455 mixed-case spelling. Direct map writes skip
// CanonicalMIMEHeaderKey and are emitted verbatim by http.Header.WriteSubset.
func PreserveWebSocketHeaderCase(h http.Header) {
	for canonical, rfc := range websocketHeaderCases {
		if v, ok := h[canonical]; ok {
			delete(h, canonical)
			h[rfc] = v
		}
	}
}
