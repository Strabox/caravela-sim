package util

import (
	"fmt"
	"time"
)

func FmtDuration(d time.Duration) string {
	m := d / time.Minute
	return fmt.Sprintf("%d", m)
}
