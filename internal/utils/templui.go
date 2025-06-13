package utils

import (
	"crypto/rand"
	"fmt"

	twmerge "github.com/Oudwins/tailwind-merge-go"
	"github.com/a-h/templ"
)

// TwMerge combines Tailwind classes and resolves conflicts.
func TwMerge(classes ...string) string {
	return twmerge.Merge(classes...)
}

// If returns value if condition is true, otherwise an empty value of type T.
func If[T comparable](condition bool, value T) T {
	var empty T
	if condition {
		return value
	}
	return empty
}

// IfElse returns trueValue if condition is true, otherwise falseValue.
func IfElse[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

// MergeAttributes combines multiple Attributes into one.
func MergeAttributes(attrs ...templ.Attributes) templ.Attributes {
	merged := templ.Attributes{}
	for _, attr := range attrs {
		for k, v := range attr {
			merged[k] = v
		}
	}
	return merged
}

// RandomID generates a random ID string.
func RandomID() string {
	return fmt.Sprintf("id-%s", rand.Text())
}
