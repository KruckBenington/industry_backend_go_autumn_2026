package main

import (
	"errors"
	"strconv"
)

var ErrZero = errors.New("zero is not allowed")

func fizzBuzz(n int) (string, error) {
	nIsMod3 := n % 3 == 0
	nIsMod5 := n % 5 == 0

	switch {

	case n == 0:
		return "", ErrZero
	case nIsMod3 && nIsMod5:
		return "FizzBuzz", nil
	case nIsMod5:
		return "Buzz", nil
	case nIsMod3:
		return "Fizz", nil
	default:
		return strconv.Itoa(n), nil
	}

}
