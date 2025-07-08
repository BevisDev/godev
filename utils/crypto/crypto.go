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

// EncryptAES encrypts the given plaintext string using AES encryption in CFB mode,
// with the provided key. A random IV is generated for each encryption, and the
// result is returned as a base64-encoded string.
//
// Parameters:
//   - plaintext: the string to encrypt.
//   - key: the AES key (must be 16, 24, or 32 bytes long for AES-128, AES-192, or AES-256).
//
// Returns:
//   - The base64-encoded ciphertext (IV + encrypted data).
//   - An error if encryption fails.
//
// Example:
//
//	key := []byte("examplekey123456") // 16 bytes
//	encrypted, err := EncryptAES("Hello, world!", key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Encrypted:", encrypted)
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

// DecryptAES decrypts the given base64-encoded ciphertext string using AES encryption in CFB mode,
// with the provided key. The ciphertext must include the IV (initialization vector) prepended.
//
// Parameters:
//   - ciphertext: the base64-encoded string containing IV + encrypted data.
//   - key: the AES key (must be 16, 24, or 32 bytes long).
//
// Returns:
//   - The decrypted plaintext string.
//   - An error if decryption fails.
//
// Example:
//
//	key := []byte("examplekey123456") // 16 bytes
//	encrypted, _ := EncryptAES("Hello, world!", key)
//	plaintext, err := DecryptAES(encrypted, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Decrypted:", plaintext)
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
		return "", errors.New("ciphertext quá ngắn")
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
//   - filePath: path to the PEM-encoded public key file.
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
func ReadPublicKey(filePath string) (*rsa.PublicKey, error) {
	pubKeyBytes, err := os.ReadFile(filePath)
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

// EncodeBase64 encodes the input string to a Base64-encoded string.
//
// Example:
//
//	EncodeBase64("hello") -> "aGVsbG8="
func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// DecodeBase64 decodes a Base64-encoded string to its original content.
// If decoding fails, it returns an empty string.
//
// Example:
//
//	DecodeBase64("aGVsbG8=") -> "hello"
func DecodeBase64(str string) string {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return ""
	}
	return string(data)
}

// HexSha256 returns the SHA-256 hash of the input string, encoded as a hexadecimal string.
//
// Example:
//
//	HexSha256("hello") -> "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
func HexSha256(str string) string {
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

// HexMd5 returns the MD5 hash of the input string, encoded as a hexadecimal string.
//
// Example:
//
//	HexMd5("hello") -> "5d41402abc4b2a76b9719d911017c592"
func HexMd5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// HmacSha256 computes the HMAC-SHA256 of a message using the provided secret key.
// The result is encoded as a hexadecimal string.
//
// Example:
//
//	HmacSha256("my message", "my secret")
func HmacSha256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}
