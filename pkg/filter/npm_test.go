package filter

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFilterNPM(t *testing.T) {
	now := time.Now().UTC()
	oldVer := now.AddDate(0, 0, -10).Format(time.RFC3339)
	newVer := now.AddDate(0, 0, -2).Format(time.RFC3339)

	input := map[string]interface{}{
		"time": map[string]interface{}{
			"1.0.0": oldVer,
			"1.1.0": newVer,
		},
		"versions": map[string]interface{}{
			"1.0.0": map[string]interface{}{},
			"1.1.0": map[string]interface{}{},
		},
		"dist-tags": map[string]interface{}{
			"latest": "1.1.0",
		},
	}

	body, _ := json.Marshal(input)
	filtered, _ := FilterNPM(body)

	var output map[string]interface{}
	json.Unmarshal(filtered, &output)

	versions := output["versions"].(map[string]interface{})
	if _, ok := versions["1.1.0"]; ok {
		t.Errorf("Expected 1.1.0 to be filtered out")
	}
	distTags := output["dist-tags"].(map[string]interface{})
	if distTags["latest"] != "1.0.0" {
		t.Errorf("Expected latest to be downgraded to 1.0.0, got %v", distTags["latest"])
	}
}

func TestFilterNPM_EmptySet(t *testing.T) {
	now := time.Now().UTC()
	newVer := now.AddDate(0, 0, -2).Format(time.RFC3339)

	input := map[string]interface{}{
		"time": map[string]interface{}{
			"1.1.0": newVer,
		},
		"versions": map[string]interface{}{
			"1.1.0": map[string]interface{}{},
		},
		"dist-tags": map[string]interface{}{
			"latest": "1.1.0",
		},
	}
	body, _ := json.Marshal(input)
	filtered, _ := FilterNPM(body)

	var output map[string]interface{}
	json.Unmarshal(filtered, &output)
	versions := output["versions"].(map[string]interface{})
	if len(versions) != 0 {
		t.Errorf("Expected versions to be empty, got %d", len(versions))
	}
	distTags := output["dist-tags"].(map[string]interface{})
	if distTags["latest"] != nil {
		t.Errorf("Expected latest tag to be cleared or unchanged, got %v", distTags["latest"])
	}
}
