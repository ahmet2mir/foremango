package foreman

import "net/url"

// Server is the Foreman instance a Client talks to.
type Server struct {
	// URL is the server's base URL, such as https://foreman.example.com.
	// If Foreman is served under a sub-path, include it: it is kept as a
	// prefix of every API path.
	URL url.URL
}
