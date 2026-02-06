package httpapi

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var reqIDCounter uint64

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := fmt.Sprintf("req-%d", atomic.AddUint64(&reqIDCounter, 1))
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		rec.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(rec, r)
		log.Printf("event=request method=%s path=%s status=%d duration_ms=%d request_id=%s remote=%s", r.Method, r.URL.Path, rec.status, time.Since(start).Milliseconds(), reqID, r.RemoteAddr)
	})
}
