package main

import "strings"

func greet(name string) string {
	currentName := strings.TrimSpace(name)

	if len(currentName) == 0 {
		return "Hello, World!"
	}

	return "Hello, " + currentName + "!"
}
