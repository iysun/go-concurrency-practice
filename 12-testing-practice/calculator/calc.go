package calculator

import "errors"

var ErrDivisionByZero = errors.New("division by zero")

// Calculator performs basic arithmetic.
// These functions are intentionally simple — the interesting part is the tests.
type Calculator struct{}

func New() *Calculator { return &Calculator{} }

func (c *Calculator) Add(a, b float64) float64      { return a + b }
func (c *Calculator) Subtract(a, b float64) float64 { return a - b }
func (c *Calculator) Multiply(a, b float64) float64 { return a * b }

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Fibonacci returns the n-th Fibonacci number (n >= 0).
// TODO: implement iteratively (not recursively — benchmark the difference)
func (c *Calculator) Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	// TODO: implement
	return 0
}
