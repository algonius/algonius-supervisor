package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

// MatchType represents the type of pattern matching to perform
type MatchType int

const (
	MatchExact MatchType = iota // Exact string match
	MatchPrefix                 // Prefix match
	MatchSuffix                 // Suffix match
	MatchContains               // Contains match
	MatchRegex                  // Regular expression match
	MatchWildcard                // Wildcard pattern match (* and ?)
)

// Pattern represents a pattern that can be used to match agent names
type Pattern struct {
	Type     MatchType `json:"type"`
	Pattern  string    `json:"pattern"`
	Compiled *regexp.Regexp `json:"-"` // Compiled regex (only for regex patterns)
}

// NewPattern creates a new Pattern from a string pattern
func NewPattern(patternStr string) (*Pattern, error) {
	if patternStr == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	pattern := &Pattern{
		Pattern: patternStr,
	}

	// Determine pattern type and compile if necessary
	if strings.HasPrefix(patternStr, "regex:") {
		pattern.Type = MatchRegex
		regexStr := strings.TrimPrefix(patternStr, "regex:")
		compiled, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}
		pattern.Compiled = compiled
	} else if strings.Contains(patternStr, "*") || strings.Contains(patternStr, "?") {
		pattern.Type = MatchWildcard
		// Convert wildcard to regex for easier matching
		regexStr := wildcardToRegex(patternStr)
		compiled, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid wildcard pattern: %v", err)
		}
		pattern.Compiled = compiled
	} else if strings.HasPrefix(patternStr, "prefix:") {
		pattern.Type = MatchPrefix
		pattern.Pattern = strings.TrimPrefix(patternStr, "prefix:")
	} else if strings.HasPrefix(patternStr, "suffix:") {
		pattern.Type = MatchSuffix
		pattern.Pattern = strings.TrimPrefix(patternStr, "suffix:")
	} else if strings.HasPrefix(patternStr, "contains:") {
		pattern.Type = MatchContains
		pattern.Pattern = strings.TrimPrefix(patternStr, "contains:")
	} else if strings.HasPrefix(patternStr, "exact:") {
		pattern.Type = MatchExact
		pattern.Pattern = strings.TrimPrefix(patternStr, "exact:")
	} else {
		// Default to exact match
		pattern.Type = MatchExact
	}

	return pattern, nil
}

// Matches checks if the given string matches the pattern
func (p *Pattern) Matches(s string) bool {
	switch p.Type {
	case MatchExact:
		return p.Pattern == s
	case MatchPrefix:
		return strings.HasPrefix(s, p.Pattern)
	case MatchSuffix:
		return strings.HasSuffix(s, p.Pattern)
	case MatchContains:
		return strings.Contains(s, p.Pattern)
	case MatchRegex, MatchWildcard:
		return p.Compiled != nil && p.Compiled.MatchString(s)
	default:
		return false
	}
}

// String returns a string representation of the pattern
func (p *Pattern) String() string {
	switch p.Type {
	case MatchExact:
		return fmt.Sprintf("exact:%s", p.Pattern)
	case MatchPrefix:
		return fmt.Sprintf("prefix:%s", p.Pattern)
	case MatchSuffix:
		return fmt.Sprintf("suffix:%s", p.Pattern)
	case MatchContains:
		return fmt.Sprintf("contains:%s", p.Pattern)
	case MatchRegex:
		return fmt.Sprintf("regex:%s", p.Pattern)
	case MatchWildcard:
		return p.Pattern
	default:
		return p.Pattern
	}
}

// wildcardToRegex converts a wildcard pattern to a regex pattern
func wildcardToRegex(wildcard string) string {
	regex := regexp.QuoteMeta(wildcard)
	regex = strings.ReplaceAll(regex, "\\*", ".*")
	regex = strings.ReplaceAll(regex, "\\?", ".")
	regex = "^" + regex + "$"
	return regex
}

// Matcher represents a collection of patterns for matching agent names
type Matcher struct {
	patterns []*Pattern
}

// NewMatcher creates a new Matcher with the given patterns
func NewMatcher(patterns []string) (*Matcher, error) {
	matcher := &Matcher{
		patterns: make([]*Pattern, 0, len(patterns)),
	}

	for _, patternStr := range patterns {
		pattern, err := NewPattern(patternStr)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %v", patternStr, err)
		}
		matcher.patterns = append(matcher.patterns, pattern)
	}

	return matcher, nil
}

// Matches checks if any of the patterns match the given string
func (m *Matcher) Matches(s string) bool {
	for _, pattern := range m.patterns {
		if pattern.Matches(s) {
			return true
		}
	}
	return false
}

