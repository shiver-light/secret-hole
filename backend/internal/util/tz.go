package util

import (
	"time"
)

var AsiaShanghai, _ = time.LoadLocation("Asia/Shanghai")

// shanghaiDate returns yyyy-mm-dd in Asia/shanghai for quota partitioning.
func ShanghaiDate(t time.Time) string {
	return t.In(AsiaShanghai).Format("2006-01-02")
}
