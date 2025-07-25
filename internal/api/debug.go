package api

import (
	"log"
	"net/http"
	"net/http/httputil"
)

type debugTransport struct {
	rt http.RoundTripper
}

func (t debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dumpReq, _ := httputil.DumpRequestOut(req, true)
	log.Printf("REQUEST:\n%s\n", dumpReq)
	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	dumpResp, _ := httputil.DumpResponse(resp, true)
	log.Printf("RESPONSE:\n%s\n", dumpResp)
	return resp, nil
}

func (c *Client) EnableDebug() {
	c.http = &http.Client{Transport: debugTransport{rt: http.DefaultTransport}}
}
