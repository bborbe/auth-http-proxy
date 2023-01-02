// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth_test

import (
	"github.com/bborbe/auth-http-proxy/auth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Crypter", func() {
	It("Compiles", func() {
		var err error
		crypter := auth.NewCrypter([]byte("AES256Key-32Characters1234567890"))
		encrypted, err := crypter.Encrypt("hello world")
		Expect(err).To(BeNil())
		plain, err := crypter.Decrypt(encrypted)
		Expect(err).To(BeNil())
		Expect(plain).To(Equal("hello world"))
	})
})
