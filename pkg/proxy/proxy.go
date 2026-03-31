package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/buffer7/buffer7/pkg/filter"
)

type BufferProxy struct {
	Target *url.URL
	Proxy  *httputil.ReverseProxy
}

func NewProxy(targetURL string) (*BufferProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	bp := &BufferProxy{
		Target: target,
	}

	p := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = bp.Target.Scheme
			req.URL.Host = bp.Target.Host
			req.Host = bp.Target.Host
		},
	}

	// Modify response for metadata
	p.ModifyResponse = func(resp *http.Response) error {
		// 只有 200 OK 且是 JSON 时才过滤
		if resp.StatusCode == 200 && strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			body, _ := io.ReadAll(resp.Body)
			var filtered []byte
			var err error

			if strings.Contains(bp.Target.Host, "pypi.org") {
				filtered, err = filter.FilterPyPI(body)
			} else {
				filtered, err = filter.FilterNPM(body)
			}

			if err == nil {
				resp.Body = io.NopCloser(bytes.NewReader(filtered))
				resp.ContentLength = int64(len(filtered))
				resp.Header.Set("Content-Length", strconv.Itoa(len(filtered)))
			}
		}
		return nil
	}

	bp.Proxy = p
	return bp, nil
}

func (p *BufferProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 302 Redirect for binary files (.tgz, .whl)
	if strings.HasSuffix(r.URL.Path, ".tgz") || strings.HasSuffix(r.URL.Path, ".whl") {
		http.Redirect(w, r, p.Target.String()+r.URL.Path, http.StatusFound)
		return
	}
	p.Proxy.ServeHTTP(w, r)
}
