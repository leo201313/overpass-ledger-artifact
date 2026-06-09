package common

import "testing"

func TestVerifySignature(t *testing.T) {
	krad, err := GenerateKeyRandom()
	if err != nil {
		t.Fatalf("%v", err)
	}

	testMessage := []byte("This is a test message")
	signature, err := krad.Sign(testMessage)
	if err != nil {
		t.Fatalf("%v", err)
	}

	t.Logf("%x", krad.PublicKeyBytes())

	if VerifySignature(testMessage, krad.PubKeyBytes, signature) {
		return
	} else {
		t.Fatalf("fail in verify")
	}
}
