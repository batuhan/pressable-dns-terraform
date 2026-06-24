package pressable

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientUsesClientCredentialsAndDecodesEnvelope(t *testing.T) {
	var sawBearer bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/token":
			if got := r.FormValue("grant_type"); got != "client_credentials" {
				t.Fatalf("grant_type = %q", got)
			}
			_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: "token", ExpiresIn: 3599})
		case "/v1/sites":
			sawBearer = r.Header.Get("Authorization") == "Bearer token"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Success",
				"data":    []map[string]any{{"id": 1, "name": "example"}},
				"errors":  nil,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, "", "client", "secret")
	if err != nil {
		t.Fatal(err)
	}
	envelope, status, err := client.Request(context.Background(), http.MethodGet, "/v1/sites", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("status = %d", status)
	}
	if !sawBearer {
		t.Fatal("request did not include bearer token")
	}
	if envelope.Message != "Success" {
		t.Fatalf("message = %q", envelope.Message)
	}
}
