package config

import "testing"

func TestPlaceholder(t *testing.T) {
	t.Run("loads valid config", func(t *testing.T) {
		// TODO:
		// - create a minimal valid JSON config fixture
		// - call the future Load function
		// - assert no error
		// - assert parsed fields match expectations
	})

	t.Run("rejects invalid config", func(t *testing.T) {
		// TODO:
		// - create malformed or incomplete JSON
		// - call the future Load function
		// - assert an error is returned
	})

	t.Run("rejects route without backends", func(t *testing.T) {
		// TODO:
		// - define a config with a route but no backends
		// - call the future Validate function
		// - assert validation fails
	})
}

func BenchmarkPlaceholder(b *testing.B) {
	// TODO:
	// - prepare a representative config payload once before b.ResetTimer()
	// - benchmark the future Load or Validate path
	// - use b.ReportAllocs() when useful

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// TODO: call the function being benchmarked
	}
}
