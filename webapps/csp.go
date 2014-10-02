package webapps

import (
	"net/http"
	"strings"
)

type CspConfig struct {
	Policies []string
}

var cspHeaderNames = []string{
	"Content-Security-Policy",
	"X-Content-Security-Policy",
}

var cspPolicy string

func InitCsp(config CspConfig) {
	cspPolicy = strings.Join(config.Policies, "; ")
}

func SetCspHeader(w http.ResponseWriter) {
	header := w.Header()

	for _, name := range cspHeaderNames {
		header.Set(name, cspPolicy)
	}
}
