package forward

import (
	"net/http"
	"testing"

	. "github.com/bborbe/assert"
	"github.com/golang/glog"
	"os"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestImplementsHandler(t *testing.T) {
	object := New("target:80", nil)
	var expected *http.Handler
	err := AssertThat(object, Implements(expected))
	if err != nil {
		t.Fatal(err)
	}
}
