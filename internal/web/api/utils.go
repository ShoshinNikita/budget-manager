package api

import "strings"

// trimSpacePointer is like 'strings.TrimPointer' but for pointers
func trimSpacePointer(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	return &trimmed
}
