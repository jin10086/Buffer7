package filter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestFilterPyPI(t *testing.T) {
	now := time.Now().UTC()
	oldVer := now.AddDate(0, 0, -10).Format(time.RFC3339)
	newVer := now.AddDate(0, 0, -2).Format(time.RFC3339)

	input := map[string]interface{}{
		"releases": map[string]interface{}{
			"1.0.0": []interface{}{
				map[string]interface{}{"upload_time_iso_8601": oldVer},
			},
			"1.1.0": []interface{}{
				map[string]interface{}{"upload_time_iso_8601": newVer},
			},
		},
	}

	body, _ := json.Marshal(input)
	filtered, _ := FilterPyPI(body)

	var output map[string]interface{}
	json.Unmarshal(filtered, &output)

	releases := output["releases"].(map[string]interface{})
	if _, ok := releases["1.1.0"]; ok {
		t.Errorf("Expected 1.1.0 to be filtered out")
	}
	if _, ok := releases["1.0.0"]; !ok {
		t.Errorf("Expected 1.0.0 to be kept")
	}
}

func TestFilterPyPI_MultiFile(t *testing.T) {
	now := time.Now().UTC()
	oldVer := now.AddDate(0, 0, -10).Format(time.RFC3339)
	newVer := now.AddDate(0, 0, -2).Format(time.RFC3339)

	input := map[string]interface{}{
		"releases": map[string]interface{}{
			"2.0.0": []interface{}{
				map[string]interface{}{"upload_time_iso_8601": oldVer}, // 安全文件
				map[string]interface{}{"upload_time_iso_8601": newVer}, // 新上传文件
			},
		},
	}

	body, _ := json.Marshal(input)
	filtered, _ := FilterPyPI(body)

	var output map[string]interface{}
	json.Unmarshal(filtered, &output)
	releases := output["releases"].(map[string]interface{})
	if _, ok := releases["2.0.0"]; !ok {
		t.Errorf("Expected 2.0.0 to be kept because it has at least one safe file")
	}
}

func TestFilterPyPI_Empty(t *testing.T) {
	input := map[string]interface{}{
		"releases": map[string]interface{}{
			"3.0.0": []interface{}{}, // 空版本
		},
	}

	body, _ := json.Marshal(input)
	filtered, _ := FilterPyPI(body)

	var output map[string]interface{}
	json.Unmarshal(filtered, &output)
	releases := output["releases"].(map[string]interface{})
	if _, ok := releases["3.0.0"]; ok {
		t.Errorf("Expected empty release 3.0.0 to be filtered out")
	}
}

