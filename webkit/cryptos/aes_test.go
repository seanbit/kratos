package cryptos

import (
	"crypto/aes"
	"os"
	"strings"
	"testing"
)

func TestEncrypt2Decrypt(t *testing.T) {
	data := "hello world"
	validKey := "1234567890123456"
	enc, err := EncryptByKey(validKey, data)
	if err != nil {
		t.Error(err)
	}
	dec, err := DecryptByKey(validKey, enc)
	if err != nil {
		t.Error(err)
	}
	if string(dec) != string(data) {
		t.Error("Decrypt error")
	}
	t.Logf("Decrypt success")
}

func TestPkcs7Pad(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		wantLen   int
	}{
		{
			name:      "empty data",
			data:      []byte{},
			blockSize: aes.BlockSize,
			wantLen:   aes.BlockSize,
		},
		{
			name:      "data shorter than block",
			data:      []byte("hello"),
			blockSize: aes.BlockSize,
			wantLen:   aes.BlockSize,
		},
		{
			name:      "data exactly one block",
			data:      []byte("1234567890123456"),
			blockSize: aes.BlockSize,
			wantLen:   aes.BlockSize * 2,
		},
		{
			name:      "data longer than one block",
			data:      []byte("12345678901234567"),
			blockSize: aes.BlockSize,
			wantLen:   aes.BlockSize * 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pkcs7Pad(tt.data, tt.blockSize)
			if len(result) != tt.wantLen {
				t.Errorf("pkcs7Pad() len = %d, want %d", len(result), tt.wantLen)
			}
			if len(result)%tt.blockSize != 0 {
				t.Errorf("pkcs7Pad() result not multiple of block size")
			}
			// verify padding bytes
			padding := int(result[len(result)-1])
			for i := 0; i < padding; i++ {
				if result[len(result)-1-i] != byte(padding) {
					t.Errorf("pkcs7Pad() invalid padding byte at position %d", len(result)-1-i)
				}
			}
		})
	}
}

func TestPkcs7Unpad(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "valid padding 1",
			data:    []byte{0x01},
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "valid padding with data",
			data:    append([]byte("hello"), 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b),
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid padding size zero",
			data:    []byte{0x00},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid padding size too large",
			data:    []byte{0x05, 0x05},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "inconsistent padding bytes",
			data:    []byte{0x01, 0x02, 0x03, 0x03, 0x02}, // last byte says 2 padding, but second-to-last is 0x03
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkcs7Unpad(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Unpad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != string(tt.want) {
				t.Errorf("pkcs7Unpad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncryptByKey(t *testing.T) {
	validKey := "1234567890123456"

	tests := []struct {
		name    string
		key     string
		text    string
		wantErr bool
	}{
		{
			name:    "valid encryption",
			key:     validKey,
			text:    "hello world",
			wantErr: false,
		},
		{
			name:    "empty text",
			key:     validKey,
			text:    "",
			wantErr: false,
		},
		{
			name:    "key too short",
			key:     "short",
			text:    "hello",
			wantErr: true,
		},
		{
			name:    "key too long",
			key:     "12345678901234567890",
			text:    "hello",
			wantErr: true,
		},
		{
			name:    "long text",
			key:     validKey,
			text:    strings.Repeat("a", 1000),
			wantErr: false,
		},
		{
			name:    "unicode text",
			key:     validKey,
			text:    "你好世界🌍",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptByKey(tt.key, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptByKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("EncryptByKey() returned empty string for valid input")
			}
		})
	}
}

func TestDecryptByKey(t *testing.T) {
	validKey := "1234567890123456"

	tests := []struct {
		name       string
		key        string
		ciphertext string
		wantErr    bool
	}{
		{
			name:       "invalid base64",
			key:        validKey,
			ciphertext: "not-valid-base64!!!",
			wantErr:    true,
		},
		{
			name:       "ciphertext too short",
			key:        validKey,
			ciphertext: "YWJj",
			wantErr:    true,
		},
		{
			name:       "wrong key length",
			key:        "short",
			ciphertext: "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY3ODkw",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptByKey(tt.key, tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptByKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := "1234567890123456"

	tests := []struct {
		name string
		text string
	}{
		{"empty string", ""},
		{"short text", "hi"},
		{"exact block size", "1234567890123456"},
		{"longer than block", "12345678901234567"},
		{"multiple blocks", strings.Repeat("x", 100)},
		{"unicode", "你好世界 Hello 🌍"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"newlines", "line1\nline2\r\nline3"},
		{"whitespace", "  \t  \n  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptByKey(key, tt.text)
			if err != nil {
				t.Fatalf("EncryptByKey() error = %v", err)
			}

			decrypted, err := DecryptByKey(key, encrypted)
			if err != nil {
				t.Fatalf("DecryptByKey() error = %v", err)
			}

			if decrypted != tt.text {
				t.Errorf("Round trip failed: got %q, want %q", decrypted, tt.text)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	key := "1234567890123456"
	text := "same plaintext"

	encrypted1, err := EncryptByKey(key, text)
	if err != nil {
		t.Fatalf("First EncryptByKey() error = %v", err)
	}

	encrypted2, err := EncryptByKey(key, text)
	if err != nil {
		t.Fatalf("Second EncryptByKey() error = %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Error("EncryptByKey() should produce different ciphertext due to random IV")
	}

	// Both should decrypt to same plaintext
	decrypted1, _ := DecryptByKey(key, encrypted1)
	decrypted2, _ := DecryptByKey(key, encrypted2)

	if decrypted1 != text || decrypted2 != text {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}

func TestGetStringFromEnv(t *testing.T) {
	t.Run("existing env var", func(t *testing.T) {
		key := "TEST_AES_ENV_VAR"
		value := "test_value"
		os.Setenv(key, value)
		defer os.Unsetenv(key)

		got := GetStringFromEnv(key)
		if got != value {
			t.Errorf("GetStringFromEnv() = %v, want %v", got, value)
		}
	})

	t.Run("missing env var panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("GetStringFromEnv() should panic for missing env var")
			}
		}()
		GetStringFromEnv("NON_EXISTENT_ENV_VAR_12345")
	})
}

func BenchmarkEncryptByKey(b *testing.B) {
	key := "1234567890123456"
	text := strings.Repeat("benchmark text ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncryptByKey(key, text)
	}
}

func BenchmarkDecryptByKey(b *testing.B) {
	key := "1234567890123456"
	text := strings.Repeat("benchmark text ", 100)
	encrypted, _ := EncryptByKey(key, text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecryptByKey(key, encrypted)
	}
}
