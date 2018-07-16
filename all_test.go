package pvc

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	var exit int
	defer func() { os.Exit(exit) }()
	if os.Getenv("VAULT_DEV_ALREADY_RUNNING") == "" {
		log.Printf("running vault server")
		if err := exec.Command("/bin/bash", "-c", "docker run --name vault -d -p 8200:8200 -e VAULT_USE_APP_ID=1 -v ${PWD}/testing/vault_app_ids.json:/opt/app-id.json -v ${PWD}/testing/vault_secrets.json:/opt/secrets.json -v ${PWD}/testing/vault_policies.json:/opt/policies.json quay.io/dollarshaveclub/vault-dev:master").Run(); err != nil {
			log.Fatalf("error running vault: %v", err)
		}
		defer func() {
			log.Printf("stopping vault server")
			exec.Command("/bin/bash", "-c", "docker stop vault && docker rm -v vault").Run()
		}()
	}
	exit = m.Run()
}
