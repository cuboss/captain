package net

import (
	"bytes"
	"encoding/json"
	"io"
	"k8s.io/klog"
	"net/http"
	"net/url"
)

// DoHTTPRequest 发起http请求
func DoHTTPRequest(client *http.Client, reqURL, method string, header *http.Header, urlParam map[string]string, data interface{}) (resp *http.Response, err error) {
	if _, err := url.Parse(reqURL); err != nil {
		return nil, err
	}
	var request *http.Request
	var body io.Reader
	// process data
	if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(d)
	}
	if request, err = http.NewRequest(method, reqURL, body); err != nil {
		return nil, err
	}
	// process header
	if header != nil {
		for k, vs := range *header {
			for _, v :=  range vs {
				request.Header.Add(k, v)
			}
		}
	}
	// encoding
	if _, ok := request.Header["Accept-Encoding"]; ok {
		request.Header.Set("Accept-Encoding", "*")
	} else {
		request.Header.Add("Accept-Encoding", "*")
	}
	// url param
	urlValues := request.URL.Query()
	for key, value := range urlParam {
		urlValues.Add(key, value)
	}
	request.URL.RawQuery = urlValues.Encode()
	klog.Infof("request method/url: [%s] %s", method, reqURL)
	klog.Infof("request header: %v", request.Header)
	resp, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	klog.Infof("response: %v", resp)
	return resp, nil
}