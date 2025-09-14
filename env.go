package service

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/FlowSeer/fail"
)

// nameKey is an unexported type used as a context key for storing the service name.
type nameKey struct{}

// versionKey is an unexported type used as a context key for storing the service version.
type versionKey struct{}

// namespaceKey is an unexported type used as a context key for storing the service namespace.
type namespaceKey struct{}

// WithName returns a new context derived from ctx that carries the provided service name.
// The name can later be retrieved using the Name function.
func WithName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, nameKey{}, name)
}

// Name extracts the service name from the context, if present.
// If no name is set, it returns the empty string.
func Name(ctx context.Context) string {
	if name, ok := ctx.Value(nameKey{}).(string); ok {
		return name
	}
	return ""
}

// WithVersion returns a new context derived from ctx that carries the provided service version.
// The version can later be retrieved using the Version function.
func WithVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, versionKey{}, version)
}

// Version extracts the service version from the context, if present.
// If no version is set, it returns the empty string.
func Version(ctx context.Context) string {
	if version, ok := ctx.Value(versionKey{}).(string); ok {
		return version
	}
	return ""
}

// WithNamespace returns a new context derived from ctx that carries the provided service namespace.
// The namespace can later be retrieved using the Namespace function.
func WithNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, namespaceKey{}, namespace)
}

// Namespace extracts the service namespace from the context, if present.
// If no namespace is set, it returns the empty string.
func Namespace(ctx context.Context) string {
	if namespace, ok := ctx.Value(namespaceKey{}).(string); ok {
		return namespace
	}
	return ""
}

// EnvName constructs a normalized environment variable name by combining a prefix and a name.
// If the prefix is empty, it defaults to "SERVICE_".
// If the prefix does not end with an underscore, one is appended.
// The resulting string (prefix + name) is then normalized using NormalizeEnvName.
//
// Example usage:
//
//		EnvName("MYAPP", "LOG_LEVEL")   	// returns "MYAPP_LOG_LEVEL"
//		EnvName("", "CONFIG_PATH")      	// returns "SERVICE_CONFIG_PATH"
//	    EnvName("MYAPP_", "0config-dtest") 	// returns "MYAPP_CONFIG_TEST"
//	    EnvName("0APP", "config") 			// returns "_0APP_CONFIG"
func EnvName(prefix string, name string) string {
	if prefix == "" {
		prefix = "SERVICE_"
	} else {
		if !strings.HasSuffix(prefix, "_") {
			prefix += "_"
		}
	}

	return NormalizeEnvName(prefix + name)
}

// LookupEnv retrieves the value of the environment variable constructed from the given prefix and name.
// The environment variable name is normalized using EnvName(prefix, name).
// Returns the value and a boolean indicating whether the variable was present.
func LookupEnv(prefix string, name string) (string, bool) {
	return os.LookupEnv(EnvName(prefix, name))
}

// GetEnv returns the value of the environment variable constructed from the given prefix and name.
// The environment variable name is normalized using EnvName(prefix, name).
// If the variable is not present, GetEnv returns the empty string.
func GetEnv(prefix string, name string) string {
	return os.Getenv(EnvName(prefix, name))
}

// GetEnvOrDefault returns the value of the environment variable constructed from the given prefix and name,
// or the provided defaultValue if the variable is not set.
// The environment variable name is normalized using EnvName(prefix, name).
func GetEnvOrDefault(prefix string, name string, defaultValue string) string {
	if value, ok := LookupEnv(prefix, name); ok {
		return value
	}
	return defaultValue
}

// MustGetEnv retrieves the value of the environment variable constructed from the given prefix and name.
// The environment variable name is normalized using EnvName(prefix, name).
// If the variable is not set, MustGetEnv returns an error indicating that the environment variable must be set.
// This function is useful for required configuration values that must be present at runtime.
//
// Example usage:
//
//	val, err := MustGetEnv("MYAPP", "API_KEY")
//	if err != nil {
//	    return fail.WithContext(ctx, err)
//	}
func MustGetEnv(prefix string, name string) (string, error) {
	envName := EnvName(prefix, name)

	value, ok := LookupEnv(prefix, name)
	if !ok {
		return "", fail.New().
			Attribute("envName", envName).
			Msgf("environment variable must be set: %s", envName)
	}

	return value, nil
}

// NormalizeEnvName transforms an input string into a valid, conventional environment variable name.
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
func NormalizeEnvName(name string) string {
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
