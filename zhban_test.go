package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ServerTestRequests(w http.ResponseWriter, r *http.Request) {
	ua := r.Header.Get("User-Agent")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, ua)
}

func TestBroweserHeadersData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(ServerTestRequests))

	clientData := &ClientData{
		settings: Settings{keyParamEnable: true, keyParam: "right_key"},
	}

	key := "right_key"

	/*
	   check headers (bh not set)
	*/
	req, err := http.NewRequest("GET", "http://ya.ru/12", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("key", key)
	req.Header.Add("url", ts.URL)

	handler := http.HandlerFunc(clientData.Process)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	status := rr.Code

	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "Go-http-client/1.1"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	/*
	   check headers (bh is set)
	*/

	clientData = &ClientData{
		settings: Settings{keyParamEnable: true, keyParam: "right_key", browserHeadersGen: true},
	}

	req, err = http.NewRequest("GET", "http://ya.ru/123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("key", key)
	req.Header.Add("url", ts.URL)

	handler = http.HandlerFunc(clientData.Process)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	status = rr.Code

	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected = "Go-http-client/1.1"
	if rr.Body.String() == expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	ts.Close()
}

func TestKey(t *testing.T) {

	clientData := &ClientData{
		settings: Settings{keyParamEnable: true, keyParam: "right_key"},
	}

	key := "wrong_key"

	/*
	   check wrong key
	*/
	req, err := http.NewRequest("GET", "http://ya.ru/123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("key", key)
	req.Header.Add("url", "http://ya.ru/123")

	handler := http.HandlerFunc(clientData.Process)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	status := rr.Code

	if status != http.StatusGone {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusGone)
	}

	expected := nginxError
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	/*
	   check right key
	*/

	key = "right_key"

	req, err = http.NewRequest("GET", "http://ya.ru/123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("key", key)
	req.Header.Add("url", "http://ya.ru/123")

	handler = http.HandlerFunc(clientData.Process)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	status = rr.Code

	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGRPCWithAUTH(t *testing.T) {
	want := "yandex.ru"
	clientData := &ClientData{
		settings: Settings{keyParamEnable: true, keyParam: "right_key"},
	}
	req := &DataRequestKey{Key: "right_key", Url: "http://yandex.ru"}
	resp, err := clientData.GetDataKey(context.Background(), req)
	if err != nil {
		t.Errorf("got unexpected error")
	}
	if !strings.Contains(resp.Data, want) {
		t.Errorf("got %v want %v", resp.Data, want)
	}
}

func TestGRPCKey(t *testing.T) {
	clientData := &ClientData{
		settings: Settings{keyParamEnable: true, keyParam: "right_key"},
	}
	req := &DataRequestKey{Key: "wrong_key", Url: "http://yandex.ru"}
	_, err := clientData.GetDataKey(context.Background(), req)
	if err == nil {
		t.Errorf("got unexpected error")
	}
}
