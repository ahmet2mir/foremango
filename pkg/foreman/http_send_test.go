package foreman

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// failingDoer is an HTTPDoer that always fails, to exercise Send()'s
// transport-error branch - the dummy server used everywhere else in this
// package can only return successful HTTP responses, never a transport
// failure (DNS, connection refused, timeout, ...), so this is the only way
// to reach that branch without an actual broken network.
type failingDoer struct{ err error }

func (f failingDoer) Do(_ *http.Request) (*http.Response, error) {
	return nil, f.err
}

func TestClient_Send_TransportError(t *testing.T) {
	client := NewClient(Server{}, ClientCredentials{}, ClientConfig{
		HTTPClient: failingDoer{err: errors.New("connection refused")},
	})

	req, err := client.NewRequestWithContext(context.Background(), http.MethodGet, "/foo", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}

	statusCode, body, sendErr := client.Send(req)
	if sendErr == nil {
		t.Fatalf("expected a transport error, got nil")
	}
	if statusCode != -1 {
		t.Errorf("statusCode = %d, want -1", statusCode)
	}
	if len(body) != 0 {
		t.Errorf("body = %q, want empty", body)
	}
}
