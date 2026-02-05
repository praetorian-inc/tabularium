package external

// derefString safely dereferences a string pointer.
// Returns empty string if the pointer is nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
