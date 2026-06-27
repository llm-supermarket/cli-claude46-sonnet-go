package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/rfjakob/eme"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

const (
	fileMagic       = "RCLONE\x00\x00"
	fileMagicSize   = 8
	fileNonceSize   = 24
	fileHeaderSize  = fileMagicSize + fileNonceSize
	blockDataSize   = 64 * 1024
	blockHeaderSize = secretbox.Overhead
	blockSize       = blockDataSize + blockHeaderSize
	aesBlockSize    = aes.BlockSize
)

// defaultSalt is the salt rclone uses when no salt is specified.
var defaultSalt = []byte{
	0xA8, 0x0D, 0xF4, 0x3A, 0x8F, 0xBD, 0x03, 0x08,
	0xA7, 0xCA, 0xB8, 0x3E, 0x58, 0x1F, 0x86, 0xB1,
}

// Encoding controls how encrypted filenames are encoded.
type Encoding int

const (
	EncodingBase32 Encoding = iota // base32hex lowercase, no padding (rclone default)
	EncodingBase64                 // base64url, no padding
)

// ParseEncoding parses the encoding string, returning an error for unknown values.
func ParseEncoding(s string) (Encoding, error) {
	switch strings.ToLower(s) {
	case "base32", "":
		return EncodingBase32, nil
	case "base64":
		return EncodingBase64, nil
	default:
		return 0, fmt.Errorf("unknown encoding %q: must be base32 or base64", s)
	}
}

// Cipher holds the derived keys and AES block cipher for rclone crypt operations.
type Cipher struct {
	dataKey   [32]byte
	nameKey   [32]byte
	nameTweak [aesBlockSize]byte
	block     cipher.Block
	enc       Encoding
}

// New derives keys from password and salt and returns a Cipher.
// When salt is empty the rclone default salt is used.
func New(password, salt string, enc Encoding) (*Cipher, error) {
	c := &Cipher{enc: enc}

	saltBytes := defaultSalt
	if salt != "" {
		saltBytes = []byte(salt)
	}

	keySize := len(c.dataKey) + len(c.nameKey) + len(c.nameTweak) // 80 bytes

	var key []byte
	if password == "" {
		key = make([]byte, keySize)
	} else {
		var err error
		key, err = scrypt.Key([]byte(password), saltBytes, 16384, 8, 1, keySize)
		if err != nil {
			return nil, fmt.Errorf("key derivation: %w", err)
		}
	}

	copy(c.dataKey[:], key[0:32])
	copy(c.nameKey[:], key[32:64])
	copy(c.nameTweak[:], key[64:80])

	var err error
	c.block, err = aes.NewCipher(c.nameKey[:])
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	return c, nil
}

// EncryptFileName encrypts a filename segment and encodes it.
func (c *Cipher) EncryptFileName(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	padded := pkcs7Pad([]byte(plaintext), aesBlockSize)
	encrypted := eme.Transform(c.block, c.nameTweak[:], padded, eme.DirectionEncrypt)
	return c.encodeFilename(encrypted), nil
}

// DecryptFileName decodes and decrypts a filename segment.
func (c *Cipher) DecryptFileName(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	raw, err := c.decodeFilename(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decoding filename: %w", err)
	}
	if len(raw) == 0 || len(raw)%aesBlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext length %d", len(raw))
	}
	decrypted := eme.Transform(c.block, c.nameTweak[:], raw, eme.DirectionDecrypt)
	plain, err := pkcs7Unpad(decrypted, aesBlockSize)
	if err != nil {
		return "", fmt.Errorf("unpadding: %w", err)
	}
	return string(plain), nil
}

func (c *Cipher) encodeFilename(data []byte) string {
	switch c.enc {
	case EncodingBase64:
		return base64.RawURLEncoding.EncodeToString(data)
	default:
		s := base32.HexEncoding.EncodeToString(data)
		s = strings.TrimRight(s, "=")
		return strings.ToLower(s)
	}
}

func (c *Cipher) decodeFilename(name string) ([]byte, error) {
	switch c.enc {
	case EncodingBase64:
		return base64.RawURLEncoding.DecodeString(name)
	default:
		upper := strings.ToUpper(name)
		// Re-add stripped padding
		if rem := len(upper) % 8; rem != 0 {
			upper += strings.Repeat("=", 8-rem)
		}
		return base32.HexEncoding.DecodeString(upper)
	}
}

// EncryptContent reads from r, encrypts using rclone's format, and writes to w.
func (c *Cipher) EncryptContent(w io.Writer, r io.Reader) error {
	var nonce [fileNonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fmt.Errorf("generating nonce: %w", err)
	}
	if _, err := w.Write([]byte(fileMagic)); err != nil {
		return err
	}
	if _, err := w.Write(nonce[:]); err != nil {
		return err
	}

	buf := make([]byte, blockDataSize)
	for {
		n, readErr := io.ReadFull(r, buf)
		if n > 0 {
			encrypted := secretbox.Seal(nil, buf[:n], &nonce, &c.dataKey)
			if _, err := w.Write(encrypted); err != nil {
				return err
			}
			incrementNonce(&nonce)
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}

// DecryptContent reads from r, decrypts rclone-formatted ciphertext, and writes to w.
func (c *Cipher) DecryptContent(w io.Writer, r io.Reader) error {
	header := make([]byte, fileHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return fmt.Errorf("reading file header: %w", err)
	}
	if string(header[:fileMagicSize]) != fileMagic {
		return fmt.Errorf("not an rclone encrypted file (wrong magic bytes)")
	}
	var nonce [fileNonceSize]byte
	copy(nonce[:], header[fileMagicSize:])

	buf := make([]byte, blockSize)
	for {
		n, readErr := io.ReadFull(r, buf)
		if n > 0 {
			decrypted, ok := secretbox.Open(nil, buf[:n], &nonce, &c.dataKey)
			if !ok {
				return fmt.Errorf("decryption failed: incorrect password or corrupted data")
			}
			if _, err := w.Write(decrypted); err != nil {
				return err
			}
			incrementNonce(&nonce)
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}

func incrementNonce(nonce *[fileNonceSize]byte) {
	for i := range nonce {
		nonce[i]++
		if nonce[i] != 0 {
			break
		}
	}
}

func pkcs7Pad(data []byte, blockLen int) []byte {
	padLen := blockLen - (len(data) % blockLen)
	result := make([]byte, len(data)+padLen)
	copy(result, data)
	for i := len(data); i < len(result); i++ {
		result[i] = byte(padLen)
	}
	return result
}

func pkcs7Unpad(data []byte, blockLen int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockLen != 0 {
		return nil, fmt.Errorf("data length %d is not a multiple of block length %d", len(data), blockLen)
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockLen {
		return nil, fmt.Errorf("invalid padding value %d", padLen)
	}
	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return nil, fmt.Errorf("inconsistent padding bytes")
		}
	}
	return data[:len(data)-padLen], nil
}
