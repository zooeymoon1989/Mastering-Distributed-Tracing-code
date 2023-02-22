package main

import (
	"Mastering-Distributed-Tracing-code/lib/tracing"
	"Mastering-Distributed-Tracing-code/people"
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

	name := strings.TrimPrefix(r.URL.Path, "/sayHello/")
	greeting, err := SayHello(name, span)
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
// add opentracing span
func SayHello(name string, span opentracing.Span) (string, error) {
	person, err := repo.GetPerson(name, span)
	if err != nil {
		return "", err
	}

	// add k-v pair to span
	span.LogKV(
		"name", person.Name,
		"title", person.Title,
		"description", person.Description,
	)

	return FormatGreeting(
		person.Name,
		person.Title,
		person.Description,
		span,
	), nil
}

// FormatGreeting combines information about a person into a greeting string.
// add span for context
func FormatGreeting(name, title, description string, span opentracing.Span) string {
	// add span in this
	span = opentracing.GlobalTracer().StartSpan(
		"format-greeting",
		opentracing.ChildOf(span.Context()),
	)
	span.Finish()

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
