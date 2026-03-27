package storage

import "testing"

func TestSetAndGet(t *testing.T) {
	Reset()

	if err := Set("name", "gox"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	var name string
	if err := Get("name", &name); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if name != "gox" {
		t.Errorf("got %q, want %q", name, "gox")
	}
}

func TestGetMissing(t *testing.T) {
	Reset()

	var val string
	err := Get("nonexistent", &val)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestSetStruct(t *testing.T) {
	Reset()

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	u := User{Name: "Alice", Age: 30}
	if err := Set("user", u); err != nil {
		t.Fatalf("Set: %v", err)
	}

	var got User
	if err := Get("user", &got); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Alice" || got.Age != 30 {
		t.Errorf("got %+v, want {Alice 30}", got)
	}
}

func TestDelete(t *testing.T) {
	Reset()

	Set("key", "value")
	if err := Delete("key"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	var val string
	err := Get("key", &val)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestSetOverwrite(t *testing.T) {
	Reset()

	Set("counter", 1)
	Set("counter", 42)

	var val int
	if err := Get("counter", &val); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != 42 {
		t.Errorf("got %d, want 42", val)
	}
}
