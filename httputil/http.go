package httputil

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/three-plus-three/modules/errors"
)

type HandleFunc func(req *http.Request, resp *http.Response) error

var InsecureHttpTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}

var InsecureHttpClent = &http.Client{Transport: InsecureHttpTransport}

func InvokeHttp(action, url string, body interface{}, exceptedCode int, result interface{}, cachedBuffer *bytes.Buffer) errors.RuntimeError {
	return InvokeHttpWithContext(nil, action, url, body, exceptedCode, result, cachedBuffer)
}

func InvokeHttpWithContext(ctx context.Context, action, url string, body interface{}, exceptedCode int, result interface{}, cachedBuffer *bytes.Buffer) errors.RuntimeError {
	var req *http.Request
	var e error

	if body == nil {
		req, e = http.NewRequest(action, url, nil)
	} else {
		switch value := body.(type) {
		case []byte:
			req, e = http.NewRequest(action, url, bytes.NewReader(value))
		case string:
			req, e = http.NewRequest(action, url, strings.NewReader(value))
		case io.Reader:
			req, e = http.NewRequest(action, url, value)
		default:
			if cachedBuffer == nil {
				cachedBuffer = bytes.NewBuffer(make([]byte, 0, 1000))
			} else {
				cachedBuffer.Reset()
			}
			e = json.NewEncoder(cachedBuffer).Encode(body)
			if nil != e {
				return errors.NewApplicationError(http.StatusBadRequest, e.Error())
			}
			req, e = http.NewRequest(action, url, cachedBuffer)
		}
	}
	if e != nil {
		return errors.NewApplicationError(http.StatusBadRequest, e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "Keep-Alive")

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return errors.NewApplicationError(http.StatusServiceUnavailable, e.Error())
	}

	// Install closing the request body (if any)
	defer func() {
		if nil != resp.Body {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != exceptedCode {
		if exceptedCode != 0 || resp.StatusCode < http.StatusOK || resp.StatusCode > 299 {
			var respBody []byte
			respBody, e = ioutil.ReadAll(resp.Body)
			if nil != e {
				panic(e.Error())
			}
			if 0 == len(respBody) {
				return errors.NewApplicationError(resp.StatusCode, fmt.Sprintf("request '%v' fail: %v: read_error", url, resp.StatusCode))
			}
			return errors.NewApplicationError(resp.StatusCode, "request '"+url+"' fail: "+resp.Status+": "+string(respBody))
		}
	}

	if nil == result {
		return nil
	}

	if cb, ok := result.(HandleFunc); ok {
		e = cb(req, resp)
		if e != nil {
			return errors.ToRuntimeError(e)
		}
		return nil
	}

	if w, ok := result.(io.Writer); ok {
		_, e = io.Copy(w, resp.Body)
		if e != nil {
			return errors.ToRuntimeError(e)
		}
		return nil
	}

	if nil == cachedBuffer {
		if s, ok := result.(*string); ok {
			cachedBuffer = bytes.NewBuffer(make([]byte, 0, 1024))
			if _, e = io.Copy(cachedBuffer, resp.Body); nil != e {
				return errors.NewApplicationError(http.StatusInternalServerError,
					fmt.Sprintf("%v: %s", http.StatusInternalServerError, e.Error()))
			}
			*s = cachedBuffer.String()
			return nil
		}

		if bs, ok := result.(*[]byte); ok {
			cachedBuffer = bytes.NewBuffer(make([]byte, 0, 1024))
			if _, e = io.Copy(cachedBuffer, resp.Body); nil != e {
				return errors.NewApplicationError(http.StatusInternalServerError,
					fmt.Sprintf("%v: %s", http.StatusInternalServerError, e.Error()))
			}

			*bs = cachedBuffer.Bytes()
			return nil
		}

		//bs, _ := ioutil.ReadAll(resp.Body)
		//cachedBuffer = bytes.NewBuffer(make([]byte, 0, 1024))
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		e = decoder.Decode(result)
		if nil != e {
			return errors.NewApplicationError(http.StatusInternalServerError, e.Error())
		}
		return nil
	}

	cachedBuffer.Reset()

	if _, e = io.Copy(cachedBuffer, resp.Body); nil != e {
		return errors.NewApplicationError(http.StatusInternalServerError,
			fmt.Sprintf("%v: %s", http.StatusInternalServerError, e.Error()))
	}
	if s, ok := result.(*string); ok {
		*s = cachedBuffer.String()
		return nil
	}

	if bs, ok := result.(*[]byte); ok {
		*bs = cachedBuffer.Bytes()
		return nil
	}

	if 0 == cachedBuffer.Len() {
		return errors.NewApplicationError(resp.StatusCode,
			fmt.Sprintf("%v: read empty error", resp.StatusCode))
	}

	if e = json.Unmarshal(cachedBuffer.Bytes(), result); nil != e {
		return errors.NewApplicationError(http.StatusInternalServerError,
			fmt.Sprintf("umarshal '%s' to %T failed, %s",
				cachedBuffer.String(), result, e.Error()))
	}
	return nil
}

func Post(url string, body, result interface{}, cachedBuffer *bytes.Buffer) errors.RuntimeError {
	return InvokeHttp("POST", url, body, http.StatusOK, result, cachedBuffer)
}

func Get(url string, result interface{}, cachedBuffer *bytes.Buffer) errors.RuntimeError {
	return InvokeHttp("GET", url, nil, http.StatusOK, result, cachedBuffer)
}

func Put(url string, body, result interface{}, cachedBuffer *bytes.Buffer) errors.RuntimeError {
	return InvokeHttp("PUT", url, body, http.StatusOK, result, cachedBuffer)
}

func Delete(url string, body, result interface{}, cachedBuffer *bytes.Buffer) errors.RuntimeError {
	return InvokeHttp("DELETE", url, body, http.StatusOK, result, cachedBuffer)
}

func Do(method, url string, body, statusCode int, result interface{}, cachedBuffer *bytes.Buffer) errors.RuntimeError {
	return InvokeHttp(method, url, body, statusCode, result, cachedBuffer)
}
