// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bborbe/auth-http-proxy/auth"
	flag "github.com/bborbe/flagenv"
	http_client_builder "github.com/bborbe/http/client_builder"
	"github.com/bborbe/http_handler/auth_basic"
	"github.com/bborbe/http_handler/auth_html"
	"github.com/bborbe/http_handler/check"
	"github.com/bborbe/http_handler/debug"
	"github.com/bborbe/http_handler/forward"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.jona.me/crowd"
)

var (
	portPtr             = flag.Int("port", 8080, "port")
	basicAuthRealmPtr   = flag.String("basic-auth-realm", "", "basic auth realm")
	targetAddressPtr    = flag.String("target-address", "", "target address")
	targetHealthzUrlPtr = flag.String("target-healthz-url", "", "target healthz address")
	verifierPtr         = flag.String("verifier", "", "verifier (file,ldap,crowd,auth)")
	secretPtr           = flag.String("secret", "", "aes secret key (length: 32")
	kindPtr             = flag.String("kind", "", "(basic,html)")
	configPtr           = flag.String("config", "", "config")
	requiredGroupsPtr   = flag.String("required-groups", "", "required groups reperated by comma")
	cacheTTLPtr         = flag.Duration("cache-ttl", 5*time.Minute, "cache ttl")

	// file params
	fileUseresPtr = flag.String("file-users", "", "users")

	// ldap params
	ldapBaseDnPtr       = flag.String("ldap-base-dn", "", "ldap-base-dn")
	ldapHostPtr         = flag.String("ldap-host", "", "ldap-host")
	ldapServerNamePtr   = flag.String("ldap-servername", "", "ldap-servername")
	ldapPortPtr         = flag.Int("ldap-port", 0, "ldap-port")
	ldapUseSSLPtr       = flag.Bool("ldap-use-ssl", false, "ldap-use-ssl")
	ldapSkipTlsPtr      = flag.Bool("ldap-skip-tls", false, "ldap-skip-tls")
	ldapBindDNPtr       = flag.String("ldap-bind-dn", "", "ldap-bind-dn")
	ldapBindPasswordPtr = flag.String("ldap-bind-password", "", "ldap-bind-password")
	ldapUserFilterPtr   = flag.String("ldap-user-filter", "", "ldap-user-filter")
	ldapGroupFilterPtr  = flag.String("ldap-group-filter", "", "ldap-group-filter")
	ldapUserDnPtr       = flag.String("ldap-user-dn", "", "ldap-user-dn")
	ldapGroupDnPtr      = flag.String("ldap-group-dn", "", "ldap-group-dn")
	ldapUserFieldPtr    = flag.String("ldap-user-field", "", "ldap-user-field")
	ldapGroupFieldPtr   = flag.String("ldap-group-field", "", "ldap-group-field")

	// crowd
	crowdURLPtr     = flag.String("crowd-url", "", "crowd url")
	crowdAppNamePtr = flag.String("crowd-app-name", "", "crowd app name")
	crowdAppPassPtr = flag.String("crowd-app-password", "", "crowd app password")
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := &application{}
	flag.Parse()

	if err := app.parseConfig(); err != nil {
		glog.Exit(err)
	}

	if err := app.validate(); err != nil {
		glog.Exit(err)
	}

	if err := app.run(); err != nil {
		glog.Exit(err)
	}
}

type application struct {
	Port             Port                  `json:"port"`
	CacheTTL         auth.CacheTTL         `json:"cache-ttl"`
	TargetAddress    TargetAddress         `json:"target-address"`
	TargetHealthzUrl TargetHealthzUrl      `json:"target-healthz-url"`
	BasicAuthRealm   BasicAuthRealm        `json:"basic-auth-realm"`
	Secret           Secret                `json:"secret"`
	RequiredGroups   []auth.GroupName      `json:"required-groups"`
	VerifierType     VerifierType          `json:"verifier"`
	UserFile         auth.UserFile         `json:"file-users"`
	Kind             Kind                  `json:"kind"`
	LdapHost         auth.LdapHost         `json:"ldap-host"`
	LdapServerName   auth.LdapServerName   `json:"ldap-servername"`
	LdapPort         auth.LdapPort         `json:"ldap-port"`
	LdapUseSSL       auth.LdapUseSSL       `json:"ldap-use-ssl"`
	LdapSkipTls      auth.LdapSkipTls      `json:"ldap-skip-tls"`
	LdapBindDN       auth.LdapBindDN       `json:"ldap-bind-dn"`
	LdapBindPassword auth.LdapBindPassword `json:"ldap-bind-password"`
	LdapBaseDn       auth.LdapBaseDn       `json:"ldap-base-dn"`
	LdapUserDn       auth.LdapUserDn       `json:"ldap-user-dn"`
	LdapGroupDn      auth.LdapGroupDn      `json:"ldap-group-dn"`
	LdapUserFilter   auth.LdapUserFilter   `json:"ldap-user-filter"`
	LdapGroupFilter  auth.LdapGroupFilter  `json:"ldap-group-filter"`
	LdapUserField    auth.LdapUserField    `json:"ldap-user-field"`
	LdapGroupField   auth.LdapGroupField   `json:"ldap-group-field"`
	CrowdURL         CrowdURL              `json:"crowd-url"`
	CrowdAppName     CrowdAppName          `json:"crowd-app-name"`
	CrowdAppPassword CrowdAppPassword      `json:"crowd-app-password"`
}

