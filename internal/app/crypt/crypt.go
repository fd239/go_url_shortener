package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"github.com/fd239/go_url_shortener/config"
	"log"
)

var CryptURL *CipherCrypt

// CipherCrypt struct for Decrypting
type CipherCrypt struct {
	nonce  []byte
	aesGCM cipher.AEAD
}

func Decrypt(userID string) (string, error) {
	err := initCrypt()
	if err != nil {
		return "", err
	}
	b, err := hex.DecodeString(userID)
	if err != nil {
		log.Printf("Decrypt decode string error: %v", err)
		return "", err
	}

	decrypted, terr := CryptURL.aesGCM.Open(nil, CryptURL.nonce, b, nil)

	if terr != nil {
		log.Printf("aesGCM open error: %v", terr)
	}

	return string(decrypted), err

}

func Encrypt(userID string) (string, error) {
	err := initCrypt()
	if err != nil {
		return "", err
	}

	encrypted := CryptURL.aesGCM.Seal(nil, CryptURL.nonce, []byte(userID), nil)
	return hex.EncodeToString(encrypted), nil

}

func initCrypt() error {
	aesblock, err := aes.NewCipher(config.SecretKey)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return err
	}

	nonce := config.SecretKey[len(config.SecretKey)-aesgcm.NonceSize():]

	CryptURL = &CipherCrypt{
		aesGCM: aesgcm,
		nonce:  nonce,
	}

	return nil

}
