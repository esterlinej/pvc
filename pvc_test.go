package pvc

import (
	"strings"
	"testing"
)

type fakeVaultIO struct{}

func (fv *fakeVaultIO) TokenAuth(token string) error {
	return nil
}
func (fv *fakeVaultIO) AppRoleAuth(roleid string) error {
	return nil
}
func (fv *fakeVaultIO) K8sAuth(jwt, roleid string) error {
	return nil
}
func (fv *fakeVaultIO) GetValue(path string) ([]byte, error) {
	return nil, nil
}

func newFakeVaultClient(_ *vaultBackend) (vaultIO, error) {
	return &fakeVaultIO{}, nil
}

var _ vaultIO = &fakeVaultIO{}

func TestNewSecretsClientVaultBackend(t *testing.T) {
	getVaultClient = newFakeVaultClient
	defer func() { getVaultClient = newVaultClient }()
	sc, err := NewSecretsClient(
		WithVaultBackend(TokenVaultAuth, "foo"),
	)
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
	getVaultClient = newFakeVaultClient
	defer func() { getVaultClient = newVaultClient }()
	ops := []SecretsClientOption{
		WithVaultAuthRetries(2),
		WithMapping("{{ .ID }}"),
		WithVaultBackend(TokenVaultAuth, "foo"),
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
		WithJSONFileBackend("example/secrets.json"),
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
	sc, err := NewSecretsClient(
		WithJSONFileBackend("example/secrets.json"),
	)
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
	sc, err := NewSecretsClient(
		WithJSONFileBackend("example/secrets.json"),
		WithMapping(mapping))
	if err != nil {
		t.Fatalf("error getting SecretsClient: %v", err)
	}
	if sc.backend.(*jsonFileBackendGetter).config.mapping != mapping {
		t.Fatalf("mapping did not match: %v", sc.backend.(*jsonFileBackendGetter).config.mapping)
	}
}

func TestNewSecretsClientMultipleBackends(t *testing.T) {
	_, err := NewSecretsClient(
		WithVaultBackend(TokenVaultAuth, "foo"),
		WithEnvVarBackend(),
		WithJSONFileBackend("example/secrets.json"))
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

func TestNewSecretsClientInvalidBackend(t *testing.T) {
	badBackendType := func() SecretsClientOption {
		return func(s *secretsClientConfig) {
			s.betype = 9999 // invalid
			s.backendCount++
		}
	}
	_, err := NewSecretsClient(badBackendType())
	if err == nil {
		t.Fatalf("should have failed")
	}
	if !strings.Contains(err.Error(), "invalid or unknown backend type") {
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
