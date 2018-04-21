package auth_html

import (
	"html/template"
	"net/http"
	"time"

	"github.com/bborbe/http/header"
	"github.com/golang/glog"
)

const (
	fieldNameLogin    = "login"
	fieldNamePassword = "password"
	cookieName        = "auth-http-proxy-token"
	loginDuration     = 24 * time.Hour
)

type Check func(username string, password string) (bool, error)

type Crypter interface {
	Encrypt(text string) (string, error)
	Decrypt(text string) (string, error)
}

type handler struct {
	handler http.HandlerFunc
	check   Check
	crypter Crypter
}

func New(
	subhandler http.HandlerFunc,
	check Check,
	crypter Crypter,
) *handler {
	h := new(handler)
	h.handler = subhandler
	h.check = check
	h.crypter = crypter
	return h
}

func (h *handler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	glog.V(4).Infof("check html auth")
	if err := h.serveHTTP(responseWriter, request); err != nil {
		glog.Warningf("check html auth failed: %v", err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *handler) serveHTTP(responseWriter http.ResponseWriter, request *http.Request) error {
	glog.V(4).Infof("check html auth")
	valid, err := h.validateLogin(request)
	if err != nil {
		glog.V(2).Infof("validate login failed: %v", err)
		return err
	}
	if valid {
		glog.V(4).Infof("login is valid, forward request")
		h.handler(responseWriter, request)
		return nil
	}
	return h.validateLoginParams(responseWriter, request)
}

func (h *handler) validateLogin(request *http.Request) (bool, error) {
	if valid, _ := h.validateLoginBasic(request); valid {
		return true, nil
	}
	valid, err := h.validateLoginCookie(request)
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (h *handler) validateLoginBasic(request *http.Request) (bool, error) {
	glog.V(4).Infof("validate login via basic")
	user, pass, err := header.ParseAuthorizationBasisHttpRequest(request)
	if err != nil {
		glog.V(2).Infof("parse basic authorization header failed: %v", err)
		return false, err
	}
	result, err := h.check(user, pass)
	if err != nil {
		glog.Warningf("check auth for user %v failed: %v", user, err)
		return false, err
	}
	glog.V(4).Infof("validate login via basic => %v", result)
	return result, nil
}

func (h *handler) validateLoginCookie(request *http.Request) (bool, error) {
	glog.V(4).Infof("validate login via cookie")
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		glog.V(2).Infof("get cookie %v failed: %v", cookieName, err)
		return false, nil
	}
	data, err := h.crypter.Decrypt(cookie.Value)
	if err != nil {
		glog.V(2).Infof("decrypt cookie value failed: %v", err)
		return false, nil
	}
	user, pass, err := header.ParseAuthorizationToken(data)
	if err != nil {
		glog.V(2).Infof("parse cookie failed: %v", err)
		return false, nil
	}
	result, err := h.check(user, pass)
	if err != nil {
		glog.Warningf("check auth for user %v failed: %v", user, err)
		return false, err
	}
	glog.V(4).Infof("validate login via cookie => %v", result)
	return result, nil
}

func (h *handler) validateLoginParams(responseWriter http.ResponseWriter, request *http.Request) error {
	glog.V(4).Infof("validate login via params")
	login := request.FormValue(fieldNameLogin)
	password := request.FormValue(fieldNamePassword)
	if len(login) == 0 || len(password) == 0 {
		glog.V(4).Infof("login or password empty => skip")
		return h.loginForm(responseWriter)
	}
	valid, err := h.check(login, password)
	if err != nil {
		glog.V(2).Infof("check login failed: %v", err)
		return err
	}
	if !valid {
		glog.V(4).Infof("login failed, show login form")
		return h.loginForm(responseWriter)
	}
	glog.V(4).Infof("login success, set cookie")
	data, err := h.crypter.Encrypt(header.CreateAuthorizationToken(login, password))
	if err != nil {
		glog.V(2).Infof("encrypt failed: %v", err)
		return err
	}
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     cookieName,
		Value:    data,
		Expires:  createExpires(),
		Path:     "/",
		Domain:   request.URL.Host,
		HttpOnly: true,
		Secure:   isSecureRequest(request),
	})
	target := request.RequestURI
	glog.V(4).Infof("login success, redirect to %v", target)
	return h.redirect(responseWriter, target)
}

