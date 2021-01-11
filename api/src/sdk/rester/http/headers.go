package http

import (
	"io"
	"io/ioutil"
	"net/http"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

func handleHeaders(resp *http.Response, output interface{}, limit int64) error {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return merr.From(err).Desc("could not read response body")
	}

	// we consider an error occurred below code 200 and above code 400
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return merr.TransformHTTPCode(resp.StatusCode).Desc(string(data))
	}

	headers, ok := output.(map[string][]string)
	if !ok {
		return merr.Internal().Desc("output on headers must be a map[string]string")
	}
	if headers != nil {
		for k, v := range resp.Header {
			headers[k] = v
		}
	}
	return nil
}
