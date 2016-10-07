package main

import (
	"fmt"
	"net"
	"net/http"
	"runtime"

	"time"

	auth_client "github.com/bborbe/auth/client/verify_group_service"
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/crypter"
	"github.com/bborbe/http_handler/forward"
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/bborbe/auth_http_proxy/verifier"
	auth_verifier "github.com/bborbe/auth_http_proxy/verifier/auth"
	"github.com/bborbe/auth_http_proxy/verifier/cache"
	crowd_verifier "github.com/bborbe/auth_http_proxy/verifier/crowd"
	file_verifier "github.com/bborbe/auth_http_proxy/verifier/file"
	ldap_verifier "github.com/bborbe/auth_http_proxy/verifier/ldap"
	flag "github.com/bborbe/flagenv"
	http_client_builder "github.com/bborbe/http/client_builder"
	http_requestbuilder "github.com/bborbe/http/requestbuilder"
	"github.com/bborbe/http_handler/auth_basic"
	"github.com/bborbe/http_handler/auth_html"
	"github.com/bborbe/http_handler/check"
	debug_handler "github.com/bborbe/http_handler/debug"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"go.jona.me/crowd"
)

const (
	defaultPort                      int = 8080
	parameterPort                        = "port"
	parameterTargetAddress               = "target-address"
	parameterTargetHealthzUrl            = "target-healthz-url"
	parameterBasicAuthRealm              = "basic-auth-realm"
	parameterRequiredGroups              = "required-groups"
	parameterVerifierType                = "verifier"
	parameterKind                        = "kind"
	parameterConfig                      = "config"
	parameterFileUsers                   = "file-users"
	parameterAuthUrl                     = "auth-url"
	parameterAuthApplicationName         = "auth-application-name"
	parameterAuthApplicationPassword     = "auth-application-password"
	parameterLdapBase                    = "ldap-base"
	parameterLdapUserDn                  = "ldap-user-dn"
	parameterLdapGroupDn                 = "ldap-group-dn"
	parameterLdapHost                    = "ldap-host"
	parameterLdapPort                    = "ldap-port"
	parameterLdapUseSSL                  = "ldap-use-ssl"
	parameterLdapBindDN                  = "ldap-bind-dn"
	parameterLdapBindPassword            = "ldap-bind-password"
	parameterLdapUserFilter              = "ldap-user-filter"
	parameterLdapGroupFilter             = "ldap-group-filter"
	parameterCacheTTL                    = "cache-ttl"
	parameterLdapServerName              = "ldap-servername"
	parameterSecret                      = "secret"
	parameterCrowdURL                    = "crowd-url"
	parameterCrowdAppName                = "crowd-app-name"
	parameterCrowdAppPassword            = "crowd-app-password"
)

var (
	portPtr             = flag.Int(parameterPort, defaultPort, "port")
	authUrlPtr          = flag.String(parameterAuthUrl, "", "auth url")
	basicAuthRealmPtr   = flag.String(parameterBasicAuthRealm, "", "basic auth realm")
	targetAddressPtr    = flag.String(parameterTargetAddress, "", "target address")
	targetHealthzUrlPtr = flag.String(parameterTargetHealthzUrl, "", "target healthz address")
	verifierPtr         = flag.String(parameterVerifierType, "", "verifier (file,ldap,crowd,auth)")
	secretPtr           = flag.String(parameterSecret, "", "aes secret key (length: 32")
	kindPtr             = flag.String(parameterKind, "", "(basic,html)")
	configPtr           = flag.String(parameterConfig, "", "config")
	requiredGroupsPtr   = flag.String(parameterRequiredGroups, "", "required groups reperated by comma")
	cacheTTLPtr         = flag.Duration(parameterCacheTTL, 5*time.Minute, "cache ttl")
	// file params
	fileUseresPtr = flag.String(parameterFileUsers, "", "users")
	// auth params
	authApplicationNamePtr     = flag.String(parameterAuthApplicationName, "", "auth application name")
	authApplicationPasswordPtr = flag.String(parameterAuthApplicationPassword, "", "auth application password")
	// ldap params
	ldapBasePtr         = flag.String(parameterLdapBase, "", "ldap-base")
	ldapHostPtr         = flag.String(parameterLdapHost, "", "ldap-host")
	ldapServerNamePtr   = flag.String(parameterLdapServerName, "", "ldap-servername")
	ldapPortPtr         = flag.Int(parameterLdapPort, 0, "ldap-port")
	ldapUseSSLPtr       = flag.Bool(parameterLdapUseSSL, false, "ldap-use-ssl")
	ldapBindDNPtr       = flag.String(parameterLdapBindDN, "", "ldap-bind-dn")
	ldapBindPasswordPtr = flag.String(parameterLdapBindPassword, "", "ldap-bind-password")
	ldapUserFilterPtr   = flag.String(parameterLdapUserFilter, "", "ldap-user-filter")
	ldapGroupFilterPtr  = flag.String(parameterLdapGroupFilter, "", "ldap-group-filter")
	ldapUserDnPtr       = flag.String(parameterLdapUserDn, "", "ldap-user-dn")
	ldapGroupDnPtr      = flag.String(parameterLdapGroupDn, "", "ldap-group-dn")
	// crowd
	crowdURLPtr     = flag.String(parameterCrowdURL, "", "crowd url")
	crowdAppNamePtr = flag.String(parameterCrowdAppName, "", "crowd app name")
	crowdAppPassPtr = flag.String(parameterCrowdAppPassword, "", "crowd app password")
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := do(); err != nil {
		glog.Exit(err)
	}
}

