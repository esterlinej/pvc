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

// 32 random bytes, base64-encoded
var binarydata = Base64Prefix + "jbsZSSkdDfAGtVR+9QjigOv7B8zjbCnF5GsQPKZIvzc="

func TestEnvVarBackendGetterGet(t *testing.T) {
	eb := &envVarBackend{
		mapping: "{{ .ID }}",
	}

	sid := "MY_SECRET"
	value := "foo"
	if err := os.Setenv(sid, value); err != nil {
		t.Fatalf("error setting env var: %v", err)
	}
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

	if err := os.Setenv(sid, binarydata); err != nil {
		t.Fatalf("error setting env var: %v", err)
	}
	s, err = evb.Get(sid)
	if err != nil {
		t.Fatalf("binary get failed: %v", err)
	}
	if len(s) != 32 {
		t.Fatalf("bad binary data length %v (wanted 32)", len(s))
	}
}

func TestEnvVarBackendGetterGetFilteredName(t *testing.T) {
	eb := &envVarBackend{
		mapping: "SECRET_{{ .ID }}",
	}

	sid := "foo/bar_value"
	envvar := "SECRET_FOO_BAR_VALUE"
	value := "foo"
	if err := os.Setenv(envvar, value); err != nil {
		t.Fatalf("error setting env var: %v", err)
	}
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
