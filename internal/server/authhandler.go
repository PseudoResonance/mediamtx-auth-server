package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/pseudoresonance/authserver/internal/database"
)

type AuthHandler struct {
	PrivateIps    []string
	NetPrivateIps []net.IPNet
	ApiIps        []string
	NetApiIps     []net.IPNet

	QueryTokenKey string
	Database      *database.DatabaseManager
}

func (a *AuthHandler) Init() {
	// Parse CIDR strings to Golang IPNets
	a.NetPrivateIps = make([]net.IPNet, len(a.PrivateIps))
	for i, entry := range a.PrivateIps {
		_, cidr, err := net.ParseCIDR(entry)
		if err != nil {
			log.Fatalf("Invalid CIDR %v\n", entry)
		}
		a.NetPrivateIps[i] = *cidr
	}

	a.NetApiIps = make([]net.IPNet, len(a.ApiIps))
	for i, entry := range a.ApiIps {
		_, cidr, err := net.ParseCIDR(entry)
		if err != nil {
			log.Fatalf("Invalid CIDR %v\n", entry)
		}
		a.NetApiIps[i] = *cidr
	}
}

type authRequestBody struct {
	User     *string `json:"user,omitempty"`     // Ignored
	Password *string `json:"password,omitempty"` // Ignored
	Token    *string `json:"token,omitempty"`    // Ignored
	Ip       *string `json:"ip,omitempty"`
	Action   *string `json:"action,omitempty"`
	Path     *string `json:"path,omitempty"`
	Protocol *string `json:"protocol,omitempty"` // Different from connect/disconnect protocol
	Id       *string `json:"id,omitempty"`       // To match up with connect/disconnect protocol
	Query    *string `json:"query,omitempty"`
}

func (a AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Decode
	request := authRequestBody{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request.Ip == nil || request.Action == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip := net.ParseIP(*request.Ip)

	// Special logic for API access
	if *request.Action == "api" || *request.Action == "metrics" || *request.Action == "pprof" {
		if listContainsIp(a.NetApiIps, ip) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		return
	}

	// Other access from private networks is accepted - generally for container networks
	if listContainsIp(a.NetPrivateIps, ip) {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Other access
	if request.Query == nil || len(*request.Query) == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if request.Path == nil || len(*request.Path) == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	queryUrl, err := url.Parse(fmt.Sprintf("?%v", *request.Query))
	if err != nil {
		log.Printf("Error parsing query string: (%v)\n%v\n", *request.Query, err)
	}
	queryParsed, err := url.ParseQuery(queryUrl.RawQuery)
	if err != nil {
		log.Printf("Error parsing query string: (%v)\n%v\n", *request.Query, err)
	}
	token := queryParsed.Get(a.QueryTokenKey)

	var conn *database.Connection
	if request.Protocol != nil && request.Id != nil {
		conn = &database.Connection{Id: *request.Id, Protocol: *request.Protocol}
	}
	res, err := a.Database.ValidateAuth(&database.Credentials{
		Action:     *request.Action,
		Path:       *request.Path,
		QueryToken: token,
	}, conn)
	if err != nil {
		log.Printf("Error while validating auth\n%v\n", err)
	}
	if res {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusForbidden)
}

func listContainsIp(list []net.IPNet, ip net.IP) bool {
	for _, r := range list {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}
