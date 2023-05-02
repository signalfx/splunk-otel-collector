package fakek8s

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Fake K8s API server request: %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
