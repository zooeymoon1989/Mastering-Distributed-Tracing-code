package main

import (
	xhttp "Mastering-Distributed-Tracing-code/chapter4/lib/http"
	"Mastering-Distributed-Tracing-code/chapter4/lib/model"
	"Mastering-Distributed-Tracing-code/chapter4/lib/tracing"
	"Mastering-Distributed-Tracing-code/chapter4/othttp"
	"context"
	"encoding/json"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"net/http"
	"net/url"
	"strings"
)

var client = &http.Client{Transport: &nethttp.Transport{}}

func main() {
	tracer, closer := tracing.Init("go-4-hello")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	http.HandleFunc("/sayHello/", handleSayHello)

	othttp.ListenAndServe(":8080", "/sayHello")
}

func handleSayHello(w http.ResponseWriter, r *http.Request) {

	span := opentracing.SpanFromContext(r.Context())
	name := strings.TrimPrefix(r.URL.Path, "/sayHello/")
	greeting, err := SayHello(r.Context(), name)
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(otlog.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	span.SetTag("response", greeting)
	w.Write([]byte((greeting)))
}

func SayHello(ctx context.Context, name string) (string, error) {
	person, err := getPerson(ctx, name)
	if err != nil {
		return "", err
	}
	return formatGreeting(ctx, person)
}

func formatGreeting(ctx context.Context, person *model.Person) (string, error) {
	v := url.Values{}
	v.Set("name", person.Name)
	v.Set("title", person.Title)
	v.Set("description", person.Description)
	url := "http://localhost:8082/formatGreeting?" + v.Encode()
	res, err := get(ctx, "formatGreeting", url)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func getPerson(ctx context.Context, name string) (*model.Person, error) {
	res, err := get(ctx, "getPerson", "http://localhost:8081/getPerson/"+name)
	if err != nil {
		return nil, err
	}
	var person model.Person
	if err := json.Unmarshal(res, &person); err != nil {
		return nil, err
	}
	return &person, nil
}

func get(ctx context.Context, operationName, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req)
	defer ht.Finish()

	return xhttp.DoWithClient(req, client)
}
