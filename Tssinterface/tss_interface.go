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

func GenerateKey(th, part int, id string) PrivateKeyCert {
	key := GenerateKeys(th, part, id)
	pks := PrivateKeyCert{}
	ecdsaPk := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     key.ECDSAPub.X(),
		Y:     key.ECDSAPub.Y(),
	}
	pks.Pk = ecdsaPk
	pks.Threshold = th
	pks.Participants = part
	pks.Id = id
	return pks
}
