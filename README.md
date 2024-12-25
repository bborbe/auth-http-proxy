# Auth-Http-Proxy

Is a http-proxy to add authentication to applications without own. You can choose between basic-auth and html-form based. Posible user storages are ldap, file, crowd or [auth](https://github.com/bborbe/auth).  

## Install va sources 

```
go get github.com/bborbe/auth-http-proxy
```

## Usage

Start sample you want protect

```
go get github.com/bborbe/server/bin/file_server
```

```
file_server \
-logtostderr \
-v=2 \
-port=7777 \
-root=/tmp
```

### With file backend

_only for testing_

`echo 'admin:tester' > sample/sample_users`

Start auth-http-proxy

```
auth-http-proxy \
-logtostderr \
-v=2 \
-port=8888 \
-kind=basic \
-basic-auth-realm=TestAuth \
-target-address=localhost:7777 \
-verifier=file \
-file-users=sample/sample_users
```

### With crowd backend

Start auth-http-proxy

```
auth-http-proxy \
-logtostderr \
-v=2 \
-port=8888 \
-kind=basic \
-basic-auth-realm=TestAuth \
-target-address=localhost:7777 \
-verifier=crowd \
-crowd-url="https://crowd.example.com/" \
-crowd-app-name="user" \
-crowd-app-password="pass" 
```

### With ldap backend

Start auth-http-proxy

```
auth-http-proxy \
-logtostderr \
-v=2 \
-port=8888 \
-kind=basic \
-basic-auth-realm=TestAuth \
-target-address=localhost:7777 \
-verifier=ldap \
-ldap-host="ldap.example.com" \
-ldap-port=389 \
-ldap-use-ssl=false \
-ldap-skip-tls=true \
-ldap-bind-dn="cn=root,dc=example,dc=com" \
-ldap-bind-password="S3CR3T" \
-ldap-base-dn="dc=example,dc=com" \
-ldap-user-db="ou=users" \
-ldap-group-db="ou=groups" \
-ldap-user-filter="(uid=%s)" \
-ldap-group-filter="(member=uid=%s,ou=users,dc=example,dc=com)" \
-ldap-user-field="uid" \
-ldap-group-field="ou" \
-required-groups="admin"
```

Start auth-http-proxy with config

`vi config.json`

```
{
  "port": 8888,
  "target-address": "localhost:7777",
  "kind": "html",
  "secret": "AES256Key-32Characters1234567890",
  "verifier": "ldap",
  "required-groups": ["Admins"],
  "ldap-host": "ldap.example.com",
  "ldap-port": 389,
  "ldap-use-ssl": false,
  "ldap-user-dn": "ou=users",
  "ldap-group-dn": "ou=groups",
  "ldap-bind-dn": "cn=root,dc=example,dc=com",
  "ldap-bind-password": "S3CR3T",
  "ldap-base-dn": "dc=example,dc=com",
  "ldap-user-dn": "ou=users",
  "ldap-group-dn": "ou=groups",
  "ldap-user-filter": "(uid=%s)",
  "ldap-group-filter": "(member=uid=%s,ou=users,dc=example,dc=com)",
  "ldap-user-field": "uid",
  "ldap-group-field": "ou",
  "required-groups": "admin"
}
```

```
auth-http-proxy \
-logtostderr \
-v=2 \
-config=sample/config_ldap.json
```
