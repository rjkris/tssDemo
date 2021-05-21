package main

import (
	"fmt"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	s256k1 "github.com/btcsuite/btcd/btcec"
	"math/big"
	"strconv"
	"time"
)

func main() {
	tss.SetCurve(s256k1.S256())
	preParams, _ := keygen.GeneratePreParams(1 * time.Minute)
	fmt.Printf("preparams: %+v", preParams)

	var parties tss.UnSortedPartyIDs
	for i := 0; i < 3; i++ {
		var id, moniker = strconv.Itoa(i), strconv.Itoa(i)
		uniqueKey := big.NewInt(int64(i))
		thisParty := tss.NewPartyID(id, moniker, uniqueKey)
		parties = append(parties, thisParty)
	}

	sortParties := tss.SortPartyIDs(parties)

	ctx := tss.NewPeerContext(sortParties)
	curPartid := parties[0]
	params := tss.NewParameters(ctx, curPartid, len(parties), 3)

	partyIDMap := make(map[string]*tss.PartyID)
	for _, id := range parties {
		partyIDMap[id.Id] = id
	}

	tssMessageChan := make(chan tss.Message)
	endChan := make(chan keygen.LocalPartySaveData)
	go func() {
		party := keygen.NewLocalParty(params, tssMessageChan, endChan, *preParams) // Omit the last arg to compute the pre-params in round 1
		if err := party.Start(); err != nil {
			println(err)
		}
	}()

	keyGenData := <-endChan
	fmt.Printf("keyGenData: %+v", keyGenData)
	//for {
	//	select {
	//	case keygenData := <-endChan:
	//		// signer := &ThresholdSigner{
	//		// 	groupInfo:    s.groupInfo,
	//		// 	thresholdKey: ThresholdKey(keygenData),
	//		// }
	//		fmt.Println("get keygenData ",keygenData)
	//		pkX, pkY := keygenData.ECDSAPub.X(), keygenData.ECDSAPub.Y()
	//
	//		curve := tss.EC()
	//		publicKey := ecdsa.PublicKey{
	//			Curve: curve,
	//			X:     pkX,
	//			Y:     pkY,
	//		}
	//		fmt.Println("get publicKey ",publicKey)
	//		//ethPublicKey, err := eth.SerializePublicKey((*ecdsa.PublicKey)(&publicKey))
	//		//if err != nil {
	//		//	fmt.Println("SerializePublicKey error ",err)
	//		//}
	//		//fmt.Println("eth Public key ",	)
	//		return //signer, nil
	//	case tmp := <-tssMessageChan://如何把message路由给其他的chan?
	//			fmt.Println("get message ",tmp)
	//			continue
	//		//_, routing, _ := tmp.WireBytes()
	//		//senderPartyID := sortedPartyIDs.FindByKey(MemberID(routing.From.GetKey()).bigInt())
	//
	//		// _, err := party.UpdateFromBytes(
	//		// 	bytes,
	//		// 	senderPartyID,
	//		// 	true,
	//		// )
	//		// if err != nil {
	//		// 	fmt.Println("UpdateFromBytes error ",err)
	//		// 	continue
	//		// }
	//		//if senderPartyID == party.PartyID() {
	//		//	tssGlobalMessageChan<-tmp
	//		//	//				fmt.Println("senderPartyID equal partyID ",senderPartyID)
	//		//}
	//		// fmt.Println("UpdateFromBytes succ ")
	//		//nil, timeoutError{KeyGenerationProtocolTimeout, "key generation", memberIDs}
	//	}
	//}
}