func TestFilterPyPI_EdgeCases(t *testing.T) {
	// Case 1: Invalid JSON
	_, err := FilterPyPI([]byte(`{invalid}`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Case 2: Missing releases field
	body, _ := FilterPyPI([]byte(`{"info":{"name":"test"}}`))
	if !json.Valid(body) {
		t.Error("Expected valid JSON for missing releases")
	}

	// Case 3: All files unsafe
	now := time.Now().UTC()
	newVer := now.AddDate(0, 0, -2).Format(time.RFC3339)
	input := map[string]interface{}{
		"releases": map[string]interface{}{
			"4.0.0": []interface{}{
				map[string]interface{}{"upload_time_iso_8601": newVer},
			},
		},
	}
	body, _ = json.Marshal(input)
	filtered, _ := FilterPyPI(body)
	var output map[string]interface{}
	json.Unmarshal(filtered, &output)
	releases := output["releases"].(map[string]interface{})
	if _, ok := releases["4.0.0"]; ok {
		t.Error("Expected 4.0.0 to be filtered out as all files are unsafe")
	}

	// Case 4: Invalid files element
	input = map[string]interface{}{
		"releases": map[string]interface{}{
			"5.0.0": []interface{}{"not-a-map"},
		},
	}
	body, _ = json.Marshal(input)
	filtered, _ = FilterPyPI(body)
	json.Unmarshal(filtered, &output)
	releases = output["releases"].(map[string]interface{})
	if _, ok := releases["5.0.0"]; ok {
		t.Error("Expected 5.0.0 to be filtered out with invalid file element")
	}

	// Case 5: Missing or empty upload_time
	input = map[string]interface{}{
		"releases": map[string]interface{}{
			"6.0.0": []interface{}{
				map[string]interface{}{"something": "else"},
				map[string]interface{}{"upload_time_iso_8601": ""},
			},
		},
	}
	body, _ = json.Marshal(input)
	filtered, _ = FilterPyPI(body)
	json.Unmarshal(filtered, &output)
	releases = output["releases"].(map[string]interface{})
	if _, ok := releases["6.0.0"]; ok {
		t.Error("Expected 6.0.0 to be filtered out with missing upload_time")
	}
}

func TestPyPICache(t *testing.T) {
	pkgName := "test-pkg"
	entry := pypiCacheEntry{
		forbiddenVersions: map[string]bool{"1.0.0": true},
	}
	pypiMetadataCache.Store(pkgName, entry)

	val, ok := pypiMetadataCache.Load(pkgName)
	if !ok {
		t.Fatal("Cache miss")
	}
	cachedEntry := val.(pypiCacheEntry)
	if !cachedEntry.forbiddenVersions["1.0.0"] {
		t.Error("Wrong cache value")
	}
}

func TestFilterPyPISimpleCacheHit(t *testing.T) {
	packageName := "cached-pkg-test"
	forbiddenVersion := "2.0.0"

	// 1. 手动向缓存注入禁止版本
	pypiMetadataCache.Store(packageName, pypiCacheEntry{
		forbiddenVersions: map[string]bool{forbiddenVersion: true},
	})

	// 2. 模拟 Simple API 的响应 (HTML 格式)
	inputHTML := `
<!DOCTYPE html>
<html>
  <body>
    <a href="https://files.pythonhosted.org/packages/source/c/cached-pkg-test/cached-pkg-test-1.0.0.tar.gz">cached-pkg-test-1.0.0.tar.gz</a>
    <a href="https://files.pythonhosted.org/packages/source/c/cached-pkg-test/cached-pkg-test-2.0.0.tar.gz">cached-pkg-test-2.0.0.tar.gz</a>
  </body>
</html>`

	// 3. 调用 FilterPyPISimple，应该命中缓存且不需要执行 http.Get
	// 由于我们注入了一个不存在的包名，如果它尝试 http.Get，会返回 404 或错误
	// 但如果命中缓存，它将直接使用我们注入的 forbiddenVersions
	filtered, err := FilterPyPISimple(packageName, []byte(inputHTML))
	if err != nil {
		t.Fatalf("FilterPyPISimple failed: %v", err)
	}

	filteredStr := string(filtered)
	if !strings.Contains(filteredStr, "cached-pkg-test-1.0.0.tar.gz") {
		t.Errorf("Expected 1.0.0 to be kept")
	}
	if strings.Contains(filteredStr, "cached-pkg-test-2.0.0.tar.gz") {
		t.Errorf("Expected 2.0.0 to be filtered out via cache")
	}

	// 4. 测试 JSON 格式 (Simple JSON PEP 691)
	inputJSON := `{"files": [{"filename": "cached-pkg-test-1.0.0.tar.gz"}, {"filename": "cached-pkg-test-2.0.0.tar.gz"}]}`
	filteredJSON, err := FilterPyPISimple(packageName, []byte(inputJSON))
	if err != nil {
		t.Fatalf("FilterPyPISimple JSON failed: %v", err)
	}

	var outputJSON struct {
		Files []map[string]interface{} `json:"files"`
	}
	if err := json.Unmarshal(filteredJSON, &outputJSON); err != nil {
		t.Fatalf("Failed to unmarshal filtered JSON: %v", err)
	}

	if len(outputJSON.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(outputJSON.Files))
	}
	if filename, ok := outputJSON.Files[0]["filename"].(string); !ok || filename != "cached-pkg-test-1.0.0.tar.gz" {
		t.Errorf("Expected cached-pkg-test-1.0.0.tar.gz, got %v", filename)
	}
}
