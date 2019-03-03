# zhban
lightweight http proxy

* Support http 1.1
* Optional proxy Headers
* Automatic optional modification of response content encoding to utf8
* Embedded browser headers with random User-Agent to remote host
* key protected

## How it works

```
   +------+                                 +-----+               +-----------+
   |client|                                 |proxy|               |destination|
   +------+                                 +-----+               +-----------+
                     --Req-->       
             HTTP req. with POST data
           url=http://ya.ru&key=qwqerty123             --Req-->
                                                       <--Res--
                     <--Res--
```

client send HTTP request with POST data:

```
url=http://ya.ru
```

**url** - address of the requested resource

**key** (optional) - The key that will be checked for each request with the original key specified when starting the proxy

```
key=qwerty123
```

if key is wrong, or HTTP request is not POST metod, then returned nginx GONE page

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
./zhban -key qwerty123 -p 3002
```

**cmd args:**
```
  -bh
        Generate browser headers to final host. User-Agent Header is a random UA (default true)
  -k string
        Security key. If not set - insecure connection
  -p int
        port for waiting connects with POST requests (default 3000)
  -ph
        Enable proxing headers to final host
  -utf8
        convert output data to utf8 encoding
  -v    Verbose output
```

### Docker

```
docker-compose up -d --build
```

to run in production, use -key option for secure connections

