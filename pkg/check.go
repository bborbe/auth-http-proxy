// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

//counterfeiter:generate -o ../mocks/check.go --fake-name Check . Check
type Check interface {
	Check(username string, password string) (bool, error)
}

type CheckFunc func(username string, password string) (bool, error)

func (c CheckFunc) Check(username string, password string) (bool, error) {
	return c(username, password)
}
