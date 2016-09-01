package model

type VerifierType string

func (v VerifierType) String() string {
	return string(v)
}

type UserName string

func (u UserName) String() string {
	return string(u)
}

type Password string

func (p Password) String() string {
	return string(p)
}

type UserFile string
