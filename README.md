# PVC
[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]
[![CircleCI](https://circleci.com/gh/dollarshaveclub/pvc.svg?style=svg)](https://circleci.com/gh/dollarshaveclub/pvc)

[godocs]: https://pkg.go.dev/github.com/dollarshaveclub/pvc

PVC (polyvinyl chloride) is a simple, generic secret retrieval library that supports
multiple backends.

PVC lets applications access secrets without caring too much about where they
happen to be stored. The use case is to allow secrets to come from local/insecure
backends during development and testing, and then from Vault in production without
significant code changes required.

## Backends

- [Vault KV Version 1](https://www.vaultproject.io/docs/secrets/kv)
- Environment variables
- JSON file

## Secret Values

PVC makes some assumptions about how your secrets are stored in the various backends:

- If using Vault, there must be exactly one key called "value" for any given secret path (this can be overridden with 
`WithVaultValueKey("foo")`). The data associated with the value key will be retrieved and returned literally to the 
client as a byte slice.
- If using JSON or environment variables, the value will be treated as a string unless Base64-encoded prefixed with `Base64Prefix`, in which case 
it will be decoded when retrieved. In both cases, the value is returned to the client as a byte slice. This allows binary secrets to be stored in
backends that only hold strings.

## Vault Authentication

PVC supports token, Kubernetes, AppID (deprecated) and AppRole authentication.

## Example

```go
package main

import (
	"fmt"
    "github.com/dollarshaveclub/pvc"
)

func main() {

// environment variable backend
sc, _ := pvc.NewSecretsClient(pvc.WithEnvVarBackend(), pvc.WithMapping("SECRET_MYAPP_{{ .ID }}"))
secret, _ := sc.Get("foo") // fetches the env var "SECRET_MYAPP_FOO"

// JSON file backend
sc, _ = pvc.NewSecretsClient(pvc.WithJSONFileBackend(),pvc.WithJSONFileLocation("secrets.json"))
secret, _ = sc.Get("foo") // fetches the value in secrets.json under the key "foo"

fmt.Printf("foo: %v\n", string(secret))

// Vault backend
sc, _ = pvc.NewSecretsClient(
    pvc.WithVaultBackend(), 
    pvc.WithVaultAuthentication(pvc.Token), 
    pvc.WithVaultToken("some token"), 
    pvc.WithVaultHost("http://vault.example.com:8200"), 
    pvc.WithMapping("secret/development/{{ .ID }}"))
secret, _ = sc.Get("foo") // fetches the value from Vault (using token auth) from path secret/development/foo

fmt.Printf("foo: %v\n", string(secret))

// Automatic struct filling
type Secrets struct {
    Username string `secret:"secret/username"`  // secret id: secret/username
    Password string `secret:"secret/password"`
    EncryptionKey []byte `secret:"secret/enc_key"` // fields can be strings or byte slices
}

secrets := Secrets{}

// Fill automatically fills the fields in the secrets struct that have "secret" tags
err := sc.Fill(&secrets)
if err != nil {
    panic(err)
}

fmt.Printf("my username is: %v\n", secrets.Username)
fmt.Printf("my password is: %v\n", secrets.Password)
fmt.Printf("my key length is %d\n", len(secrets.EncryptionKey))
}
```

See also `example/`
