package authtest

import (
	"context"
	_ "embed"
	"net"
	"net/http"
	"testing"
)

//go:embed testdata/response.xml
var response []byte

func SetupAuth(listener net.Listener, t *testing.T) {
	server := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(response)
		}),
	}
	t.Cleanup(func() {
		_ = listener.Close()
		_ = server.Shutdown(context.Background())
	})
	go func() {
		_ = server.Serve(listener)
	}()
}
