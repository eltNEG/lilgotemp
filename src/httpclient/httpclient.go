package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	u "net/url"
)

type proxy struct {
	host     string
	user     string
	password string
}

// HTTPClient ...
type HTTPClient struct {
	Headers map[string]string
	proxy   *proxy
}

// Call ...
func (h *HTTPClient) Call(method, url string, payload []byte) (status int, body []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, nil, err
	}
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	if h.proxy != nil {
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(&u.URL{
					Scheme: "http",
					User:   u.UserPassword(h.proxy.user, h.proxy.password),
					Host:   h.proxy.host,
				}),
			},
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, body, err
}

// Get ...
func (h *HTTPClient) Get(url string, params map[string]string) (status int, body []byte, err error) {
	_url := url
	n := 0
	for k, v := range params {
		joiner := "?"
		if n > 0 {
			joiner = "&"
		}
		_url = fmt.Sprintf("%s%s%s=%s", _url, joiner, k, v)
		n++
	}
	return h.Call("GET", _url, nil)
}

func (h *HTTPClient) Post(_url string, _body interface{}) (status int, body []byte, err error) {
	byteBody, err := json.Marshal(_body)
	if err != nil {
		return status, body, err
	}
	return h.Call("POST", _url, byteBody)
}

func (h *HTTPClient) Delete(_url string, _body interface{}) (status int, body []byte, err error) {
	byteBody, err := json.Marshal(_body)
	if err != nil {
		return status, body, err
	}
	return h.Call("DELETE", _url, byteBody)
}

// SetDefaultHeader ...
func (h *HTTPClient) SetDefaultHeader() {
	h.SetHeader("Content-Type", "application/json")
}

// SetHeader ...
func (h *HTTPClient) SetHeader(key, value string) {
	if h.Headers == nil {
		h.Headers = make(map[string]string)
	}
	h.Headers[key] = value
}

// GetNewHTTPClient ...
func GetNewHTTPClient() *HTTPClient {
	return &HTTPClient{}
}

// SetProxy ...
func (h *HTTPClient) SetProxy(host, user, password string) {
	h.proxy = &proxy{
		host:     host,
		user:     user,
		password: password,
	}
}
