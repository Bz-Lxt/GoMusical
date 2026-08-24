package download

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ServeRange is a thin wrapper around http.ServeContent so callers can
// inject metering without reimplementing If-Range / ETag / 206.
func ServeRange(w http.ResponseWriter, r *http.Request, name string, mod time.Time, f *os.File, size int64) {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, no-store")
	http.ServeContent(w, r, name, mod, f)
}

type Metered struct {
	W      http.ResponseWriter
	OnByte func(n int)
}

func (m *Metered) Header() http.Header         { return m.W.Header() }
func (m *Metered) WriteHeader(status int)      { m.W.WriteHeader(status) }
func (m *Metered) Write(p []byte) (int, error) {
	n, err := m.W.Write(p)
	if n > 0 && m.OnByte != nil {
		m.OnByte(n)
	}
	return n, err
}

type RateLimited struct {
	Src   io.ReadSeeker
	Wait  func(n int)
	chunk int
}

func (r *RateLimited) Seek(offset int64, whence int) (int64, error) {
	return r.Src.Seek(offset, whence)
}

func (r *RateLimited) Read(p []byte) (int, error) {
	if r.chunk <= 0 {
		r.chunk = 32 * 1024
	}
	if len(p) > r.chunk {
		p = p[:r.chunk]
	}
	n, err := r.Src.Read(p)
	if n > 0 && r.Wait != nil {
		r.Wait(n)
	}
	return n, err
}

func ParseSingleRange(h string, size int64) (start, end int64, ok bool) {
	h = strings.TrimSpace(h)
	if !strings.HasPrefix(h, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(h, "bytes=")
	if strings.Contains(spec, ",") {
		return 0, 0, false
	}
	parts := strings.Split(spec, "-")
	if len(parts) != 2 {
		return 0, 0, false
	}
	if parts[0] == "" {
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix <= 0 {
			return 0, 0, false
		}
		if suffix > size {
			suffix = size
		}
		return size - suffix, size - 1, true
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false
	}
	if parts[1] == "" {
		return start, size - 1, true
	}
	end, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil || end < start {
		return 0, 0, false
	}
	if end >= size {
		end = size - 1
	}
	return start, end, true
}

func BytesOfRange(h string, size int64) int64 {
	start, end, ok := ParseSingleRange(h, size)
	if !ok {
		return size
	}
	return end - start + 1
}
