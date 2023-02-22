package main

import (
	"Mastering-Distributed-Tracing-code/lib/tracing"
	"Mastering-Distributed-Tracing-code/people"
	"context"
	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"log"
	"net/http"
	"strings"
)

var repo *people.Repository
var tracer opentracing.Tracer

func main() {
	repo = people.NewRepository()
	defer repo.Close()

	// init in application
	tr, closer := tracing.Init("go-2-hello")
	tracer = tr
	defer closer.Close()
	// set global tracer
	opentracing.SetGlobalTracer(tracer)

	http.HandleFunc("/sayHello/", handleSayHello)

	log.Print("Listening on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSayHello(w http.ResponseWriter, r *http.Request) {
	// add name
	span := tracer.StartSpan("say-hello")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(r.Context(), span)

	name := strings.TrimPrefix(r.URL.Path, "/sayHello/")
	greeting, err := SayHello(ctx, name)
	if err != nil {
		// add opentracing log
		span.SetTag("error", true)
		span.LogFields(otlog.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// add tag response
	span.SetTag("response", greeting)
	w.Write([]byte(greeting))
}

// SayHello creates a greeting for the named person.
func SayHello(ctx context.Context, name string) (string, error) {
	person, err := repo.GetPerson(ctx, name)
	if err != nil {
		return "", err
	}

	// add logKV from context passed by parameters
	opentracing.SpanFromContext(ctx).LogKV(
		"name", person.Name,
		"title", person.Title,
		"description", person.Description,
	)

	return FormatGreeting(
		ctx,
		person.Name,
		person.Title,
		person.Description,
	), nil
}

// FormatGreeting combines information about a person into a greeting string.
// add span for context
func FormatGreeting(ctx context.Context, name, title, description string) string {

	span, _ := opentracing.StartSpanFromContext(ctx, "format-greeting")
	defer span.Finish()

	response := "Hello, "
	if title != "" {
		response += title + " "
	}
	response += name + "!"
	if description != "" {
		response += " " + description
	}
	return response
}