func do() error {
	config, err := createConfig()
	if err != nil {
		return err
	}
	server, err := createServer(config)
	if err != nil {
		return err
	}
	glog.V(2).Infof("start server")
	return gracehttp.Serve(server)
}

func createConfig() (*model.Config, error) {
	var config *model.Config
	var err error
	configPath := model.ConfigPath(*configPtr)
	if configPath.IsValue() {
		glog.V(2).Infof("parse config from file: %v", configPath)
		config, err = configPath.Parse()
		if err != nil {
			glog.Warningf("parse config failed: %v", err)
			return nil, err
		}
	} else {
		glog.V(2).Infof("create empty config")
		config = new(model.Config)
	}
	if config.Port <= 0 {
		config.Port = model.Port(*portPtr)
	}
	if len(config.Kind) == 0 {
		config.Kind = model.Kind(*kindPtr)
	}
	if len(config.UserFile) == 0 {
		config.UserFile = model.UserFile(*fileUseresPtr)
	}
	if len(config.VerifierType) == 0 {
		config.VerifierType = model.VerifierType(*verifierPtr)
	}
	if len(config.Secret) == 0 {
		config.Secret = model.Secret(*secretPtr)
	}
	if len(config.BasicAuthRealm) == 0 {
		config.BasicAuthRealm = model.BasicAuthRealm(*basicAuthRealmPtr)
	}
	if len(config.TargetHealthzUrl) == 0 {
		config.TargetHealthzUrl = model.TargetHealthzUrl(*targetHealthzUrlPtr)
	}
	if len(config.TargetAddress) == 0 {
		config.TargetAddress = model.TargetAddress(*targetAddressPtr)
	}
	if config.CacheTTL.IsEmpty() {
		config.CacheTTL = model.CacheTTL(*cacheTTLPtr)
	}
	if len(config.AuthUrl) == 0 {
		config.AuthUrl = model.AuthUrl(*authUrlPtr)
	}
	if len(config.AuthApplicationName) == 0 {
		config.AuthApplicationName = model.AuthApplicationName(*authApplicationNamePtr)
	}
	if len(config.AuthApplicationPassword) == 0 {
		config.AuthApplicationPassword = model.AuthApplicationPassword(*authApplicationPasswordPtr)
	}
	if len(config.RequiredGroups) == 0 {
		config.RequiredGroups = model.CreateGroupsFromString(*requiredGroupsPtr)
	}
	if len(config.LdapBase) == 0 {
		config.LdapBase = model.LdapBase(*ldapBasePtr)
	}
	if len(config.LdapHost) == 0 {
		config.LdapHost = model.LdapHost(*ldapHostPtr)
	}
	if len(config.LdapServerName) == 0 {
		config.LdapServerName = model.LdapServerName(*ldapServerNamePtr)
	}
	if config.LdapPort <= 0 {
		config.LdapPort = model.LdapPort(*ldapPortPtr)
	}
	if !config.LdapUseSSL {
		config.LdapUseSSL = model.LdapUseSSL(*ldapUseSSLPtr)
	}
	if len(config.LdapBindDN) == 0 {
		config.LdapBindDN = model.LdapBindDN(*ldapBindDNPtr)
	}
	if len(config.LdapBindPassword) == 0 {
		config.LdapBindPassword = model.LdapBindPassword(*ldapBindPasswordPtr)
	}
	if len(config.LdapUserFilter) == 0 {
		config.LdapUserFilter = model.LdapUserFilter(*ldapUserFilterPtr)
	}
	if len(config.LdapGroupFilter) == 0 {
		config.LdapGroupFilter = model.LdapGroupFilter(*ldapGroupFilterPtr)
	}
	if len(config.LdapUserDn) == 0 {
		config.LdapUserDn = model.LdapUserDn(*ldapUserDnPtr)
	}
	if len(config.LdapGroupDn) == 0 {
		config.LdapGroupDn = model.LdapGroupDn(*ldapGroupDnPtr)
	}
	if len(config.CrowdURL) == 0 {
		config.CrowdURL = model.CrowdURL(*crowdURLPtr)
	}
	if len(config.CrowdAppName) == 0 {
		config.CrowdAppName = model.CrowdAppName(*crowdAppNamePtr)
	}
	if len(config.CrowdAppPassword) == 0 {
		config.CrowdAppPassword = model.CrowdAppPassword(*crowdAppPassPtr)
	}
	return config, nil
}

