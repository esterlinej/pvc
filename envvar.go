package pvc

import (
	"fmt"
	"os"
	"strings"
)

// Default mapping for this backend
const (
	DefaultEnvVarMapping = "SECRET_{{ .ID }}" // DefaultEnvVarMapping is uppercased after interpolation for convenience
)

type envVarBackendGetter struct {
	mapper SecretMapper
	config *envVarBackend
}

func newEnvVarBackendGetter(eb *envVarBackend) (*envVarBackendGetter, error) {
	if eb.mapping == "" {
		eb.mapping = DefaultEnvVarMapping
	}
	sm, err := newSecretMapper(eb.mapping)
	if err != nil {
		return nil, fmt.Errorf("error with mapping: %v", err)
	}
	return &envVarBackendGetter{
		mapper: sm,
		config: eb,
	}, nil
}

func (ebg *envVarBackendGetter) Get(id string) ([]byte, error) {
	vname, err := ebg.mapper.MapSecret(id)
	vname = strings.ToUpper(vname)
	if err != nil {
		return nil, fmt.Errorf("error mapping id to var name: %v", err)
	}
	secret, exists := os.LookupEnv(vname)
	if !exists {
		return nil, fmt.Errorf("secret not found: %v", vname)
	}
	return []byte(secret), nil
}
