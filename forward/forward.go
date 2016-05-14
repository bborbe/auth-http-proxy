package forward

import (
	"net/http"

	"io"

	"github.com/bborbe/log"
)

var logger = log.DefaultLogger

type ExecuteRequest func(address string, req *http.Request) (resp *http.Response, err error)

type handler struct {
	target         string
	executeRequest ExecuteRequest
}

func New(target string, executeRequest ExecuteRequest) *handler {
	h := new(handler)
	h.target = target
	h.executeRequest = executeRequest
	return h
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger.Debugf("forward request")
	subreq := *req
	subresp, err := h.executeRequest(h.target, &subreq)
	if err != nil {
		logger.Debugf("forward request failed: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Debugf("write response")
	resp.WriteHeader(subresp.StatusCode)
	copyHeader(resp, &subresp.Header)
	io.Copy(resp, subresp.Body)
	logger.Debugf("forward request done")
}

func copyHeader(resp http.ResponseWriter, source *http.Header) {
	header := resp.Header()
	for key, values := range *source {
		for _, value := range values {
			header.Add(key, value)
		}
	}
}
