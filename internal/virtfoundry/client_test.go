package virtfoundry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestClientLoginAndPingAuth(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/v1/auth/login":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"token":"jwt-test"}`))
		case "/api/v1/auth/me":
			if r.Header.Get("Authorization") != "Bearer jwt-test" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"username":"root"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(srv.URL, false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()
	if err := client.Health(ctx); err != nil {
		t.Fatalf("Health: %v", err)
	}
	if err := client.Login(ctx, "root", "secret"); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if err := client.PingAuth(ctx); err != nil {
		t.Fatalf("PingAuth: %v", err)
	}
}

func TestClientAPIKeyAuth(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/me" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer vfd_live_test" {
			http.Error(w, `{"error":"invalid api key"}`, http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"username":"automation"}`))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(srv.URL, false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("vfd_live_test")
	if err := client.PingAuth(context.Background()); err != nil {
		t.Fatalf("PingAuth: %v", err)
	}
}

func TestClientAPIErrorUsesJSONMessageOnly(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request","secret":"do-not-leak"}`))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(srv.URL, false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = client.Health(context.Background())
	if err == nil {
		t.Fatal("expected health error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bad request") {
		t.Fatalf("expected JSON error field in %q", msg)
	}
	if strings.Contains(msg, "do-not-leak") || strings.Contains(msg, `"secret"`) {
		t.Fatalf("raw body leaked into diagnostics: %q", msg)
	}
}

func TestClientAPIErrorOmitsNonJSONBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("stack-trace: sensitive"))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(srv.URL, false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = client.Health(context.Background())
	if err == nil {
		t.Fatal("expected health error")
	}
	if got := err.Error(); got != "API error: HTTP 500" {
		t.Fatalf("got %q, want status-only message", got)
	}
}
