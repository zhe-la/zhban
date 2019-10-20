package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/corpix/uarand"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/html/charset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const nginxError string = "<html><head><title>410 Gone</title></head><body bgcolor=\"white\"><center><h1>410 Gone</h1></center><hr><center>nginx/1.13.10</center></body></html>"
const nginxErrorCode int = http.StatusGone //410

const errorCode int = http.StatusInternalServerError

// ClientData структура
type ClientData struct {
	settings   Settings
	client     http.Client
	mu         sync.Mutex
	grpcServer *grpc.Server
}

// Settings настройки из agrs
type Settings struct {
	proxyHeaders      bool
	browserHeadersGen bool
	verbose           bool
	toUtf8            bool
	keyParam          string
	bindhost          string
	bindport          int
	keyParamEnable    bool
}

type kostylError struct {
	err      string
	httpCode string
}

// Health struct
type Health struct{}

func rangeIn(low, hi int) string {
	return time.Now().Format("2006-01-02 15:04:05") + " " + strconv.Itoa(low+rand.Intn(hi-low))
}

// Check healthcheck
func (h *Health) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch healthcheck
func (h *Health) Watch(req *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}

func (clientData *ClientData) myUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// TODO
	reply, err := handler(ctx, req)
	return reply, err
}

// GetDataKey Запрос с авторизацией
func (clientData *ClientData) GetDataKey(ctx context.Context, in *DataRequestKey) (out *DataResponse, err error) {
	reqid := rangeIn(10000, 99999)
	if in.GetKey() != clientData.settings.keyParam {
		return nil, errors.New("Wrong key")
	}
	body, err := clientData.getFromRemote(reqid, in.GetUrl(), nil)
	return &DataResponse{Data: string(body)}, nil
}

func (clientData *ClientData) getFromRemote(reqid string, url string, headers http.Header) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if headers == nil {
		if clientData.settings.proxyHeaders {
			req.Header = headers
		}
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
		fmt.Println(reqid, "Headers to send:")
		for k, v := range req.Header {
			fmt.Printf(reqid+" %q => %q\n", k, v[0])
		}
	}

	resp, err := clientData.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if clientData.settings.verbose {
		fmt.Println(reqid, "Response")
		fmt.Printf(reqid+"Status: %s StatusCode: %d Uncompressed: %t Protocol version: %s TransferEncoding %q \n", resp.Status, resp.StatusCode, resp.Uncompressed, resp.Proto, resp.TransferEncoding)
	}

	data := &resp.Body

	if clientData.settings.toUtf8 {
		r, err := charset.NewReader(*data, resp.Header.Get("Content-Type"))
		if err != nil {
			fmt.Println(reqid, "Reader encoding error:", err)
			return nil, err
		}
		rr := ioutil.NopCloser(r)
		data = &rr
	}

	body, err := ioutil.ReadAll(*data)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Process Обрабатывает входящий HTTP-запрос
