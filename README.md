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
