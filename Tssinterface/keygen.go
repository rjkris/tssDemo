package Tssinterface

import (
	"crypto/elliptic"
	"fmt"
	"github.com/binance-chain/tss-lib/common"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"runtime"
	"sync/atomic"
)

func GenerateKeys(th, part int, id string) keygen.LocalPartySaveData {
	tss.SetCurve(elliptic.P256())

	threshold := th
	var output keygen.LocalPartySaveData

	fixtures, pIDs, err := LoadKeygenTest(part, id)
	if err != nil {
		common.Logger.Info("No test fixtures were found, so the safe primes will be generated from scratch. This may take a while...")
		pIDs = tss.GenerateTestPartyIDs(part)
	} else {
		return fixtures[0]
	}

	// pIDs := tssInterface.GenerateTestPartyIDs(th)
	p2pCtx := tss.NewPeerContext(pIDs)
	parties := make([]*keygen.LocalParty, 0, len(pIDs))

	errCh := make(chan *tss.Error, len(pIDs))
	outCh := make(chan tss.Message, len(pIDs))
	endCh := make(chan keygen.LocalPartySaveData, len(pIDs))

	updater := SharedPartyUpdater

	startGR := runtime.NumGoroutine()

	// init the parties
	for i := 0; i < len(pIDs); i++ {
		var P *keygen.LocalParty
		params := tss.NewParameters(p2pCtx, pIDs[i], len(pIDs), threshold)
		if i < len(fixtures) {
			P = keygen.NewLocalParty(params, outCh, endCh, fixtures[i].LocalPreParams).(*keygen.LocalParty)
		} else {
			P = keygen.NewLocalParty(params, outCh, endCh).(*keygen.LocalParty)
		}
		// P = keygen.NewLocalParty(params, outCh, endCh).(*keygen.LocalParty)
		parties = append(parties, P)
		go func(P *keygen.LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	// PHASE: keygen
	var ended int32
keygen:
	for {
		fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			common.Logger.Errorf("Error: %s", err)
			break keygen

		case msg := <-outCh:
			dest := msg.GetTo()
			if dest == nil { // broadcast!
				for _, P := range parties {
					if P.PartyID().Index == msg.GetFrom().Index {
						continue
					}
					go updater(P, msg, errCh)
				}
			} else { // point-to-point!
				if dest[0].Index == msg.GetFrom().Index {
					common.Logger.Fatalf("party %d tried to send a message to itself (%d)", dest[0].Index, msg.GetFrom().Index)
					return keygen.LocalPartySaveData{}
				}
				go updater(parties[dest[0].Index], msg, errCh)
			}

		case save := <-endCh:
			// SAVE a test fixture file for this P (if it doesn't already exist)
			// .. here comes a workaround to recover this party's index (it was removed from save data)
			index, err := save.OriginalIndex()
			if err != nil {
				common.Logger.Errorf("should not be an error getting a party's index from save data: %s", err)
			}
			TryWriteTestFixtureFile(index, save, id)

			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(pIDs)) {
				output = save
				common.Logger.Infof("Done. Received save data from %d participants", ended)
				common.Logger.Infof("Start goroutines: %d, End goroutines: %d", startGR, runtime.NumGoroutine())
				break keygen
			}
		}
	}
	return output
}

func SharedPartyUpdater(party tss.Party, msg tss.Message, errCh chan<- *tss.Error) {
	// do not send a message from this party back to itself
	if party.PartyID() == msg.GetFrom() {
		return
	}
	bz, _, err := msg.WireBytes()
	if err != nil {
		errCh <- party.WrapError(err)
		return
	}
	pMsg, err := tss.ParseWireMessage(bz, msg.GetFrom(), msg.IsBroadcast())
	if err != nil {
		errCh <- party.WrapError(err)
		return
	}
	if _, err := party.Update(pMsg); err != nil {
		errCh <- err
	}
}

