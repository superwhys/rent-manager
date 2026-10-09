package mysql

import "fmt"

// validateByteLength preserves the byte limits enforced by the previous persistence schema.
func validateByteLength(limit int, values ...string) error {
	for _, value := range values {
		if len(value) > limit {
			return fmt.Errorf("字段长度不能超过 %d 字节", limit)
		}
	}
	return nil
}
