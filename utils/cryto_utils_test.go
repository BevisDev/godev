package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestEncryptDecryptAES(t *testing.T) {
	key := []byte("examplekey123456")
	plaintext := "This is a test message"

	ciphertext, err := EncryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}

	decrypted, err := DecryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptAES failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("DecryptAES = %q; want %q", decrypted, plaintext)
	}
}

func TestBase64EncodeDecode(t *testing.T) {
	original := "Test string 123!@#"

	encoded := EncodeBase64(original)
	decoded := DecodeBase64(encoded)

	if decoded != original {
		t.Errorf("DecodeBase64(EncodeBase64(%q)) = %q; want %q", original, decoded, original)
	}
}

func TestHexSha256(t *testing.T) {
	result := HexSha256("hello")
	expected := sha256.Sum256([]byte("hello"))
	if result != hex.EncodeToString(expected[:]) {
		t.Errorf("HexSha256 = %v; want %v", result, hex.EncodeToString(expected[:]))
	}
}

func TestHexMd5(t *testing.T) {
	result := HexMd5("hello")
	expected := md5.Sum([]byte("hello"))
	if result != hex.EncodeToString(expected[:]) {
		t.Errorf("HexMd5 = %v; want %v", result, hex.EncodeToString(expected[:]))
	}
}

func TestHmacSha256(t *testing.T) {
	message := "data"
	secret := "key"
	expected := hmac.New(sha256.New, []byte(secret))
	expected.Write([]byte(message))
	expectedHash := hex.EncodeToString(expected.Sum(nil))

	result := HmacSha256(message, secret)

	if result != expectedHash {
		t.Errorf("HmacSha256 = %v; want %v", result, expectedHash)
	}
}
