package pvc

import (
	"strings"
	"testing"
)

func TestNewSecretsClientVaultBackend(t *testing.T) {
	sc, err := NewSecretsClient(WithVaultBackend(), WithVaultHost("foo"), WithVaultAuthentication(None))
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend == nil {
		t.Fatalf("backend is nil")
	}
	switch sc.backend.(type) {
	case *vaultBackendGetter:
		break
	default:
		t.Fatalf("wrong backend type: %T", sc.backend)
	}
}

func TestNewSecretsClientVaultBackendOptionOrdering(t *testing.T) {
	ops := []SecretsClientOption{
		WithVaultAuthentication(None),
		WithVaultHost("asdf"),
		WithMapping("{{ .ID }}"),
		WithVaultBackend(),
	}
	sc, err := NewSecretsClient(ops...)
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend == nil {
		t.Fatalf("backend is nil")
	}
	switch sc.backend.(type) {
	case *vaultBackendGetter:
		break
	default:
		t.Fatalf("wrong backend type: %T", sc.backend)
	}
}

func TestNewSecretsClientJSONFileBackendOptionOrdering(t *testing.T) {
	ops := []SecretsClientOption{
		WithMapping("{{ .ID }}"),
		WithJSONFileLocation("example/secrets.json"),
		WithJSONFileBackend(),
	}
	sc, err := NewSecretsClient(ops...)
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend == nil {
		t.Fatalf("backend is nil")
	}
	switch sc.backend.(type) {
	case *jsonFileBackendGetter:
		break
	default:
		t.Fatalf("wrong backend type: %T", sc.backend)
	}
}

func TestNewSecretsClientEnvVarBackendOptionOrdering(t *testing.T) {
	ops := []SecretsClientOption{
		WithMapping("{{ .ID }}"),
		WithEnvVarBackend(),
	}
	sc, err := NewSecretsClient(ops...)
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend == nil {
		t.Fatalf("backend is nil")
	}
	switch sc.backend.(type) {
	case *envVarBackendGetter:
		break
	default:
		t.Fatalf("wrong backend type: %T", sc.backend)
	}
}

func TestNewSecretsClientEnvVarBackend(t *testing.T) {
	sc, err := NewSecretsClient(WithEnvVarBackend())
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend == nil {
		t.Fatalf("backend is nil")
	}
	switch sc.backend.(type) {
	case *envVarBackendGetter:
		break
	default:
		t.Fatalf("wrong backend type: %T", sc.backend)
	}
}

func TestNewSecretsClientJSONFileBackend(t *testing.T) {
	sc, err := NewSecretsClient(WithJSONFileBackend(), WithJSONFileLocation("example/secrets.json"))
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend == nil {
		t.Fatalf("backend is nil")
	}
	switch sc.backend.(type) {
	case *jsonFileBackendGetter:
		break
	default:
		t.Fatalf("wrong backend type: %T", sc.backend)
	}
}

func TestWithMapping(t *testing.T) {
	mapping := "{{ .ID }}"
	sc, err := NewSecretsClient(WithJSONFileBackend(), WithJSONFileLocation("example/secrets.json"), WithMapping(mapping))
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend.(*jsonFileBackendGetter).config.mapping != mapping {
		t.Fatalf("mapping did not match: %v", sc.backend.(*jsonFileBackendGetter).config.mapping)
	}
}

func TestNewSecretsClientMultipleBackends(t *testing.T) {
	_, err := NewSecretsClient(WithVaultBackend(), WithEnvVarBackend(), WithJSONFileBackend())
	if err == nil {
		t.Fatalf("should have failed")
	}
	if !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("expected multiple backends error, received: %v", err)
	}
}

func TestNewSecretsClientNoBackends(t *testing.T) {
	_, err := NewSecretsClient()
	if err == nil {
		t.Fatalf("should have failed")
	}
	if !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("expected no backends error, received: %v", err)
	}
}

func TestNewSecretMapper(t *testing.T) {
	sc, err := newSecretMapper("foo/{{ .ID }}/bar")
	if err != nil {
		t.Fatalf("error getting secret mapper")
	}
	v, err := sc.MapSecret("asdf")
	if err != nil {
		t.Fatalf("error mapping: %v", err)
	}
	if v != "foo/asdf/bar" {
		t.Fatalf("incorrect value: %v", v)
	}
}

func TestNewSecretMapperMissingID(t *testing.T) {
	_, err := newSecretMapper("{{ .Foo }}")
	if err == nil {
		t.Fatalf("should have failed")
	}
}

func TestNewSecretMapperInvalidTemplate(t *testing.T) {
	_, err := newSecretMapper("{{ .%#$")
	if err == nil {
		t.Fatalf("should have failed")
	}
}
