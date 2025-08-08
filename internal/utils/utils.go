package utils

import (
	"time"
)

func GetNowLocalAdjusted() time.Time {
	return time.Now().Add(time.Hour * 7) // hardcoded for Asia/Bangkok
}
