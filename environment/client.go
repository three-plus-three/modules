package environment

import (
	"bytes"
	"net/url"

	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/httputil"
	"github.com/three-plus-three/modules/urlutil"
)

type HttpClient struct {
	cfg          *ServiceConfig
	basePath     string
	params       url.Values
	Body         interface{}
	cached       *bytes.Buffer
	exceptedCode int
}

func (hc *HttpClient) Clone() *HttpClient {
	copied := &HttpClient{}
	*copied = *hc

	if len(hc.params) > 0 {
		copied.params = url.Values{}
		for key, values := range hc.params {
			switch len(values) {
			case 0:
				copied.params[key] = values
			case 1:
				copied.params.Set(key, values[0])
			default:
				newValues := make([]string, len(values))
				copy(newValues, values)
				copied.params[key] = newValues
			}
		}
	}
	copied.cached = nil
	return copied
}

func (hc *HttpClient) ExceptedCode(exceptedCode int) *HttpClient {
	hc.exceptedCode = exceptedCode
	return hc
}

func (hc *HttpClient) Buffer(buffer *bytes.Buffer) *HttpClient {
	hc.cached = buffer
	return hc
}

func (hc *HttpClient) RequestURL() string {
	if len(hc.params) > 0 {
		return hc.basePath + "?" + hc.params.Encode()
	}
	return hc.basePath
}

func (hc *HttpClient) Path(paths ...string) *HttpClient {
	if hc.basePath != "" {
		hc.basePath = urlutil.JoinWith(hc.basePath, paths)
	} else {
		hc.basePath = urlutil.Join(paths...)
	}
	return hc
}

func (hc *HttpClient) SetParam(key, value string) *HttpClient {
	if hc.params == nil {
		hc.params = url.Values{}
	}
	hc.params.Set(key, value)
	return hc
}

func (hc *HttpClient) SetParams(values map[string]string) *HttpClient {
	if hc.params == nil {
		hc.params = url.Values{}
	}

	for key, value := range values {
		hc.params.Set(key, value)
	}
	return hc
}

func (hc *HttpClient) SetBody(body interface{}) *HttpClient {
	hc.Body = body
	return hc
}

func (hc HttpClient) Do(action string, result interface{}) errors.RuntimeError {
	urlStr := hc.cfg.UrlFor(hc.basePath)
	if len(hc.params) != 0 {
		urlStr = urlStr + "?" + hc.params.Encode()
	}
	return httputil.InvokeHttp(action, urlStr, hc.Body, hc.exceptedCode, result, hc.cached)
}

func (hc HttpClient) POST(result interface{}) errors.RuntimeError {
	return hc.Do("POST", result)
}

func (hc HttpClient) Create(result interface{}) errors.RuntimeError {
	return hc.Do("POST", result)
}

func (hc HttpClient) GET(result interface{}) errors.RuntimeError {
	return hc.Do("GET", result)
}

func (hc HttpClient) PUT(result interface{}) errors.RuntimeError {
	return hc.Do("PUT", result)
}

func (hc HttpClient) DELETE(result interface{}) errors.RuntimeError {
	return hc.Do("DELETE", result)
}
