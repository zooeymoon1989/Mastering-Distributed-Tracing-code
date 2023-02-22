package main

import (
	"Mastering-Distributed-Tracing-code/chapter4/lib/tracing"
	"Mastering-Distributed-Tracing-code/chapter4/othttp"
	"context"
	"github.com/opentracing/opentracing-go"
	"net/http"
)

func main() {
	tracer, closer := tracing.Init("go-4-formatter")
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)

	http.HandleFunc("/formatGreeting/", handleFormatGreeting)
	othttp.ListenAndServe(":8082", "/formatGreeting")
}

func handleFormatGreeting(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	title := r.FormValue("title")
	descr := r.FormValue("description")

	greeting := FormatGreeting(r.Context(), name, title, descr)
	w.Write([]byte(greeting))
}

// FormatGreeting combines information about a person into a greeting string.
func FormatGreeting(
	ctx context.Context,
	name, title, description string,
) string {
	span := opentracing.SpanFromContext(ctx)

	greeting := span.BaggageItem("greeting")
	if greeting == "" {
		greeting = "Hello"
	}
	response := greeting + ", "
	if title != "" {
		response += title + " "
	}
	response += name + "!"
	if description != "" {
		response += " " + description
	}
	return response
}
