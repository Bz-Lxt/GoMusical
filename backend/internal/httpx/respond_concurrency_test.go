package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"runtime"
	"runtime/debug"
	"testing"

	"gomusical/internal/httpx"
)

type blockingResponseWriter struct {
	header  http.Header
	body    bytes.Buffer
	ready   chan<- struct{}
	release <-chan struct{}
	status  int
}

func (w *blockingResponseWriter) Header() http.Header {
	return w.header
}

func (w *blockingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ready <- struct{}{}
	<-w.release
}

func (w *blockingResponseWriter) Write(p []byte) (int, error) {
	return w.body.Write(p)
}

func TestJSONConcurrentResponsesRemainIsolated(t *testing.T) {
	previous := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(previous)
	previousGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previousGC)

	for attempt := 0; attempt < 32; attempt++ {
		ready := make(chan struct{}, 2)
		release := make(chan struct{})
		done := make(chan struct{}, 2)
		leftMarker := "left"
		rightMarker := "right"
		left := &blockingResponseWriter{header: make(http.Header), ready: ready, release: release}
		right := &blockingResponseWriter{header: make(http.Header), ready: ready, release: release}

		go func() {
			httpx.JSON(left, http.StatusOK, map[string]string{"request": leftMarker})
			done <- struct{}{}
		}()
		<-ready

		go func() {
			httpx.JSON(right, http.StatusOK, map[string]string{"request": rightMarker})
			done <- struct{}{}
		}()
		<-ready

		close(release)
		<-done
		<-done

		if left.status != http.StatusOK || right.status != http.StatusOK {
			t.Fatalf("unexpected statuses: left=%d right=%d", left.status, right.status)
		}
		if got := responseMarker(t, left.body.Bytes()); got != leftMarker {
			t.Fatalf("left response contains %q, want %q", got, leftMarker)
		}
		if got := responseMarker(t, right.body.Bytes()); got != rightMarker {
			t.Fatalf("right response contains %q, want %q", got, rightMarker)
		}
	}
}

func responseMarker(t *testing.T, body []byte) string {
	t.Helper()
	var response struct {
		OK   bool `json:"ok"`
		Data struct {
			Request string `json:"request"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode response %q: %v", body, err)
	}
	if !response.OK {
		t.Fatalf("response is not successful: %s", body)
	}
	return response.Data.Request
}
