package file

import (
	"testing"

	"os"

	. "github.com/bborbe/assert"
	"github.com/bborbe/auth-http-proxy/verifier"
	"github.com/golang/glog"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestImplementsVerifier(t *testing.T) {
	object := New("users")
	var expected *verifier.Verifier
	err := AssertThat(object, Implements(expected))
	if err != nil {
		t.Fatal(err)
	}
}
