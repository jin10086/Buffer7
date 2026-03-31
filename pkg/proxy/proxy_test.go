package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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
	p, _ := NewProxy(backend.URL, "npm")
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
	p, _ := NewProxy(backend.URL, "pypi")
	// 强制注入 pypi.org 到 host 以触发分流
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

func TestProxyNPMIntegration_Downgrade(t *testing.T) {
	// 1. Mock NPM Registry
	now := time.Now().UTC()
	oldVerTime := now.AddDate(0, 0, -10).Format(time.RFC3339)
	newVerTime := now.AddDate(0, 0, -1).Format(time.RFC3339)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"name": "test-pkg",
			"dist-tags": {"latest": "1.1.0"},
			"time": {
				"1.0.0": "` + oldVerTime + `",
				"1.1.0": "` + newVerTime + `"
			},
			"versions": {
				"1.0.0": {},
				"1.1.0": {}
			}
		}`))
	}))
	defer backend.Close()

	// 2. Setup Proxy
	p, _ := NewProxy(backend.URL, "npm")
	proxyServer := httptest.NewServer(p)
	defer proxyServer.Close()

	// 3. Request Metadata
	resp, _ := http.Get(proxyServer.URL + "/test-pkg")
	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	// 验证 1.1.0 被过滤
	versions := data["versions"].(map[string]interface{})
	if _, ok := versions["1.1.0"]; ok {
		t.Errorf("Expected 1.1.0 to be filtered out")
	}

	// 验证 latest 标签被降级到 1.0.0
	distTags := data["dist-tags"].(map[string]interface{})
	if distTags["latest"] != "1.0.0" {
		t.Errorf("Expected latest to be downgraded to 1.0.0, got %v", distTags["latest"])
	}
}

func TestProxyPyPIIntegration_Filtering(t *testing.T) {
	// 1. Mock PyPI Registry
	now := time.Now().UTC()
	oldTime := now.AddDate(0, 0, -10).Format(time.RFC3339)
	newTime := now.AddDate(0, 0, -1).Format(time.RFC3339)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// 模拟 PyPI 响应格式
		w.Write([]byte(`{
			"info": {"name": "test-pkg"},
			"releases": {
				"1.0.0": [
					{"upload_time_iso_8601": "` + oldTime + `", "url": "https://example.com/1.0.0.tar.gz"}
				],
				"2.0.0": [
					{"upload_time_iso_8601": "` + newTime + `", "url": "https://example.com/2.0.0.tar.gz"}
				]
			}
		}`))
	}))
	defer backend.Close()

	// 2. Setup Proxy
	p, _ := NewProxy(backend.URL, "pypi")
	// 强制注入 pypi.org 到 host 以触发分流
	p.Target, _ = url.Parse("https://pypi.org")
	// 覆盖 Director 逻辑以便请求发往 Mock 后端
	p.Proxy.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.Replace(backend.URL, "http://", "", 1)
	}
	
	proxyServer := httptest.NewServer(p)
	defer proxyServer.Close()

	// 3. Request
	resp, _ := http.Get(proxyServer.URL + "/pypi/test-pkg/json")
	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	releases := data["releases"].(map[string]interface{})
	
	// 验证 2.0.0 被过滤
	if _, ok := releases["2.0.0"]; ok {
		t.Errorf("Expected 2.0.0 to be filtered out")
	}
	// 验证 1.0.0 被保留
	if _, ok := releases["1.0.0"]; !ok {
		t.Errorf("Expected 1.0.0 to be kept")
	}
}

func TestNewProxy_Error(t *testing.T) {
	_, err := NewProxy(":%:invalid", "npm")
	if err == nil {
		t.Error("Expected error for invalid target URL")
	}
}

func TestProxy_NonJSON(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("not json"))
	}))
	defer backend.Close()

	p, _ := NewProxy(backend.URL, "npm")
	proxyServer := httptest.NewServer(p)
	defer proxyServer.Close()

	resp, _ := http.Get(proxyServer.URL + "/test")
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "not json" {
		t.Errorf("Expected unchanged body for non-JSON, got %s", body)
	}
}

func TestProxy_Non200(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer backend.Close()

	p, _ := NewProxy(backend.URL, "npm")
	proxyServer := httptest.NewServer(p)
	defer proxyServer.Close()

	resp, _ := http.Get(proxyServer.URL + "/test")
	if resp.StatusCode != 404 {
		t.Errorf("Expected 404, got %d", resp.StatusCode)
	}
}
