// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import "net/http"

//counterfeiter:generate -o ../mocks/http-handler.go --fake-name HttpHandler . HttpHandler
type HttpHandler http.Handler
