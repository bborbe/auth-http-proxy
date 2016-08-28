package forward

import (
	"net/http"

	"io"

	"fmt"

	"github.com/golang/glog"
)

type executeRequest func(address string, req *http.Request) (resp *http.Response, err error)

type handler struct {
	target         string
	executeRequest executeRequest
}

func New(target string, executeRequest executeRequest) *handler {
	h := new(handler)
	h.target = target
	h.executeRequest = executeRequest
	return h
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	glog.V(2).Infof("forward request")
	err := h.serveHTTP(resp, req)
	if err != nil {
		glog.V(2).Infof("forward request failed: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) serveHTTP(resp http.ResponseWriter, req *http.Request) error {
	glog.V(2).Infof("%v", req)
	urlStr := fmt.Sprintf("http://%s%s", req.Host, req.RequestURI)
	glog.V(2).Infof("forward request %s %s", req.Method, urlStr)
	subreq, err := http.NewRequest(req.Method, urlStr, req.Body)
	if err != nil {
		return err
	}
	subreq.Header = req.Header
	subresp, err := h.executeRequest(h.target, subreq)
	if err != nil {
		return err
	}
	glog.V(2).Infof("write response")
	copyHeader(resp, &subresp.Header)
	resp.WriteHeader(subresp.StatusCode)
	if _, err := io.Copy(resp, subresp.Body); err != nil {
		return err
	}
	glog.V(2).Infof("forward request done")
	return nil
}

func copyHeader(resp http.ResponseWriter, source *http.Header) {
	for key, values := range *source {
		for _, value := range values {
			resp.Header().Add(key, value)
		}
	}
}