func (a *application) parseConfig() error {
	if len(*configPtr) > 0 {
		file, err := os.Open(*configPtr)
		if err != nil {
			return errors.Wrapf(err, "read config %v failed", *configPtr)
		}
		err = json.NewDecoder(file).Decode(a)
		if err != nil {
			return errors.Wrap(err, "parse config json failed")
		}
	}
	if a.Port <= 0 {
		a.Port = Port(*portPtr)
	}
	if len(a.Kind) == 0 {
		a.Kind = Kind(*kindPtr)
	}
	if len(a.UserFile) == 0 {
		a.UserFile = auth.UserFile(*fileUseresPtr)
	}
	if len(a.VerifierType) == 0 {
		a.VerifierType = VerifierType(*verifierPtr)
	}
	if len(a.Secret) == 0 {
		a.Secret = Secret(*secretPtr)
	}
	if len(a.BasicAuthRealm) == 0 {
		a.BasicAuthRealm = BasicAuthRealm(*basicAuthRealmPtr)
	}
	if len(a.TargetHealthzUrl) == 0 {
		a.TargetHealthzUrl = TargetHealthzUrl(*targetHealthzUrlPtr)
	}
	if len(a.TargetAddress) == 0 {
		a.TargetAddress = TargetAddress(*targetAddressPtr)
	}
	if a.CacheTTL.IsEmpty() {
		a.CacheTTL = auth.CacheTTL(*cacheTTLPtr)
	}
	if len(a.RequiredGroups) == 0 {
		for _, groupName := range strings.Split(*requiredGroupsPtr, ",") {
			if len(groupName) > 0 {
				a.RequiredGroups = append(a.RequiredGroups, auth.GroupName(groupName))
			}
		}
		glog.V(1).Infof("required groups: %v", a.RequiredGroups)
	}
	if len(a.LdapBaseDn) == 0 {
		a.LdapBaseDn = auth.LdapBaseDn(*ldapBaseDnPtr)
	}
	if len(a.LdapHost) == 0 {
		a.LdapHost = auth.LdapHost(*ldapHostPtr)
	}
	if len(a.LdapServerName) == 0 {
		a.LdapServerName = auth.LdapServerName(*ldapServerNamePtr)
	}
	if a.LdapPort <= 0 {
		a.LdapPort = auth.LdapPort(*ldapPortPtr)
	}
	if !a.LdapUseSSL {
		a.LdapUseSSL = auth.LdapUseSSL(*ldapUseSSLPtr)
	}
	if !a.LdapSkipTls {
		a.LdapSkipTls = auth.LdapSkipTls(*ldapSkipTlsPtr)
	}
	if len(a.LdapBindDN) == 0 {
		a.LdapBindDN = auth.LdapBindDN(*ldapBindDNPtr)
	}
	if len(a.LdapBindPassword) == 0 {
		a.LdapBindPassword = auth.LdapBindPassword(*ldapBindPasswordPtr)
	}
	if len(a.LdapUserFilter) == 0 {
		a.LdapUserFilter = auth.LdapUserFilter(*ldapUserFilterPtr)
	}
	if len(a.LdapGroupFilter) == 0 {
		a.LdapGroupFilter = auth.LdapGroupFilter(*ldapGroupFilterPtr)
	}
	if len(a.LdapUserField) == 0 {
		a.LdapUserField = auth.LdapUserField(*ldapUserFieldPtr)
	}
	if len(a.LdapGroupField) == 0 {
		a.LdapGroupField = auth.LdapGroupField(*ldapGroupFieldPtr)
	}
	if len(a.LdapUserDn) == 0 {
		a.LdapUserDn = auth.LdapUserDn(*ldapUserDnPtr)
	}
	if len(a.LdapGroupDn) == 0 {
		a.LdapGroupDn = auth.LdapGroupDn(*ldapGroupDnPtr)
	}
	if len(a.CrowdURL) == 0 {
		a.CrowdURL = CrowdURL(*crowdURLPtr)
	}
	if len(a.CrowdAppName) == 0 {
		a.CrowdAppName = CrowdAppName(*crowdAppNamePtr)
	}
	if len(a.CrowdAppPassword) == 0 {
		a.CrowdAppPassword = CrowdAppPassword(*crowdAppPassPtr)
	}
	return nil
}

