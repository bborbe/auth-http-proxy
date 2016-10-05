package forward

import (
	"net/http"
	"testing"

	"os"

	"fmt"

	. "github.com/bborbe/assert"
	"github.com/bborbe/http/mock"
	"github.com/golang/glog"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestImplementsHandler(t *testing.T) {
	object := New("target:80", nil)
	var expected *http.Handler
	if err := AssertThat(object, Implements(expected)); err != nil {
		t.Fatal(err)
	}
}

func TestCopyHeaderEmpyHeader(t *testing.T) {
	resp := mock.NewHttpResponseWriterMock()
	req := make(http.Header)
	copyHeader(resp, &req)
	if err := AssertThat(len(resp.Header()), Is(0)); err != nil {
		t.Fatal(err)
	}
}

func TestCopyHeaderHeader(t *testing.T) {
	resp := mock.NewHttpResponseWriterMock()
	req := make(http.Header)
	req.Add("A", "b")
	copyHeader(resp, &req)
	if err := AssertThat(len(resp.Header()), Is(1)); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v", resp.Header())
	values, ok := resp.Header()["A"]
	if err := AssertThat(ok, Is(true)); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(len(values), Is(1)); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(values[0], Is("b")); err != nil {
		t.Fatal(err)
	}
}
