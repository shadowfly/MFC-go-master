////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: block.go
// @Date: 2018/05/08 15:18:08
////////////////////////////////////////////////////////////////////////////////

package block

import (
	"mjoy.io/core/transaction"
	"mjoy.io/common/types"
	"mjoy.io/trie"
	"encoding/binary"
	"bytes"
	"sync/atomic"
	"math/big"
	"time"
	"sort"
	"mjoy.io/common"
	"github.com/tinylib/msgp/msgp"
	"fmt"
	"mjoy.io/utils/bloom"
	"mjoy.io/common/types/util/hex"
)

type DerivableList interface {
	Len() int
	GetMsgp(i int) []byte
}

func DeriveSha(list DerivableList) types.Hash {
	keyBytesBuf := bytes.NewBuffer([]byte{})
	trie := new(trie.Trie)
	for i := 0; i < list.Len(); i++ {
		keyBytesBuf.Reset()
		binary.Write(keyBytesBuf, binary.BigEndian, i)
		trie.Update(keyBytesBuf.Bytes(), list.GetMsgp(i))
	}
	return trie.Hash()
}

//go:generate msgp


type ConsensusData struct{
	Id     string
	Para   []byte
}

// block header
type Header struct {
	ParentHash  		types.Hash            `json:"parentHash" `
	StateRootHash   	types.Hash            `json:"stateRoot" `
	TxRootHash      	types.Hash            `json:"transactionsRoot" `
	ReceiptRootHash 	types.Hash            `json:"receiptsRoot" `
	Bloom       		types.Bloom           `json:"logsBloom" `
	Number      		*types.BigInt         `json:"number" `
	Time        		*types.BigInt         `json:"timestamp" `

	//BlockProducer is not used in protocol.
	BlockProducer   	types.Address         `json:"blockProducer" msg:"-"`
	ConsensusData     	ConsensusData         `json:"consensusData" `
	//Signature values
	V                 	*types.BigInt         `json:"v"`
	R                 	*types.BigInt         `json:"r"`
	S                 	*types.BigInt         `json:"s"`
}

type Body struct {
	Transactions []*transaction.Transaction
}
type Block struct {
	B_header      *Header               // block header
	B_body        Body                  // all transactions in this block

	// caches
	hash atomic.Value
	size atomic.Value

	// These fields are used by package mjoy to track
	// inter-peer block relay.
	ReceivedAt   time.Time      `msg:"-"`
	ReceivedFrom interface{}    `msg:"-"`
}

func (h *Header) Hash() types.Hash {
	hash, err := common.MsgpHash(h)
	if err != nil {
		return types.Hash{}
	}
	return hash
}

var (
	EmptyRootHash  = DeriveSha(transaction.Transactions{})
)
// NewBlock creates a new block. The input data is copied,
// changes to header and to the field values will not affect the
// block.
//
// The values of TxHash, ReceiptHash and Bloom in header
// are ignored and set to values derived from the given txs and receipts.
func NewBlock(header *Header, txs []*transaction.Transaction, receipts []*transaction.Receipt) *Block {
	b := &Block{B_header: CopyHeader(header)}

	// TODO: panic if len(txs) != len(receipts)
	if len(txs) == 0 {
		b.B_header.TxRootHash = EmptyRootHash
	} else {
		b.B_header.TxRootHash = DeriveSha(transaction.Transactions(txs))
		b.B_body.Transactions = make(transaction.Transactions, len(txs))
		copy(b.B_body.Transactions, txs)
	}

	if len(receipts) == 0 {
		b.B_header.ReceiptRootHash = EmptyRootHash
	} else {
		b.B_header.ReceiptRootHash = DeriveSha(transaction.Receipts(receipts))
		bloomIn := []bloom.BloomByte{}
		for _, receipt := range receipts {
			for _, log := range receipt.Logs {
				bloomIn = append(bloomIn, log.Address)
				for _, b := range log.Topics {
					bloomIn = append(bloomIn, b)
				}
			}
		}
		b.B_header.Bloom = bloom.CreateBloom(bloomIn)
	}
	return b
}
// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h
	if cpy.Time = new(types.BigInt); h.Time != nil {
		cpy.Time.Put(h.Time.IntVal)
	}

	if cpy.Number = new(types.BigInt); h.Number != nil {
		cpy.Number.Put(h.Number.IntVal)
	}

	if cpy.V = new(types.BigInt); h.V != nil {
		cpy.V.Put(h.V.IntVal)
	}

	if cpy.R = new(types.BigInt); h.R != nil {
		cpy.R.Put(h.R.IntVal)
	}

	if cpy.S = new(types.BigInt); h.S != nil {
		cpy.S.Put(h.S.IntVal)
	}

	return &cpy
}

func NewBlockWithHeader(header *Header) *Block {
	return &Block{B_header: CopyHeader(header)}
}

func (header *Header) HashNoSig() types.Hash {
	v := &HeaderNoSig{
		header.ParentHash,
		header.StateRootHash,
		header.TxRootHash,
		header.ReceiptRootHash,
		header.Bloom,
		header.Number,
		header.Time,
		header.BlockProducer,
		header.ConsensusData,
	}
	return v.Hash()
}

// WithSignature returns a new header with the given signature.
func (h *Header) WithSignature(signer Signer, sig []byte) (*Header, error) {
	r, s, v, err := signer.SignatureValues(h, sig)
	if err != nil {
		return nil, err
	}
	cpy := CopyHeader(h)
	cpy.R, cpy.S, cpy.V = &types.BigInt{*r}, &types.BigInt{*s}, &types.BigInt{*v}
	return cpy, nil
}

