package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/corpix/uarand"
	"golang.org/x/net/html/charset"
)

const nginxError string = "<html><head><title>410 Gone</title></head><body bgcolor=\"white\"><center><h1>410 Gone</h1></center><hr><center>nginx/1.13.10</center></body></html>"
const nginxErrorCode int = http.StatusGone //410

const errorCode int = http.StatusInternalServerError

type ClientData struct {
	settings Settings
	client   http.Client
}

type Settings struct {
	proxyHeaders      bool
	browserHeadersGen bool
	verbose           bool
	toUtf8            bool
	keyParam          string
	port              int
	keyParamEnable    bool
}

func rangeIn(low, hi int) string {
	return time.Now().Format("2006-01-02 15:04:05") + " " + strconv.Itoa(low+rand.Intn(hi-low))
}

func (clientData *ClientData) GetData(w http.ResponseWriter, r *http.Request) {
	Reqid := rangeIn(10000, 99999) + " "
	if clientData.settings.keyParamEnable {
		key := r.Header.Get("key")
		if key != clientData.settings.keyParam {
			fmt.Println(Reqid + "Wrong key!")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(nginxErrorCode)
			fmt.Fprintf(w, nginxError)
			return
		}
	}

	url := r.Header.Get("url")
	if url == "" {
		fmt.Println(Reqid + "empty url header")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(nginxErrorCode)
		fmt.Fprintf(w, nginxError)
	}

	if clientData.settings.verbose {
		fmt.Println("\n" + Reqid + "NEW Request")
		fmt.Println(Reqid+"URL:", url)
		fmt.Printf(Reqid+"HOST: %s URL: %s RequestURI: %s Protocol version: %s TransferEncoding %q \n", r.Host, r.URL.Path, r.RequestURI, r.Proto, r.TransferEncoding)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(Reqid+"Request construction error", err.Error())
		w.WriteHeader(errorCode)
		fmt.Fprintf(w, "Request construction error")
		return
	}

	if clientData.settings.proxyHeaders {
		req.Header = r.Header
	}

	if clientData.settings.browserHeadersGen {
		req.Header.Set("Host", req.Host)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Referer", url)
		req.Header.Set("User-Agent", uarand.GetRandom())
		req.Header.Set("Accept-Encoding", "*")
		req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	}

	if clientData.settings.verbose {
		fmt.Println(Reqid + "Headers to send:")
		for k, v := range req.Header {
			fmt.Printf(Reqid+"%q => %q\n", k, v[0])
		}
	}

	resp, err := clientData.client.Do(req)
	if err != nil {
		fmt.Println(Reqid+"Making request error:", err.Error())
		w.WriteHeader(errorCode)
		fmt.Fprintf(w, "Making request error")
		return
	}

	defer resp.Body.Close()

	if clientData.settings.verbose {
		fmt.Println(Reqid + "Response")
		fmt.Printf(Reqid+"Status: %s StatusCode: %d Uncompressed: %t Protocol version: %s TransferEncoding %q \n", resp.Status, resp.StatusCode, resp.Uncompressed, resp.Proto, resp.TransferEncoding)
	}

	data := &resp.Body

	if clientData.settings.toUtf8 {
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

	body, err := ioutil.ReadAll(*data)
	if err != nil {
		fmt.Println(Reqid+"Couldn't parse response body", err.Error())
		w.WriteHeader(errorCode)
		fmt.Fprintf(w, "Couldn't parse response body")
		return
	}

	if clientData.settings.verbose {
		fmt.Println(Reqid + "send result")
	}

	fmt.Fprint(w, string(body))
}

func main() {

	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		TLSHandshakeTimeout: 0 * time.Second,
	}

	clientData := &ClientData{
		settings: Settings{
			keyParamEnable: false,
		},
		client: http.Client{Transport: transport},
	}

	flag.BoolVar(&clientData.settings.proxyHeaders, "ph", false, "Enable proxing headers to final host")
	flag.BoolVar(&clientData.settings.browserHeadersGen, "bh", true, "Generate browser headers to final host. User-Agent Header is a random UA")
	flag.BoolVar(&clientData.settings.verbose, "v", false, "Verbose output")
	flag.BoolVar(&clientData.settings.toUtf8, "u", false, "Convert response data to utf8 encoding")
	flag.StringVar(&clientData.settings.keyParam, "k", "", "Security key. If set, the request must contain the header \"Key\"")
	flag.IntVar(&clientData.settings.port, "p", 3000, "Port for waiting connections")
	flag.Parse()

	// show options
	fmt.Println("Proxy headers enable:", clientData.settings.proxyHeaders)
	fmt.Println("Browser headers enable:", clientData.settings.browserHeadersGen)
	fmt.Println("Verbose output:", clientData.settings.verbose)

	if clientData.settings.keyParam != "" {
		clientData.settings.keyParamEnable = true
	}
	fmt.Println("Key enable:", clientData.settings.keyParamEnable)

	http.HandleFunc("/", clientData.GetData)

	fmt.Println("Sever listening on 0.0.0.0:" + strconv.Itoa(clientData.settings.port))
	http.ListenAndServe(":"+strconv.Itoa(clientData.settings.port), nil)
}
