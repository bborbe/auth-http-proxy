// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/auth-http-proxy/mocks"
	"github.com/bborbe/auth-http-proxy/pkg"
)

var _ = Describe("AuthHtmlHandler", func() {
	var ctx context.Context
	var err error
	var basicHandler http.Handler
	var recorder *httptest.ResponseRecorder
	var req *http.Request
	var subhandler *mocks.HttpHandler
	var check *mocks.Check
	var crypter *mocks.Crypter
	BeforeEach(func() {
		ctx = context.Background()

		subhandler = &mocks.HttpHandler{}

		check = &mocks.Check{}
		check.CheckReturns(true, nil)

		crypter = &mocks.Crypter{}

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		Expect(err).To(BeNil())
		recorder = httptest.NewRecorder()
	})
	JustBeforeEach(func() {
		basicHandler = pkg.NewAuthHtmlHandler(subhandler, check, crypter)
		basicHandler.ServeHTTP(recorder, req)
	})
	Context("Success", func() {
		BeforeEach(func() {
			req.SetBasicAuth("myuser", "mypass")
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("calls subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
			argResp, argRequest := subhandler.ServeHTTPArgsForCall(0)
			Expect(argResp).NotTo(BeNil())
			Expect(argRequest).NotTo(BeNil())
			Expect(argRequest.Header.Get(pkg.ForwardForUserHeader)).To(Equal("myuser"))
		})
	})
})
