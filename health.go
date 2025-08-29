package service

// Health represents the current health status of a service or component.
// It provides both machine-readable and human-readable information, making it suitable
// for monitoring, diagnostics, and external health checks.
type Health struct {
	// Status indicates the current health state of the service or component.
	// This should always be set to a value from the HealthStatus enumeration.
	Status HealthStatus
	// Reason provides a concise, human-readable explanation for the current status.
	// This field is intended for operators or users to quickly understand the cause of a non-healthy state.
	Reason string
	// Details contains additional structured information about the health status.
	// This can be any type (commonly a map or struct) that can be serialized to JSON for external reporting.
	// Use this field to provide context, metrics, or other diagnostic data.
	Details any
	// Error is an optional error value associated with the health status.
	// It should be set if the health status is due to an error condition, or nil otherwise.
	Error error
}
