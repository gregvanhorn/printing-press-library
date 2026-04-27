package client

import "strings"

// calComVersionMap maps Cal.com v2 API path prefixes to their required
// cal-api-version header values. Cal.com defaults to a v1-compatible
// fallback shape when the header is missing — bookings/list returns []
// even when bookings exist, /slots returns shape changes, etc. Source:
// per-endpoint `parameters` in cal-com-openapi-v2.json where
// cal-api-version is required.
//
// pp-note: This is a Cal.com-specific patch to compensate for the
// generator not yet emitting per-endpoint required headers from the
// OpenAPI spec. Logged for retro as a Printing Press bug.
var calComVersionMap = []struct {
	prefix  string
	version string
}{
	// Order matters: longer prefixes first
	{"/v2/bookings", "2024-08-13"},
	{"/v2/event-types", "2024-06-14"},
	{"/v2/schedules", "2024-06-11"},
	{"/v2/slots", "2024-09-04"},
	{"/v2/teams", "2024-08-13"},
	{"/v2/organizations", "2024-08-13"},
	// Fallback for other v2 endpoints — most accept default version
}

// requiredHeadersForPath returns headers that must be on every request
// to the given path, regardless of caller-supplied headerOverrides.
// Returns nil if no headers are required.
func requiredHeadersForPath(path string) map[string]string {
	for _, entry := range calComVersionMap {
		if strings.HasPrefix(path, entry.prefix) {
			return map[string]string{"cal-api-version": entry.version}
		}
	}
	return nil
}
