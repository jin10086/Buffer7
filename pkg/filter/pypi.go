package filter

import (
	"encoding/json"
)

// FilterPyPI 过滤 PyPI 的 JSON 元数据
// PyPI JSON API: https://pypi.org/pypi/<package>/json
func FilterPyPI(body []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body, err
	}

	releases, ok := data["releases"].(map[string]interface{})
	if !ok {
		return body, nil
	}

	var forbidden []string
	for v, filesObj := range releases {
		files, ok := filesObj.([]interface{})
		if !ok || len(files) == 0 {
			forbidden = append(forbidden, v)
			continue
		}

		// 一个版本可能有多个文件（sdist, wheel），只要有一个发布于 7 天前即可认为该版本安全
		isSafe := false
		for _, f := range files {
			fileMap, ok := f.(map[string]interface{})
			if !ok { continue }
			
			uploadTime, _ := fileMap["upload_time_iso_8601"].(string)
			if uploadTime != "" && IsSafe(uploadTime) {
				isSafe = true
				break
			}
		}

		if !isSafe {
			forbidden = append(forbidden, v)
		}
	}

	for _, v := range forbidden {
		delete(releases, v)
	}

	return json.Marshal(data)
}
