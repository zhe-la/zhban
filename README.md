# zhban
lightweight http proxy with browser random header UA

### Build

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
### Usage

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

to run in production, use -key option for secure connections

