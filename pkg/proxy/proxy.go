package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
    "github.com/buffer7/buffer7/pkg/filter"
    "io"
    "bytes"
    "strconv"
)

type BufferProxy struct {
	Target *url.URL
	Proxy  *httputil.ReverseProxy
}

func NewProxy(targetURL string) (*BufferProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil { return nil, err }
	
	p := httputil.NewSingleHostReverseProxy(target)
	
	// Modify response for metadata
	p.ModifyResponse = func(resp *http.Response) error {
		// 只有 200 OK 且是 JSON 时才过滤 (NPM)
		if resp.StatusCode == 200 && strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			body, _ := io.ReadAll(resp.Body)
			filtered, _ := filter.FilterNPM(body)
			resp.Body = io.NopCloser(bytes.NewReader(filtered))
			resp.ContentLength = int64(len(filtered))
			resp.Header.Set("Content-Length", strconv.Itoa(len(filtered)))
		}
		return nil
	}

	return &BufferProxy{Target: target, Proxy: p}, nil
}

func (p *BufferProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 302 Redirect for binary files (.tgz, .whl)
    if strings.HasSuffix(r.URL.Path, ".tgz") || strings.HasSuffix(r.URL.Path, ".whl") {
        http.Redirect(w, r, p.Target.String()+r.URL.Path, http.StatusFound)
        return
    }
	r.Host = p.Target.Host
	p.Proxy.ServeHTTP(w, r)
}
