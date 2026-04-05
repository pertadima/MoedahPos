package service

// stringToPtr returns a pointer to the string if it's not empty, otherwise nil.
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ptrToString returns the string value if the pointer is not nil, otherwise "".
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
