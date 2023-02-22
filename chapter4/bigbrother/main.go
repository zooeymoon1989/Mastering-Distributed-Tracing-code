package main

import (
	"Mastering-Distributed-Tracing-code/chapter4/lib/tracing"
	"Mastering-Distributed-Tracing-code/chapter4/othttp"
	"Mastering-Distributed-Tracing-code/chapter4/people"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"net/http"
	"strings"
)

var repo *people.Repository

func main() {
	tracer, closer := tracing.Init("go-4-bigbrother")
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)
	repo = people.NewRepository()
	defer repo.Close()

	http.HandleFunc("/getPerson/", handleGetPerson)
	othttp.ListenAndServe(":8081", "/getPerson")
}

func handleGetPerson(w http.ResponseWriter, r *http.Request) {

	span := opentracing.SpanFromContext(r.Context())

	name := strings.TrimPrefix(r.URL.Path, "/getPerson/")
	person, err := repo.GetPerson(r.Context(), name)
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(otlog.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.LogKV(
		"name", person.Name,
		"title", person.Title,
		"description", person.Description,
	)

	bytes, _ := json.Marshal(person)
	w.Write(bytes)
}
