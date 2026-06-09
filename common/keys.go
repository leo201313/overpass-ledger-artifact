package common

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"opl/coes"
)

// KeyPair is wrapped for private key
type KeyPair struct {
	PrvKey      *ecdsa.PrivateKey
	PubKeyBytes [coes.PUBKEY_LENGTH]byte
}

// PublicKeyToBytes generates the public key in the form of byte array
func PublicKeyToBytes(pubkey ecdsa.PublicKey) [coes.PUBKEY_LENGTH]byte {
	byteskey := [coes.PUBKEY_LENGTH]byte{}
	copy(byteskey[:], append(pubkey.X.Bytes(), pubkey.Y.Bytes()...))
	return byteskey
}

// BytesToPublickey reform the ecdsa.PublicKey from a bytes array
func BytesToPublickey(byteskey [coes.PUBKEY_LENGTH]byte) ecdsa.PublicKey {
	x := big.Int{}
	y := big.Int{}
	x.SetBytes(byteskey[:coes.PUBKEY_SEPARATE])
	y.SetBytes(byteskey[coes.PUBKEY_SEPARATE:])
	pubKey := ecdsa.PublicKey{
		Curve: coes.DEFAULT_CURVE,
		X:     &x,
		Y:     &y,
	}
	return pubKey
}

// PublicKeyBytes generates the public key in the form of byte array, return it and save it
func (kp *KeyPair) PublicKeyBytes() [coes.PUBKEY_LENGTH]byte {
	if kp.PubKeyBytes != [coes.PUBKEY_LENGTH]byte{} {
		return kp.PubKeyBytes
	} else {
		pubkey := PublicKeyToBytes(kp.PrvKey.PublicKey)
		kp.PubKeyBytes = pubkey
		return pubkey
	}
}

// GenerateKeyRandom generates a random key pair based on the default curve
func GenerateKeyRandom() (*KeyPair, error) {
	key, err := ecdsa.GenerateKey(coes.DEFAULT_CURVE, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("fail in common.GenerateKeyRandom: %v", err)
	}

	kp := &KeyPair{
		PrvKey: key,
	}

	kp.PublicKeyBytes()
	return kp, nil
}

// Sign use the private key to sign a text and back the signature in a bytes array
func (kp *KeyPair) Sign(text []byte) ([coes.SIG_LENGTH]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, kp.PrvKey, text)
	if err != nil {
		return [coes.SIG_LENGTH]byte{}, fmt.Errorf("fail in common.Sign: %v", err)
	}
	signature := [coes.SIG_LENGTH]byte{}
	rBytes := make([]byte, coes.SIG_SEPARATE)
	sBytes := make([]byte, coes.SIG_SEPARATE)
	copy(signature[:], append(r.FillBytes(rBytes), s.FillBytes(sBytes)...))
	return signature, nil
}

// VerifySignature verify the signature use pubkey (in bytes array form)
func VerifySignature(text []byte, byteskey [coes.PUBKEY_LENGTH]byte, signature [coes.SIG_LENGTH]byte) bool {
	r := big.Int{}
	s := big.Int{}
	r.SetBytes(signature[:coes.SIG_SEPARATE])
	s.SetBytes(signature[coes.SIG_SEPARATE:])

	pubkey := BytesToPublickey(byteskey)
	return ecdsa.Verify(&pubkey, text, &r, &s)
}