func isSecureRequest(request *http.Request) bool {
	return request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https"
}

func createExpires() time.Time {
	return time.Now().Add(loginDuration)
}

func (h *handler) loginForm(responseWriter http.ResponseWriter) error {
	glog.V(4).Infof("login form")
	var t = template.Must(template.New("loginForm").Parse(HTML_LOGIN_FORM))
	data := struct {
		FieldNameLogin    string
		FieldNamePassword string
	}{
		FieldNameLogin:    fieldNameLogin,
		FieldNamePassword: fieldNamePassword,
	}
	responseWriter.Header().Add("Content-Type", "text/html")
	responseWriter.WriteHeader(http.StatusUnauthorized)
	return t.Execute(responseWriter, data)
}

func (h *handler) redirect(responseWriter http.ResponseWriter, target string) error {
	glog.V(4).Infof("login form")
	var t = template.Must(template.New("loginForm").Parse(HTML_REDIRECT))
	data := struct {
		Target string
	}{
		Target: target,
	}
	responseWriter.Header().Add("Content-Type", "text/html")
	responseWriter.WriteHeader(http.StatusUnauthorized)
	return t.Execute(responseWriter, data)
}

const HTML_REDIRECT = `<!DOCTYPE html>
<html>
<title>Login Success</title>
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta http-equiv="Content-Language" content="en">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="author" content="Benjamin Borbe">
<meta name="description" content="Login Success">
<meta http-equiv="refresh" content="0;URL={{.Target}}">
<link rel="icon" href="data:;base64,=">
<link rel="stylesheet" type="text/css" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
<link rel="stylesheet" type="text/css" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap-theme.min.css">
<script type="text/javascript">
window.location.href='{{.Target}}';
</script>
<style>
html {
	position: relative;
	min-height: 100%;
}
body {
	margin-top: 60px;
}
</style>
</script>
</head>
<body>
<div class="view-container">
	<div class="container">
		<div class="starter-template">
			<h1>Login Success</h1>
			<a href="{{.Target}}">{{.Target}}</a>
		</div>
	</div>
</div>
</body>
</html>
`

const HTML_LOGIN_FORM = `<!DOCTYPE html>
<html>
<title>Login Form</title>
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta http-equiv="Content-Language" content="en">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="author" content="Benjamin Borbe">
<meta name="description" content="Login Form">
<link rel="icon" href="data:;base64,=">
<link rel="stylesheet" type="text/css" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
<link rel="stylesheet" type="text/css" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap-theme.min.css">
<style>
html {
	position: relative;
	min-height: 100%;
}
body {
	margin-top: 60px;
}
</style>
</script>
</head>
<body>
<div class="view-container">
	<div class="container">
		<div class="starter-template">
			<form name="loginForm" class="form-horizontal" action="" method="post">
				<fieldset>
					<legend>Login required</legend>

					<div class="form-group">
						<label class="col-md-3 control-label" for="{{.FieldNameLogin}}">Login</label>
						<div class="col-md-3">
							<input type="text" id="{{.FieldNameLogin}}" name="{{.FieldNameLogin}}" min="1" max="255" required="" placeholder="login" class="form-control input-md">
						</div>
					</div>
					<div class="form-group">
						<label class="col-md-3 control-label" for="{{.FieldNamePassword}}">Password</label>
						<div class="col-md-3">
							<input type="password" id="{{.FieldNamePassword}}" name="{{.FieldNamePassword}}" min="1" max="255" required="" placeholder="password" class="form-control input-md">
						</div>
					</div>
					<div class="form-group">
						<label class="col-md-3 control-label" for="singlebutton"></label>
						<div class="col-md-3">
							<input type="submit" id="singlebutton" name="singlebutton" class="btn btn-primary" value="login">
						</div>
					</div>
				</fieldset>
			</form>
		</div>
	</div>
</div>
</body>
</html>`
