# Auth-Http-Proxy

## Install

`go get github.com/bborbe/auth_http_proxy/bin/auth_http_proxy_server`

`go get github.com/bborbe/auth/bin/auth_server`

`go get go get github.com/siddontang/ledisdb/cmd/ledis-server`

`go get github.com/bborbe/server/bin/file_server`

## Usage

Start sample you want protect

```
file_server \
-logtostderr \
-v=2 \
-port=7777 \
-root=/tmp
```

### With Auth backend 

Start ledis database

```
ledis-server \
-databases=1 \
-addr=localhost:5555
```

Start auth-server

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
http://localhost:6666/user
```

Start auth_http_proxy_server

```
auth_http_proxy_server \
-logtostderr \
-v=2 \
-port=8888 \
-basic-auth-realm=TestAuth \
-target-address=localhost:7777 \
-verifier=auth \
-auth-url=http://localhost:6666 \
-auth-application-name=auth \
-auth-application-password=test123
```

### With file backend

`echo 'admin:tester' > sample_users`

Start auth_http_proxy_server

```
auth_http_proxy_server \
-logtostderr \
-v=2 \
-port=8888 \
-basic-auth-realm=TestAuth \
-target-address=localhost:7777 \
-verifier=file \
-file-users=sample_users
```

### With ldap backend (TBD)

Start auth_http_proxy_server

```
auth_http_proxy_server \
-logtostderr \
-v=2 \
-port=8888 \
-basic-auth-realm=TestAuth \
-target-address=localhost:7777 \
-verifier=ldap \
-ldap-foo=bar
```

## Continuous integration

[Jenkins](https://www.benjamin-borbe.de/jenkins/job/Go-Auth-Http-Proxy/)

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
