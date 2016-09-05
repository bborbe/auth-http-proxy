package crypter

import (
	"testing"

	"os"

	. "github.com/bborbe/assert"
	"github.com/golang/glog"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestImplementsCrypter(t *testing.T) {
	object := New([]byte("AES256Key-32Characters1234567890"))
	var expected *Crypter
	err := AssertThat(object, Implements(expected))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCrypter(t *testing.T) {
	crypter := New([]byte("AES256Key-32Characters1234567890"))
	encrypted, err := crypter.Encrypt("hello world")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := crypter.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(plain, Is("hello world")); err != nil {
		t.Fatal(err)
	}
}
