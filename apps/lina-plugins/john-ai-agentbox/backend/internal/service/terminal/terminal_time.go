// This file contains terminal timestamp projection helpers. Public API DTOs use
// Unix millisecond values while generated entities keep database time types.

package terminal

import "time"

func unixMilliFromTimePtr(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.UnixMilli()
}

func unixMilliPtrFromTimePtr(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	out := value.UnixMilli()
	return &out
}
