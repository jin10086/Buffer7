package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProxyIntegration(t *testing.T) {
	// 1. Mock Backend Registry
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只有非 .tgz 的请求才验证 Header
		if !strings.HasSuffix(r.URL.Path, ".tgz") {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("Authorization header not forwarded on path %s", r.URL.Path)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"name":"test"}`))
	}))
	defer backend.Close()

	// 2. Setup Proxy
	p, _ := NewProxy(backend.URL)
	proxyServer := httptest.NewServer(p)
	defer proxyServer.Close()

	// 3. Test Metadata Request
	req, _ := http.NewRequest("GET", proxyServer.URL+"/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	// 4. Test 302 Redirect
	req, _ = http.NewRequest("GET", proxyServer.URL+"/test.tgz", nil)
	resp, _ = http.DefaultClient.Do(req)
	// We check for 200 here because DefaultClient follows redirects
	if resp.StatusCode != 200 || !strings.Contains(resp.Request.URL.String(), backend.URL) {
		t.Errorf("Expected redirect to backend, got %s", resp.Request.URL.String())
	}
}

func TestProxyPyPIIntegration(t *testing.T) {
	// 1. Mock PyPI Backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// 返回一个带 releases 的 PyPI 格式 JSON
		w.Write([]byte(`{"releases": {"1.1.0": [{"upload_time_iso_8601": "2026-03-30T00:00:00Z"}]}}`))
	}))
	defer backend.Close()

	// 2. Setup Proxy targeting PyPI
	p, _ := NewProxy(backend.URL)
	// 强制注入 pypi.org 到 host 以触发分流 (这里 mock 一个 targetURL 包含 pypi.org)
	p.Target, _ = url.Parse("https://pypi.org")
	p.Proxy.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.Replace(backend.URL, "http://", "", 1)
	}

	proxyServer := httptest.NewServer(p)
	defer proxyServer.Close()

	// 3. Request
	resp, _ := http.Get(proxyServer.URL + "/requests/json")
	body, _ := io.ReadAll(resp.Body)
	
	if strings.Contains(string(body), "1.1.0") {
		t.Errorf("Expected 1.1.0 to be filtered out by PyPI filter")
	}
}
