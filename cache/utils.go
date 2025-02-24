package cache

import (
	"strings"
	"unicode"
)

// Helper function to preprocess strings
func preprocessString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	// Remove special characters
	var builder strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// Levenshtein distance algorithm
func levenshteinDistance(a, b string) int {
	la := len(a)
	lb := len(b)

	if la == 0 {
		return lb
	} else if lb == 0 {
		return la
	}

	// Create a distance matrix
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
	}
	for i := 0; i <= la; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(
				dp[i-1][j]+1,      // deletion
				dp[i][j-1]+1,      // insertion
				dp[i-1][j-1]+cost, // substitution
			)
		}
	}
	return dp[la][lb]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	} else if b < c {
		return b
	}
	return c
}
