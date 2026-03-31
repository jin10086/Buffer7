package filter

import (
	"encoding/json"
	"strings"
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

func TestFilterNPM_EdgeCases(t *testing.T) {
	// Case 1: Invalid JSON
	_, err := FilterNPM([]byte(`{invalid}`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Case 2: Missing fields
	body, _ := FilterNPM([]byte(`{"other":"field"}`))
	if string(body) != `{"other":"field"}` {
		t.Errorf("Expected unchanged body for missing fields, got %s", body)
	}

	// Case 3: Latest is already safe
	now := time.Now().UTC()
	oldVer := now.AddDate(0, 0, -10).Format(time.RFC3339)
	input := map[string]interface{}{
		"time": map[string]interface{}{
			"1.0.0": oldVer,
		},
		"versions": map[string]interface{}{
			"1.0.0": map[string]interface{}{},
		},
		"dist-tags": map[string]interface{}{
			"latest": "1.0.0",
		},
	}
	body, _ = json.Marshal(input)
	filtered, _ := FilterNPM(body)
	var output map[string]interface{}
	json.Unmarshal(filtered, &output)
	distTags := output["dist-tags"].(map[string]interface{})
	if distTags["latest"] != "1.0.0" {
		t.Errorf("Expected latest to remain 1.0.0, got %v", distTags["latest"])
	}

	// Case 4: dist-tags is nil
	input = map[string]interface{}{
		"time": map[string]interface{}{
			"1.0.0": oldVer,
		},
		"versions": map[string]interface{}{
			"1.0.0": map[string]interface{}{},
		},
	}
	body, _ = json.Marshal(input)
	filtered, _ = FilterNPM(body)
	if !strings.Contains(string(filtered), `"1.0.0"`) {
		t.Error("Expected 1.0.0 to remain when dist-tags is missing")
	}

	// Case 5: Missing versions field
	input = map[string]interface{}{
		"time": map[string]interface{}{
			"1.0.0": oldVer,
		},
	}
	body, _ = json.Marshal(input)
	filtered, _ = FilterNPM(body)
	if !strings.Contains(string(filtered), `"time"`) {
		t.Error("Expected time field to remain when versions is missing")
	}

	// Case 6: Invalid types in time field (triggers !ok and time.Parse err)
	input = map[string]interface{}{
		"time": map[string]interface{}{
			"1.0.0": 123, // not string
			"1.1.0": "invalid-time",
			"latest": now.AddDate(0, 0, -2).Format(time.RFC3339),
		},
		"versions": map[string]interface{}{
			"1.0.0": map[string]interface{}{},
			"1.1.0": map[string]interface{}{},
			"latest": map[string]interface{}{},
		},
		"dist-tags": map[string]interface{}{
			"latest": "latest",
		},
	}
	body, _ = json.Marshal(input)
	FilterNPM(body)

	// Case 7: latest tag not a string
	input = map[string]interface{}{
		"time": map[string]interface{}{
			"1.0.0": oldVer,
		},
		"versions": map[string]interface{}{
			"1.0.0": map[string]interface{}{},
		},
		"dist-tags": map[string]interface{}{
			"latest": 123, // not string
		},
	}
	body, _ = json.Marshal(input)
	FilterNPM(body)

	// Case 8: dist-tags exists but no latest
	input = map[string]interface{}{
		"time": map[string]interface{}{
			"1.0.0": oldVer,
		},
		"versions": map[string]interface{}{
			"1.0.0": map[string]interface{}{},
		},
		"dist-tags": map[string]interface{}{
			"other": "tag",
		},
	}
	body, _ = json.Marshal(input)
	FilterNPM(body)
}
