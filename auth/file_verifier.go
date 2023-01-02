// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/golang/glog"
)

type UserFile string

func (u UserFile) String() string {
	return string(u)
}

type fileAuth struct {
	userFile UserFile
}

func NewFileAuth(userFile UserFile) Verifier {
	return &fileAuth{
		userFile: userFile,
	}
}

func (a *fileAuth) Verify(username UserName, password Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	file, err := os.Open(a.userFile.String())
	if err != nil {
		glog.Warningf("open user file %v failed: %v", a.userFile.String(), err)
		return false, err
	}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil && err != io.EOF {
			glog.Warningf("read line of file %v failed: %v", a.userFile.String(), err)
			return false, err
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && parts[0] == username.String() {
			if parts[1] == password.String() {
				glog.V(2).Infof("found user and password is valid")
				return true, nil
			} else {
				glog.V(1).Infof("found user and password is invalid")
				return false, nil
			}
		}
		if err == io.EOF {
			glog.V(1).Infof("reach eof, user %v not found", username)
			return false, nil
		}
	}
}