func (a *application) validate() error {
	if a.Port <= 0 {
		return fmt.Errorf("parameter Port missing")
	}
	if len(a.TargetAddress) == 0 {
		return fmt.Errorf("parameter TargetAddress missing")
	}
	if len(a.Kind) == 0 {
		return fmt.Errorf("parameter Kind missing")
	}
	if a.Kind != "basic" && a.Kind != "html" {
		return fmt.Errorf("parameter Kind invalid")
	}
	if len(a.VerifierType) == 0 {
		return fmt.Errorf("parameter VerifierType missing")
	}
	if a.VerifierType != "auth" && a.VerifierType != "ldap" && a.VerifierType != "file" && a.VerifierType != "crowd" {
		return fmt.Errorf("parameter VerifierType invalid")
	}
	if a.VerifierType == "ldap" {
		if len(a.LdapHost) == 0 {
			return fmt.Errorf("parameter LdapHost missing")
		}
		if a.LdapPort == 0 {
			return fmt.Errorf("parameter LdapPort missing")
		}
		if len(a.LdapBindDN) == 0 {
			return fmt.Errorf("parameter LdapBindDN missing")
		}
		if len(a.LdapBindPassword) == 0 {
			return fmt.Errorf("parameter LdapBindPassword missing")
		}
		if len(a.LdapBaseDn) == 0 {
			return fmt.Errorf("parameter LdapBaseDn missing")
		}
		if len(a.LdapUserFilter) == 0 {
			return fmt.Errorf("parameter LdapUserFilter missing")
		}
		if len(a.LdapGroupFilter) == 0 {
			return fmt.Errorf("parameter LdapGroupFilter missing")
		}
	}
	if a.VerifierType == "crowd" {
		if len(a.CrowdAppName) == 0 {
			return fmt.Errorf("parameter CrowdAppName missing")
		}
		if len(a.CrowdAppPassword) == 0 {
			return fmt.Errorf("parameter CrowdAppPassword missing")
		}
		if len(a.CrowdURL) == 0 {
			return fmt.Errorf("parameter CrowdURL missing")
		}
	}
	if a.VerifierType == "file" {
		if len(a.UserFile) == 0 {
			return fmt.Errorf("parameter UserFile missing")
		}
	}
	if a.Kind == "html" {
		if len(a.Secret) == 0 {
			return fmt.Errorf("parameter Secret missing")
		}
		if len(a.Secret)%16 != 0 {
			return fmt.Errorf("parameter Secret invalid length")
		}
	}
	if a.Kind == "basic" {
		if len(a.BasicAuthRealm) == 0 {
			return fmt.Errorf("parameter BasicAuthRealm missing")
		}
	}
	return nil
}

func (a *application) run() error {
	glog.V(2).Infof("create http server on %s", a.Port.Address())

	dialer := &net.Dialer{
		Timeout: http_client_builder.DEFAULT_TIMEOUT,
	}
	forwardHandler := forward.New(a.TargetAddress.String(),
		func(address string, req *http.Request) (resp *http.Response, err error) {
			return http_client_builder.New().WithoutProxy().WithoutRedirects().WithDialFunc(
				func(network, address string) (net.Conn, error) {
					return dialer.Dial(network, a.TargetAddress.String())
				}).BuildRoundTripper().RoundTrip(req)
		})

	glog.V(2).Infof("get auth filter for: %v", a.Kind)
	v, err := a.createVerifier()
	if err != nil {
		return err
	}
	var httpFilter http.Handler
	switch a.Kind {
	case "html":
		httpFilter = auth_html.New(forwardHandler.ServeHTTP, func(username string, password string) (bool, error) {
			return v.Verify(auth.UserName(username), auth.Password(password))
		}, auth.NewCrypter(a.Secret.Bytes()))
	case "basic":
		httpFilter = auth_basic.New(forwardHandler.ServeHTTP, func(username string, password string) (bool, error) {
			return v.Verify(auth.UserName(username), auth.Password(password))
		}, a.BasicAuthRealm.String())
	default:
		return errors.Errorf("unknown kind %v", a.Kind)
	}

	router := mux.NewRouter()
	router.Path("/healthz").Handler(a.checkHandler())
	router.Path("/readiness").Handler(a.checkHandler())
	router.NotFoundHandler = httpFilter
	if err != nil {
		return err
	}
	var handler http.Handler = router
	if glog.V(4) {
		glog.Infof("add debug handler")
		handler = debug.New(handler)
	}
	return gracehttp.Serve(&http.Server{Addr: a.Port.Address(), Handler: handler})
}

