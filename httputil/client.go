package httputil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/urlutil"
)

type HttpClient struct {
	Url      string
	Warnings interface{}
}

func (self *HttpClient) CreateUrl() *urlutil.URLBuilder {
	return NewURLBuilder(self.Url)
}

func (self *HttpClient) InvokeWithObject(action, url string, body interface{}, exceptedCode int) Result {
	if nil == body {
		return self.InvokeWith(action, url, nil, exceptedCode)
	} else {
		buffer := bytes.NewBuffer(make([]byte, 0, 1000))
		e := json.NewEncoder(buffer).Encode(body)
		if nil != e {
			return ReturnError(http.StatusBadRequest, e.Error())
		}
		return self.InvokeWith(action, url, buffer, exceptedCode)
	}
}

func (self *HttpClient) InvokeWithBytes(action, url string, msg []byte, exceptedCode int) Result {
	if nil == msg {
		return self.InvokeWith(action, url, nil, exceptedCode)
	} else {
		return self.InvokeWith(action, url, bytes.NewBuffer(msg), exceptedCode)
	}
}

func (self *HttpClient) InvokeWith(action, url string, body io.Reader, exceptedCode int) Result {
	self.Warnings = nil
	var result ResultImpl
	err := InvokeHttp(action, url, body, exceptedCode, &result, nil)
	if nil != err {
		result.SetError(err.Code(), err.Error())
	}
	return &result
}

type Client struct {
	HttpClient
}

func NewClient(url, target string) *Client {
	if 0 == len(url) {
		panic("'url' is empty")
	}

	if 0 == len(target) {
		panic("'target' is empty")
	}

	return &Client{HttpClient: HttpClient{Url: NewURLBuilder(url).Concat(target).ToUrl()}}
}

func (self *Client) Create(params map[string]string, body interface{}) Result {
	return self.InvokeWithObject("POST", self.CreateUrl().WithQueries(params, "").ToUrl(), body, 201)
}

func (self *Client) Put(params map[string]string, body interface{}) Result {
	id := params["id"]
	if 0 == len(id) {
		return ReturnWithIsRequired("id")
	}
	delete(params, "id")

	return self.InvokeWithObject("PUT", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), body, 200)
}

func (self *Client) Delete(params map[string]string) Result {
	id := params["id"]

	if 0 == len(id) {
		return self.InvokeWith("DELETE", self.CreateUrl().WithQueries(params, "").ToUrl(), nil, 200)
	}

	delete(params, "id")
	return self.InvokeWith("DELETE", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), nil, 200)
}

func (self *Client) Get(params map[string]string) Result {
	id := params["id"]

	if 0 == len(id) {
		return self.InvokeWith("GET", self.CreateUrl().WithQueries(params, "").ToUrl(), nil, 200)
	}

	delete(params, "id")
	return self.InvokeWith("GET", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), nil, 200)
}

func ReadLinesFromWeb(url string, cachedBuffer *bytes.Buffer, cb func(line []byte) error) errors.RuntimeError {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.NewApplicationError(http.StatusBadRequest, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "Keep-Alive")
	resp, e := InsecureHttpClent.Do(req)
	if nil != e {
		return errors.NewApplicationError(http.StatusServiceUnavailable, e.Error())
	}
	return ReadLinesFromResponse(resp, cachedBuffer, cb)
}

func ReadLinesFromResponse(resp *http.Response, cachedBuffer *bytes.Buffer, cb func(line []byte) error) errors.RuntimeError {
	defer func() {
		if nil != resp.Body {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != 200 {
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		respBody, _ := ioutil.ReadAll(resp.Body)
		if 0 == len(respBody) {
			return errors.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: read_error", resp.StatusCode))
		}
		return errors.NewApplicationError(resp.StatusCode, string(respBody))
	}

	lineReader := bufio.NewReaderSize(resp.Body, 1024)
	for {
		line, _, e := lineReader.ReadLine()
		if nil != e {
			if e == io.EOF {
				return nil
			}
			return errors.ToRuntimeError(e)
		}
		if e = cb(line); nil != e {
			return errors.ToRuntimeError(e)
		}
	}
}
