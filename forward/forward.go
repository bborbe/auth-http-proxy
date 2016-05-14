package forward

import (
	"net/http"

	"io"

	"github.com/bborbe/log"
	"fmt"
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
	err := h.serveHTTP(resp, req)
	if err != nil {
		logger.Debugf("forward request failed: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) serveHTTP(resp http.ResponseWriter, req *http.Request) error {
	logger.Debugf("%v", req)
	urlStr := fmt.Sprintf("http://%s%s", req.Host, req.RequestURI)
	logger.Debugf("forward request %s %s", req.Method, urlStr)
	subreq, err := http.NewRequest(req.Method, urlStr, req.Body)
	if err != nil {
		return err
	}
	subreq.Header = req.Header
	subresp, err := h.executeRequest(h.target, subreq)
	if err != nil {
		return err
	}
	logger.Debugf("write response")
	copyHeader(resp, &subresp.Header)
	resp.WriteHeader(subresp.StatusCode)
	io.Copy(resp, subresp.Body)
	logger.Debugf("forward request done")
	return nil
}

func copyHeader(resp http.ResponseWriter, source *http.Header) {
	for key, values := range *source {
		for _, value := range values {
			resp.Header().Add(key, value)
		}
	}
}
