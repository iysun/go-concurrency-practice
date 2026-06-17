package calculator_test

import (
	"testing"

	"go-concurrency-practice/12-testing-practice/calculator"
)

// TestAdd demonstrates the table-driven test pattern — the idiomatic Go testing style.
// Each case has a name (shown on failure), inputs, and expected output.
func TestAdd(t *testing.T) {
	c := calculator.New()
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positive", 1, 2, 3},
		{"negative", -1, -2, -3},
		{"mixed", -1, 2, 1},
		{"zeros", 0, 0, 0},
		{"floats", 0.1, 0.2, 0.3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Add(tt.a, tt.b)
			// Use a small epsilon for float comparison
			if abs(got-tt.want) > 1e-9 {
				t.Errorf("Add(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	c := calculator.New()

	t.Run("normal division", func(t *testing.T) {
		got, err := c.Divide(10, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 5 {
			t.Errorf("got %v, want 5", got)
		}
	})

	t.Run("division by zero", func(t *testing.T) {
		_, err := c.Divide(1, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Check for the specific error type — don't just check err != nil
		if err != calculator.ErrDivisionByZero {
			t.Errorf("got %v, want ErrDivisionByZero", err)
		}
	})
}

// TestFibonacci is a TODO — implement after writing Fibonacci in calc.go
func TestFibonacci(t *testing.T) {
	c := calculator.New()
	tests := []struct {
		n    int
		want int
	}{
		{0, 0}, {1, 1}, {2, 1}, {3, 2}, {5, 5}, {10, 55},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := c.Fibonacci(tt.n)
			if got != tt.want {
				t.Errorf("Fibonacci(%d) = %d, want %d", tt.n, got, tt.want)
			}
		})
	}
}

// BenchmarkFibonacci measures performance — run with go test -bench=. -benchmem
// TODO: add a recursive implementation to compare against
func BenchmarkFibonacci(b *testing.B) {
	c := calculator.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Fibonacci(30)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
