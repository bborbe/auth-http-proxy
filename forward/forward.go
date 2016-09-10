package forward

import (
	"net/http"

	"io"

	"fmt"

	"github.com/bborbe/auth_http_proxy/model"
	"github.com/golang/glog"
)

type executeRequest func(address string, req *http.Request) (resp *http.Response, err error)

type handler struct {
	target         model.TargetAddress
	executeRequest executeRequest
}

func New(target model.TargetAddress, executeRequest executeRequest) *handler {
	h := new(handler)
	h.target = target
	h.executeRequest = executeRequest
	return h
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	glog.V(2).Infof("forward request")
	if err := h.serveHTTP(resp, req); err != nil {
		glog.V(1).Infof("forward request failed: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	glog.V(2).Infof("request forward succesful")
}

func (h *handler) serveHTTP(resp http.ResponseWriter, req *http.Request) error {
	glog.V(2).Infof("%v", req)
	urlStr := fmt.Sprintf("http://%s%s", req.Host, req.RequestURI)
	glog.V(2).Infof("forward request %s %s", req.Method, urlStr)
	subreq, err := http.NewRequest(req.Method, urlStr, req.Body)
	if err != nil {
		glog.V(1).Infof("create request to %v failed: %v", urlStr, err)
		return err
	}
	subreq.Header = req.Header
	subresp, err := h.executeRequest(h.target.String(), subreq)
	if err != nil {
		glog.V(2).Infof("execute request to %v failed: %v", h.target, err)
		return err
	}
	glog.V(2).Infof("write response")
	copyHeader(resp, &subresp.Header)
	resp.WriteHeader(subresp.StatusCode)
	if _, err := io.Copy(resp, subresp.Body); err != nil {
		glog.V(1).Infof("copy body failed: %v", err)
		return err
	}
	glog.V(2).Infof("forward request done")
	return nil
}

func copyHeader(resp http.ResponseWriter, req *http.Header) {
	for key, values := range *req {
		for _, value := range values {
			resp.Header().Add(key, value)
		}
	}
}
