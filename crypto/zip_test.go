package crypto

import (
	"bytes"
	"io"
	"testing"
)

func TestZip(t *testing.T) {
	// Test case 1: Basic encryption and decryption
	passphrase := []byte("password")
	zc := NewZipCrypto(passphrase)
	zc2 := NewZipCrypto(passphrase)
	originalData := []byte("Hello, World!")
	encryptedData := zc.Encrypt(originalData)
	decryptedData := zc2.Decrypt(encryptedData)

	if string(originalData) != string(decryptedData) {
		t.Errorf("Decrypted data does not match original data. Got: %s, Expected: %s", decryptedData, originalData)
	}

	// Test case 2: Encryption with different passphrase
	passphrase2 := []byte("different_password")
	zipCrypto2 := NewZipCrypto(passphrase2)

	encryptedData2 := zipCrypto2.Encrypt(originalData)
	decryptedData2 := zc.Decrypt(encryptedData2)

	if string(originalData) == string(decryptedData2) {
		t.Errorf("Decrypted data with wrong passphrase should not match original data. Got: %s, Expected: %s", decryptedData2, originalData)
	}
}

func TestZipCryptoStream(t *testing.T) {
	passphrase := []byte("stream_password")
	plain := []byte("stream hello world, zip crypto test!")

	// 加密
	src := bytes.NewReader(plain)
	encBuf := &bytes.Buffer{}
	zcEnc := NewZipCrypto(passphrase)
	zw := NewZipCryptoWriter(encBuf, zcEnc)
	_, err := io.Copy(zw, src)
	if err != nil {
		t.Fatalf("Encrypt stream failed: %v", err)
	}

	// 解密
	zcDec := NewZipCrypto(passphrase)
	decBuf := &bytes.Buffer{}
	zr := NewZipCryptoReader(bytes.NewReader(encBuf.Bytes()), zcDec)
	_, err = io.Copy(decBuf, zr)
	if err != nil {
		t.Fatalf("Decrypt stream failed: %v", err)
	}

	if !bytes.Equal(plain, decBuf.Bytes()) {
		t.Errorf("Stream decrypted data mismatch. Got: %s, Want: %s", decBuf.Bytes(), plain)
	}

	// 错误密码解密
	wrongZc := NewZipCrypto([]byte("wrong_password"))
	wrongDecBuf := &bytes.Buffer{}
	wrongZr := NewZipCryptoReader(bytes.NewReader(encBuf.Bytes()), wrongZc)
	_, err = io.Copy(wrongDecBuf, wrongZr)
	if err != nil {
		t.Fatalf("Decrypt with wrong password failed: %v", err)
	}
	if bytes.Equal(plain, wrongDecBuf.Bytes()) {
		t.Errorf("Stream decrypted data with wrong password should not match original")
	}
}

func TestZipCryptoWriterReadFrom(t *testing.T) {
	passphrase := []byte("readfrom_password")
	plain := []byte("this is a test for ReadFrom interface!")

	// 加密
	encBuf := &bytes.Buffer{}
	zcEnc := NewZipCrypto(passphrase)
	zw := NewZipCryptoWriter(encBuf, zcEnc)
	n, err := zw.ReadFrom(bytes.NewReader(plain))
	if err != nil {
		t.Fatalf("ReadFrom encrypt failed: %v", err)
	}
	if n != int64(len(plain)) {
		t.Errorf("ReadFrom encrypt length mismatch: got %d, want %d", n, len(plain))
	}

	// 解密
	zcDec := NewZipCrypto(passphrase)
	decBuf := &bytes.Buffer{}
	zr := NewZipCryptoReader(bytes.NewReader(encBuf.Bytes()), zcDec)
	n2, err := io.Copy(decBuf, zr)
	if err != nil {
		t.Fatalf("Decrypt stream failed: %v", err)
	}
	if n2 != int64(len(plain)) {
		t.Errorf("Decrypt length mismatch: got %d, want %d", n2, len(plain))
	}
	if !bytes.Equal(plain, decBuf.Bytes()) {
		t.Errorf("Decrypted data mismatch. Got: %s, Want: %s", decBuf.Bytes(), plain)
	}
}

func TestZipCryptoReaderWriteTo(t *testing.T) {
	passphrase := []byte("writeto_password")
	plain := []byte("this is a test for WriteTo interface!")

	// 加密
	encBuf := &bytes.Buffer{}
	zcEnc := NewZipCrypto(passphrase)
	zw := NewZipCryptoWriter(encBuf, zcEnc)
	_, err := io.Copy(zw, bytes.NewReader(plain))
	if err != nil {
		t.Fatalf("Encrypt stream failed: %v", err)
	}

	// 解密
	zcDec := NewZipCrypto(passphrase)
	zr := NewZipCryptoReader(bytes.NewReader(encBuf.Bytes()), zcDec)
	decBuf := &bytes.Buffer{}
	n, err := zr.WriteTo(decBuf)
	if err != nil {
		t.Fatalf("WriteTo decrypt failed: %v", err)
	}
	if n != int64(len(plain)) {
		t.Errorf("WriteTo decrypt length mismatch: got %d, want %d", n, len(plain))
	}
	if !bytes.Equal(plain, decBuf.Bytes()) {
		t.Errorf("WriteTo decrypted data mismatch. Got: %s, Want: %s", decBuf.Bytes(), plain)
	}
}