// AddSignature returns modify the  header( R, S, V) with the given signature.
func (h *Header) AddSignature(signer Signer, sig []byte) ( error) {
	r, s, v, err := signer.SignatureValues(h, sig)
	if err != nil {
		return err
	}
	h.R, h.S, h.V = &types.BigInt{*r}, &types.BigInt{*s}, &types.BigInt{*v}
	return  nil
}

func (b *Block) Transactions() transaction.Transactions { return b.B_body.Transactions }

func (b *Block) Transaction(hash types.Hash) *transaction.Transaction {
	for _, transaction := range b.B_body.Transactions {
		if transaction.Hash() == hash {
			return transaction
		}
	}
	return nil
}


func (b *Block) Number() *big.Int     { return new(big.Int).Set(&b.B_header.Number.IntVal) }
func (b *Block) Time() *big.Int       { return new(big.Int).Set(&b.B_header.Time.IntVal) }

func (b *Block) NumberU64() uint64       { return b.B_header.Number.IntVal.Uint64() }
func (b *Block) Bloom() types.Bloom      { return b.B_header.Bloom }
func (b *Block) Coinbase() types.Address { return b.B_header.BlockProducer }
func (b *Block) Root() types.Hash        { return b.B_header.StateRootHash }
func (b *Block) ParentHash() types.Hash  { return b.B_header.ParentHash }
func (b *Block) TxHash() types.Hash      { return b.B_header.TxRootHash }
func (b *Block) ReceiptHash() types.Hash { return b.B_header.ReceiptRootHash }

func (b *Block) Header() *Header { return CopyHeader(b.B_header) }

// Body returns the non-header content of the block.
func (b *Block) Body() *Body { return &Body{b.B_body.Transactions} }


func (b *Block) HashNoSig() types.Hash {
	return b.B_header.HashNoSig()
}

func (b *Block) Size() common.StorageSize {
	if size := b.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := 0
	var buf bytes.Buffer
	msgp.Encode(&buf, b)
	c = buf.Len()
	b.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}


// WithSeal returns a new block with the data from b but the header replaced with
// the sealed one.
func (b *Block) WithSeal(header *Header) *Block {
	cpy := *header

	return &Block{
		B_header:       &cpy,
		B_body:         b.B_body,
	}
}

// WithBody returns a new block with the given transaction  contents.
func (b *Block) WithBody(body *Body) *Block {
	block := &Block{
		B_header:       CopyHeader(b.B_header),
	}
	block.B_body.Transactions = make([]*transaction.Transaction, len(body.Transactions))
	copy(block.B_body.Transactions, body.Transactions)
	return block
}

// Hash returns the keccak256 hash of b's header.
// The hash is computed on the first call and cached thereafter.
func (b *Block) Hash() types.Hash {
	if hash := b.hash.Load(); hash != nil {
		return hash.(types.Hash)
	}
	v := b.B_header.Hash()
	b.hash.Store(v)
	return v
}

func (b *Block) String() string {
	str := fmt.Sprintf(`Block(#%v): Size: %v {
BlockproducerHash: %x
%v
Transactions:
%v
}
`, b.Number(), b.Size(), b.B_header.HashNoSig(), b.B_header, b.B_body.Transactions)
	return str
}

func (h *Header) String() string {
	return fmt.Sprintf(`Header(%x):
[
	ParentHash:         %x
	BlockProducer:      %x
	StateRootHash:      %x
	TxRootHash          %x
	ReceiptRootHash:    %x
	Bloom:              %x
	Number:	            %v
	Time:               %v
	ConsensusData:      %v
	R:                  %v
	S:                  %v
    V:                  %v
]`, h.Hash(), h.ParentHash, h.BlockProducer, h.StateRootHash, h.TxRootHash, h.ReceiptRootHash, h.Bloom, (*hex.Big)(&h.Number.IntVal), (*hex.Big)(&h.Time.IntVal), h.ConsensusData, (*hex.Big)(&h.R.IntVal), (*hex.Big)(&h.S.IntVal), (*hex.Big)(&h.V.IntVal))
}


//blocks part

type Blocks []*Block

type BlockBy func(b1, b2 *Block) bool

func (self BlockBy) Sort(blocks Blocks) {
	bs := blockSorter{
		blocks: blocks,
		by:     self,
	}
	sort.Sort(bs)
}

type blockSorter struct {
	blocks Blocks
	by     func(b1, b2 *Block) bool
}

func (self blockSorter) Len() int { return len(self.blocks) }
func (self blockSorter) Swap(i, j int) {
	self.blocks[i], self.blocks[j] = self.blocks[j], self.blocks[i]
}
func (self blockSorter) Less(i, j int) bool { return self.by(self.blocks[i], self.blocks[j]) }

func Number(b1, b2 *Block) bool { return b1.B_header.Number.IntVal.Cmp(&b2.B_header.Number.IntVal) < 0 }

// header wihtout signature
type HeaderNoSig struct {
	ParentHash  		types.Hash            `json:"parentHash" `
	StateRootHash   	types.Hash            `json:"stateRoot" `
	TxRootHash      	types.Hash            `json:"transactionsRoot"`
	ReceiptRootHash 	types.Hash            `json:"receiptsRoot" `
	Bloom       		types.Bloom           `json:"logsBloom" `
	Number      		*types.BigInt         `json:"number" `
	Time        		*types.BigInt         `json:"timestamp" `

	//BlockProducer is not used in protocol.
	BlockProducer   	types.Address         `json:"blockProducer" msg:"-" `
	ConsensusData     	ConsensusData         `json:"consensusData" `
}
func (h *HeaderNoSig) Hash() types.Hash {
	hash, err := common.MsgpHash(h)
	if err != nil {
		return types.Hash{}
	}
	return hash
}

//type Headers []*Header
type Headers struct {
	Headers []*Header
}
