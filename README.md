# PVC
[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]
[![CircleCI](https://circleci.com/gh/dollarshaveclub/pvc.svg?style=svg)](https://circleci.com/gh/dollarshaveclub/pvc)

[godocs]: https://godoc.org/github.com/dollarshaveclub/pvc

PVC (polyvinyl chloride) is a simple, generic secret retrieval library that supports
multiple backends.

PVC lets applications access secrets without caring too much about where they
happen to be stored. The use case is to allow secrets to come from local/insecure
backends during development and testing, and then from Vault in production without
significant code changes required.

## Backends

- Vault
- Environment variables
- JSON file

## Vault Authentication

PVC supports token, AppID and AppRole authentication.

## Example

```go
// environment variable backend
sc, _ := pvc.NewSecretsClient(pvc.WithEnvVarBackend(), pvc.WithMapping("SECRET_MYAPP_{{ .ID }}"))
secret, _ := sc.Get("foo")

// JSON file backend
sc, _ := pvc.NewSecretsClient(pvc.WithJSONFileBackend(),pvc.WithJSONFileLocation("secrets.json"))
secret, _ := sc.Get("foo")

// Vault backend
sc, _ := pvc.NewSecretsClient(pvc.WithVaultBackend(), pvc.WithVaultAuthentication(pvc.Token), pvc.WithVaultToken(vaultToken), pvc.WithVaultHost(vaultHost), pvc.WithMapping("secret/development/{{ .ID }}"))
secret, _ := sc.Get("foo")
```

See also `example/`
