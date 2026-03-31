package proxy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNPME2E(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not found, skipping E2E test")
	}

	now := time.Now().UTC()
	oldTime := now.AddDate(0, 0, -10).Format(time.RFC3339)
	newTime := now.AddDate(0, 0, -1).Format(time.RFC3339)

	// 1. Mock Registry
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "test-pkg") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"name": "test-pkg",
				"dist-tags": {"latest": "2.0.0"},
				"time": {
					"1.0.0": "` + oldTime + `",
					"2.0.0": "` + newTime + `"
				},
				"versions": {
					"1.0.0": {
						"name": "test-pkg",
						"version": "1.0.0",
						"dist": {"tarball": "http://example.com/test-pkg-1.0.0.tgz"}
					},
					"2.0.0": {
						"name": "test-pkg",
						"version": "2.0.0",
						"dist": {"tarball": "http://example.com/test-pkg-2.0.0.tgz"}
					}
				}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer backend.Close()

	// 2. Setup Temp Workspace
	tmpDir, err := os.MkdirTemp("", "buffer7-npm-e2e")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 3. Run buffer7 + npm install
	_, filename, _, _ := runtime.Caller(0)
	rootPath := filepath.Join(filepath.Dir(filename), "../../")

	// 使用 go run main.go 启动代理并让其代理 npm 命令
	// 我们添加 --dry-run 虽然 npm install 没有标准的 --dry-run，但我们可以使用 --package-lock-only 或类似的。
	// 但 plan 里是用普通的 npm install。
	cmd := exec.Command("go", "run", "main.go", "npm", "install", "test-pkg", "--prefix", tmpDir, "--no-bin-links")
	cmd.Dir = rootPath
	cmd.Env = append(os.Environ(), 
		"BUFFER7_UPSTREAM_REGISTRY="+backend.URL,
		"CGO_ENABLED=0",
	)
	
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	// fmt.Printf("NPM Output: %s\n", outputStr)

	// 4. Verify output contains 1.0.0 and NOT 2.0.0
	// 即使 npm 最终因为找不到 tarball 报错，它的输出也会显示它决定安装哪个版本。
	if strings.Contains(outputStr, "test-pkg@2.0.0") {
		t.Errorf("Expected 2.0.0 to be filtered, but it appeared in npm output. Output: %s", outputStr)
	}
	
	if !strings.Contains(outputStr, "test-pkg@1.0.0") && !strings.Contains(outputStr, "1.0.0") {
		t.Errorf("Expected npm to try installing 1.0.0, but it did not appear in output. Output: %s", outputStr)
	}
}

func TestPyPIE2E(t *testing.T) {
	// Check if pip or pip3 is available
	pipCmd := "pip"
	if _, err := exec.LookPath(pipCmd); err != nil {
		pipCmd = "pip3"
		if _, err := exec.LookPath(pipCmd); err != nil {
			t.Skip("pip/pip3 not found, skipping E2E test")
		}
	}

	now := time.Now().UTC()
	// Use slightly different time format if needed, but PyPI filter uses IsSafe which checks if it's > 7 days ago.
	oldTime := now.AddDate(0, 0, -10).Format(time.RFC3339)
	newTime := now.AddDate(0, 0, -1).Format(time.RFC3339)

	// 1. Mock Registry
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// PyPI pip calls /simple/<pkg>/ or /pypi/<pkg>/json
		if strings.HasSuffix(r.URL.Path, "/json") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"info": {"name": "test-pkg"},
				"releases": {
					"1.0.0": [{"upload_time_iso_8601": "` + oldTime + `", "url": "https://example.com/test_pkg-1.0.0-py3-none-any.whl"}],
					"2.0.0": [{"upload_time_iso_8601": "` + newTime + `", "url": "https://example.com/test_pkg-2.0.0-py3-none-any.whl"}]
				}
			}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/simple/test-pkg") {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
<!DOCTYPE html>
<html>
  <body>
    <a href="https://example.com/test_pkg-1.0.0-py3-none-any.whl">test_pkg-1.0.0-py3-none-any.whl</a>
    <a href="https://example.com/test_pkg-2.0.0-py3-none-any.whl">test_pkg-2.0.0-py3-none-any.whl</a>
  </body>
</html>
			`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer backend.Close()

	// 2. Run buffer7 + pip install
	_, filename, _, _ := runtime.Caller(0)
	rootPath := filepath.Join(filepath.Dir(filename), "../../")

	// We use --index-url via BUFFER7_UPSTREAM_REGISTRY in main.go
	cmd := exec.Command("go", "run", "main.go", pipCmd, "install", "test-pkg", "--dry-run", "--no-cache-dir", "-vvv")
	cmd.Dir = rootPath
	cmd.Env = append(os.Environ(), 
		"BUFFER7_UPSTREAM_REGISTRY="+backend.URL,
		"CGO_ENABLED=0",
	)
	
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	// pip might fail with 404 because we used example.com, 
	// but we only care if it found the correct version before failing.
	if err != nil && !strings.Contains(outputStr, "1.0.0") {
		t.Fatalf("pip install failed and could not find 1.0.0: %v\nOutput: %s", err, outputStr)
	}

	// 3. Verify output contains 1.0.0 and NOT 2.0.0
	if strings.Contains(outputStr, "2.0.0") {
		t.Errorf("Expected 2.0.0 to be filtered, but it appeared in pip output. Output: %s", outputStr)
	}
	if !strings.Contains(outputStr, "1.0.0") {
		t.Errorf("Expected pip to try installing 1.0.0, but it did not appear in output. Output: %s", outputStr)
	}
}
