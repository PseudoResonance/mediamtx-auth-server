package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/pseudoresonance/authserver/internal/database"
)

type ConnectHandler struct {
	Database *database.DatabaseManager
}

func (a ConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Since no curl in container by default, just use GET with wget instead...
	parsed, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Printf("Error parsing query string: (%v)\n%v\n", r.URL.RawQuery, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	action := parsed.Get("action")
	id := parsed.Get("id")
	connType := parsed.Get("type")

	if action == "connect" {
		a.Database.Connect(database.Connection{Id: id, Protocol: connType})
	} else {
		a.Database.Disconnect(database.Connection{Id: id, Protocol: connType})
	}

	w.WriteHeader(http.StatusOK)
}
