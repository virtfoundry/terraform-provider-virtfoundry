package virtfoundry

import (
	"errors"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	if IsNotFound(nil) {
		t.Fatal("nil error should not be not-found")
	}
	if !IsNotFound(errors.New("API error: HTTP 404: {\"error\":\"missing\"}")) {
		t.Fatal("HTTP 404 should be not-found")
	}
	if !IsNotFound(errors.New(`resource "abc" not found`)) {
		t.Fatal("not found message should match")
	}
	if IsNotFound(errors.New("permission denied")) {
		t.Fatal("other errors should not match")
	}
}

func TestFindByID(t *testing.T) {
	items := []Tenant{
		{ID: "a", Name: "one"},
		{ID: "b", Name: "two"},
	}
	got, err := findByID(items, "b", func(t Tenant) string { return t.ID })
	if err != nil || got.Name != "two" {
		t.Fatalf("findByID: got %+v err %v", got, err)
	}
	if _, err := findByID(items, "z", func(t Tenant) string { return t.ID }); err == nil {
		t.Fatal("expected error for missing id")
	}
}
