package wallet

import (
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// ChannelReservation...
type ChannelReservation struct {
	FundingType FundingType

	FundingAmount btcutil.Amount
	ReserveAmount btcutil.Amount
	MinFeePerKb   btcutil.Amount

	sync.RWMutex // All fields below owned by the wallet.

	theirInputs []*wire.TxIn
	ourInputs   []*wire.TxIn

	theirChange []*wire.TxOut
	ourChange   []*wire.TxOut

	ourKey   *btcec.PrivateKey
	theirKey *btcec.PublicKey

	// In order of sorted inputs. Sorting is done in accordance
	// to BIP-69: https://github.com/bitcoin/bips/blob/master/bip-0069.mediawiki.
	theirSigs [][]byte
	ourSigs   [][]byte

	normalizedTxID wire.ShaHash

	fundingTx *wire.MsgTx
	// TODO(roasbef): time locks, who pays fee etc.
	// TODO(roasbeef): record Bob's ln-ID?

	completedFundingTx *btcutil.Tx

	reservationID uint64
	wallet        *LightningWallet
}

// newChannelReservation...
func newChannelReservation(t FundingType, fundingAmt btcutil.Amount,
	minFeeRate btcutil.Amount, wallet *LightningWallet, id uint64) *ChannelReservation {
	return &ChannelReservation{
		FundingType:   t,
		FundingAmount: fundingAmt,
		MinFeePerKb:   minFeeRate,
		wallet:        wallet,
		reservationID: id,
	}
}

// OurFunds...
func (r *ChannelReservation) OurFunds() ([]*wire.TxIn, []*wire.TxOut, *btcec.PublicKey) {
	r.RLock()
	defer r.Unlock()
	return r.ourInputs, r.ourChange, r.ourKey.PubKey()
}

// AddCounterPartyFunds...
func (r *ChannelReservation) AddFunds(theirInputs []*wire.TxIn, theirChangeOutputs []*wire.TxOut, multiSigKey *btcec.PublicKey) error {
	errChan := make(chan error, 1)

	r.wallet.msgChan <- &addCounterPartyFundsMsg{
		pendingFundingID:   r.reservationID,
		theirInputs:        theirInputs,
		theirChangeOutputs: theirChangeOutputs,
		theirKey:           multiSigKey,
	}

	return <-errChan
}

// OurSigs...
func (r *ChannelReservation) OurSigs() [][]byte {
	r.RLock()
	defer r.Unlock()
	return r.ourSigs
}

// TheirFunds...
func (r *ChannelReservation) TheirFunds() ([]*wire.TxIn, []*wire.TxOut, *btcec.PublicKey) {
	r.RLock()
	defer r.Unlock()
	return r.theirInputs, r.theirChange, r.theirKey
}

// CompleteFundingReservation...
func (r *ChannelReservation) CompleteReservation(reservationID uint64, theirSigs [][]byte) error {
	errChan := make(chan error, 1)

	r.wallet.msgChan <- &addCounterPartySigsMsg{
		pendingFundingID: reservationID,
		theirSigs:        theirSigs,
		err:              errChan,
	}

	return <-errChan
}

// FinalFundingTransaction...
func (r *ChannelReservation) FinalFundingTx() *btcutil.Tx {
	r.RLock()
	defer r.Unlock()
	return r.completedFundingTx
}

// RequestFundingReserveCancellation...
// TODO(roasbeef): also return mutated state?
func (r *ChannelReservation) Cancel(reservationID uint64) {
	doneChan := make(chan struct{}, 1)
	r.wallet.msgChan <- &fundingReserveCancelMsg{
		pendingFundingID: reservationID,
		done:             doneChan,
	}

	<-doneChan
}