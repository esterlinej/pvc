package pvc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

// Default mapping for this backend
const (
	DefaultVaultMapping = "secret/{{ .ID }}"
)

// VaultAuthentication enumerates the supported Vault authentication methods
type VaultAuthentication int

// Various Vault authentication methods
const (
	None    VaultAuthentication = iota // No authentication at all
	AppID                              // AppID
	Token                              // Token authentication
	AppRole                            // AppRole
	K8s                                // Kubernetes
)

type vaultBackendGetter struct {
	vc     vaultIO
	mapper SecretMapper
	config *vaultBackend
}

func newVaultBackendGetter(vb *vaultBackend, vc vaultIO) (*vaultBackendGetter, error) {
	var err error
	if vb.host == "" {
		return nil, fmt.Errorf("Vault host is required")
	}
	switch vb.authentication {
	case None:
		break
	case Token:
		err = vc.TokenAuth(vb.token)
		if err != nil {
			return nil, fmt.Errorf("error authenticating with supplied token: %v", err)
		}
	case AppID:
		err = vc.AppIDAuth(vb.appid, vb.userid, vb.useridpath)
		if err != nil {
			return nil, fmt.Errorf("error performing AppID authentication: %v", err)
		}
	case AppRole:
		return nil, fmt.Errorf("AppRole authentication not implemented")
	case K8s:
		err = vc.K8sAuth(vb.k8sjwt, vb.roleid)
		if err != nil {
			return nil, fmt.Errorf("error performing Kubernetes authentication: %v", err)
		}
	default:
		return nil, fmt.Errorf("unknown authentication method: %v", vb.authentication)
	}
	if vb.mapping == "" {
		vb.mapping = DefaultVaultMapping
	}
	sm, err := newSecretMapper(vb.mapping)
	if err != nil {
		return nil, fmt.Errorf("error with mapping: %v", err)
	}
	return &vaultBackendGetter{
		vc:     vc,
		mapper: sm,
		config: vb,
	}, nil
}

func (vbg *vaultBackendGetter) Get(id string) ([]byte, error) {
	path, err := vbg.mapper.MapSecret(id)
	if err != nil {
		return nil, fmt.Errorf("error mapping id to path: %v", err)
	}
	v, err := vbg.vc.GetStringValue(path)
	if err != nil {
		return nil, fmt.Errorf("error reading value: %v", err)
	}
	return []byte(v), nil
}

// vaultIO describes an object capable of interacting with Vault
type vaultIO interface {
	TokenAuth(token string) error
	AppIDAuth(appid string, userid string, useridpath string) error
	AppRoleAuth(roleid string) error
	K8sAuth(jwt, roleid string) error
	GetStringValue(path string) (string, error)
	GetBase64Value(path string) ([]byte, error)
}

// vaultClient is the concrete implementation of vaultIO interacting with a real Vault server
type vaultClient struct {
	client *api.Client
	config *vaultBackend
	token  string
}

var _ vaultIO = &vaultClient{}

// newVaultClient returns a vaultClient object or error
func newVaultClient(config *vaultBackend) (*vaultClient, error) {
	vc := vaultClient{}
	c, err := api.NewClient(&api.Config{Address: config.host})
	vc.client = c
	vc.config = config
	return &vc, err
}

// tokenAuth sets the client token but doesn't check validity
func (c *vaultClient) TokenAuth(token string) error {
	c.token = token
	c.client.SetToken(token)
	ta := c.client.Auth().Token()
	var err error
	for i := 0; i <= int(c.config.authRetries); i++ {
		_, err = ta.LookupSelf()
		if err == nil {
			break
		}
		log.Printf("Token auth failed: %v, retrying (%v/%v)", err, i+1, c.config.authRetries)
		time.Sleep(time.Duration(c.config.authRetryDelaySecs) * time.Second)
	}
	if err != nil {
		return fmt.Errorf("error performing auth call to Vault (retries exceeded): %v", err)
	}
	return nil
}

func (c *vaultClient) getTokenAndConfirm(route string, payload interface{}) error {
	var resp *api.Response
	var err error
	for i := 0; i <= int(c.config.authRetries); i++ {
		req := c.client.NewRequest("POST", route)
		jerr := req.SetJSONBody(payload)
		if jerr != nil {
			return fmt.Errorf("error setting auth JSON body: %v", jerr)
		}
		resp, err = c.client.RawRequest(req)
		if err == nil {
			break
		}
		log.Printf("auth failed: %v, retrying (%v/%v)", err, i+1, c.config.authRetries)
		time.Sleep(time.Duration(c.config.authRetryDelaySecs) * time.Second)
	}
	if err != nil {
		return fmt.Errorf("error performing auth call to Vault (retries exceeded): %v", err)
	}

	var output interface{}
	jd := json.NewDecoder(resp.Body)
	err = jd.Decode(&output)
	if err != nil {
		return fmt.Errorf("error unmarshaling Vault auth response: %v", err)
	}
	body := output.(map[string]interface{})
	auth := body["auth"].(map[string]interface{})
	c.token = auth["client_token"].(string)
	return nil
}

// appIDAuth attempts to perform app-id authorization.
func (c *vaultClient) AppIDAuth(appid string, userid string, useridpath string) error {
	if userid == "" {
		uidb, err := ioutil.ReadFile(useridpath)
		if err != nil {
			return fmt.Errorf("error reading useridpath: %v: %v", useridpath, err)
		}
		userid = string(uidb)
	}
	bodystruct := struct {
		AppID  string `json:"app_id"`
		UserID string `json:"user_id"`
	}{
		AppID:  appid,
		UserID: string(userid),
	}
	return c.getTokenAndConfirm("/v1/auth/app-id/login", &bodystruct)
}

func (c *vaultClient) AppRoleAuth(roleid string) error {
	return nil
}

func (c *vaultClient) K8sAuth(jwt, roleid string) error {
	payload := struct {
		JWT  string `json:"jwt"`
		Role string `json:"role"`
	}{
		JWT:  jwt,
		Role: roleid,
	}
	return c.getTokenAndConfirm("/v1/auth/kubernetes/login", &payload)
}

// getValue retrieves value at path
func (c *vaultClient) getValue(path string) (interface{}, error) {
	c.client.SetToken(c.token)
	lc := c.client.Logical()
	s, err := lc.Read(path)
	if err != nil {
		return nil, fmt.Errorf("error reading secret from Vault: %v: %v", path, err)
	}
	if s == nil {
		return nil, fmt.Errorf("secret not found")
	}
	if _, ok := s.Data["value"]; !ok {
		return nil, fmt.Errorf("secret missing 'value' key")
	}
	return s.Data["value"], nil
}

// GetStringValue retrieves a value expected to be a string
func (c *vaultClient) GetStringValue(path string) (string, error) {
	val, err := c.getValue(path)
	if err != nil {
		return "", err
	}
	switch val := val.(type) {
	case string:
		return val, nil
	default:
		return "", fmt.Errorf("unexpected type for %v value: %T", path, val)
	}
}

// GetBase64Value retrieves and decodes a value expected to be base64-encoded binary
func (c *vaultClient) GetBase64Value(path string) ([]byte, error) {
	val, err := c.GetStringValue(path)
	if err != nil {
		return []byte{}, err
	}
	decoded, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return []byte{}, fmt.Errorf("vault path: %v: error decoding base64 value: %v", path, err)
	}
	return decoded, nil
}
