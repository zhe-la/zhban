package zhban

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "time"
  "github.com/corpix/uarand"
  "golang.org/x/net/html/charset"
  "flag"
  "math/rand"
  "strconv"
)

const nginxError string = "<html><head><title>410 Gone</title></head><body bgcolor=\"white\"><center><h1>410 Gone</h1></center><hr><center>nginx/1.13.10</center></body></html>"
const nginxErrorCode int = http.StatusGone //410

const errorCode int = http.StatusInternalServerError

var tr = &http.Transport{
  MaxIdleConns:       10,
  IdleConnTimeout:    30 * time.Second,
  DisableCompression: true,
  TLSHandshakeTimeout: 0 * time.Second,
}
var client = &http.Client{Transport: tr}

type Settings struct {
  proxyHeaders bool
  browserHeaders bool
  verbose bool
  utf8 bool
  keyParam string
  port int
  keyParamEnable bool
}

func rangeIn(low, hi int) string {
    return time.Now().Format("2006-01-02 15:04:05")+" "+strconv.Itoa(low + rand.Intn(hi-low))
}

func (settings *Settings) GetData(w http.ResponseWriter, r *http.Request) {
  Reqid := rangeIn(10000,99999)+" "
  if (r.Method == "POST") {
    if err := r.ParseForm(); err != nil {
      w.WriteHeader(errorCode)
      fmt.Fprintf(w, "ParseForm() err: %v", err)
      return
    }

    if (settings.keyParamEnable) {
      key := r.FormValue("key")
      if (key != settings.keyParam) {
        fmt.Println(Reqid+"Wrong key!")
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
          w.WriteHeader(nginxErrorCode)
          fmt.Fprintf(w, nginxError)
        return
      }         
    }

    url := r.FormValue("url")
    if(url=="") {
      fmt.Println(Reqid+"No url param")
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.WriteHeader(nginxErrorCode)
        fmt.Fprintf(w, nginxError)
        return
    }

    if settings.verbose {
      fmt.Println("\n"+Reqid+"NEW Request")
      fmt.Println(Reqid+"URL:",url)
      fmt.Printf(Reqid+"HOST: %s URL: %s RequestURI: %s Protocol version: %s TransferEncoding %q \n", r.Host, r.URL.Path,r.RequestURI,r.Proto,r.TransferEncoding)
    }

    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        fmt.Println(Reqid+"Request construction error",err.Error())
        w.WriteHeader(errorCode)
          fmt.Fprintf(w, "Request construction error")
        return
    }

    if settings.proxyHeaders {
      req.Header = r.Header
    }

    if settings.browserHeaders {
      req.Header.Set("Host", req.Host)
      req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
      req.Header.Set("Referer", url)
      req.Header.Set("User-Agent", uarand.GetRandom())
      req.Header.Set("Accept-Encoding", "*")
      req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
    }

    if settings.verbose {
      fmt.Println(Reqid+"Headers to send:");
      for k, v := range req.Header {
        fmt.Printf(Reqid+"%q => %q\n",k, v[0])
      }
    }

    resp, err := client.Do(req)
    if err != nil {
      fmt.Println(Reqid+"Making request error:",err.Error())
      w.WriteHeader(errorCode)
      fmt.Fprintf(w, "Making request error")
      return
    }

    defer resp.Body.Close()

    if settings.verbose {
      fmt.Println(Reqid+"Response")
      fmt.Printf(Reqid+"Status: %s StatusCode: %d Uncompressed: %t Protocol version: %s TransferEncoding %q \n", resp.Status, resp.StatusCode,resp.Uncompressed,resp.Proto,resp.TransferEncoding)
    }

    data := &resp.Body

    if settings.utf8 {
      r, err := charset.NewReader(*data, resp.Header.Get("Content-Type"))
      if err != nil {
        fmt.Println(Reqid+"Reader encoding error:", err)
        w.WriteHeader(errorCode)
        fmt.Fprintf(w, "Reader encoding error")
        return
      }
      rr := ioutil.NopCloser(r)
      data = &rr
    }

    body, err := ioutil.ReadAll(*data);
    if err != nil {
      fmt.Println(Reqid+"Couldn't parse response body", err.Error())
      w.WriteHeader(errorCode)
      fmt.Fprintf(w, "Couldn't parse response body")
      return
    }

    if settings.verbose {
      fmt.Println(Reqid+"send result");
    }

    fmt.Fprint(w, string(body))
  } else {
    fmt.Println(Reqid+"NOT POST METHOD")
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(nginxErrorCode)
    fmt.Fprintf(w, nginxError)
    return
  }
}

func main() {
  fmt.Println("")
  fmt.Println("  |-| ")
  fmt.Println("  | | ")
  fmt.Println(" /   \\")
  fmt.Println(" | Z |")
  fmt.Println(" | H |")
  fmt.Println(" | B |")
  fmt.Println(" | A |")
  fmt.Println(" | N |")
  fmt.Println(" |___|")
  fmt.Println("")

  settings := &Settings{}

  flag.BoolVar(&settings.proxyHeaders,"ph", false, "Enable proxing headers to final host")
  flag.BoolVar(&settings.browserHeaders,"bh", true, "Generate browser headers to final host. User-Agent Header is a random UA")
  flag.BoolVar(&settings.verbose,"v", false, "Verbose output")
  flag.BoolVar(&settings.utf8,"utf8", false, "Convert output data to utf8 encoding")
  flag.StringVar(&settings.keyParam,"k", "", "Security key. If not set - insecure connection")
  flag.IntVar(&settings.port,"p", 3000, "Port for waiting connections with POST requests")
  flag.Parse()

  // show options
  fmt.Println("Proxy headers enable:", settings.proxyHeaders);
  fmt.Println("Browser headers enable:", settings.browserHeaders);
  fmt.Println("Verbose output:", settings.verbose);

  settings.keyParamEnable = false
  if (settings.keyParam != "") {
    settings.keyParamEnable = true
  }
  fmt.Println("Key enable:",settings.keyParamEnable);

  fmt.Println("");

  http.HandleFunc("/", settings.GetData)
  fmt.Println("---------------------------------")
  fmt.Println(" Sever listening on 0.0.0.0:"+strconv.Itoa(settings.port))
  fmt.Println("---------------------------------")
  fmt.Println("")
  http.ListenAndServe(":"+strconv.Itoa(settings.port), nil)
}