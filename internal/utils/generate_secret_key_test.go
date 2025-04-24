package utils

import (
	"encoding/base64"
	"testing"
)

func TestGenerateRandomSecretKey_DecodeLength(t *testing.T) {
	key := GenerateRandomSecretKey()
	data, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		t.Fatalf("expected valid base64 URL-encoded string, got error: %v", err)
	}
	if len(data) != secretKeyLength {
		t.Errorf("expected decoded byte length %d, got %d", secretKeyLength, len(data))
	}
}

func TestGenerateRandomSecretKey_Unique(t *testing.T) {
	key1 := GenerateRandomSecretKey()
	key2 := GenerateRandomSecretKey()
	if key1 == key2 {
		t.Errorf("expected unique keys, but got identical values: %q", key1)
	}
}
