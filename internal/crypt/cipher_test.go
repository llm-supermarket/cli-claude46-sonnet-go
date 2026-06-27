package crypt

import (
	"bytes"
	"strings"
	"testing"
)

const (
	testPassword = "hunter2"
	testSalt     = "mysalt"
)

func newCipher(t *testing.T, password, salt string, enc Encoding) *Cipher {
	t.Helper()
	c, err := New(password, salt, enc)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// --- Filename encryption ---

func TestFilenameRoundTrip_Base32_NoSalt(t *testing.T) {
	c := newCipher(t, testPassword, "", EncodingBase32)
	plain := "my_document.txt"

	enc, err := c.EncryptFileName(plain)
	if err != nil {
		t.Fatalf("EncryptFileName: %v", err)
	}
	if enc == plain {
		t.Fatal("encrypted name equals plaintext")
	}

	dec, err := c.DecryptFileName(enc)
	if err != nil {
		t.Fatalf("DecryptFileName: %v", err)
	}
	if dec != plain {
		t.Fatalf("want %q, got %q", plain, dec)
	}
}

func TestFilenameRoundTrip_Base64_NoSalt(t *testing.T) {
	c := newCipher(t, testPassword, "", EncodingBase64)
	plain := "my_document.txt"

	enc, err := c.EncryptFileName(plain)
	if err != nil {
		t.Fatalf("EncryptFileName: %v", err)
	}

	dec, err := c.DecryptFileName(enc)
	if err != nil {
		t.Fatalf("DecryptFileName: %v", err)
	}
	if dec != plain {
		t.Fatalf("want %q, got %q", plain, dec)
	}
}

func TestFilenameRoundTrip_Base32_WithSalt(t *testing.T) {
	c := newCipher(t, testPassword, testSalt, EncodingBase32)
	plain := "secret_file.pdf"

	enc, err := c.EncryptFileName(plain)
	if err != nil {
		t.Fatalf("EncryptFileName: %v", err)
	}

	dec, err := c.DecryptFileName(enc)
	if err != nil {
		t.Fatalf("DecryptFileName: %v", err)
	}
	if dec != plain {
		t.Fatalf("want %q, got %q", plain, dec)
	}
}

func TestFilenameEncoding_Base32IsLowercaseHex(t *testing.T) {
	c := newCipher(t, testPassword, "", EncodingBase32)
	enc, err := c.EncryptFileName("test.txt")
	if err != nil {
		t.Fatalf("EncryptFileName: %v", err)
	}
	if enc != strings.ToLower(enc) {
		t.Fatalf("base32 output should be lowercase, got %q", enc)
	}
	// base32hex alphabet is 0-9 and a-v
	for _, ch := range enc {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'v')) {
			t.Fatalf("unexpected character %q in base32hex output %q", ch, enc)
		}
	}
}

func TestFilenameDifferentSaltsDifferentOutput(t *testing.T) {
	c1 := newCipher(t, testPassword, "", EncodingBase32)
	c2 := newCipher(t, testPassword, testSalt, EncodingBase32)
	plain := "test.txt"

	enc1, _ := c1.EncryptFileName(plain)
	enc2, _ := c2.EncryptFileName(plain)
	if enc1 == enc2 {
		t.Fatal("different salts should produce different encrypted filenames")
	}
}

func TestFilenameDecrypt_WrongPassword(t *testing.T) {
	c1 := newCipher(t, testPassword, "", EncodingBase32)
	c2 := newCipher(t, "wrongpassword", "", EncodingBase32)

	enc, _ := c1.EncryptFileName("test.txt")
	_, err := c2.DecryptFileName(enc)
	// May succeed (EME has no authentication) but the result should differ.
	// At minimum it should not error on the encoding step.
	if err != nil {
		t.Logf("DecryptFileName with wrong password returned error (acceptable): %v", err)
	}
}

// --- Content encryption ---

func TestContentRoundTrip_NoSalt(t *testing.T) {
	c := newCipher(t, testPassword, "", EncodingBase32)
	original := []byte("Hello, World! This is a test file with some content.")

	var encrypted bytes.Buffer
	if err := c.EncryptContent(&encrypted, bytes.NewReader(original)); err != nil {
		t.Fatalf("EncryptContent: %v", err)
	}

	var decrypted bytes.Buffer
	if err := c.DecryptContent(&decrypted, &encrypted); err != nil {
		t.Fatalf("DecryptContent: %v", err)
	}

	if !bytes.Equal(decrypted.Bytes(), original) {
		t.Fatalf("decrypted content mismatch\nwant: %q\ngot:  %q", original, decrypted.Bytes())
	}
}

