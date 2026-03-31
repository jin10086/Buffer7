package filter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type pypiCacheEntry struct {
	forbiddenVersions map[string]bool
}

var pypiMetadataCache sync.Map // key: string (packageName), value: pypiCacheEntry

// FilterPyPI 过滤 PyPI 的 JSON 元数据
func FilterPyPI(body []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body, err
	}

	info, _ := data["info"].(map[string]interface{})
	packageName, _ := info["name"].(string)

	releases, ok := data["releases"].(map[string]interface{})
	if !ok {
		return body, nil
	}

	var forbidden []string
	forbiddenVersions := make(map[string]bool)
	for v, filesObj := range releases {
		files, ok := filesObj.([]interface{})
		if !ok || len(files) == 0 {
			forbidden = append(forbidden, v)
			forbiddenVersions[v] = true
			continue
		}

		isSafe := false
		for _, f := range files {
			fileMap, ok := f.(map[string]interface{})
			if !ok {
				continue
			}

			uploadTime, _ := fileMap["upload_time_iso_8601"].(string)
			if uploadTime != "" && IsSafe(uploadTime) {
				isSafe = true
				break
			}
		}

		if !isSafe {
			forbidden = append(forbidden, v)
			forbiddenVersions[v] = true
		}
	}

	for _, v := range forbidden {
		delete(releases, v)
	}

	if packageName != "" {
		pypiMetadataCache.Store(packageName, pypiCacheEntry{forbiddenVersions: forbiddenVersions})
	}

	return json.Marshal(data)
}

// FilterPyPISimple 过滤 PyPI 的 Simple API (HTML 或 JSON)
func FilterPyPISimple(packageName string, body []byte) ([]byte, error) {
	// 1. 获取该包的 JSON 元数据以确定不安全版本
	jsonURL := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)
	resp, err := http.Get(jsonURL)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return body, nil
	}

	jsonBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	// 2. 解析 JSON 元数据获取禁用的版本列表
	var metaData map[string]interface{}
	if err := json.Unmarshal(jsonBody, &metaData); err != nil {
		return body, err
	}

	releases, ok := metaData["releases"].(map[string]interface{})
	if !ok {
		return body, nil
	}

	forbiddenVersions := make(map[string]bool)
	for v, filesObj := range releases {
		files, _ := filesObj.([]interface{})
		isSafe := false
		for _, f := range files {
			fileMap, _ := f.(map[string]interface{})
			if uploadTime, _ := fileMap["upload_time_iso_8601"].(string); IsSafe(uploadTime) {
				isSafe = true
				break
			}
		}
		if !isSafe {
			forbiddenVersions[v] = true
		}
	}

	if len(forbiddenVersions) == 0 {
		return body, nil
	}

	// 3. 尝试解析为 Simple JSON
	var simpleData struct {
		Files []map[string]interface{} `json:"files"`
	}
	if err := json.Unmarshal(body, &simpleData); err == nil && len(simpleData.Files) > 0 {
		// 这是 Simple JSON 格式
		var filteredFiles []map[string]interface{}
		for _, file := range simpleData.Files {
			filename, _ := file["filename"].(string)
			keep := true
			for v := range forbiddenVersions {
				if strings.Contains(filename, "-"+v+"-") || strings.Contains(filename, "-"+v+".") {
					keep = false
					break
				}
			}
			if keep {
				filteredFiles = append(filteredFiles, file)
			}
			}
			simpleData.Files = filteredFiles
			return json.Marshal(simpleData)
			}


	// 4. 否则退回到 HTML 处理
	lines := strings.Split(string(body), "\n")
	var filteredLines []string
	for _, line := range lines {
		keep := true
		for v := range forbiddenVersions {
			if strings.Contains(line, "-"+v+"-") || strings.Contains(line, "-"+v+".") {
				keep = false
				break
			}
		}
		if keep {
			filteredLines = append(filteredLines, line)
		}
	}
	return []byte(strings.Join(filteredLines, "\n")), nil
}
