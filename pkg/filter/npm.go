package filter

import (
	"encoding/json"
	"time"
)

func FilterNPM(body []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body, err
	}

	times, ok := data["time"].(map[string]interface{})
	if !ok { return body, nil }
	versions, ok := data["versions"].(map[string]interface{})
	if !ok { return body, nil }
	distTags, _ := data["dist-tags"].(map[string]interface{})

	var forbidden []string
	for v, tStr := range times {
		if v == "created" || v == "modified" { continue }
		if !IsSafe(tStr.(string)) {
			forbidden = append(forbidden, v)
		}
	}

	for _, v := range forbidden {
		delete(versions, v)
		delete(times, v)
	}

	if distTags != nil {
		latest, _ := distTags["latest"].(string)
		isForbidden := false
		for _, v := range forbidden {
			if v == latest { isForbidden = true; break }
		}
		if isForbidden {
			// Find newest safe version
			var newestSafe string
			var newestTime time.Time
			for v, tStr := range times {
				if v == "created" || v == "modified" { continue }
				tStrVal, ok := tStr.(string)
				if !ok { continue }
				t, err := time.Parse(time.RFC3339, tStrVal)
				if err != nil { continue }
				if t.After(newestTime) {
					newestTime = t
					newestSafe = v
				}
			}
			if newestSafe != "" {
				distTags["latest"] = newestSafe
			} else {
				delete(distTags, "latest")
			}
		}
	}

	return json.Marshal(data)
}
