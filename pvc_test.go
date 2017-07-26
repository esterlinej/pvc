package pvc

import "testing"

func TestNewSecretsClientVaultBackend(t *testing.T) {
	sc, err := NewSecretsClient(WithVaultBackend(), WithVaultAuthentication(Token), WithVaultToken("asdf"), WithVaultHost("https://foo.bar.com:8300"))
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

func TestNewSecretsClientEnvVarBackend(t *testing.T) {
	_, err := NewSecretsClient(WithEnvVarBackend())
	if err == nil {
		t.Fatalf("should have failed with not implemented")
	}
}

func TestNewSecretsClientJSONFileBackend(t *testing.T) {
	_, err := NewSecretsClient(WithJSONFileBackend())
	if err == nil {
		t.Fatalf("should have failed with not implemented")
	}
}
