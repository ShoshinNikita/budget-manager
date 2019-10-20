package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

type route struct {
	methods string
	path    string
	handler http.Handler
}

func (s Server) addRoutes(router *mux.Router) {
	routes := []route{
		// TODO
	}

	for _, r := range routes {
		router.Methods(r.methods).Path(r.path).Handler(r.handler)
	}
}
