package zhban

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "strings"
)

func TestGetData(t *testing.T) {
  settings := &Settings{keyParamEnable: true, keyParam:"123"}
  /* 
    check no url, key exist
  */
  req, err := http.NewRequest("POST", "/", strings.NewReader("key=123"))
  req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  if err != nil {
      t.Fatal(err)
  }
  rr := httptest.NewRecorder()
  handler := http.HandlerFunc(settings.GetData)
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
    check wrong key
  */
  req, err = http.NewRequest("POST", "/", strings.NewReader("url=http://ya.ru&key=111"))
  req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  if err != nil {
      t.Fatal(err)
  }
  rr = httptest.NewRecorder()
  handler = http.HandlerFunc(settings.GetData)
  handler.ServeHTTP(rr, req)

  status = rr.Code
  if status != http.StatusGone {
      t.Errorf("handler returned wrong status code: got %v want %v",
          status, http.StatusGone)
  }

  expected = nginxError
  if rr.Body.String() != expected {
      t.Errorf("handler returned unexpected body: got %v want %v",
          rr.Body.String(), expected)
  }

  /* 
    check all
  */
  req, err = http.NewRequest("POST", "/", strings.NewReader("url=http://ya.ru&key=123"))
  req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  if err != nil {
      t.Fatal(err)
  }
  rr = httptest.NewRecorder()
  handler = http.HandlerFunc(settings.GetData)
  handler.ServeHTTP(rr, req)

  status = rr.Code
  if status != http.StatusOK {
      t.Errorf("handler returned wrong status code: got %v want %v",
          status, http.StatusOK)
  }
}