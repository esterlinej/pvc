package pvc

import (
	"encoding/json"
	"fmt"
	"os"
)

// Default mapping for this backend
const (
	DefaultJSONFileMapping = "{{ .ID }}"
)

type jsonFileBackendGetter struct {
	mapper   SecretMapper
	config   *jsonFileBackend
	contents map[string]string
}

func newjsonFileBackendGetter(jb *jsonFileBackend) (*jsonFileBackendGetter, error) {
	f, err := os.Open(jb.fileLocation)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()
	c := map[string]string{}
	d := json.NewDecoder(f)
	err = d.Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("error decoding file (must be a JSON object): %v", err)
	}
	if jb.mapping == "" {
		jb.mapping = DefaultJSONFileMapping
	}
	sm, err := newSecretMapper(jb.mapping)
	if err != nil {
		return nil, fmt.Errorf("error with mapping: %v", err)
	}
	return &jsonFileBackendGetter{
		mapper:   sm,
		config:   jb,
		contents: c,
	}, nil
}

func (jbg *jsonFileBackendGetter) Get(id string) ([]byte, error) {
	key, err := jbg.mapper.MapSecret(id)
	if err != nil {
		return nil, fmt.Errorf("error mapping id to object key: %v", err)
	}
	if val, ok := jbg.contents[key]; ok {
		return []byte(val), nil
	}
	return nil, fmt.Errorf("secret not found: %v", key)
}
