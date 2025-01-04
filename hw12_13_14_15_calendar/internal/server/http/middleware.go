package internalhttp

import (
	"log"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("%s [%s] %s %s %s %d %dms \"%s\"\n",
			r.RemoteAddr,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.Path,
			r.Proto,
			rw.status,
			duration.Milliseconds(),
			r.UserAgent(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
