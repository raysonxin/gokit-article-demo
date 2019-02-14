package main

import "errors"

// Service Define a service interface
type Service interface {

	// Add calculate a+b
	Add(a, b int) int

	// Subtract calculate a-b
	Subtract(a, b int) int

	// Multiply calculate a*b
	Multiply(a, b int) int

	// Divide calculate a/b
	Divide(a, b int) (int, error)
}

//ArithmeticService implement Service interface
type ArithmeticService struct {
}

// Add implement Add method
func (s ArithmeticService) Add(a, b int) int {
	return a + b
}

// Substract implement Substract method
func (s ArithmeticService) Substract(a, b int) int {
	return a - b
}

// Multiply implement Multiply method
func (s ArithmeticService) Multiply(a, b int) int {
	return a * b
}

// Divide implement Divide method
func (s ArithmeticService) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("the dividend can not be zero!")
	}

	return a / b, nil
}
