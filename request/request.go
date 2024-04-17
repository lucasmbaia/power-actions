package request

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"

	ContextBasicAuth int = 2
)

type Response struct {
	Header http.Header
	Code   int
	Body   []byte
}

type Options struct {
	Ctx     context.Context
	Body    interface{}
	Headers map[string]string
	Params  url.Values
}

type BasicAuth struct {
	Username string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}

type Client struct {
	client *http.Client
}

type ClientConfiguration struct {
	EnableCookieJar  bool
	CookieJarOptions *cookiejar.Options
	CustomHttpClient *http.Client
}

func NewClient(cfg ClientConfiguration) (c *Client, err error) {
	var jar *cookiejar.Jar

	c = &Client{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
			},
		},
	}

	if cfg.CustomHttpClient != nil {
		c = &Client{
			client: cfg.CustomHttpClient,
		}
	}

	if cfg.EnableCookieJar {
		//the function New never is return error because this return a nil error.
		//So, its impossible test a situation that return error.
		if jar, err = cookiejar.New(cfg.CookieJarOptions); err != nil {
			return
		}

		c.client.Jar = jar
	}

	return
}

func (c *Client) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.client.Jar.SetCookies(u, cookies)
}

func (c *Client) GetCookies(u *url.URL) []*http.Cookie {
	return c.client.Jar.Cookies(u)
}

func (c *Client) Request(method, path string, o Options) (r Response, err error) {
	var (
		req   *http.Request
		resp  *http.Response
		b     []byte
		pb    io.Reader
		query url.Values
		uq    *url.URL
	)

	if o.Body != nil {
		if _, ok := o.Body.(*bytes.Buffer); ok {
			pb = o.Body.(*bytes.Buffer)
		} else {
			var body []byte

			if body, err = json.Marshal(o.Body); err != nil {
				return
			}

			pb = bytes.NewReader(body)
		}
	}

	if uq, err = url.Parse(path); err != nil {
		return
	}

	query = uq.Query()
	for k, v := range o.Params {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	uq.RawQuery = query.Encode()

	if o.Body != nil {
		if req, err = http.NewRequest(method, uq.String(), pb); err != nil {
			return
		}
	} else {
		if req, err = http.NewRequest(method, uq.String(), nil); err != nil {
			return
		}
	}

	if o.Headers != nil {
		for k, v := range o.Headers {
			req.Header.Set(k, v)
		}
	}

	if o.Ctx != nil {
		if auth, ok := o.Ctx.Value(ContextBasicAuth).(BasicAuth); ok {
			req.SetBasicAuth(auth.Username, auth.Password)
		}
	}

	if resp, err = c.client.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if b, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	r = Response{Header: resp.Header, Code: resp.StatusCode, Body: b}
	return
}