func (a *application) checkHandler() http.Handler {
	if len(a.TargetHealthzUrl) > 0 {
		return check.New(a.checkHttp)
	}
	return check.New(a.checkTcp)
}

func (a *application) checkHttp() error {
	resp, err := http.Get(a.TargetHealthzUrl.String())
	if err != nil {
		glog.V(1).Infof("check url %v failed: %v", a.TargetHealthzUrl, err)
		return err
	}
	if resp.StatusCode/100 != 2 {
		glog.V(1).Infof("check url %v has wrong status: %v", a.TargetHealthzUrl, resp.Status)
		return fmt.Errorf("check url %v has wrong status: %v", a.TargetHealthzUrl, resp.Status)
	}
	glog.V(4).Infof("http check to %v success", a.TargetHealthzUrl)
	return nil
}

func (a *application) checkTcp() error {
	conn, err := net.Dial("tcp", a.TargetAddress.String())
	if err != nil {
		glog.V(1).Infof("tcp connection to %v failed: %v", a.TargetAddress, err)
		return err
	}
	glog.V(4).Infof("tcp connection to %v success", a.TargetAddress)
	return conn.Close()
}

func (a *application) createVerifier() (auth.Verifier, error) {
	glog.V(2).Infof("get verifier for: %v", a.VerifierType)
	switch a.VerifierType {
	case "ldap":
		return auth.NewCacheAuth(&auth.LdapAuth{
			LdapAuthenticator: auth.NewLdapAuthenticator(
				a.LdapBaseDn,
				a.LdapHost,
				a.LdapServerName,
				a.LdapPort,
				a.LdapUseSSL,
				a.LdapSkipTls,
				a.LdapBindDN,
				a.LdapBindPassword,
				a.LdapUserDn,
				a.LdapUserFilter,
				a.LdapUserField,
				a.LdapGroupDn,
				a.LdapGroupFilter,
				a.LdapGroupField,
			),
			RequiredGroups: a.RequiredGroups,
		}, a.CacheTTL), nil
	case "file":
		return auth.NewCacheAuth(auth.NewFileAuth(a.UserFile), a.CacheTTL), nil
	case "crowd":
		crowdClient, err := crowd.New(a.CrowdAppName.String(), a.CrowdAppPassword.String(), a.CrowdURL.String())
		if err != nil {
			glog.V(2).Infof("create crowd client failed: %v", err)
			return nil, errors.Wrap(err, "create crowd client failed")
		}
		return auth.NewCacheAuth(auth.NewCrowdAuth(crowdClient.Authenticate), a.CacheTTL), nil
	default:
		return nil, errors.Errorf("unknown verifier type: %v", a.VerifierType)
	}
}

type Port int

func (p Port) Address() string {
	return fmt.Sprintf(":%d", p)
}

type TargetAddress string

func (t TargetAddress) String() string {
	return string(t)
}

type TargetHealthzUrl string

func (t TargetHealthzUrl) String() string {
	return string(t)
}

type BasicAuthRealm string

func (b BasicAuthRealm) String() string {
	return string(b)
}

type Secret string

func (s Secret) String() string {
	return string(s)
}

func (s Secret) Bytes() []byte {
	return []byte(s)
}

type Kind string

func (k Kind) String() string {
	return string(k)
}

type VerifierType string

func (v VerifierType) String() string {
	return string(v)
}

type CrowdURL string

func (c CrowdURL) String() string {
	return string(c)
}

type CrowdAppName string

func (c CrowdAppName) String() string {
	return string(c)
}

type CrowdAppPassword string

func (c CrowdAppPassword) String() string {
	return string(c)
}
