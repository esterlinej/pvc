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

// VaultAuthentication enumerates the supported Vault authentication methods
type VaultAuthentication int

// Various Vault authentication methods
const (
	AppID   VaultAuthentication = iota // AppID
	Token                              // Token authentication
	AppRole                            // AppRole
)

type vaultBackendGetter struct {
	vc     *vaultClient
	config *vaultBackend
}

func newVaultBackendGetter(vb *vaultBackend) (*vaultBackendGetter, error) {
	vc, err := newVaultClient(vb)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %v", err)
	}
	switch vb.authentication {
	case Token:
		err = vc.tokenAuth(vb.token)
		if err != nil {
			return nil, fmt.Errorf("error authenticating with supplied token: %v", err)
		}
	case AppID:
		err = vc.appIDAuth(vb.appid, vb.userid, vb.useridpath)
		if err != nil {
			return nil, fmt.Errorf("error performing AppID authentication: %v", err)
		}
	case AppRole:
		return nil, fmt.Errorf("AppRole authentication not implemented")
	}
	return &vaultBackendGetter{
		vc:     vc,
		config: vb,
	}, nil
}

func (vbg *vaultBackendGetter) Get(secret *SecretDefinition) ([]byte, error) {
	if secret.VaultPath == "" {
		return nil, fmt.Errorf("VaultPath is empty")
	}
	v, err := vbg.vc.getStringValue(secret.VaultPath)
	if err != nil {
		return nil, fmt.Errorf("error reading value: %v", err)
	}
	return []byte(v), nil
}

type vaultClient struct {
	client *api.Client
	config *vaultBackend
	token  string
}

// newVaultClient returns a vaultClient object or error
func newVaultClient(config *vaultBackend) (*vaultClient, error) {
	vc := vaultClient{}
	c, err := api.NewClient(&api.Config{Address: config.host})
	vc.client = c
	vc.config = config
	return &vc, err
}

// tokenAuth sets the client token but doesn't check validity
func (c *vaultClient) tokenAuth(token string) error {
	c.token = token
	c.client.SetToken(token)
	ta := c.client.Auth().Token()
	var err error
	for i := 0; i < int(c.config.authRetries); i++ {
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

// appIDAuth attempts to perform app-id authorization.
func (c *vaultClient) appIDAuth(appid string, userid string, useridpath string) error {
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
	var resp *api.Response
	var err error
	for i := 0; i < int(c.config.authRetries); i++ {
		req := c.client.NewRequest("POST", "/v1/auth/app-id/login")
		jerr := req.SetJSONBody(bodystruct)
		if jerr != nil {
			return fmt.Errorf("error setting auth JSON body: %v", jerr)
		}
		resp, err = c.client.RawRequest(req)
		if err == nil {
			break
		}
		log.Printf("App-ID auth failed: %v, retrying (%v/%v)", err, i+1, c.config.authRetries)
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

// getStringValue retrieves a value expected to be a string
func (c *vaultClient) getStringValue(path string) (string, error) {
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

// getBase64Value retrieves and decodes a value expected to be base64-encoded binary
func (c *vaultClient) getBase64Value(path string) ([]byte, error) {
	val, err := c.getStringValue(path)
	if err != nil {
		return []byte{}, err
	}
	decoded, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return []byte{}, fmt.Errorf("vault path: %v: error decoding base64 value: %v", path, err)
	}
	return decoded, nil
}
