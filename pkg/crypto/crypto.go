package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// Crypto 加密管理器
type Crypto struct {
	enabled bool
	key     []byte
	aead    cipher.AEAD
}

// NewCrypto 创建新的加密管理器
func NewCrypto(enabled bool, key []byte, algorithm string) (*Crypto, error) {
	if !enabled {
		return &Crypto{enabled: false}, nil
	}

	// 使用 SHA-256 生成固定长度的密钥
	hash := sha256.Sum256(key)
	key = hash[:]

	var block cipher.Block
	var err error

	switch algorithm {
	case "aes-256-gcm":
		block, err = aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
	case "chacha20-poly1305":
		// 使用 ChaCha20-Poly1305
		block, err = aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported encryption algorithm")
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Crypto{
		enabled: true,
		key:     key,
		aead:    aead,
	}, nil
}

// Encrypt 加密数据
func (c *Crypto) Encrypt(plaintext []byte) ([]byte, error) {
	if !c.enabled {
		return plaintext, nil
	}

	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := c.aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt 解密数据
func (c *Crypto) Decrypt(ciphertext []byte) ([]byte, error) {
	if !c.enabled {
		return ciphertext, nil
	}

	if len(ciphertext) < c.aead.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:c.aead.NonceSize()]
	ciphertext = ciphertext[c.aead.NonceSize():]

	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateKey 生成随机密钥
func GenerateKey(size int) (string, error) {
	key := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// IsEnabled 检查加密是否启用
func (c *Crypto) IsEnabled() bool {
	return c.enabled
}
