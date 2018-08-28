package main

import (
	"context"
	"github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	addr    string
	token   string
	timeout time.Duration
}

func NewClient(addr string, token string) *Client {
	return &Client{addr: addr, token: token, timeout: 10 * time.Second}
}

func (c *Client) request(module string, method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.addr+url, body)
	if err != nil {
		return req, err
	}
	req.Header.Set("x-token", c.token)
	req.Header.Set("x-module", module)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	req.WithContext(ctx)
	return req, nil
}

func (c *Client) Response(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, err
	}
	return res, err
}

func (c *Client) ResponseBytes(req *http.Request) (*http.Response, []byte, error) {
	res, err := c.Response(req)
	if err != nil {
		return res, nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return res, body, err
	}
	return res, body, err
}

func (c *Client) ResponseJson(req *http.Request) (*http.Response, jsoniter.Any, error) {
	res, body, err := c.ResponseBytes(req)
	if err != nil {
		return res, nil, err
	}
	data := c.BytesToJson(body)
	return res, data, err
}

func (c *Client) BytesToJson(body []byte) jsoniter.Any {
	return jsoniter.Get(body)
}

func (c *Client) Get(module string, url string) (*http.Request, error) {
	return c.request(module, "GET", url, nil)
}

func (c *Client) Put(module string, url string, body io.Reader) (*http.Request, error) {
	return c.request(module, "PUT", url, body)
}

func (c *Client) Delete(module string, url string, body io.Reader) (*http.Request, error) {
	return c.request(module, "DELETE", url, body)
}
