package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/buffer7/buffer7/pkg/filter"
)

type BufferProxy struct {
	Target       *url.URL
	Proxy        *httputil.ReverseProxy
	RegistryType string // "npm" or "pypi"
}

func NewProxy(targetURL string, registryType string) (*BufferProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	bp := &BufferProxy{
		Target:       target,
		RegistryType: registryType,
	}

	p := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = bp.Target.Scheme
			req.URL.Host = bp.Target.Host
			req.Host = bp.Target.Host
			// 禁用压缩以简化 Body 过滤
			req.Header.Set("Accept-Encoding", "identity")
		},
	}

	// Modify response for metadata
	p.ModifyResponse = func(resp *http.Response) error {
		contentType := resp.Header.Get("Content-Type")
		// 只要是 JSON、HTML 或者是 PyPI 的自定义格式就参与过滤
		isJSON := strings.Contains(contentType, "json")
		isHTML := strings.Contains(contentType, "html")

		// 只有 200 OK 且满足格式时才过滤
		if resp.StatusCode == 200 && (isJSON || isHTML) {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			
			var filtered []byte
			if bp.RegistryType == "pypi" {
				if strings.HasSuffix(resp.Request.URL.Path, "/json") {
					filtered, err = filter.FilterPyPI(body)
				} else if strings.HasPrefix(resp.Request.URL.Path, "/simple/") {
					// 提取包名 (例如 /simple/requests/ -> requests)
					parts := strings.Split(strings.Trim(resp.Request.URL.Path, "/"), "/")
					if len(parts) >= 2 {
						packageName := parts[1]
						filtered, err = filter.FilterPyPISimple(packageName, body, bp.Target.String())
					} else {
						filtered = body
					}
				} else {
					filtered = body
				}
			} else {
				filtered, err = filter.FilterNPM(body)
			}

			// 如果过滤失败，回退到原始 Body 以确保客户端不会收到空数据
			if err != nil {
				filtered = body
			}

			resp.Body = io.NopCloser(bytes.NewReader(filtered))
			resp.ContentLength = int64(len(filtered))
			resp.Header.Set("Content-Length", strconv.Itoa(len(filtered)))
		}
		return nil
	}

	bp.Proxy = p
	return bp, nil
}

func (p *BufferProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[DEBUG] Request: %s %s\n", r.Method, r.URL.Path)

	// 302 Redirect for binary files (.tgz, .whl)
	if strings.HasSuffix(r.URL.Path, ".tgz") || strings.HasSuffix(r.URL.Path, ".whl") {
		http.Redirect(w, r, p.Target.String()+r.URL.Path, http.StatusFound)
		return
	}
	p.Proxy.ServeHTTP(w, r)
}
