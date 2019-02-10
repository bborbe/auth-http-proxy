package auth

type UserName string

func (u UserName) String() string {
	return string(u)
}

type Password string

func (p Password) String() string {
	return string(p)
}

type Verifier interface {
	Verify(UserName, Password) (bool, error)
}
