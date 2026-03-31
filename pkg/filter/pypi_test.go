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
