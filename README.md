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

`echo 'admin:tester' > sample_users`

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
-file-users=sample_users
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
-config=config.json
```

### With Auth backend 

Start ledis database

`go get github.com/siddontang/ledisdb/cmd/ledis-server`

```
ledis-server \
-databases=1 \
-addr=localhost:5555
```

Start auth-server

`go get github.com/bborbe/auth/bin/auth_server`

```
auth_server \
-logtostderr \
-v=2 \
-port=6666 \
-ledisdb-address=localhost:5555 \
-auth-application-password=test123
```

Register user

`echo -n 'tester:secret' | base64`

```
curl \
-X POST \
-d '{ "authToken":"dGVzdGVyOnNlY3JldA==","user":"tester" }' \
-H "Authorization: Bearer YXV0aDp0ZXN0MTIz" \
http://localhost:6666/api/1.0/user
```

Start auth-http-proxy

```
auth-http-proxy \
-logtostderr \
-v=2 \
-port=8888 \
-kind=basic \
-basic-auth-realm=TestAuth \
-target-address=localhost:7777 \
-verifier=auth \
-auth-url=http://localhost:6666 \
-auth-application-name=auth \
-auth-application-password=test123
```

## Continuous integration

[Jenkins](https://jenkins.benjamin-borbe.de/job/Go-Auth-Http-Proxy/)

## Copyright and license

    Copyright (c) 2016, Benjamin Borbe <bborbe@rocketnews.de>
    All rights reserved.
    
    Redistribution and use in source and binary forms, with or without
    modification, are permitted provided that the following conditions are
    met:
    
       * Redistributions of source code must retain the above copyright
         notice, this list of conditions and the following disclaimer.
       * Redistributions in binary form must reproduce the above
         copyright notice, this list of conditions and the following
         disclaimer in the documentation and/or other materials provided
         with the distribution.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
    "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
    LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
    A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
    OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
    SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
    LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
    DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
    THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
    (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
    OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
