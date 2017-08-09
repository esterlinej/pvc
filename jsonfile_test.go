package pvc

import "testing"

func TestNewjsonFileBackendGetter(t *testing.T) {
	jb := &jsonFileBackend{
		fileLocation: "example/secrets.json",
	}
	_, err := newjsonFileBackendGetter(jb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
}

func TestJSONFileBackendGetterGet(t *testing.T) {
	jb := &jsonFileBackend{
		fileLocation: "example/secrets.json",
	}
	jbg, err := newjsonFileBackendGetter(jb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
	sid := "foo"
	value := "bar"
	s, err := jbg.Get(sid)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(s) != value {
		t.Fatalf("bad value: %v (expected %v)", string(s), value)
	}
}
