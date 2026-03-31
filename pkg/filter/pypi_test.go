package filter

import (
	"encoding/json"
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
