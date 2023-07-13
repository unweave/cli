package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/unweave/cli/vars"
)

type loggedRoundTripper struct {
	rt http.RoundTripper
}

func (c *loggedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if vars.Debug {
		fmt.Printf("=> %s %s\n", request.Method, request.URL.String())
	}
	startTime := time.Now()
	response, err := c.rt.RoundTrip(request)
	duration := time.Since(startTime)
	if vars.Debug {
		if err != nil {
			fmt.Printf("<= (%s) ERROR! %s\n", duration.String(), err.Error())
		} else {
			fmt.Printf("<= (%s) %s\n", duration.String(), response.Status)
		}
	}
	return response, err
}

// newLoggedTransport takes an http.RoundTripper and returns a new one that logs requests and responses
func newLoggedTransport(rt http.RoundTripper) http.RoundTripper {
	return &loggedRoundTripper{rt: rt}
}
