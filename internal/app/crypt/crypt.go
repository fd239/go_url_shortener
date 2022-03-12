package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"log"
)

var CryptURL *CipherCrypt

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
		log.Println(fmt.Sprintf("Decrypt decode string error: %v", err))
		return "", err
	}

	decrypted, terr := CryptURL.aesGCM.Open(nil, CryptURL.nonce, b, nil)

	if terr != nil {
		log.Println(fmt.Sprintf("aesGCM open error: %v", terr))
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
	aesblock, err := aes.NewCipher(common.SecretKey)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return err
	}

	nonce := common.SecretKey[len(common.SecretKey)-aesgcm.NonceSize():]

	CryptURL = &CipherCrypt{
		aesGCM: aesgcm,
		nonce:  nonce,
	}

	return nil

}
