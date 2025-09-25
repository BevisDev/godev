package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

// EncodeBase64 encodes the input string to a Base64-encoded string.
func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// DecodeBase64 decodes a Base64-encoded string into its original content.
func DecodeBase64(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncodeBase64Bytes encodes a byte slice to a Base64-encoded string.
func EncodeBase64Bytes(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64Bytes decodes a base64-encoded string and returns raw bytes.
// Returns an error if decoding fails.
func DecodeBase64Bytes(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// HexSha256 returns the SHA-256 hash of the input string, encoded as a hexadecimal string.
func HexSha256(str string) string {
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

// HexMd5 returns the MD5 hash of the input string, encoded as a hexadecimal string.
func HexMd5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// HmacSha256 computes the HMAC-SHA256 of a message using the provided secret key.
func HmacSha256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}

// EncryptAES encrypts plaintext using AES in CFB mode and returns base64-encoded ciphertext.
func EncryptAES(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts base64-encoded ciphertext using AES in CFB mode.
func DecryptAES(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return string(data), nil
}

// ReadPublicKey reads an RSA public key from a PEM-encoded file.
//
// The function expects the file to contain a PEM block in PKIX (SubjectPublicKeyInfo) format,
// such as:
//
//	-----BEGIN PUBLIC KEY-----
//	...base64 data...
//	-----END PUBLIC KEY-----
//
// Parameters:
//   - path: path to the PEM-encoded public key file.
//
// Returns:
//   - *rsa.PublicKey if parsing succeeds.
//   - An error if reading, decoding, or parsing fails.
//
// Example:
//
//	pubKey, err := ReadPublicKey("public.pem")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Public key modulus size:", pubKey.N.BitLen())
func ReadPublicKey(path string) (*rsa.PublicKey, error) {
	pubKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read public key file: %w", err)
	}

	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}

	return rsaPub, nil
}

// ReadPrivateKey reads an RSA private key from a PEM-encoded file.
//
// The function expects the file to contain a PEM block in PKCS#1 format,
// such as:
//
//	-----BEGIN RSA PRIVATE KEY-----
//	...base64 data...
//	-----END RSA PRIVATE KEY-----
//
// Parameters:
//   - path: path to the PEM-encoded private key file.
//
// Returns:
//   - *rsa.PrivateKey if parsing succeeds.
//   - An error if reading, decoding, or parsing fails.
//
// Example:
//
//	privKey, err := ReadPrivateKey("private.key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Private key modulus size:", privKey.N.BitLen())
func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// EncryptPKCS1v15 encrypts plaintext using RSA PKCS#1 v1.5 and returns base64-encoded ciphertext.
func EncryptPKCS1v15(pub *rsa.PublicKey, plainText string) (string, error) {
	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plainText))
	if err != nil {
		return "", fmt.Errorf("PKCS1 encryption failed: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// DecryptPKCS1v15 decrypts a base64-encoded ciphertext using RSA PKCS#1 v1.5.
func DecryptPKCS1v15(priv *rsa.PrivateKey, cipherTextB64 string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	plainBytes, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipherBytes)
	if err != nil {
		return "", fmt.Errorf("PKCS1 decryption failed: %w", err)
	}

	return string(plainBytes), nil
}

// EncryptOAEP encrypts plaintext using RSA-OAEP with SHA-256 and returns base64-encoded ciphertext.
func EncryptOAEP(pub *rsa.PublicKey, plainText string) (string, error) {
	hash := sha256.New()

	encryptedBytes, err := rsa.EncryptOAEP(hash, rand.Reader, pub, []byte(plainText), nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// DecryptOAEP decrypts a base64-encoded RSA-OAEP ciphertext using SHA-256.
func DecryptOAEP(priv *rsa.PrivateKey, cipherTextB64 string) (string, error) {
	hash := sha256.New()
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := rsa.DecryptOAEP(hash, rand.Reader, priv, cipherBytes, nil)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}
