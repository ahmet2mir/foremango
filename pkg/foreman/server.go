package foreman

import (
	"net/url"
)

// Server definition. The server represents the Foreman API endpoint that a
// Client directs all of its requests to.
type Server struct {
	// The URL of the API gateway
	URL url.URL
}
