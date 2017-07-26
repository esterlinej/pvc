package pvc

import (
	"log"
	"os"
	"testing"
)

const (
	testSecretPath = "secret/development/test_value"
)

var testvb = vaultBackend{
	host:   os.Getenv("VAULT_ADDR"),
	appid:  os.Getenv("VAULT_TEST_APPID"),
	userid: os.Getenv("VAULT_TEST_USERID"),
	token:  os.Getenv("VAULT_TEST_TOKEN"),
}

func testGetVaultClient(t *testing.T) *vaultClient {
	vc, err := newVaultClient(&testvb)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}
	return vc
}

func TestVaultIntegrationAppIDAuth(t *testing.T) {
	vc := testGetVaultClient(t)
	err := vc.appIDAuth(testvb.appid, testvb.userid, "")
	if err != nil {
		log.Fatalf("error authenticating: %v", err)
	}
}

func TestVaultIntegrationTokenAuth(t *testing.T) {
	vc := testGetVaultClient(t)
	err := vc.tokenAuth(testvb.token)
	if err != nil {
		log.Fatalf("error authenticating: %v", err)
	}
}

func TestVaultIntegrationGetValue(t *testing.T) {
	vc := testGetVaultClient(t)
	err := vc.appIDAuth(testvb.appid, testvb.userid, "")
	if err != nil {
		t.Fatalf("error authenticating: %v", err)
	}
	_, err = vc.getValue(testSecretPath)
	if err != nil {
		t.Fatalf("error getting value: %v", err)
	}
}
