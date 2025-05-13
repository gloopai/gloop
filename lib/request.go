package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type request struct{}

var Request request

// http post json
func (r *request) HttpPostJson(url string, data interface{}, header map[string]interface{}) ([]byte, error) {
	// b, _ := json.Marshal(data)
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	Log.Info(fmt.Sprintf("HTTP POST JSON: %s data:%v", url, b))
	request, _ := http.NewRequest("POST", url, b)
	request.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		request.Header.Set(k, fmt.Sprintf("%v", v))
	}
	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, nil
}

// http post json resultHeader
func (r *request) HttpPostJsonResultHeader(url string, data interface{}, header map[string]interface{}) ([]byte, http.Header, error) {
	// b, _ := json.Marshal(data)
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	if err != nil {
		return nil, nil, err
	}

	Log.Info(fmt.Sprintf("HTTP POST JSON: %s data:%v", url, b))
	request, _ := http.NewRequest("POST", url, b)
	request.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		request.Header.Set(k, fmt.Sprintf("%v", v))
	}
	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return nil, resp.Header, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.Header, nil
}

// http get
func (r *request) HttpGet(url string, data map[string]interface{}, header map[string]interface{}) ([]byte, error) {
	_queryString := ""
	for k, v := range data {
		if v != "" {
			_queryString = fmt.Sprintf("%v&%v=%v", _queryString, k, v)
		}

	}
	Log.Info(fmt.Sprintf("HTTP GET: %s  %s", url, _queryString))
	getUrl := ""
	if strings.Contains(url, "?") {
		getUrl = fmt.Sprintf("%v&%v", url, _queryString)
	} else {
		getUrl = fmt.Sprintf("%v?%v", url, _queryString)
	}
	request, _ := http.NewRequest("GET", getUrl, strings.NewReader(""))
	for k, v := range header {
		if v != "" {
			request.Header.Set(k, fmt.Sprintf("%v", v))
		}

	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return body, nil
}
