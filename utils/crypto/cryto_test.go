package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"os"
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

func TestDecryptAES_WrongKey(t *testing.T) {
	key := []byte("examplekey123456")
	plaintext := "secret"
	ciphertext, err := EncryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptAES: %v", err)
	}

	wrongKey := []byte("wrongkey1234567890")
	_, err = DecryptAES(ciphertext, wrongKey)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}

func TestBase64EncodeDecode(t *testing.T) {
	original := "Test string 123!@#"

	encoded := EncodeBase64(original)

	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatalf("DecodeBase64 returned error: %v", err)
	}

	if decoded != original {
		t.Errorf("DecodeBase64(EncodeBase64(%q)) = %q; want %q", original, decoded, original)
	}
}

func TestDecodeBase64_Invalid(t *testing.T) {
	_, err := DecodeBase64("!!!not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
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

// Mocks a temporary private key file for testing
func writeTempPrivateKeyFile(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	tmpFile, err := os.CreateTemp("", "test-private-*.pem")
	if err != nil {
		t.Fatalf("failed to create temp private key file: %v", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(pemData); err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}

	return tmpFile.Name()
}

func TestRSAEncryptionFlow(t *testing.T) {
	// 1. Generate key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	pubKey := &privKey.PublicKey

	// 2. Write private key to temp file
	privFilePath := writeTempPrivateKeyFile(t, privKey)
	defer os.Remove(privFilePath)

	// 3. Read private key back
	readKey, err := ReadPrivateKey(privFilePath)
	if err != nil {
		t.Fatalf("failed to read private key: %v", err)
	}

	// 4. Encrypt plaintext
	plaintext := "secure message"
	cipherText, err := EncryptOAEP(pubKey, plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// 5. Decrypt ciphertext
	decrypted, err := DecryptOAEP(readKey, cipherText)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// 6. Validate
	if decrypted != plaintext {
		t.Errorf("decryption mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecryptPKCS1v15(t *testing.T) {
	// Generate test RSA key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	pubKey := &privKey.PublicKey
	original := "This is a test message"

	// Encrypt
	encrypted, err := EncryptPKCS1v15(pubKey, original)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptPKCS1v15(privKey, encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != original {
		t.Errorf("Decrypted text mismatch. Got %q, want %q", decrypted, original)
	}
}