func (clientData *ClientData) Process(w http.ResponseWriter, r *http.Request) {
	reqid := rangeIn(10000, 99999)
	if r.Method != http.MethodGet {
		return
	}

	if r.URL.String() == "/healthcheck" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if clientData.settings.keyParamEnable {
		key := r.Header.Get("key")
		if key != clientData.settings.keyParam {
			fmt.Println(reqid, "Wrong key!")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(nginxErrorCode)
			fmt.Fprintf(w, nginxError)
			return
		}
	}

	url := r.Header.Get("url")
	if url == "" {
		fmt.Println(reqid, "empty url header")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(nginxErrorCode)
		fmt.Fprintf(w, nginxError)
	}

	if clientData.settings.verbose {
		fmt.Println(reqid, "NEW Request")
		fmt.Println(reqid, "URL:", url)
		fmt.Printf(reqid+" HOST: %s URL: %s RequestURI: %s Protocol version: %s TransferEncoding %q \n", r.Host, r.URL.Path, r.RequestURI, r.Proto, r.TransferEncoding)
	}

	body, err := clientData.getFromRemote(reqid, url, r.Header)
	if err != nil {
		if err.Error() == "ServiceUnavailable" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		fmt.Println(reqid, "Making request error:", err.Error())
		w.WriteHeader(errorCode)
		fmt.Fprintf(w, "Making request error")
		return
	}

	if clientData.settings.verbose {
		fmt.Println(reqid, "send result")
	}

	fmt.Fprint(w, string(body))
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
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
		mu:     sync.Mutex{},
	}
	grpcPort := flag.Int("grpc", 4000, "gRPC listen port (example 4000)")
	consulAddr := flag.String("consul", "", "consul addr host:port (example 127.0.0.1:8500)")
	flag.BoolVar(&clientData.settings.proxyHeaders, "ph", false, "Enable proxing headers to final host")
	flag.BoolVar(&clientData.settings.browserHeadersGen, "bh", true, "Generate browser headers to final host. User-Agent Header is a random UA")
	flag.BoolVar(&clientData.settings.verbose, "v", false, "Verbose output")
	flag.BoolVar(&clientData.settings.toUtf8, "u", false, "Convert response data to utf8 encoding")
	flag.StringVar(&clientData.settings.keyParam, "k", "", "Security key. If set, the request must contain the header \"Key\"")
	flag.StringVar(&clientData.settings.bindhost, "bind", "", "Address for listen (example: 127.0.0.1)")
	flag.IntVar(&clientData.settings.bindport, "http", 3000, "HTTP port for client connections")
	flag.Parse()

	// show options
	if isFlagPassed("grpc") {
		fmt.Println("Proxy headers enable :", false, "-not supported for gRPC")
	} else {
		fmt.Println("Proxy headers enable:", clientData.settings.proxyHeaders)
	}
	fmt.Println("Browser headers enable:", clientData.settings.browserHeadersGen)
	fmt.Println("Verbose output:", clientData.settings.verbose)

	if clientData.settings.keyParam != "" {
		clientData.settings.keyParamEnable = true
	}
	if isFlagPassed("grpc") {
		fmt.Println("Key enable:", true, "-forced for gRPC")
	} else {
		fmt.Println("Key enable:", clientData.settings.keyParamEnable)
	}

	http.HandleFunc("/", clientData.Process)

	// go run zhban.go -grpc 4000 -consul="127.0.0.1:8500"

	if isFlagPassed("consul") {
		// Consul
		config := api.DefaultConfig()
		config.Address = *consulAddr
		consul, err := api.NewClient(config)

		// регистрируем сервис в консуле
		fmt.Println("registered in consul...")
		portCunsulPass := 0
		accessPass := ""

		// Проверяем, какой порт gRPC или HTTP использовать
		if isFlagPassed("grpc") {
			portCunsulPass = *grpcPort
			accessPass = "gRPC"
		} else {
			portCunsulPass = clientData.settings.bindport
			accessPass = "HTTP"
		}
		host := clientData.settings.bindhost
		if host == "" {
			host = "127.0.0.1"
		}
		serviceID := "zhban_" + accessPass + "_" + strconv.Itoa(portCunsulPass)
		serviceSettings := api.AgentServiceRegistration{}
		serviceSettings.ID = serviceID
		serviceSettings.Name = "zhban"
		serviceSettings.Port = portCunsulPass
		serviceSettings.Address = clientData.settings.bindhost
		if !isFlagPassed("grpc") {
			serviceSettings.Check = &api.AgentServiceCheck{
				Interval:                       "30s",
				DeregisterCriticalServiceAfter: "5s",
				HTTP:                           "http://" + fmt.Sprintf("%s:%d", host, clientData.settings.bindport) + "/healthcheck",
			}
		} else {
			serviceSettings.Check = &api.AgentServiceCheck{
				Interval:                       "30s",
				DeregisterCriticalServiceAfter: "5s",
				GRPC:                           fmt.Sprintf("%s:%d", host, *grpcPort),
			}
		}
		err = consul.Agent().ServiceRegister(&serviceSettings)

		if err != nil {
			fmt.Println("cant add service to consul", err)
			return
		}
		fmt.Println("registered in consul", serviceID)

		cleanup := func() {
			err := consul.Agent().ServiceDeregister(serviceID)
			if err != nil {
				fmt.Println("cant add service to consul", err)
				return
			}
			fmt.Println("deregistered in consul", serviceID)
		}
		defer cleanup()

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cleanup()
			os.Exit(1)
		}()
	}

	//gRPC server
	if isFlagPassed("grpc") {
		fmt.Println("gRPC server (service) listen on ", fmt.Sprintf("%s:%d", clientData.settings.bindhost, *grpcPort))
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", clientData.settings.bindhost, *grpcPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		clientData.grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(clientData.myUnaryInterceptor),
		)
		RegisterZhbanServer(clientData.grpcServer, clientData)
		grpc_health_v1.RegisterHealthServer(clientData.grpcServer, &Health{})
		clientData.grpcServer.Serve(lis)
		fmt.Println("Server Stoped")
	} else {
		fmt.Println("HTTP Sever listening on " + fmt.Sprintf("%s:%d", clientData.settings.bindhost, clientData.settings.bindport))
		http.ListenAndServe(fmt.Sprintf("%s:%d", clientData.settings.bindhost, clientData.settings.bindport), nil)
	}
}
