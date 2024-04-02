// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/auth-http-proxy/pkg"
)

var _ = Describe("Crypter", func() {
	It("Compiles", func() {
		var err error
		crypter := pkg.NewCrypter([]byte("AES256Key-32Characters1234567890"))
		encrypted, err := crypter.Encrypt("hello world")
		Expect(err).To(BeNil())
		plain, err := crypter.Decrypt(encrypted)
		Expect(err).To(BeNil())
		Expect(plain).To(Equal("hello world"))
	})
})
