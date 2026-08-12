package virtfoundry

import "testing"

func TestFindByIDOrName(t *testing.T) {
	items := []VMTemplate{
		{ID: "a", Name: "ubuntu-2204"},
		{ID: "b", Name: "fedora-39"},
	}
	got, err := findByIDOrName(items, "ubuntu-2204", func(t VMTemplate) string { return t.ID }, func(t VMTemplate) string { return t.Name })
	if err != nil || got.ID != "a" {
		t.Fatalf("by name: got %+v err %v", got, err)
	}
	got, err = findByIDOrName(items, "b", func(t VMTemplate) string { return t.ID }, func(t VMTemplate) string { return t.Name })
	if err != nil || got.Name != "fedora-39" {
		t.Fatalf("by id: got %+v err %v", got, err)
	}
	if _, err = findByIDOrName(items, "missing", func(t VMTemplate) string { return t.ID }, func(t VMTemplate) string { return t.Name }); err == nil {
		t.Fatal("expected error for missing template")
	}
}
