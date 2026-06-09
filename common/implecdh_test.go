package common

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func TestKeyPair_ComputeSharedAESKey(t *testing.T) {
	Krad, _ := GenerateKeyRandom()
	Leo, _ := GenerateKeyRandom()
	Krad_PubKey := Krad.PublicKeyBytes()
	Leo_PubKey := Leo.PublicKeyBytes()

	nonce := []byte(strconv.Itoa(rand.Int()))

	Krad_AES, err := Krad.ComputeSharedAESKey(Leo_PubKey, nonce)
	if err != nil {
		t.Fatalf("%v", err)
	}

	Leo_AES, err := Leo.ComputeSharedAESKey(Krad_PubKey, nonce)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(Krad_AES, Leo_AES) {
		t.Fatalf("The two aes key is not same!")
	}

	message := []byte("This is the test message!")
	cipherMessage, err := AES_Encrypt(message, Krad_AES)
	if err != nil {
		t.Fatalf("%v", err)
	}

	decryptMessage, err := AES_Decrypt(cipherMessage, Leo_AES)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(message, decryptMessage) {
		t.Fatalf("Messages are not the same!")
	}
}
