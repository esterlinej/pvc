package pvc

import (
	"bytes"
	"fmt"
	"testing"
)

type fakeBackend struct {
	GetFunc func(id string) ([]byte, error)
}

func (fb *fakeBackend) Get(id string) ([]byte, error) {
	if fb.GetFunc != nil {
		return fb.GetFunc(id)
	}
	return []byte{}, nil
}

var _ secretBackend = &fakeBackend{}

type secrets struct {
	Name       string `secret:"name"`
	Key        []byte `secret:"key"`
	IgnoredKey string `secret:"-"`
	Unused     string
}

type badsecrets struct {
	Age uint `secret:"age"`
}

type badsecrets2 struct {
	RandomThing []rune `secret:"random"`
}

var somestr = "asdf"
var nilptr *int = nil
var ifacewithnil interface{} = nilptr

func TestSecretsClient_Fill(t *testing.T) {
	tests := []struct {
		name      string
		backend   secretBackend
		s         interface{}
		wantErr   bool
		validatef func(s interface{}) error
	}{
		{
			name: "success",
			backend: &fakeBackend{
				GetFunc: func(id string) ([]byte, error) {
					switch id {
					case "name":
						return []byte("Frank"), nil
					case "key":
						return []byte("asdf"), nil
					default:
						return nil, fmt.Errorf("unknown id: %v", id)
					}
				},
			},
			s: &secrets{},
			validatef: func(s interface{}) error {
				v, ok := s.(*secrets)
				if !ok {
					return fmt.Errorf("bad type: %T", s)
				}
				if v == nil {
					return fmt.Errorf("s is nil")
				}
				if v.Name != "Frank" {
					return fmt.Errorf("bad name: %v", v.Name)
				}
				if !bytes.Equal(v.Key, []byte("asdf")) {
					return fmt.Errorf("bad key: %v", string(v.Key))
				}
				if v.IgnoredKey != "" {
					return fmt.Errorf("ignored field should be an empty string: %v", v.IgnoredKey)
				}
				if v.Unused != "" {
					return fmt.Errorf("unused should be an empty string: %v", v.Unused)
				}
				return nil
			},
		},
		{
			name: "bad type",
			backend: &fakeBackend{
				GetFunc: func(id string) ([]byte, error) {
					return []byte("anything"), nil
				},
			},
			s:       &badsecrets{},
			wantErr: true,
		},
		{
			name: "bad slice type",
			backend: &fakeBackend{
				GetFunc: func(id string) ([]byte, error) {
					return []byte("anything"), nil
				},
			},
			s:       &badsecrets2{},
			wantErr: true,
		},
		{
			name: "not a struct",
			backend: &fakeBackend{
				GetFunc: func(id string) ([]byte, error) {
					return []byte("anything"), nil
				},
			},
			s:       &somestr,
			wantErr: true,
		},
		{
			name: "not a pointer",
			backend: &fakeBackend{
				GetFunc: func(id string) ([]byte, error) {
					return []byte("anything"), nil
				},
			},
			s:       somestr,
			wantErr: true,
		},
		{
			name: "nil",
			backend: &fakeBackend{
				GetFunc: func(id string) ([]byte, error) {
					return []byte("anything"), nil
				},
			},
			wantErr: true,
		},
		{
			name: "interface with nil",
			backend: &fakeBackend{
				GetFunc: func(id string) ([]byte, error) {
					return []byte("anything"), nil
				},
			},
			s:       ifacewithnil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &SecretsClient{
				backend: tt.backend,
			}
			err := sc.Fill(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fill() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				t.Logf("err: %v", err)
				return
			}
			if tt.validatef != nil {
				if err := tt.validatef(tt.s); err != nil {
					t.Errorf("validate failed: %v", err)
				}
			}
		})
	}
}
