package parse

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type ctxT struct {
	ts            *httptest.Server
	oldHost       string
	oldHttpClient *http.Client
}

var ctx = ctxT{}
var testClient *Client

func setupTestServer(handler http.HandlerFunc) *httptest.Server {
	ts := httptest.NewTLSServer(handler)
	ctx.ts = ts

	_url, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	ctx.oldHost = testClient.host
	ctx.oldHttpClient = testClient.httpClient

	testClient.host = _url.Host
	testClient.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return ts
}

func teardownTestServer() {
	ctx.ts.Close()
	testClient.host = ctx.oldHost
	testClient.httpClient = ctx.oldHttpClient
}

func TestMain(m *testing.M) {
	testClient = NewClient("app_id", "rest_key", "master_key", "https://api.parse.com", "/1")
	os.Exit(m.Run())
}
