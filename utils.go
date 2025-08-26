package service

import (
	"regexp"
	"strings"
)

// envName constructs a normalized environment variable name by combining a prefix and a name.
// If the prefix is empty, it defaults to "SERVICE_".
// If the prefix does not end with an underscore, one is appended.
// The resulting string (prefix + name) is then normalized to conform to environment variable naming conventions
// using normalizeEnvName, which ensures uppercase, replaces invalid characters, collapses underscores, etc.
//
// Example usage:
//
//	envName("MYAPP", "LOG_LEVEL")   // returns "MYAPP_LOG_LEVEL"
//	envName("", "CONFIG_PATH")      // returns "SERVICE_CONFIG_PATH"
func envName(prefix string, name string) string {
	if prefix == "" {
		prefix = "SERVICE_"
	} else {
		if !strings.HasSuffix(prefix, "_") {
			prefix += "_"
		}
	}

	return normalizeEnvName(prefix + name)
}

// normalizeEnvName transforms an input string into a valid, conventional environment variable name.
// The normalization process performs the following steps:
//  1. Converts the input to uppercase for conventional env var style.
//  2. Replaces any character that is not an uppercase letter, digit, or underscore with an underscore.
//  3. Collapses consecutive underscores into a single underscore to avoid redundant separators.
//  4. Trims any leading or trailing underscores for a clean result.
//  5. If the resulting name starts with a digit, prepends an underscore to ensure validity.
//
// Examples:
//
//	"my-config" → "MY_CONFIG"
//	"test@123" → "TEST_123"
//	"1st-place" → "_1ST_PLACE"
//	"hello__world" → "HELLO_WORLD"
func normalizeEnvName(name string) string {
	if name == "" {
		return name
	}

	// Convert to uppercase for conventional env var style
	name = strings.ToUpper(name)

	// Replace any non-alphanumeric characters (except underscores) with underscores
	re := regexp.MustCompile(`[^A-Z0-9_]`)
	name = re.ReplaceAllString(name, "_")

	// Collapse consecutive underscores into a single underscore
	re = regexp.MustCompile(`_+`)
	name = re.ReplaceAllString(name, "_")
	// Trim leading and trailing underscores
	name = strings.Trim(name, "_")

	// If the name starts with a digit, prepend an underscore
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}

	return name
}
