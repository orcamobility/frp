// Copyright 2025 The frp Authors
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
	"bytes"
	"net/http"
	"testing"
)

func TestPreserveWebSocketHeaderCase(t *testing.T) {
	h := http.Header{}
	// Headers arrive in the Go canonical form after textproto parsing.
	h["Sec-Websocket-Key"] = []string{"dGhlIHNhbXBsZSBub25jZQ=="}
	h["Sec-Websocket-Version"] = []string{"13"}
	h["Sec-Websocket-Protocol"] = []string{"chat"}
	h["Sec-Websocket-Extensions"] = []string{"permessage-deflate"}
	h["Sec-Websocket-Accept"] = []string{"s3pPLMBiTxaQ9kYGzzhZRbK+xOo="}
	h["Upgrade"] = []string{"websocket"}
	h["Connection"] = []string{"Upgrade"}

	PreserveWebSocketHeaderCase(h)

	for _, canonical := range []string{
		"Sec-Websocket-Key",
		"Sec-Websocket-Version",
		"Sec-Websocket-Protocol",
		"Sec-Websocket-Extensions",
		"Sec-Websocket-Accept",
	} {
		if _, ok := h[canonical]; ok {
			t.Errorf("canonical key %q still present after preserve", canonical)
		}
	}
	for _, rfc := range []string{
		"Sec-WebSocket-Key",
		"Sec-WebSocket-Version",
		"Sec-WebSocket-Protocol",
		"Sec-WebSocket-Extensions",
		"Sec-WebSocket-Accept",
	} {
		if _, ok := h[rfc]; !ok {
			t.Errorf("RFC key %q missing after preserve", rfc)
		}
	}

	// Direct map writes survive http.Header serialization verbatim — this is
	// what the proxy ultimately writes onto the wire.
	var buf bytes.Buffer
	if err := h.WriteSubset(&buf, nil); err != nil {
		t.Fatalf("write headers: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==",
		"Sec-WebSocket-Version: 13",
		"Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=",
	} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("expected header line %q in serialized output, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Sec-Websocket-Key:",
		"Sec-Websocket-Version:",
		"Sec-Websocket-Accept:",
	} {
		if bytes.Contains([]byte(out), []byte(unwanted)) {
			t.Errorf("unexpected canonical header %q in serialized output:\n%s", unwanted, out)
		}
	}
}

func TestPreserveWebSocketHeaderCase_NoOp(t *testing.T) {
	h := http.Header{
		"Content-Type": []string{"application/json"},
		"X-Custom":     []string{"value"},
	}
	PreserveWebSocketHeaderCase(h)

	if got := h.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type mutated: %q", got)
	}
	if got := h.Get("X-Custom"); got != "value" {
		t.Errorf("X-Custom mutated: %q", got)
	}
	if len(h) != 2 {
		t.Errorf("unexpected header count: %d", len(h))
	}
}
