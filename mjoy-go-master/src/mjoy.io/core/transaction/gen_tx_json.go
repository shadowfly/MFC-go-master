// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package transaction

import (
	"encoding/json"
	"errors"
	"mjoy.io/common/types"
)

func (t Txdata) MarshalJSON() ([]byte, error) {
	type Txdata struct {
		AccountNonce	uint64		`json:"nonce"   gencodec:"required"`
		Actions		ActionSlice	`json:"actions" gencodec:"required"`
		V		*types.BigInt	`json:"v"       gencodec:"required"`
		R		*types.BigInt	`json:"r"       gencodec:"required"`
		S		*types.BigInt	`json:"s"       gencodec:"required"`
		Hash		*types.Hash	`json:"hash"    msg:"-"`
	}
	var enc Txdata
	enc.AccountNonce = t.AccountNonce
	enc.Actions = t.Actions
	enc.V = t.V
	enc.R = t.R
	enc.S = t.S
	enc.Hash = t.Hash
	return json.Marshal(&enc)
}

func (t *Txdata) UnmarshalJSON(input []byte) error {
	type Txdata struct {
		AccountNonce	*uint64		`json:"nonce"   gencodec:"required"`
		Actions		ActionSlice	`json:"actions" gencodec:"required"`
		V		*types.BigInt	`json:"v"       gencodec:"required"`
		R		*types.BigInt	`json:"r"       gencodec:"required"`
		S		*types.BigInt	`json:"s"       gencodec:"required"`
		Hash		*types.Hash	`json:"hash"    msg:"-"`
	}
	var dec Txdata
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.AccountNonce == nil {
		return errors.New("missing required field 'nonce' for Txdata")
	}
	t.AccountNonce = *dec.AccountNonce
	if dec.Actions == nil {
		return errors.New("missing required field 'actions' for Txdata")
	}
	t.Actions = dec.Actions
	if dec.V == nil {
		return errors.New("missing required field 'v' for Txdata")
	}
	t.V = dec.V
	if dec.R == nil {
		return errors.New("missing required field 'r' for Txdata")
	}
	t.R = dec.R
	if dec.S == nil {
		return errors.New("missing required field 's' for Txdata")
	}
	t.S = dec.S
	if dec.Hash != nil {
		t.Hash = dec.Hash
	}
	return nil
}