func createServer(config *model.Config) (*http.Server, error) {
	glog.V(2).Infof("create server")
	handler, err := createHandler(config)
	if err != nil {
		return nil, err
	}
	if config.Port <= 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterPort)
	}
	return &http.Server{Addr: fmt.Sprintf(":%d", config.Port), Handler: handler}, nil
}

func createHealthzCheck(config *model.Config) func() error {
	if len(config.TargetHealthzUrl) > 0 {
		return func() error {
			resp, err := http.Get(config.TargetHealthzUrl.String())
			if err != nil {
				glog.V(2).Infof("check url %v failed: %v", config.TargetHealthzUrl, err)
				return err
			}
			if resp.StatusCode/100 != 2 {
				glog.V(2).Infof("check url %v has wrong status: %v", config.TargetHealthzUrl, resp.Status)
				return fmt.Errorf("check url %v has wrong status: %v", config.TargetHealthzUrl, resp.Status)
			}
			return nil
		}
	}
	return func() error {
		conn, err := net.Dial("tcp", config.TargetAddress.String())
		if err != nil {
			glog.V(2).Infof("tcp connection to %v failed: %v", config.TargetAddress, err)
			return err
		}
		glog.V(2).Infof("tcp connection to %v success", config.TargetAddress)
		return conn.Close()
	}
}

func createHandler(config *model.Config) (http.Handler, error) {
	glog.V(2).Infof("create handler")

	filter, err := createHttpFilter(config)
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	router.NotFoundHandler = filter
	router.Path("/healthz").Handler(check.New(createHealthzCheck(config)))

	var handler http.Handler = router

	if glog.V(4) {
		glog.V(2).Infof("add debug handler")
		handler = debug_handler.New(handler)
	}
	return handler, nil
}

func createHttpFilter(config *model.Config) (http.Handler, error) {
	if len(config.Kind) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterKind)
	}
	glog.V(2).Infof("get auth filter for: %v", config.Kind)
	switch config.Kind {
	case "html":
		return createHtmlAuthHttpFilter(config)
	case "basic":
		return createBasicAuthHttpFilter(config)
	}
	return nil, fmt.Errorf("parameter %s invalid", parameterKind)
}

