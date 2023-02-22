package othttp

import (
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"log"
	"net/http"
)

// ListenAndServe starts an instrumented server with a single endpoint.
func ListenAndServe(hostPort string, endpoint string) {
	mw := nethttp.Middleware(
		opentracing.GlobalTracer(),
		http.DefaultServeMux,
		nethttp.OperationNameFunc(func(r *http.Request) string {
			return "HTTP " + r.Method + ":" + endpoint
		}),
	)

	log.Print("Listening on http://" + hostPort)
	log.Fatal(http.ListenAndServe(hostPort, mw))
}
