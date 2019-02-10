package auth_test

import (
	"testing"

	. "github.com/bborbe/assert"
	"github.com/bborbe/auth-http-proxy/auth"
)

func TestCrypter(t *testing.T) {
	crypter := auth.NewCrypter([]byte("AES256Key-32Characters1234567890"))
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
