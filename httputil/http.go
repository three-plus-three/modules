package httputil

import (
	"net/http"

	"github.com/runner-mei/resty"
	"github.com/three-plus-three/modules/netutil"
)

var InsecureHttpTransport = resty.InsecureHttpTransport
var InsecureHttpClent = resty.InsecureHttpClent

func init() {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		t.DialContext = netutil.WrapDialContext(t.DialContext)
		InsecureHttpTransport.DialContext = t.DialContext
	}
}
