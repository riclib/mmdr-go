package mmdr

import "testing"

// Representative small diagram used across the render/validate benchmarks.
const benchSrc = `flowchart LR
    A[Start] --> B{Decision}
    B -->|yes| C[Do thing]
    B -->|no| D[Skip]
    C --> E[End]
    D --> E`

func BenchmarkRender(b *testing.B) {
	Render(benchSrc) // warm fonts/regex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Render(benchSrc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidate(b *testing.B) {
	Validate(benchSrc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(benchSrc)
	}
}

func BenchmarkRenderChecked(b *testing.B) {
	RenderChecked(benchSrc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := RenderChecked(benchSrc); err != nil {
			b.Fatal(err)
		}
	}
}
