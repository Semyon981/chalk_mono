package middleware

import (
	"chalk/pkg/log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return &loggingMiddleware{next: next}
}

type loggingMiddleware struct {
	next http.Handler
}

func (mw *loggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	mw.next.ServeHTTP(w, r)
	log.Infof("method: %s, time: %s", r.URL.Path, time.Since(start))
}
