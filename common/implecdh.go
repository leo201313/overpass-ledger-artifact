package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"opl/coes"
)

// ComputeSharedAESKey uses the given public key (in bytes array form) to run the ECDH algorithm and back an AES key.
// [nonce] is for differentiating the final AES key, and it might be timestamp or something others.
func (kp *KeyPair) ComputeSharedAESKey(byteskey [coes.PUBKEY_LENGTH]byte, nonce []byte) (cipher.Block, error) {
	ecdhPrvKey, err := kp.PrvKey.ECDH()
	if err != nil {
		return nil, fmt.Errorf("fail in common.ComputeSharedAESKey: %v", err)
	}
	pubkey := BytesToPublickey(byteskey)
	ecdhPubKey, err := pubkey.ECDH()
	if err != nil {
		return nil, fmt.Errorf("fail in common.ComputeSharedAESKey: %v", err)
	}
	shared, _ := ecdhPrvKey.ECDH(ecdhPubKey)
	concat := append(shared, nonce...)
	hashShared := sha256.Sum256(concat)
	aesKey, err := aes.NewCipher(hashShared[:])
	if err != nil {
		return nil, fmt.Errorf("fail in common.ComputeSharedAESKey: %v", err)
	}
	return aesKey, nil
}

// ComputeSharedAES uses the given public key (in bytes array form) to run the ECDH algorithm and back an AES key in
// byte array form. [nonce] is for differentiating the final AES key, and it might be timestamp or something others.
func (kp *KeyPair) ComputeSharedAES(byteskey [coes.PUBKEY_LENGTH]byte, nonce []byte) ([32]byte, error) {
	ecdhPrvKey, err := kp.PrvKey.ECDH()
	if err != nil {
		return [32]byte{}, fmt.Errorf("fail in common.ComputeSharedAESKey: %v", err)
	}
	pubkey := BytesToPublickey(byteskey)
	ecdhPubKey, err := pubkey.ECDH()
	if err != nil {
		return [32]byte{}, fmt.Errorf("fail in common.ComputeSharedAESKey: %v", err)
	}
	shared, _ := ecdhPrvKey.ECDH(ecdhPubKey)
	concat := append(shared, nonce...)
	hashShared := sha256.Sum256(concat)

	return hashShared, nil
}

// AES_Encrypt uses the ase key to encrypt test in a GCM mode
func AES_Encrypt(text []byte, key cipher.Block) ([]byte, error) {
	gcm, err := cipher.NewGCM(key)
	if err != nil {
		return nil, fmt.Errorf("fail in common.AES_Encrypt: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("fail in common.AES_Encrypt: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, text, nil)

	return ciphertext, nil
}

// AES_Decrypt decrypts a ciphertext using the GCM mode
func AES_Decrypt(ciphertext []byte, key cipher.Block) ([]byte, error) {
	gcm, err := cipher.NewGCM(key)
	if err != nil {
		return nil, fmt.Errorf("fail in common.AES_Decrypt: %v", err)
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("fail in common.AES_Decrypt: %v", err)
	}

	return plaintext, nil
}
