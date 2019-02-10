package auth

type Verifier interface {
	Verify(UserName, Password) (bool, error)
}
