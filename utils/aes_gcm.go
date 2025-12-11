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
	"encoding/json"
	"errors"
	"hash"
	"io"
)

/* ---------------------
   PBKDF2(HMAC-SHA256)
   与 JS WebCrypto 完全一致
--------------------- */

func pbkdf2Key(password, salt []byte, iter, keyLen int) []byte {
	h := sha256.New
	hLen := h().Size()
	numBlocks := (keyLen + hLen - 1) / hLen

	var buf bytes.Buffer

	for block := 1; block <= numBlocks; block++ {
		// salt || INT(block)
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(block))

		// ⚠️ 必须复制 salt，避免 append 修改原 slice
		saltBlock := append(append([]byte(nil), salt...), b...)

		U := hmacSum(password, saltBlock, h)
		T := make([]byte, len(U))
		copy(T, U)

		for i := 1; i < iter; i++ {
			U = hmacSum(password, U, h)
			for j := 0; j < len(T); j++ {
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

/* ---------------------
   AES-GCM PARAMS
--------------------- */

const (
	saltSize  = 16
	nonceSize = 12
	keySize   = 32 // AES-256 (256bit)
	iterCount = 100000
)

/* ---------------------
   EncryptAny (保持原样)
--------------------- */

func EncryptAny(passphrase string, obj any) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return EncryptString(passphrase, data)
}

/* ---------------------
   EncryptString
   参数类型保持不动：plaintext []byte
   返回 string (base64)
--------------------- */

func EncryptString(passphrase string, plaintext []byte) (string, error) {
	// salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// PBKDF2 derive 32-byte key
	key := pbkdf2Key([]byte(passphrase), salt, iterCount, keySize)

	// AES-256-GCM
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

	// Seal(plaintext)
	ct := aesgcm.Seal(nil, nonce, plaintext, nil)

	// output = salt || nonce || ciphertext
	out := make([]byte, 0, len(salt)+len(nonce)+len(ct))
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ct...)

	return base64.StdEncoding.EncodeToString(out), nil
}

/* ---------------------
   DecryptString
   参数与返回类型完全保留
--------------------- */

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
	ct := raw[saltSize+nonceSize:]

	key := pbkdf2Key([]byte(passphrase), salt, iterCount, keySize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
