package forward

import (
	"net/http"

	"github.com/bborbe/log"
)

var logger = log.DefaultLogger

type handler struct {
	target string
}

func New(target string) *handler {
	h := new(handler)
	h.target = target
	return h
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger.Debugf("forward request")
}
