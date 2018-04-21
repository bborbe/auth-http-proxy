package file

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/bborbe/auth-http-proxy/model"
	"github.com/golang/glog"
)

type auth struct {
	userFile model.UserFile
}

func New(userFile model.UserFile) *auth {
	a := new(auth)
	a.userFile = userFile
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	file, err := os.Open(a.userFile.String())
	if err != nil {
		glog.Warningf("open user file %v failed", a.userFile.String(), err)
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
