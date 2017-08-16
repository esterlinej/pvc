package pvc

import (
	"os"
	"testing"
)

func TestNewEnvVarBackendGetter(t *testing.T) {
	eb := &envVarBackend{}
	_, err := newEnvVarBackendGetter(eb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
}

func TestNewEnvVarBackendGetterBadMapping(t *testing.T) {
	eb := &envVarBackend{
		mapping: "{{a0301!!!",
	}
	_, err := newEnvVarBackendGetter(eb)
	if err == nil {
		t.Fatalf("should have failed with bad mapping")
	}
}

func TestEnvVarBackendGetterGet(t *testing.T) {
	eb := &envVarBackend{
		mapping: "{{ .ID }}",
	}

	sid := "MY_SECRET"
	value := "foo"
	os.Setenv(sid, value)
	defer os.Unsetenv(sid)

	evb, err := newEnvVarBackendGetter(eb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}

	s, err := evb.Get(sid)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(s) != value {
		t.Fatalf("bad value: %v (expected %v)", string(s), value)
	}
}

func TestEnvVarBackendGetterGetFilteredName(t *testing.T) {
	eb := &envVarBackend{
		mapping: "SECRET_{{ .ID }}",
	}

	sid := "foo/bar_value"
	envvar := "SECRET_FOO_BAR_VALUE"
	value := "foo"
	os.Setenv(envvar, value)
	defer os.Unsetenv(envvar)

	evb, err := newEnvVarBackendGetter(eb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}

	s, err := evb.Get(sid)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(s) != value {
		t.Fatalf("bad value: %v (expected %v)", string(s), value)
	}
}
