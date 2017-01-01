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

func setupTestServer(handler http.HandlerFunc) *httptest.Server {
	ts := httptest.NewTLSServer(handler)
	ctx.ts = ts

	_url, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	ctx.oldHost = defaultHost
	ctx.oldHttpClient = defaultClient.httpClient

	defaultHost = _url.Host
	defaultClient.httpClient = &http.Client{
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
	defaultHost = ctx.oldHost
	defaultClient.httpClient = ctx.oldHttpClient
}

func TestMain(m *testing.M) {
	Initialize("app_id", "rest_key", "master_key", "api.parse.com", "/1")
	os.Exit(m.Run())
}
