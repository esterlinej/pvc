package main

import (
	"fmt"
	"os"

	"github.com/dollarshaveclub/pvc"
)

var vaultHost = os.Getenv("VAULT_ADDR")
var vaultToken = os.Getenv("VAULT_TOKEN")

func fatal(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

// populate the secrets env vars
func setEnvSecrets() {
	os.Setenv("SECRET_MYAPP_FOO", "bar")
	os.Setenv("SECRET_MYAPP_BIZ", "asdf")
}

func getEnvClient() *pvc.SecretsClient {
	sc, err := pvc.NewSecretsClient(pvc.WithEnvVarBackend(), pvc.WithMapping("SECRET_MYAPP_{{ .ID }}"))
	if err != nil {
		fatal("error getting client: %v", err)
	}
	return sc
}

// secrets.json
func getJSONClient() *pvc.SecretsClient {
	sc, err := pvc.NewSecretsClient(pvc.WithJSONFileBackend(), pvc.WithJSONFileLocation("secrets.json")) // default mapping is fine
	if err != nil {
		fatal("error getting client: %v", err)
	}
	return sc
}

func getVaultClient() *pvc.SecretsClient {
	sc, err := pvc.NewSecretsClient(pvc.WithVaultBackend(), pvc.WithVaultAuthentication(pvc.Token), pvc.WithVaultToken(vaultToken), pvc.WithVaultHost(vaultHost), pvc.WithMapping("secret/development/{{ .ID }}"))
	if err != nil {
		fatal("error getting client: %v", err)
	}
	return sc
}

// get secrets without caring which backend is being used
func getSecrets(sc *pvc.SecretsClient) {
	foo, err := sc.Get("foo")
	if err != nil {
		fatal("error getting secret")
	}
	fmt.Printf("secret named 'foo': %v\n", foo)
	bar, err := sc.Get("bar")
	if err != nil {
		fatal("error getting secret")
	}
	fmt.Printf("secret named 'bar': %v\n", bar)
}

func main() {
	setEnvSecrets()
	fmt.Println("Env client:")
	getSecrets(getEnvClient())
	fmt.Println("JSON client:")
	getSecrets(getJSONClient())
	fmt.Println("Vault client:")
	getSecrets(getVaultClient())
}
