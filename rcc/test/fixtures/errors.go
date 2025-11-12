package fixtures

import (
	"errors"
)

// ErrorScenario represents a standardized error condition for testing
type ErrorScenario struct {
	Name         string
	ErrorCode    string
	HTTPStatus   int
	Description  string
	ChannelIndex int
	PowerLevel   int
	RadioID      string
}

// BusyError returns a BUSY error scenario for testing
func BusyError() ErrorScenario {
	return ErrorScenario{
		Name:         "Radio Busy",
		ErrorCode:    "BUSY",
		HTTPStatus:   503,
		Description:  "Radio is currently busy with another operation",
		ChannelIndex: 6,
		PowerLevel:   5,
		RadioID:      "silvus-001",
	}
}

// RangeError returns an INVALID_RANGE error scenario for testing
func RangeError() ErrorScenario {
	return ErrorScenario{
		Name:         "Invalid Range",
		ErrorCode:    "INVALID_RANGE",
		HTTPStatus:   400,
		Description:  "Parameter value is outside valid range",
		ChannelIndex: 999, // Invalid channel
		PowerLevel:   15,  // Invalid power level
		RadioID:      "silvus-001",
	}
}

// InternalError returns an INTERNAL error scenario for testing
func InternalError() ErrorScenario {
	return ErrorScenario{
		Name:         "Internal Error",
		ErrorCode:    "INTERNAL",
		HTTPStatus:   500,
		Description:  "Internal system error occurred",
		ChannelIndex: 6,
		PowerLevel:   5,
		RadioID:      "faulty-radio-001",
	}
}

// UnavailableError returns an UNAVAILABLE error scenario for testing
func UnavailableError() ErrorScenario {
	return ErrorScenario{
		Name:         "Service Unavailable",
		ErrorCode:    "UNAVAILABLE",
		HTTPStatus:   503,
		Description:  "Service is temporarily unavailable",
		ChannelIndex: 6,
		PowerLevel:   5,
		RadioID:      "offline-radio-001",
	}
}

// ErrorMapping returns the standard error code mapping per Architecture §8.5
func ErrorMapping() map[string]int {
	return map[string]int{
		"INVALID_RANGE": 400,
		"BUSY":          503,
		"UNAVAILABLE":   503,
		"INTERNAL":      500,
	}
}

// CreateError creates a standardized error with the given code
func CreateError(code string) error {
	scenarios := map[string]ErrorScenario{
		"BUSY":          BusyError(),
		"INVALID_RANGE": RangeError(),
		"INTERNAL":      InternalError(),
		"UNAVAILABLE":   UnavailableError(),
	}

	scenario, exists := scenarios[code]
	if !exists {
		return errors.New("unknown error code: " + code)
	}

	return errors.New(scenario.Description)
}

// TimeoutError returns a timeout error scenario for testing
func TimeoutError() ErrorScenario {
	return ErrorScenario{
		Name:         "Operation Timeout",
		ErrorCode:    "BUSY",
		HTTPStatus:   503,
		Description:  "Operation timed out",
		ChannelIndex: 6,
		PowerLevel:   5,
		RadioID:      "slow-radio-001",
	}
}

// ValidationError returns a validation error scenario for testing
func ValidationError() ErrorScenario {
	return ErrorScenario{
		Name:         "Validation Failed",
		ErrorCode:    "INVALID_RANGE",
		HTTPStatus:   400,
		Description:  "Request validation failed",
		ChannelIndex: -1, // Invalid negative index
		PowerLevel:   0,  // Invalid zero power
		RadioID:      "", // Empty radio ID
	}
}

// ConcurrencyError returns a concurrency error scenario for testing
func ConcurrencyError() ErrorScenario {
	return ErrorScenario{
		Name:         "Concurrency Conflict",
		ErrorCode:    "BUSY",
		HTTPStatus:   503,
		Description:  "Concurrent operation conflict",
		ChannelIndex: 6,
		PowerLevel:   5,
		RadioID:      "concurrent-radio-001",
	}
}

// ValidToken returns a valid JWT token for testing
func ValidToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJyb2xlcyI6WyJjb250cm9sbGVyIl0sInNjb3BlcyI6WyJyZWFkIiwic3RhdGUiLCJjb250cm9sIl0sImV4cCI6OTk5OTk5OTk5OX0.test-signature"
}

// ExpiredToken returns an expired JWT token for testing
func ExpiredToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJyb2xlcyI6WyJjb250cm9sbGVyIl0sInNjb3BlcyI6WyJyZWFkIiwic3RhdGUiLCJjb250cm9sIl0sImV4cCI6MTYwOTQ1NTIwMH0.expired-signature"
}

// InvalidToken returns an invalid JWT token for testing
func InvalidToken() string {
	return "invalid.token.here"
}

// AdminToken returns an admin JWT token for testing
func AdminToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZG1pbiIsInJvbGVzIjpbImFkbWluIl0sInNjb3BlcyI6WyJyZWFkIiwic3RhdGUiLCJjb250cm9sIiwic3RhdGUiLCJjb250cm9sIl0sImV4cCI6OTk5OTk5OTk5OX0.admin-signature"
}

// UserToken returns a user JWT token for testing
func UserToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyIiwicm9sZXMiOlsiY29udHJvbGxlciJdLCJzY29wZXMiOlsicmVhZCIsInN0YXRlIiwiY29udHJvbCJdLCJleHAiOjk5OTk5OTk5OTl9.user-signature"
}

// ReadOnlyToken returns a read-only JWT token for testing
func ReadOnlyToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ2aWV3ZXIiLCJyb2xlcyI6WyJ2aWV3ZXIiXSwic2NvcGVzIjpbInJlYWQiLCJzdGF0ZSJdLCJleHAiOjk5OTk5OTk5OTl9.readonly-signature"
}
