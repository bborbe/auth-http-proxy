package factory

import (
	"fmt"
	"net"
	"net/http"

	"github.com/bborbe/auth-http-proxy/crypter"
	"github.com/bborbe/auth-http-proxy/model"
	"github.com/bborbe/auth-http-proxy/verifier"
	"github.com/bborbe/auth-http-proxy/verifier/cache"
	crowd_verifier "github.com/bborbe/auth-http-proxy/verifier/crowd"
	file_verifier "github.com/bborbe/auth-http-proxy/verifier/file"
	ldap_verifier "github.com/bborbe/auth-http-proxy/verifier/ldap"
	http_client_builder "github.com/bborbe/http/client_builder"
	"github.com/bborbe/http_handler/auth_basic"
	"github.com/bborbe/http_handler/auth_html"
	"github.com/bborbe/http_handler/check"
	debug_handler "github.com/bborbe/http_handler/debug"
	"github.com/bborbe/http_handler/forward"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"go.jona.me/crowd"
	"github.com/bborbe/auth-http-proxy/ldap"
)

type authHttpProxyFactory struct {
	config      model.Config
	crowdClient crowd.Crowd
}

func New(
	config model.Config,
	crowdClient crowd.Crowd,
) *authHttpProxyFactory {
	a := new(authHttpProxyFactory)
	a.config = config
	a.crowdClient = crowdClient
	return a
}

func (a *authHttpProxyFactory) HttpServer() *http.Server {
	glog.V(2).Infof("create http server on %s", a.config.Port.Address())
	return &http.Server{Addr: a.config.Port.Address(), Handler: a.Handler()}
}

func (a *authHttpProxyFactory) createHealthzCheck() func() error {
	if len(a.config.TargetHealthzUrl) > 0 {
		return func() error {
			resp, err := http.Get(a.config.TargetHealthzUrl.String())
			if err != nil {
				glog.V(2).Infof("check url %v failed: %v", a.config.TargetHealthzUrl, err)
				return err
			}
			if resp.StatusCode/100 != 2 {
				glog.V(2).Infof("check url %v has wrong status: %v", a.config.TargetHealthzUrl, resp.Status)
				return fmt.Errorf("check url %v has wrong status: %v", a.config.TargetHealthzUrl, resp.Status)
			}
			return nil
		}
	}
	return func() error {
		conn, err := net.Dial("tcp", a.config.TargetAddress.String())
		if err != nil {
			glog.V(2).Infof("tcp connection to %v failed: %v", a.config.TargetAddress, err)
			return err
		}
		glog.V(2).Infof("tcp connection to %v success", a.config.TargetAddress)
		return conn.Close()
	}
}

func (a *authHttpProxyFactory) Handler() http.Handler {
	glog.V(2).Infof("create handler")

	checkHandler := check.New(a.createHealthzCheck())
	router := mux.NewRouter()
	router.Path("/healthz").Handler(checkHandler)
	router.Path("/readiness").Handler(checkHandler)
	router.NotFoundHandler = a.createHttpFilter()

	var handler http.Handler = router

	if glog.V(4) {
		glog.V(2).Infof("add debug handler")
		handler = debug_handler.New(handler)
	}
	return handler
}

func (a *authHttpProxyFactory) createHttpFilter() http.Handler {
	glog.V(2).Infof("get auth filter for: %v", a.config.Kind)
	switch a.config.Kind {
	case "html":
		return a.createHtmlAuthHttpFilter()
	case "basic":
		return a.createBasicAuthHttpFilter()
	}
	return nil
}

func (a *authHttpProxyFactory) createHtmlAuthHttpFilter() http.Handler {
	v := a.createVerifier()
	c := func(username string, password string) (bool, error) {
		return v.Verify(model.UserName(username), model.Password(password))
	}
	return auth_html.New(a.createForwardHandler().ServeHTTP, c, crypter.New(a.config.Secret.Bytes()))
}

func (a *authHttpProxyFactory) createBasicAuthHttpFilter() http.Handler {
	v := a.createVerifier()
	c := func(username string, password string) (bool, error) {
		return v.Verify(model.UserName(username), model.Password(password))
	}
	return auth_basic.New(a.createForwardHandler().ServeHTTP, c, a.config.BasicAuthRealm.String())
}

func (a *authHttpProxyFactory) createForwardHandler() http.Handler {
	dialer := (&net.Dialer{
		Timeout: http_client_builder.DEFAULT_TIMEOUT,
	})
	return forward.New(a.config.TargetAddress.String(),
		func(address string, req *http.Request) (resp *http.Response, err error) {
			return http_client_builder.New().WithoutProxy().WithoutRedirects().WithDialFunc(
				func(network, address string) (net.Conn, error) {
					return dialer.Dial(network, a.config.TargetAddress.String())
				}).BuildRoundTripper().RoundTrip(req)
		})
}

func (a *authHttpProxyFactory) createVerifier() verifier.Verifier {
	glog.V(2).Infof("get verifier for: %v", a.config.VerifierType)
	switch a.config.VerifierType {
	case "ldap":
		return a.createLdapVerifier()
	case "file":
		return a.createFileVerifier()
	case "crowd":
		return a.createCrowdVerifier()
	}
	return nil
}

func (a *authHttpProxyFactory) createLdapVerifier() verifier.Verifier {
	auth := &ldap_verifier.Auth{
		Authenticator: ldap.NewAuthenticator(
			a.config.LdapBaseDn,
			a.config.LdapHost,
			a.config.LdapServerName,
			a.config.LdapPort,
			a.config.LdapUseSSL,
			a.config.LdapSkipTls,
			a.config.LdapBindDN,
			a.config.LdapBindPassword,
			a.config.LdapUserDn,
			a.config.LdapUserFilter,
			a.config.LdapUserField,
			a.config.LdapGroupDn,
			a.config.LdapGroupFilter,
			a.config.LdapGroupField,
		),
		RequiredGroups: a.config.RequiredGroups,
	}
	return cache.New(auth, a.config.CacheTTL)
}

func (a *authHttpProxyFactory) createFileVerifier() verifier.Verifier {
	return cache.New(file_verifier.New(a.config.UserFile), a.config.CacheTTL)
}

func (a *authHttpProxyFactory) createCrowdVerifier() verifier.Verifier {
	return cache.New(crowd_verifier.New(a.crowdClient.Authenticate), a.config.CacheTTL)
}

func (a *authHttpProxyFactory) httpClient() *http.Client {
	return http_client_builder.New().WithoutProxy().Build()
}
