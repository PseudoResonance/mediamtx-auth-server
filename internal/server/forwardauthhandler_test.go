package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pseudoresonance/authserver/internal/config"
)

func TestFABadMethod(t *testing.T) {
	forwardAuthHandler := ForwardAuthHandler{}
	forwardAuthHandler.Init()
	req, err := http.NewRequest("POST", "/forward", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	forwardAuthHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusMethodNotAllowed)
}

func TestFANoHeaders(t *testing.T) {
	forwardAuthHandler := ForwardAuthHandler{
		QueryTokenKey: "token",
		Config: config.ForwardAuthConfig{
			UriHeader: "X-Forwarded-Uri",
			IpHeader:  "X-Forwarded-For",
			BasePath:  "/thumbnails",
		},
	}
	forwardAuthHandler.Init()
	req, err := http.NewRequest("GET", "/forward", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	forwardAuthHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusBadRequest)
}

func TestFALocal(t *testing.T) {
	forwardAuthHandler := ForwardAuthHandler{
		QueryTokenKey: "token",
		Config: config.ForwardAuthConfig{
			UriHeader: "X-Forwarded-Uri",
			IpHeader:  "X-Forwarded-For",
			BasePath:  "/thumbnails",
		},
		PrivateIps: []string{"127.0.0.1/8"},
	}
	forwardAuthHandler.Init()
	req, err := http.NewRequest("GET", "/forward", nil)
	req.Header.Add(forwardAuthHandler.Config.UriHeader, "/thumbnails/test.png?token=abc")
	req.Header.Add(forwardAuthHandler.Config.IpHeader, "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	forwardAuthHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusOK)
}

func TestFALocalMulti(t *testing.T) {
	forwardAuthHandler := ForwardAuthHandler{
		QueryTokenKey: "token",
		Config: config.ForwardAuthConfig{
			UriHeader: "X-Forwarded-Uri",
			IpHeader:  "X-Forwarded-For",
			BasePath:  "/thumbnails",
		},
		PrivateIps: []string{"127.0.0.1/8"},
	}
	forwardAuthHandler.Init()
	req, err := http.NewRequest("GET", "/forward", nil)
	req.Header.Add(forwardAuthHandler.Config.UriHeader, "/thumbnails/test.png?token=abc")
	req.Header.Add(forwardAuthHandler.Config.IpHeader, "127.0.0.1, 10.0.0.5")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	forwardAuthHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusOK)
}

func TestFARemote(t *testing.T) {
	forwardAuthHandler := ForwardAuthHandler{
		QueryTokenKey: "token",
		Config: config.ForwardAuthConfig{
			UriHeader: "X-Forwarded-Uri",
			IpHeader:  "X-Forwarded-For",
			BasePath:  "/thumbnails",
		},
		PrivateIps: []string{"127.0.0.1/8"},
	}
	forwardAuthHandler.Init()
	req, err := http.NewRequest("GET", "/forward", nil)
	req.Header.Add(forwardAuthHandler.Config.UriHeader, "/thumbnails/test.png?token=abc")
	req.Header.Add(forwardAuthHandler.Config.IpHeader, "10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	forwardAuthHandler.ServeHTTP(rr, req)
	checkStatus(t, rr.Code, http.StatusOK)
}
