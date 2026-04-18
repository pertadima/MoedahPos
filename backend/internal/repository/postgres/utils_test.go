package postgres

func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func ptr[T any](v T) *T {
	return &v
}
