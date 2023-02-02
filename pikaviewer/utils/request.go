package utils

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

type Request struct {
	RawUrl string
}

func NewRequest(url string) *Request {
	return &Request{
		RawUrl: url,
	}
}

func (r *Request) Get(params map[string]string) ([]byte, error) {
	_url, err := url.Parse(r.RawUrl)
	if err != nil {
		return nil, err
	}
	p := url.Values{}
	for k, v := range params {
		p.Set(k, v)
	}
	_url.RawQuery = p.Encode()
	urlPath := _url.String()
	resp, err := http.Get(urlPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *Request) Post(data map[string]string) ([]byte, error) {
	values := url.Values{}
	for k, v := range data {
		values.Add(k, v)
	}
	resp, err := http.PostForm(r.RawUrl, values)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}
