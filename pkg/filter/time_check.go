package filter

import (
	"time"
)

func IsSafe(publishDateStr string) bool {
	publishDate, err := time.Parse(time.RFC3339, publishDateStr)
	if err != nil {
		// 如果解析失败，保守起见认为不安全
		return false
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -7)
	return publishDate.Before(cutoff)
}
