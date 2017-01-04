package main

import (
	"runtime"
	"time"

	"github.com/bborbe/auth_http_proxy/factory"
	"github.com/bborbe/auth_http_proxy/model"
	flag "github.com/bborbe/flagenv"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/golang/glog"
	"go.jona.me/crowd"
)

const (
	defaultPort               int = 8080
	parameterPort                 = "port"
	parameterCacheTTL             = "cache-ttl"
	parameterTargetAddress        = "target-address"
	parameterTargetHealthzUrl     = "target-healthz-url"
	parameterRequiredGroups       = "required-groups"
	parameterConfig               = "config"
	// kind (html,basic)
	parameterKind           = "kind"
	parameterSecret         = "secret"
	parameterBasicAuthRealm = "basic-auth-realm"
	// verifiertype (file,auth,ldap,crowd)
	parameterVerifierType = "verifier"
	// file
	parameterFileUsers = "file-users"
	// auth
	parameterAuthUrl                 = "auth-url"
	parameterAuthApplicationName     = "auth-application-name"
	parameterAuthApplicationPassword = "auth-application-password"
	// ldap
	parameterLdapHost         = "ldap-host"
	parameterLdapPort         = "ldap-port"
	parameterLdapUseSSL       = "ldap-use-ssl"
	parameterLdapSkipTls      = "ldap-skip-tls"
	parameterLdapBindDN       = "ldap-bind-dn"
	parameterLdapBindPassword = "ldap-bind-password"
	parameterLdapBaseDn       = "ldap-base-dn"
	parameterLdapUserDn       = "ldap-user-dn"
	parameterLdapGroupDn      = "ldap-group-dn"
	parameterLdapUserFilter   = "ldap-user-filter"
	parameterLdapGroupFilter  = "ldap-group-filter"
	parameterLdapServerName   = "ldap-servername"
	parameterLdapUserField    = "ldap-user-field"
	parameterLdapGroupField   = "ldap-group-field"
	// crowd
	parameterCrowdURL         = "crowd-url"
	parameterCrowdAppName     = "crowd-app-name"
	parameterCrowdAppPassword = "crowd-app-password"
)

var (
	portPtr             = flag.Int(parameterPort, defaultPort, "port")
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
	authUrlPtr                 = flag.String(parameterAuthUrl, "", "auth url")
	authApplicationNamePtr     = flag.String(parameterAuthApplicationName, "", "auth application name")
	authApplicationPasswordPtr = flag.String(parameterAuthApplicationPassword, "", "auth application password")
	// ldap params
	ldapBaseDnPtr       = flag.String(parameterLdapBaseDn, "", "ldap-base-dn")
	ldapHostPtr         = flag.String(parameterLdapHost, "", "ldap-host")
	ldapServerNamePtr   = flag.String(parameterLdapServerName, "", "ldap-servername")
	ldapPortPtr         = flag.Int(parameterLdapPort, 0, "ldap-port")
	ldapUseSSLPtr       = flag.Bool(parameterLdapUseSSL, false, "ldap-use-ssl")
	ldapSkipTlsPtr      = flag.Bool(parameterLdapSkipTls, false, "ldap-skip-tls")
	ldapBindDNPtr       = flag.String(parameterLdapBindDN, "", "ldap-bind-dn")
	ldapBindPasswordPtr = flag.String(parameterLdapBindPassword, "", "ldap-bind-password")
	ldapUserFilterPtr   = flag.String(parameterLdapUserFilter, "", "ldap-user-filter")
	ldapGroupFilterPtr  = flag.String(parameterLdapGroupFilter, "", "ldap-group-filter")
	ldapUserDnPtr       = flag.String(parameterLdapUserDn, "", "ldap-user-dn")
	ldapGroupDnPtr      = flag.String(parameterLdapGroupDn, "", "ldap-group-dn")
	ldapUserFieldPtr    = flag.String(parameterLdapUserField, "", "ldap-user-field")
	ldapGroupFieldPtr   = flag.String(parameterLdapGroupField, "", "ldap-group-field")
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
	if err := config.Validate(); err != nil {
		return err
	}
	crowdClient, err := crowd.New(config.CrowdAppName.String(), config.CrowdAppPassword.String(), config.CrowdURL.String())
	if err != nil {
		glog.V(2).Infof("create crowd client failed: %v", err)
		return err
	}
	factory := factory.New(*config, crowdClient)
	glog.V(2).Infof("start server")
	return gracehttp.Serve(factory.HttpServer())
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
	if len(config.LdapBaseDn) == 0 {
		config.LdapBaseDn = model.LdapBaseDn(*ldapBaseDnPtr)
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
	if !config.LdapSkipTls {
		config.LdapSkipTls = model.LdapSkipTls(*ldapSkipTlsPtr)
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
	if len(config.LdapUserField) == 0 {
		config.LdapUserField = model.LdapUserField(*ldapUserFieldPtr)
	}
	if len(config.LdapGroupField) == 0 {
		config.LdapGroupField = model.LdapGroupField(*ldapGroupFieldPtr)
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