// FilterStrings filters a slice of strings, returning only those that match any pattern
func (m *Matcher) FilterStrings(strings []string) []string {
	var matched []string
	for _, s := range strings {
		if m.Matches(s) {
			matched = append(matched, s)
		}
	}
	return matched
}

// GetMatchingPatterns returns the patterns that match the given string
func (m *Matcher) GetMatchingPatterns(s string) []*Pattern {
	var matching []*Pattern
	for _, pattern := range m.patterns {
		if pattern.Matches(s) {
			matching = append(matching, pattern)
		}
	}
	return matching
}

// Count returns the number of patterns in the matcher
func (m *Matcher) Count() int {
	return len(m.patterns)
}

// IsEmpty returns true if the matcher has no patterns
func (m *Matcher) IsEmpty() bool {
	return len(m.patterns) == 0
}

// String returns a string representation of all patterns
func (m *Matcher) String() string {
	if len(m.patterns) == 0 {
		return "Matcher{empty}"
	}

	var parts []string
	for _, pattern := range m.patterns {
		parts = append(parts, pattern.String())
	}
	return fmt.Sprintf("Matcher{%s}", strings.Join(parts, ", "))
}

// ParseAgentPatterns parses agent name patterns from command line arguments
// Supports multiple patterns and various syntax:
// - exact match: "agent1" or "exact:agent1"
// - prefix: "prefix:web"
// - suffix: "suffix:-worker"
// - contains: "contains:db"
// - regex: "regex:web-.*-worker"
// - wildcard: "web-*-worker" (supports * and ?)
func ParseAgentPatterns(args []string) (*Matcher, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("at least one pattern must be provided")
	}

	return NewMatcher(args)
}

// ParseSinglePattern parses a single agent name pattern
func ParseSinglePattern(pattern string) (*Pattern, error) {
	return NewPattern(pattern)
}

// Common patterns for typical use cases
var (
	// AllAgents matches any agent name
	AllAgents = func() *Pattern {
		pattern, _ := NewPattern("*")
		return pattern
	}()

	// WebAgents matches typical web server agent names
	WebAgents = func() *Pattern {
		pattern, _ := NewPattern("prefix:web-")
		return pattern
	}()

	// WorkerAgents matches typical worker agent names
	WorkerAgents = func() *Pattern {
		pattern, _ := NewPattern("suffix:-worker")
		return pattern
	}()

	// DBAgents matches typical database-related agent names
	DBAgents = func() *Pattern {
		pattern, _ := NewPattern("contains:db")
		return pattern
	}()
)

// PredefinedMatcher creates a matcher from common predefined patterns
func PredefinedMatcher(name string) (*Matcher, error) {
	switch name {
	case "all":
		return NewMatcher([]string{"*"})
	case "web":
		return NewMatcher([]string{"prefix:web-"})
	case "workers":
		return NewMatcher([]string{"suffix:-worker"})
	case "database":
		return NewMatcher([]string{"contains:db"})
	case "api":
		return NewMatcher([]string{"prefix:api-", "contains:gateway"})
	default:
		return nil, fmt.Errorf("unknown predefined matcher: %s", name)
	}
}

// ValidatePatternString validates if a pattern string is syntactically correct
func ValidatePatternString(patternStr string) error {
	_, err := NewPattern(patternStr)
	return err
}

// GetSupportedPatternTypes returns a list of supported pattern types with descriptions
func GetSupportedPatternTypes() map[string]string {
	return map[string]string{
		"exact":    "Exact string match (e.g., 'agent1' or 'exact:agent1')",
		"prefix":   "Prefix match (e.g., 'prefix:web-')",
		"suffix":   "Suffix match (e.g., 'suffix:-worker')",
		"contains": "Contains match (e.g., 'contains:db')",
		"regex":    "Regular expression match (e.g., 'regex:web-.*-worker')",
		"wildcard": "Wildcard pattern with * and ? (e.g., 'web-*-worker')",
	}
}

// ExamplePatterns returns example patterns for documentation
func ExamplePatterns() []string {
	return []string{
		"agent1",                  // Exact match
		"exact:agent1",            // Explicit exact match
		"prefix:web-",             // All agents starting with 'web-'
		"suffix:-worker",          // All agents ending with '-worker'
		"contains:db",             // All agents containing 'db'
		"regex:web-.*-worker",     // Regex pattern
		"web-*-worker",            // Wildcard pattern
		"api-*",                   // All agents starting with 'api-'
		"*-db-*",                  // All agents with '-db-' in the middle
		"worker-?-service",        // Wildcard with single character
	}
}