package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/unweave/cli/ui"
	"github.com/unweave/cli/vars"
)

type loggedRoundTripper struct {
	rt http.RoundTripper
}

func (c *loggedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if vars.Debug {
		fmt.Fprintf(ui.Output, "=> %s %s\n", request.Method, request.URL.String())
	}
	startTime := time.Now()
	response, err := c.rt.RoundTrip(request)
	duration := time.Since(startTime)
	if vars.Debug {
		if err != nil {
			fmt.Fprintf(ui.Output, "<= (%s) ERROR! %s\n", duration.String(), err.Error())
		} else {
			fmt.Fprintf(ui.Output, "<= (%s) %s\n", duration.String(), response.Status)
		}
	}
	return response, err
}

// newLoggedTransport takes an http.RoundTripper and returns a new one that logs requests and responses
func newLoggedTransport(rt http.RoundTripper) http.RoundTripper {
	return &loggedRoundTripper{rt: rt}
}
