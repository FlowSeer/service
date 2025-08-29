package service

//go:generate go tool golang.org/x/tools/cmd/stringer -type HealthStatus -trimprefix HealthStatus

// HealthStatus defines the set of possible health states for a service or component.
// These states are used to communicate the operational condition of a service in a
// standardized and interoperable way.
type HealthStatus int

const (
	// HealthStatusUnknown indicates that the health of the service cannot be determined.
	// This is the default value and should be used when no health check has been performed,
	// or if the service does not expose health information.
	HealthStatusUnknown HealthStatus = iota
	// HealthStatusHealthy indicates that the service is fully operational and functioning as expected.
	// All critical dependencies are available, and there are no known issues.
	HealthStatusHealthy
	// HealthStatusDegraded indicates that the service is running but experiencing reduced functionality,
	// partial outages, or other non-critical issues. The service may still be available,
	// but not at full capacity, performance, or reliability.
	HealthStatusDegraded
	// HealthStatusError indicates that the service is in a failed or non-operational state.
	// This typically means a critical problem has occurred that requires immediate attention,
	// and the service is unable to fulfill its primary responsibilities.
	HealthStatusError
	// HealthStatusShutdown indicates that the service has been intentionally shut down
	// and is no longer running, but not due to an error. This status is useful for distinguishing
	// between normal shutdowns and error-induced terminations.
	HealthStatusShutdown
)