func createHtmlAuthHttpFilter(config *model.Config) (http.Handler, error) {
	verifier, err := createVerifier(config)
	if err != nil {
		return nil, err
	}
	forwardHandler, err := createForwardHandler(config)
	if err != nil {
		return nil, err
	}
	check := func(username string, password string) (bool, error) {
		return verifier.Verify(model.UserName(username), model.Password(password))
	}
	if len(config.Secret) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterSecret)
	}
	if len(config.Secret)%16 != 0 {
		return nil, fmt.Errorf("parameter %s invalid length", parameterSecret)
	}
	handler := auth_html.New(forwardHandler.ServeHTTP, check, crypter.New(config.Secret.Bytes()))
	return handler, nil
}

func createBasicAuthHttpFilter(config *model.Config) (http.Handler, error) {
	verifier, err := createVerifier(config)
	if err != nil {
		return nil, err
	}
	forwardHandler, err := createForwardHandler(config)
	if err != nil {
		return nil, err
	}
	if len(config.BasicAuthRealm) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterBasicAuthRealm)
	}
	check := func(username string, password string) (bool, error) {
		return verifier.Verify(model.UserName(username), model.Password(password))
	}
	handler := auth_basic.New(forwardHandler.ServeHTTP, check, config.BasicAuthRealm.String())
	return handler, nil
}

func createForwardHandler(config *model.Config) (http.Handler, error) {
	dialer := (&net.Dialer{
		Timeout: http_client_builder.DEFAULT_TIMEOUT,
	})
	if len(config.TargetAddress) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterTargetAddress)
	}
	forwardHandler := forward.New(config.TargetAddress.String(),
		func(address string, req *http.Request) (resp *http.Response, err error) {
			return http_client_builder.New().WithoutProxy().WithoutRedirects().WithDialFunc(
				func(network, address string) (net.Conn, error) {
					return dialer.Dial(network, config.TargetAddress.String())
				}).BuildRoundTripper().RoundTrip(req)
		})
	return forwardHandler, nil
}

func createVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.VerifierType) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterVerifierType)
	}
	glog.V(2).Infof("get verifier for: %v", config.VerifierType)
	switch config.VerifierType {
	case "auth":
		return createAuthVerifier(config)
	case "ldap":
		return createLdapVerifier(config)
	case "file":
		return createFileVerifier(config)
	case "crowd":
		return createCrowdVerifier(config)
	}
	return nil, fmt.Errorf("parameter %s invalid", parameterVerifierType)
}

func createAuthVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.AuthUrl) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthUrl)
	}
	if len(config.AuthApplicationName) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationName)
	}
	if len(config.AuthApplicationPassword) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationPassword)
	}
	httpRequestBuilderProvider := http_requestbuilder.NewHTTPRequestBuilderProvider()
	httpClient := http_client_builder.New().WithoutProxy().Build()
	authClient := auth_client.New(httpClient.Do, httpRequestBuilderProvider, auth_model.Url(config.AuthUrl), auth_model.ApplicationName(config.AuthApplicationName), auth_model.ApplicationPassword(config.AuthApplicationPassword))
	return cache.New(auth_verifier.New(
		authClient.Auth,
		config.RequiredGroups...,
	), config.CacheTTL), nil
}

func createLdapVerifier(config *model.Config) (verifier.Verifier, error) {
	return cache.New(ldap_verifier.New(
		config.LdapBase,
		config.LdapHost,
		config.LdapServerName,
		config.LdapPort,
		config.LdapUseSSL,
		config.LdapBindDN,
		config.LdapBindPassword,
		config.LdapUserFilter,
		config.LdapGroupFilter,
		config.LdapUserDn,
		config.LdapGroupDn,
		config.RequiredGroups...,
	), config.CacheTTL), nil
}

func createFileVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.UserFile) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterFileUsers)
	}
	return cache.New(file_verifier.New(config.UserFile), config.CacheTTL), nil
}

func createCrowdVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.CrowdAppName) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterCrowdAppName)
	}
	if len(config.CrowdAppPassword) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterCrowdAppPassword)
	}
	if len(config.CrowdURL) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterCrowdURL)
	}

	crowdClient, err := crowd.New(config.CrowdAppName.String(), config.CrowdAppPassword.String(), config.CrowdURL.String())
	if err != nil {
		glog.V(2).Infof("create crowd client failed: %v", err)
		return nil, err
	}

	return cache.New(crowd_verifier.New(crowdClient.Authenticate), config.CacheTTL), nil
}
