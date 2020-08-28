package pvc

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"testing"
)

const (
	testSecretPath = "secret/development/test_value"
)

var testvb = vaultBackend{
	host:               os.Getenv("TEST_VAULT_ADDR"),
	token:              os.Getenv("VAULT_TEST_TOKEN"),
	authRetries:        3,
	authRetryDelaySecs: 1,
}

func testGetVaultClient(t *testing.T) vaultIO {
	vc, err := newVaultClient(&testvb)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}
	return vc
}

func TestVaultIntegrationTokenAuth(t *testing.T) {
	if testvb.host == "" {
		t.Skipf("TEST_VAULT_ADDR undefined, skipping")
		return
	}
	vc := testGetVaultClient(t)
	err := vc.TokenAuth(testvb.token)
	if err != nil {
		log.Fatalf("error authenticating: %v", err)
	}
}

func TestVaultIntegrationAppRoleAuth(t *testing.T) {
	if testvb.host == "" {
		t.Skipf("TEST_VAULT_ADDR undefined, skipping")
		return
	}
	vc := testGetVaultClient(t)
	err := vc.AppRoleAuth("")
	if err != nil {
		log.Fatalf("error authenticating: %v", err)
	}
}

func TestVaultIntegrationGetValue(t *testing.T) {
	if testvb.host == "" {
		t.Skipf("TEST_VAULT_ADDR undefined, skipping")
		return
	}
	vc := testGetVaultClient(t)
	err := vc.TokenAuth(testvb.token)
	if err != nil {
		t.Fatalf("error authenticating: %v", err)
	}
	s, err := vc.GetValue(testSecretPath)
	if err != nil {
		t.Fatalf("error getting value: %v", err)
	}
	if string(s) != "foo" {
		t.Fatalf("bad value: %v (expected 'foo')", string(s))
	}
}

// Write a random binary value to vault, then verify that when we read it back
// it's base64-encoded and equal to the input bytes
func TestVaultIntegrationGetBinaryValue(t *testing.T) {
	// This sets up the Vault client
	if testvb.host == "" {
		t.Skipf("TEST_VAULT_ADDR undefined, skipping")
		return
	}
	vc := testGetVaultClient(t)
	err := vc.TokenAuth(testvb.token)
	if err != nil {
		t.Fatalf("error authenticating: %v", err)
	}

	// write a random binary value

	path := "secret/binval"
	key := DefaultVaultValueKey

	bsrc := make([]byte, 32)
	if n, err := rand.Read(bsrc); err != nil || n != len(bsrc) {
		t.Fatalf("error reading random bytes: %v bytes read: %v", n, err)
	}

	_, err = vc.(*vaultClient).client.Logical().Write(path, map[string]interface{}{
		key: bsrc,
	})
	if err != nil {
		t.Fatalf("error writing binary value: %v", err)
	}

	// retrieve the value
	got, err := vc.GetValue(path)
	if err != nil {
		t.Fatalf("error getting value: %v", err)
	}

	// base64 decode
	s2, _ := base64.StdEncoding.DecodeString(string(got))

	if !bytes.Equal(s2, bsrc) {
		t.Fatalf("bad value: length %v (wanted 32), bytes not equal: %v", len(s2), string(s2))
	}
}
