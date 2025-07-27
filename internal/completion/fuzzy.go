package completion

import (
	"sort"
	"strings"
)

// FuzzyMatch represents a fuzzy match with its score
type FuzzyMatch struct {
	Value string
	Score int
}

// FuzzyMatchStrings performs fuzzy matching on a slice of strings
// Returns matches sorted by score (best matches first)
func FuzzyMatchStrings(input string, candidates []string) []string {
	if input == "" {
		return candidates
	}
	
	var matches []FuzzyMatch
	
	for _, candidate := range candidates {
		score := fuzzyScore(input, candidate)
		if score > 0 {
			matches = append(matches, FuzzyMatch{
				Value: candidate,
				Score: score,
			})
		}
	}
	
	// Sort by score (descending - higher scores first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	
	// Extract just the values
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = match.Value
	}
	
	return result
}

// fuzzyScore calculates a fuzzy match score between input and candidate
// Returns 0 if no match, higher scores for better matches
func fuzzyScore(input, candidate string) int {
	input = strings.ToLower(input)
	candidate = strings.ToLower(candidate)
	
	// Exact match gets highest score
	if input == candidate {
		return 1000
	}
	
	// Prefix match gets high score
	if strings.HasPrefix(candidate, input) {
		return 900 + len(input) // Longer prefix = higher score
	}
	
	// Check if all characters in input appear in order in candidate
	inputRunes := []rune(input)
	candidateRunes := []rune(candidate)
	
	if len(inputRunes) == 0 {
		return 0
	}
	
	score := 0
	inputIdx := 0
	consecutiveMatches := 0
	
	for i, candidateChar := range candidateRunes {
		if inputIdx < len(inputRunes) && candidateChar == inputRunes[inputIdx] {
			inputIdx++
			consecutiveMatches++
			
			// Bonus for consecutive matches
			if consecutiveMatches > 1 {
				score += 10 * consecutiveMatches
			} else {
				score += 5
			}
			
			// Bonus for matches at word boundaries
			if i == 0 || candidateRunes[i-1] == '-' || candidateRunes[i-1] == '/' || candidateRunes[i-1] == '_' {
				score += 15
			}
		} else {
			consecutiveMatches = 0
		}
	}
	
	// Only return score if all input characters were matched
	if inputIdx == len(inputRunes) {
		// Bonus for shorter candidates (more specific matches)
		lengthBonus := max(0, 50 - len(candidate))
		return score + lengthBonus
	}
	
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}