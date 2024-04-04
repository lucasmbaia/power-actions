package request

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_NewClient(t *testing.T) {
	var (
		err error
	)

	var tests = []struct {
		name string
		cc   ClientConfiguration
	}{
		{
			"default configuration",
			ClientConfiguration{},
		},
		{
			"enable cookiejar",
			ClientConfiguration{
				EnableCookieJar: true,
			},
		},
		{
			"enable custom http client",
			ClientConfiguration{
				CustomHttpClient: &http.Client{},
			},
		},
		{
			"enable custom http client and cookiejar",
			ClientConfiguration{
				EnableCookieJar:  true,
				CustomHttpClient: &http.Client{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err = NewClient(tt.cc); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_Request(t *testing.T) {
	var (
		err        error
		c          *Client
		httpTest   *httptest.Server
		emptyError string
		params     = url.Values{}
	)

	httpTest = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Basic Mock Test!!!")
	}))
	defer httpTest.Close()

	params.Set("name", "mock")

	if c, err = NewClient(ClientConfiguration{}); err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name          string
		method        string
		path          string
		options       Options
		errorExpected string
	}{
		{
			"connection refused",
			GET,
			"http://127.0.0.1",
			Options{},
			`Get "http://127.0.0.1": dial tcp 127.0.0.1:80: connect: connection refused`,
		},
		{
			"basic test",
			GET,
			httpTest.URL,
			Options{},
			emptyError,
		},
		{
			"error parse url",
			GET,
			"http://127.0.0.1:errorParse",
			Options{},
			`parse "http://127.0.0.1:errorParse": invalid port ":errorParse" after host`,
		},
		{
			"post with body",
			POST,
			httpTest.URL,
			Options{
				Body: struct {
					Name string
				}{"Test"},
			},
			emptyError,
		},
		{
			"post with invalid method",
			"INVALID_METHOD",
			httpTest.URL,
			Options{
				Body: struct {
					Name string
				}{"Test"},
			},
			emptyError,
		},
		{
			"post with body error",
			POST,
			httpTest.URL,
			Options{
				Body: make(chan int),
			},
			"json: unsupported type: chan int",
		},
		{
			"get with params",
			GET,
			httpTest.URL,
			Options{
				Params: params,
			},
			emptyError,
		},
		{
			"get with headers",
			GET,
			httpTest.URL,
			Options{
				Headers: map[string]string{
					"header": "header",
				},
			},
			emptyError,
		},
		{
			"get with basic authentication",
			GET,
			httpTest.URL,
			Options{
				Ctx: context.WithValue(context.Background(), ContextBasicAuth, BasicAuth{
					Username: "username",
					Password: "password",
				}),
			},
			emptyError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err = c.Request(tt.method, tt.path, tt.options); err != nil {
				if strings.Compare(tt.errorExpected, err.Error()) != 0 {
					t.Fatal(err)
				}
			}
		})
	}
}
