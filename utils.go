package main

import (
	"fmt"
	"time"
)

// this helper extracts all the alphabetic characters from the input string
func sanitizeFilename(input string) string {
	result := ""
	for _, char := range input {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			result += string(char)
		}
		// if there is a space, we can also add it
		if char == ' ' {
			result += " "
		}
		// if there is a dash, we can also add it
		if char == '-' {
			result += "_"
		}
		// if there is an underscore, we can also add it
		if char == '_' {
			result += "_"
		}
		// if there is a dot, we can also add it
		if char == '.' {
			result += "_"
		}
		// if there is a comma, we can also add it
		if char == ',' {
			result += "_"
		}
		// if there is a semicolon, we can also add it
		if char == ';' {
			result += "_"
		}
	}
	// If there are more than one subsequent underscores, replace them with a single underscore
	for i := 0; i < len(result)-1; i++ {
		if result[i] == '_' && result[i+1] == '_' {
			result = result[:i+1] + result[i+2:]
			i--
		}
	}
	// Trim leading and trailing underscores
	if len(result) > 0 && result[0] == '_' {
		result = result[1:]
	}
	if len(result) > 0 && result[len(result)-1] == '_' {
		result = result[:len(result)-1]
	}
	// If the result is empty, return a default value
	if result == "" {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		return "default_" + timestamp

	}
	return result
}
