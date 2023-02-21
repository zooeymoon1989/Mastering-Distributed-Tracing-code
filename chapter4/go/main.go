package main

import (
	"Mastering-Distributed-Tracing-code/lib/tracing"
	"Mastering-Distributed-Tracing-code/people"
	opentracing "github.com/opentracing/opentracing-go"
	"log"
	"net/http"
	"strings"
)

var repo *people.Repository
var tracer opentracing.Tracer

func main() {
	repo = people.NewRepository()
	defer repo.Close()

	tr, closer := tracing.Init("go-2-hello")
	defer closer.Close()
	tracer = tr

	http.HandleFunc("/sayHello/", handleSayHello)

	log.Print("Listening on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSayHello(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("say-hello")
	defer span.Finish()
	name := strings.TrimPrefix(r.URL.Path, "/sayHello/")
	greeting, err := SayHello(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(greeting))
}

// SayHello creates a greeting for the named person.
func SayHello(name string) (string, error) {
	person, err := repo.GetPerson(name)
	if err != nil {
		return "", err
	}
	return FormatGreeting(
		person.Name,
		person.Title,
		person.Description,
	), nil
}

// FormatGreeting combines information about a person into a greeting string.
func FormatGreeting(name, title, description string) string {
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
