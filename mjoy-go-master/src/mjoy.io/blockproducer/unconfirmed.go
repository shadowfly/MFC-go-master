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
// @File: unconfirmed.go
// @Date: 2018/05/08 17:23:08
////////////////////////////////////////////////////////////////////////////////

package blockproducer

import (
	"container/ring"
	"sync"

	"mjoy.io/common/types"
	"mjoy.io/core/blockchain/block"
)

// headerRetriever is used by the unconfirmed block set to verify whether a previously
// produced block is part of the canonical chain or not.

type headerRetriever interface {
	// GetHeaderByNumber retrieves the canonical header associated with a block number.
	GetHeaderByNumber(number uint64) *block.Header
}

// unconfirmedBlock is a small collection of metadata about a locally produced block
// that is placed into a unconfirmed set for canonical chain inclusion tracking.

type unconfirmedBlock struct {
	index uint64
	hash  types.Hash
}

// unconfirmedBlocks implements a data structure to maintain locally produced blocks
// have have not yet reached enough maturity to guarantee chain inclusion. It is
// used by the blockproducer to provide logs to the user when a previously produced block
// has a high enough guarantee to not be reorged out of the canonical chain.

type unconfirmedBlocks struct {
	chain  headerRetriever // Blockchain to verify canonical status through
	depth  uint            // Depth after which to discard previous blocks
	blocks *ring.Ring      // Block infos to allow canonical chain cross checks
	lock   sync.RWMutex    // Protects the fields from concurrent access
}

// newUnconfirmedBlocks returns new data structure to track currently unconfirmed blocks.
func newUnconfirmedBlocks(chain headerRetriever, depth uint) *unconfirmedBlocks {
	return &unconfirmedBlocks{
		chain: chain,
		depth: depth,
	}
}
// Insert adds a new block to the set of unconfirmed ones.
func (set *unconfirmedBlocks) Insert(index uint64, hash types.Hash) {
	// If a new block was produced locally, shift out any old enough blocks
	set.Shift(index)

	// Create the new item as its own ring
	item := ring.New(1)
	item.Value = &unconfirmedBlock{
		index: index,
		hash:  hash,
	}
	// Set as the initial ring or append to the end
	set.lock.Lock()
	defer set.lock.Unlock()

	if set.blocks == nil {
		set.blocks = item
	} else {
		set.blocks.Move(-1).Link(item)
	}
	// Display a log for the user to notify of a new produced block unconfirmed
	logger.Infof("🔨 produced potential block number:%d  hash:0x%x\n" , index , hash)
}

// Shift drops all unconfirmed blocks from the set which exceed the unconfirmed sets depth
// allowance, checking them against the canonical chain for inclusion or staleness
// report.
func (set *unconfirmedBlocks) Shift(height uint64) {
	set.lock.Lock()
	defer set.lock.Unlock()

	for set.blocks != nil {

		// Retrieve the next unconfirmed block and abort if too fresh
		next := set.blocks.Value.(*unconfirmedBlock)
		if next.index+uint64(set.depth) > height {
			break
		}

		// Block seems to exceed depth allowance, check for canonical status
		header := set.chain.GetHeaderByNumber(next.index)
		switch {
		case header == nil:
			logger.Warnf("Failed to retrieve header of produced block number:%d  hash:0x%x\n" , next.index , next.hash)
		case header.Hash() == next.hash:
			logger.Infof("🔗 block reached canonical chain number:%d  hash:0x%x\n" , next.index , next.hash)
		default:
			logger.Infof("⑂ block  became a side fork number:%d  hash:0x%x\n" , next.index , next.hash)
		}
		// Drop the block out of the ring
		if set.blocks.Value == set.blocks.Next().Value {
			set.blocks = nil
		} else {
			set.blocks = set.blocks.Move(-1)
			set.blocks.Unlink(1)
			set.blocks = set.blocks.Move(1)
		}
	}
}
