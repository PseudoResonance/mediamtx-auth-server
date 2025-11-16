package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/pseudoresonance/authserver/internal/config"
	"github.com/pseudoresonance/authserver/internal/database"
)

type ForwardAuthHandler struct {
	PrivateIps    []string
	NetPrivateIps []net.IPNet

	QueryTokenKey string
	Config        config.ForwardAuthConfig
	Database      *database.DatabaseManager
}

func (a *ForwardAuthHandler) Init() {
	// Parse CIDR strings to Golang IPNets
	a.NetPrivateIps = make([]net.IPNet, len(a.PrivateIps))
	for i, entry := range a.PrivateIps {
		_, cidr, err := net.ParseCIDR(entry)
		if err != nil {
			log.Fatalf("Invalid CIDR %v\n", entry)
		}
		a.NetPrivateIps[i] = *cidr
	}
}

func (a ForwardAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ipHeader := r.Header.Get(a.Config.IpHeader)
	uriHeader := r.Header.Get(a.Config.UriHeader)

	if len(ipHeader) == 0 || len(uriHeader) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uri, found := strings.CutPrefix(uriHeader, a.Config.BasePath)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ipSplit := strings.Split(ipHeader, ",")
	ip := net.ParseIP(ipSplit[0])

	// Access from private networks is accepted - generally for container networks
	if listContainsIp(a.NetPrivateIps, ip) {
		w.WriteHeader(http.StatusOK)
		return
	}

	// External access
	queryUrl, err := url.Parse(uri)
	if err != nil {
		log.Printf("Error parsing URI: (%v)\n%v\n", uri, err)
	}
	queryParsed, err := url.ParseQuery(queryUrl.RawQuery)
	if err != nil {
		log.Printf("Error parsing URI query string: (%v)\n%v\n", uri, err)
	}
	token := queryParsed.Get(a.QueryTokenKey)

	targetFile := path.Base(queryUrl.Path)
	path := strings.TrimSuffix(filepath.Base(targetFile), filepath.Ext(targetFile))

	res, err := a.Database.ValidateAuth(&database.Credentials{
		Action:     "read",
		Path:       path,
		QueryToken: token,
	}, nil)
	if err != nil {
		log.Printf("Error while validating auth\n%v\n", err)
	}
	if res {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusForbidden)
}
