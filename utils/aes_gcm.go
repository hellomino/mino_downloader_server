package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"hash"
	"io"
)

// --- PBKDF2 implementation (HMAC-SHA256) ---
// small PBKDF2 implementation so不需要额外依赖
func pbkdf2Key(password, salt []byte, iter, keyLen int) []byte {
	h := sha256.New
	hLen := h().Size()
	numBlocks := (keyLen + hLen - 1) / hLen
	var buf bytes.Buffer
	for block := 1; block <= numBlocks; block++ {
		// U1 = HMAC(password, salt || INT(block))
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(block))
		U := hmacSum(password, append(salt, b...), h)
		T := make([]byte, len(U))
		copy(T, U)
		for i := 1; i < iter; i++ {
			U = hmacSum(password, U, h)
			for j := range T {
				T[j] ^= U[j]
			}
		}
		buf.Write(T)
	}
	return buf.Bytes()[:keyLen]
}

func hmacSum(key, data []byte, newHash func() hash.Hash) []byte {
	mac := hmac.New(newHash, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// --- AES-GCM encrypt/decrypt with PBKDF2 key derivation ---

const (
	saltSize  = 16
	nonceSize = 12 // recommended for GCM
	keySize   = 32 // AES-256
	iterCount = 100000
)

// EncryptString 使用 passphrase 加密 plaintext，返回 base64(salt||nonce||ciphertext)
func EncryptString(passphrase string, plaintext []byte) (string, error) {
	// generate salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// derive key
	key := pbkdf2Key([]byte(passphrase), salt, iterCount, keySize)

	// create AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// output salt||nonce||ciphertext
	out := append(append(salt, nonce...), ciphertext...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptString 从 base64(salt||nonce||ciphertext) 解密，返回明文
func DecryptString(passphrase string, b64 string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	if len(raw) < saltSize+nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	salt := raw[:saltSize]
	nonce := raw[saltSize : saltSize+nonceSize]
	ciphertext := raw[saltSize+nonceSize:]

	key := pbkdf2Key([]byte(passphrase), salt, iterCount, keySize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
