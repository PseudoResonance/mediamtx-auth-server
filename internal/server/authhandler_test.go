package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func strPtr[T ~string](s T) *T {
	return &s
}

func checkStatus(t *testing.T, test int, target int) {
	if test != target {
		t.Errorf("Wrong status: need (%v) got (%v)\n", target, test)
	}
}

func TestBadMethod(t *testing.T) {
	authHandler := AuthHandler{}
	authHandler.Init()
	req, err := http.NewRequest("GET", "/auth", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	authHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusMethodNotAllowed)
}

func TestBadBody(t *testing.T) {
	authHandler := AuthHandler{}
	authHandler.Init()
	body := authRequestBody{}
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/auth", &buf)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	authHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusBadRequest)
}

func TestPrivateIp(t *testing.T) {
	authHandler := AuthHandler{PrivateIps: []string{"fc00::/7"}}
	authHandler.Init()
	body := authRequestBody{
		User:     strPtr(""),
		Password: strPtr(""),
		Ip:       strPtr("fd32::"),
		Action:   strPtr("read"),
	}
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/auth", &buf)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	authHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusOK)
}

//TODO mock database
// func TestQueryParams(t *testing.T) {
// 	authHandler := AuthHandler{PrivateIps: []string{}, QueryTokenKey: "token"}
// 	authHandler.Init()
// 	body := authRequestBody{
// 		User:     strPtr(""),
// 		Password: strPtr(""),
// 		Ip:       strPtr("127.0.0.1"),
// 		Query:    strPtr("token=TOKENHERE\u0026_HLS_msn=49\u0026_HLS_part=4\u0026_HLS_skip=YES"),
// 		Action:   strPtr("read"),
// 		Path:     strPtr("streamid"),
// 	}
// 	buf := bytes.Buffer{}
// 	err := json.NewEncoder(&buf).Encode(body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req, err := http.NewRequest("POST", "/auth", &buf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rr := httptest.NewRecorder()
// 	authHandler.ServeHTTP(rr, req)
// 	checkStatus(t, rr.Code, http.StatusOK)
// }
