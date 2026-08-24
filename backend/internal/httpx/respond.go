package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"gomusical/internal/logx"
)

type Envelope struct {
	OK    bool        `json:"ok"`
	Data  any         `json:"data,omitempty"`
	Error *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var envelopePool = sync.Pool{New: func() any {
	return new(Envelope)
}}

func JSON(w http.ResponseWriter, status int, data any) {
	env := envelopePool.Get().(*Envelope)
	*env = Envelope{OK: status < 400, Data: data}
	envelopePool.Put(env)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(env); err != nil {
		logx.Error("json encode failed", "err", err)
	}
}

func JSONRaw(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Fail(w http.ResponseWriter, err error) {
	var ae *AppError
	if errors.As(err, &ae) {
		if ae.Status >= 500 {
			logx.Error("app error", "code", ae.Code, "msg", ae.Message, "err", ae.Cause)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(ae.Status)
		_ = json.NewEncoder(w).Encode(Envelope{
			OK:    false,
			Error: &ErrorBody{Code: ae.Code, Message: ae.Message},
		})
		return
	}
	logx.Error("unhandled error", "err", err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(500)
	_ = json.NewEncoder(w).Encode(Envelope{
		OK:    false,
		Error: &ErrorBody{Code: "internal", Message: "服务内部错误"},
	})
}

func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return Wrap(400, "bad_request", "JSON 无法解析或字段类型错误", err)
	}
	return nil
}

func QueryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}
