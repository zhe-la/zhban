# zhban
lightweight HTTP proxy

[![Build Status](https://travis-ci.com/poloten4ik100/zhban.svg?branch=master)](https://travis-ci.com/poloten4ik100/zhban)

* Support HTTP 1.1
* Support HTTP or gRPC clients
* Option registration / exit / health check in consul
* Optional proxying headers from client to destination without changes for HTTP clients
* Automatically transcode content response into utf8 encoding
* Generate a random User-Agent header to the destination server
* Key protection of requests

## How it works

```
   +------+                                 +-----+                     +-----------+
   |client|                                 |zhban|                     |destination|
   +------+                                 +-----+                     +-----------+
                     --Req-->       
    HTTP req. with url (and, optional, key) header
                        or
             gRPC request with key+url
                                                           --Req-->
                                                       Random UA Header
                                                        
                                                           <--Res--
                     <--Res--
```

Any client send HTTP request to zhban with **url** and **key**(optional) headers, containing **url** of destination target.

**url** - address of destination target

**key** (optional) - a header containing a key that will be verified before proxying the connection
if the key does not fit, page *nginx 410 Gone* will be given to the client

## Build

installing dependencies

```
go get golang.org/x/net/html/charset
go get github.com/corpix/uarand
go get google.golang.org/grpc
go get github.com/hashicorp/consul/api
```

to build binary file, use:

```
go build zhban.go
```

run it!

```
./zhban
```
## Usage

enable zban proxy server for HTTP clients on the port 3002

```
./zhban -k qwerty123 -http 3002
```

You can use CURL to make test HTTP request:

```
curl --header "url: http://ya.ru" --header "key: qwerty123" http://127.0.0.1:3002
```

### Other use cases:

You can also enable the zban proxy server for gRPC clients on the port 4000

```
./zhban -k qwerty123 -grpc 4000
```

Enable the zban proxy server for HTTP clients on the port 3002 and enable automatic registration / exit / health check of the service in the consul

```
./zhban -k qwerty123 -http 3002 -consul 127.0.0.1:8500
```

Enable the zban proxy server for gRPC clients on the port 4000 and enable automatic registration / exit / health check of the service in the consul

```
./zhban -k qwerty123 -grpc 4000 -consul 127.0.0.1:8500
```

**cmd args:**
```
  -bh
        Generate browser headers with random User-Agent Header to destination host (default true)
        
  -k string
        Security key. If set, the request must contain the header "Key" (must be set for gRPC) (default none)
        
  -http int
        HTTP port for waiting connections (default 3000)

  -grpc int
        gRPC listen port (example 4000)

  -bind
        Address for listen clients (HTTP or gRPC) (example: 127.0.0.1)
        
  -ph
        Enable proxing headers to destination host (only for HTTP) (default false)
        
  -u
        Convert response data to utf8 encoding (default false)
        
  -v    
        Verbose output

  -consul
        Enable register in consul at addr host:port (example 127.0.0.1:8500)
```

### Docker

```
docker-compose up -d --build
```
