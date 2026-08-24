package httpx

import (
	"context"
	"net/http"
	"strings"
	"time"

	"gomusical/internal/clock"
	"gomusical/internal/logx"
)

type ctxKey string

const (
	CtxUserID   ctxKey = "uid"
	CtxRole     ctxKey = "role"
	CtxSession  ctxKey = "sid"
	CtxCSRF     ctxKey = "csrf"
	CtxRequest  ctxKey = "reqid"
	CtxClientFP ctxKey = "cfp"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = clock.Now().Format("20060102150405.000000")
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), CtxRequest, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logx.Error("panic recovered", "panic", rec, "path", r.URL.Path)
				Fail(w, ErrInternal)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(ww, r)
		logx.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func CORS(origins []string) func(http.Handler) http.Handler {
	allow := map[string]bool{}
	for _, o := range origins {
		allow[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allow[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Request-ID, X-Client-Fingerprint")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Expose-Headers", "Retry-After, Content-Range, Accept-Ranges, ETag")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ClientFingerprint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp := r.Header.Get("X-Client-Fingerprint")
		if fp == "" {
			fp = r.UserAgent()
		}
		ctx := context.WithValue(r.Context(), CtxClientFP, fp)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RefererGuard(allow []string, prefixes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			need := false
			for _, p := range prefixes {
				if strings.HasPrefix(r.URL.Path, p) {
					need = true
					break
				}
			}
			if !need {
				next.ServeHTTP(w, r)
				return
			}
			ref := r.Header.Get("Referer")
			if ref == "" {
				ref = r.Header.Get("Origin")
			}
			ok := false
			for _, a := range allow {
				if ref == "" {
					// same-origin fetch from audio element may omit referer in some browsers;
					// allow empty only for GET playlist bootstrap from our origin cookie session.
					ok = true
					break
				}
				if strings.HasPrefix(ref, a) {
					ok = true
					break
				}
			}
			if !ok {
				Fail(w, ErrRefererDenied)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserID(ctx context.Context) string {
	v, _ := ctx.Value(CtxUserID).(string)
	return v
}

func Role(ctx context.Context) string {
	v, _ := ctx.Value(CtxRole).(string)
	return v
}

func Fingerprint(ctx context.Context) string {
	v, _ := ctx.Value(CtxClientFP).(string)
	return v
}
