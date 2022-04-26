package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	binaryContentType = "application/octet-stream"
	jsonContentType   = "application/json"
	zipContentType    = "application/zip"
)

func addIfNotEmpty(query *url.Values, key string, value string) {
	if len(key) > 0 && len(value) > 0 {
		query.Add(key, value)
	}
}

func dumpRequest(verbose bool, req *http.Request, body bool) {
	if verbose {
		dump, err := httputil.DumpRequestOut(req, body)

		if err == nil {
			fmt.Printf("\n--- Request ---\n%s\n", dump)
		}
	}
}

func dumpResponse(verbose bool, resp *http.Response, body bool) {
	if verbose {
		dump, err := httputil.DumpResponse(resp, body)

		if err == nil {
			fmt.Printf("\n--- Response ---\n%s\n", dump)
		}
	}
}
