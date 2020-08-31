package pvc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testingroot() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "testing")
}

func TestFileTreeBackendGetter(t *testing.T) {
	tb := &fileTreeBackend{
		rootPath: testingroot(),
	}
	_, err := newFileTreeBackendGetter(tb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
}

func TestFileTreeBackendGetterGet(t *testing.T) {
	expectedValue := "DrFeelgood"
	tb := &fileTreeBackend{
		rootPath: testingroot(),
	}
	tbg, err := newFileTreeBackendGetter(tb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
	sid := "username"
	s, err := tbg.Get(sid)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(s) != expectedValue {
		t.Fatalf("bad value: %v (expected %v)", string(s), expectedValue)
	}
	_, err = tbg.Get("invalid-path")
	if err == nil {
		t.Fatalf("expected a file not found error")
	}
}

func TestFileTreeBackendGetter_Get_FileTooLarge(t *testing.T) {
	tb := &fileTreeBackend{
		rootPath: testingroot(),
	}
	tbg, err := newFileTreeBackendGetter(tb)
	if err != nil {
		t.Fatalf("should have succeeded: %v", err)
	}
	sid := "username"
	oldmax := MaxFileTreeFileSizeBytes
	defer func() { MaxFileTreeFileSizeBytes = oldmax }()
	MaxFileTreeFileSizeBytes = 2
	_, err = tbg.Get(sid)
	if err == nil {
		t.Fatalf("should have returned an error")
	}
	if !strings.Contains(err.Error(), "file too large") {
		t.Fatalf("unexpected error: %v", err)
	}
}
