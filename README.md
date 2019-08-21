# zhban
lightweight HTTP proxy

[![Build Status](https://travis-ci.com/poloten4ik100/zhban.svg?branch=master)](https://travis-ci.com/poloten4ik100/zhban)

* Support HTTP 1.1
* Optional proxying headers from client to destination without changes
* Automatically transcode content response into utf8 encoding
* Generate a random User-Agent header to the destination server
* Key protection of requests

## How it works

```
   +------+                                 +-----+                     +-----------+
   |client|                                 |zhban|                     |destination|
   +------+                                 +-----+                     +-----------+
                     --Req-->       
             HTTP req. with url header
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

### cmd

```
./zhban -k qwerty123 -p 3002
```

You can use CURL to make test request:

```
curl --header "url: http://ya.ru" --header "key: qwerty123" http://127.0.0.1:3002
```

**cmd args:**
```
  -bh
        Generate browser headers with random User-Agent Header to destination host (default true)
        
  -k string
        Security key. If set, the request must contain the header "Key" (default none)
        
  -p int
        Port for waiting connections (default 3000)
        
  -ph
        Enable proxing headers to destination host (default false)
        
  -u
        Convert response data to utf8 encoding (default false)
        
  -v    
        Verbose output
```

### Docker

```
docker-compose up -d --build
```
