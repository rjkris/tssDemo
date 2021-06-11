package Tssinterface

import (
	"crypto/ecdsa"
	"crypto/elliptic"
)

type PrivateKeyCert struct {
	Pk           ecdsa.PublicKey
	Threshold    int
	Participants int
	Id           string
}

func GenerateKey(th, part int, id string) ecdsa.PublicKey {
	key := GenerateKeys(th, part, id)
	ecdsaPk := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     key.ECDSAPub.X(),
		Y:     key.ECDSAPub.Y(),
	}
	return ecdsaPk
}
