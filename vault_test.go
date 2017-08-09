package pvc

import (
	"testing"

	"github.com/dollarshaveclub/pvc/mocks"
	"github.com/golang/mock/gomock"
)

var testVaultBackend = vaultBackend{}

func TestNewVaultBackendGetter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mvc := mocks.NewMockvaultIO(ctrl)
	mvc.EXPECT().TokenAuth(gomock.Any()).Return(nil).Times(1)
	tvb := &vaultBackend{
		host:           "foo",
		authentication: Token,
	}
	_, err := newVaultBackendGetter(tvb, mvc)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
}

func TestVaultBackendGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mvc := mocks.NewMockvaultIO(ctrl)
	mvc.EXPECT().TokenAuth(gomock.Any()).Return(nil).Times(1)
	secretid := "1234"
	value := "foobar"
	mvc.EXPECT().GetStringValue(gomock.Any()).Return(value, nil).Times(1)
	tvb := &vaultBackend{
		host:           "foo",
		authentication: Token,
	}
	vbg, err := newVaultBackendGetter(tvb, mvc)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
	s, err := vbg.Get(secretid)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(s) != value {
		t.Fatalf("bad value: %v (wanted %v)", s, value)
	}
}
