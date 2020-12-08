package http

import (
	"io"
	"io/ioutil"
	"net/http"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func handleHeaders(resp *http.Response, output interface{}, limit int64) error {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return merror.Transform(err).Describe("could not read response body")
	}

	// we consider an error occurred below code 200 and above code 400
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return merror.TransformHTTPCode(resp.StatusCode).Describe(string(data))
	}

	headers, ok := output.(map[string][]string)
	if !ok {
		return merror.Internal().Describe("output on headers must be a map[string]string")
	}
	if headers != nil {
		for k, v := range resp.Header {
			headers[k] = v
		}
	}
	return nil
}
