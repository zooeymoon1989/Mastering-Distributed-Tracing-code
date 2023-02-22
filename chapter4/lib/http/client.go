package xhttp

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"io/ioutil"
	"net/http"
)

// Get executes an HTTP GET request and returns the response body.
// Any errors or non-200 status code result in an error.
func Get(ctx context.Context, operationName, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	defer span.Finish()

	// Inject() takes the `sm` SpanContext instance and injects it for
	// propagation within `carrier`.
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	return Do(req)
}

// Do executes an HTTP request and returns the response body.
// Any errors or non-200 status code result in an error.
func Do(req *http.Request) ([]byte, error) {
	return DoWithClient(req, http.DefaultClient)
}

// DoWithClient executes an HTTP request and returns the response body.
// Any errors or non-200 status code result in an error.
func DoWithClient(req *http.Request, client *http.Client) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode: %d, Body: %s", resp.StatusCode, body)
	}

	return body, nil
}