func TestContentRoundTrip_WithSalt(t *testing.T) {
	c := newCipher(t, testPassword, testSalt, EncodingBase32)
	original := []byte("Content encrypted with a custom salt value.")

	var encrypted bytes.Buffer
	if err := c.EncryptContent(&encrypted, bytes.NewReader(original)); err != nil {
		t.Fatalf("EncryptContent: %v", err)
	}

	var decrypted bytes.Buffer
	if err := c.DecryptContent(&decrypted, &encrypted); err != nil {
		t.Fatalf("DecryptContent: %v", err)
	}

	if !bytes.Equal(decrypted.Bytes(), original) {
		t.Fatalf("decrypted content mismatch")
	}
}

func TestContentEncryptionProducesRcloneMagic(t *testing.T) {
	c := newCipher(t, testPassword, "", EncodingBase32)
	var buf bytes.Buffer
	if err := c.EncryptContent(&buf, strings.NewReader("test")); err != nil {
		t.Fatalf("EncryptContent: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte(fileMagic)) {
		t.Fatalf("encrypted output does not start with rclone magic")
	}
}

func TestContentDecrypt_WrongPassword(t *testing.T) {
	c1 := newCipher(t, testPassword, "", EncodingBase32)
	c2 := newCipher(t, "wrongpassword", "", EncodingBase32)

	var encrypted bytes.Buffer
	if err := c1.EncryptContent(&encrypted, strings.NewReader("secret")); err != nil {
		t.Fatalf("EncryptContent: %v", err)
	}

	var decrypted bytes.Buffer
	err := c2.DecryptContent(&decrypted, &encrypted)
	if err == nil {
		t.Fatal("expected decryption error with wrong password, got nil")
	}
}

func TestContentDecrypt_WrongMagic(t *testing.T) {
	c := newCipher(t, testPassword, "", EncodingBase32)
	bad := bytes.NewReader([]byte("NOTMAGIC" + strings.Repeat("x", 24)))
	var out bytes.Buffer
	if err := c.DecryptContent(&out, bad); err == nil {
		t.Fatal("expected error for non-rclone file")
	}
}

func TestContentRoundTrip_LargeFile(t *testing.T) {
	c := newCipher(t, testPassword, "", EncodingBase32)
	// 3 full blocks + partial block
	original := bytes.Repeat([]byte("abcdefghij"), 3*blockDataSize/10+100)

	var encrypted bytes.Buffer
	if err := c.EncryptContent(&encrypted, bytes.NewReader(original)); err != nil {
		t.Fatalf("EncryptContent: %v", err)
	}

	var decrypted bytes.Buffer
	if err := c.DecryptContent(&decrypted, &encrypted); err != nil {
		t.Fatalf("DecryptContent: %v", err)
	}

	if !bytes.Equal(decrypted.Bytes(), original) {
		t.Fatalf("large file content mismatch")
	}
}

// --- Known-good decryption against test vectors ---

// TestDecryptKnownBase32File verifies that the pre-existing test file
// kr9tu4e1da4u3nifdd99g9tf5o (encrypted with password "Testpassword1", base32)
// decrypts to a filename of "TEST_FILE.txt".
func TestDecryptKnownFilenameBase32(t *testing.T) {
	c := newCipher(t, "Testpassword1", "", EncodingBase32)
	got, err := c.DecryptFileName("kr9tu4e1da4u3nifdd99g9tf5o")
	if err != nil {
		t.Fatalf("DecryptFileName: %v", err)
	}
	if got != "TEST_FILE.txt" {
		t.Fatalf("want %q, got %q", "TEST_FILE.txt", got)
	}
}

// TestDecryptKnownBase64File verifies that the pre-existing test file
// Iyxcijgc9bp3o5Y0npW6xqUvwWNcc3MA4SadB0sR6cY (encrypted with password "Testpassword1", base64)
// decrypts to a filename of "TEST_FILE BASE64.txt".
func TestDecryptKnownFilenameBase64(t *testing.T) {
	c := newCipher(t, "Testpassword1", "", EncodingBase64)
	got, err := c.DecryptFileName("Iyxcijgc9bp3o5Y0npW6xqUvwWNcc3MA4SadB0sR6cY")
	if err != nil {
		t.Fatalf("DecryptFileName: %v", err)
	}
	if got != "TEST_FILE BASE64.txt" {
		t.Fatalf("want %q, got %q", "TEST_FILE BASE64.txt", got)
	}
}
