package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// appendUnexpectedEOF finishes the stream after prefix bytes with io.ErrUnexpectedEOF,
// mimicking chunked HTTP bodies that truncate without a graceful EOF framing.
type appendUnexpectedEOF struct {
	prefix []byte
	pos    int
}

func (r *appendUnexpectedEOF) Read(p []byte) (int, error) {
	if r.pos < len(r.prefix) {
		n := copy(p, r.prefix[r.pos:])
		r.pos += n
		return n, nil
	}
	return 0, io.ErrUnexpectedEOF
}

type sseRoundTripFunc func(*http.Request) (*http.Response, error)

func (f sseRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func sseOK(body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}
}

func TestAgentChatSSE_truncatedUnexpectedEOFWithoutEndFails(t *testing.T) {
	prefix := []byte("id: 1\n" +
		"event: heartbeat\n" +
		"data: {}\n\n")

	c := &Client{
		http: &http.Client{
			Transport: sseRoundTripFunc(func(*http.Request) (*http.Response, error) {
				return sseOK(io.NopCloser(&appendUnexpectedEOF{prefix: prefix})), nil
			}),
			Timeout: 0,
		},
		baseURL: "http://stub",
	}

	var got []string
	err := c.AgentChatSSE(context.Background(), AgentChatRequest{AgentID: "x", Message: "y"},
		func(event string, _ json.RawMessage) error {
			got = append(got, event)
			return nil
		})
	if err == nil {
		t.Fatal("expected error when stream truncates before end event")
	}
	if !strings.Contains(err.Error(), `terminal "end"`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "heartbeat" {
		t.Fatalf("events: %v", got)
	}
}

func TestAgentChatSSE_unexpectedEOFAfterEndSucceeds(t *testing.T) {
	prefix := []byte("id: 1\n" +
		"event: end\n" +
		"data: {\"status\":\"succeeded\"}\n\n")

	c := &Client{
		http: &http.Client{
			Transport: sseRoundTripFunc(func(*http.Request) (*http.Response, error) {
				return sseOK(io.NopCloser(&appendUnexpectedEOF{prefix: prefix})), nil
			}),
			Timeout: 0,
		},
		baseURL: "http://stub",
	}

	var got []string
	err := c.AgentChatSSE(context.Background(), AgentChatRequest{AgentID: "x", Message: "y"},
		func(event string, _ json.RawMessage) error {
			got = append(got, event)
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "end" {
		t.Fatalf("events: %v", got)
	}
}
