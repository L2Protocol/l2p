// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package forkid

import (
	"bytes"
	"hash/crc32"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// TestCreation tests that different genesis and fork rule combinations result in
// the correct fork ID.
func TestCreation(t *testing.T) {
	type testcase struct {
		head uint64
		time uint64
		want ID
	}
	tests := []struct {
		config  *params.ChainConfig
		genesis *types.Block
		cases   []testcase
	}{
		// L2P mainnet test cases
		{
			params.MainnetChainConfig,
			core.DefaultGenesisBlock().ToBlock(),
			[]testcase{
				{0, 0, ID{Hash: checksumToBytes(0xa55701c9), Next: 1}},                     // Unsynced
				{1, 0, ID{Hash: checksumToBytes(0x78f790fe), Next: 2}},                     // First MirrorSync and Bruno block
				{2, 0, ID{Hash: checksumToBytes(0x4f304cae), Next: 3}},                     // First Euler block
				{3, 0, ID{Hash: checksumToBytes(0x1e9eac55), Next: 4}},                     // First Nano and Moran block
				{4, 0, ID{Hash: checksumToBytes(0x073e0118), Next: 5}},                     // First Gibbs block
				{5, 0, ID{Hash: checksumToBytes(0x9031b001), Next: 6}},                     // First Planck block
				{6, 0, ID{Hash: checksumToBytes(0x7ebec67a), Next: 7}},                     // First Luban block
				{7, 0, ID{Hash: checksumToBytes(0x0f394188), Next: 8}},                     // First Plato block
				{8, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: 1767884400}},            // First Berlin, London, Hertz and Hertzfix block
				{100, 1767884399, ID{Hash: checksumToBytes(0xa87e764d), Next: 1767884400}}, // Last pre-Shanghai block
				{100, 1767884400, ID{Hash: checksumToBytes(0xc5168552), Next: 1767884500}}, // First Shanghai and Kepler block
				{100, 1767884500, ID{Hash: checksumToBytes(0x61a5dfab), Next: 1767884600}}, // First Feynman and FeynmanFix block
				{100, 1767884600, ID{Hash: checksumToBytes(0xaf7aa1ac), Next: 1767884610}}, // First Cancun and Haber block
				{100, 1767884610, ID{Hash: checksumToBytes(0x1c7563a1), Next: 1767884620}}, // First HaberFix block
				{100, 1767884620, ID{Hash: checksumToBytes(0x9d7b15c4), Next: 1767884630}}, // First Bohr block
				{100, 1767884630, ID{Hash: checksumToBytes(0xc1265f22), Next: 1767884640}}, // First Pascal and Prague block
				{100, 1767884640, ID{Hash: checksumToBytes(0xe6abf589), Next: 0}},          // First Lorentz block
				{1000000, 2000000000, ID{Hash: checksumToBytes(0xe6abf589), Next: 0}},      // Future Lorentz block
			},
		},
	}
	for i, tt := range tests {
		for j, ttt := range tt.cases {
			if have := NewID(tt.config, tt.genesis, ttt.head, ttt.time); have != ttt.want {
				t.Errorf("test %d, case %d: fork ID mismatch: have %x, want %x", i, j, have, ttt.want)
			}
		}
	}
}

// TestValidation tests that a local peer correctly validates and accepts a remote
// fork ID.
func TestValidation(t *testing.T) {
	// Config that stops before any timestamp based fork, so the block based
	// transitions can be exercised on their own.
	blockConfig := *params.MainnetChainConfig
	blockConfig.ShanghaiTime = nil
	blockConfig.KeplerTime = nil
	blockConfig.FeynmanTime = nil
	blockConfig.FeynmanFixTime = nil
	blockConfig.CancunTime = nil
	blockConfig.HaberTime = nil
	blockConfig.HaberFixTime = nil
	blockConfig.BohrTime = nil
	blockConfig.PascalTime = nil
	blockConfig.PragueTime = nil
	blockConfig.LorentzTime = nil

	tests := []struct {
		config *params.ChainConfig
		head   uint64
		time   uint64
		id     ID
		err    error
	}{
		//------------------
		// Block based tests
		//------------------

		// Local is on the last block based fork, remote announces the same. No future fork is announced.
		{&blockConfig, 8, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: 0}, nil},

		// Local is on the last block based fork, remote announces the same. Remote also announces a next
		// fork at block 0xffffffff, but that is uncertain.
		{&blockConfig, 8, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: math.MaxUint64}, nil},

		// Local is currently in Plato only (so it's aware of Berlin), remote announces also Plato, but
		// it's not yet aware of Berlin (e.g. non updated node before the fork). In this case we don't
		// know if Berlin passed yet or not.
		{&blockConfig, 7, 0, ID{Hash: checksumToBytes(0x0f394188), Next: 0}, nil},

		// Local is currently in Plato only (so it's aware of Berlin), remote announces also Plato, and
		// it's also aware of Berlin (e.g. updated node before the fork). We don't know if Berlin passed
		// yet (will pass) or not.
		{&blockConfig, 7, 0, ID{Hash: checksumToBytes(0x0f394188), Next: 8}, nil},

		// Local is currently in Plato only (so it's aware of Berlin), remote announces also Plato, and
		// it's also aware of some random fork (e.g. misconfigured Berlin). As neither forks passed at
		// neither nodes, they may mismatch, but we still connect for now.
		{&blockConfig, 7, 0, ID{Hash: checksumToBytes(0x0f394188), Next: math.MaxUint64}, nil},

		// Local is exactly on Berlin, remote announces Plato + knowledge about Berlin. Remote is simply
		// out of sync, accept.
		{&blockConfig, 8, 0, ID{Hash: checksumToBytes(0x0f394188), Next: 8}, nil},

		// Local is past Berlin, remote announces Plato + knowledge about Berlin. Remote is simply out of
		// sync, accept.
		{&blockConfig, 100, 0, ID{Hash: checksumToBytes(0x0f394188), Next: 8}, nil},

		// Local is past Berlin, remote announces Luban + knowledge about Plato. Remote is definitely out
		// of sync. It may or may not need the Berlin update, we don't know yet.
		{&blockConfig, 100, 0, ID{Hash: checksumToBytes(0x7ebec67a), Next: 7}, nil},

		// Local is in Planck, remote announces Berlin. Local is out of sync, accept.
		{&blockConfig, 5, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: 0}, nil},

		// Local is past Berlin. Remote announces Plato but is not aware of further forks. Remote needs
		// a software update.
		{&blockConfig, 100, 0, ID{Hash: checksumToBytes(0x0f394188), Next: 0}, ErrRemoteStale},

		// Local is past Berlin, and isn't aware of more forks. Remote announces Berlin + 0xffffffff.
		// Local needs a software update, reject.
		{&blockConfig, 100, 0, ID{Hash: checksumToBytes(checksumUpdate(0xa87e764d, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is in Plato, and is aware of Berlin. Remote announces Berlin + 0xffffffff. Local needs a
		// software update, reject.
		{&blockConfig, 7, 0, ID{Hash: checksumToBytes(checksumUpdate(0xa87e764d, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is past Berlin, remote is on a completely different chain.
		{&blockConfig, 100, 0, ID{Hash: checksumToBytes(0x12345678), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is past Berlin, far in the future. Remote announces Gopherium (non existing fork) at some
		// future block 88888888, for itself, but past block for local. Local is incompatible.
		{&blockConfig, 88888888, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: 88888888}, ErrLocalIncompatibleOrStale},

		// Local is in Plato. Remote is also in Plato, but announces Gopherium (non existing fork) at
		// block 7, before Berlin. Local is incompatible.
		{&blockConfig, 7, 0, ID{Hash: checksumToBytes(0x0f394188), Next: 7}, ErrLocalIncompatibleOrStale},

		//------------------------------------
		// Block to timestamp transition tests
		//------------------------------------

		// Local is currently in Berlin only (so it's aware of Shanghai), remote announces also Berlin,
		// but it's not yet aware of Shanghai (e.g. non updated node before the fork). In this case we
		// don't know if Shanghai passed yet or not.
		{params.MainnetChainConfig, 8, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: 0}, nil},

		// Local is currently in Berlin only (so it's aware of Shanghai), remote announces also Berlin,
		// and it's also aware of Shanghai (e.g. updated node before the fork). We don't know if Shanghai
		// passed yet (will pass) or not.
		{params.MainnetChainConfig, 8, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: 1767884400}, nil},

		// Local is currently in Berlin only (so it's aware of Shanghai), remote announces also Berlin,
		// and it's also aware of some random fork (e.g. misconfigured Shanghai). As neither forks passed
		// at neither nodes, they may mismatch, but we still connect for now.
		{params.MainnetChainConfig, 8, 0, ID{Hash: checksumToBytes(0xa87e764d), Next: math.MaxUint64}, nil},

		// Local is exactly on Shanghai, remote announces Berlin + knowledge about Shanghai. Remote is
		// simply out of sync, accept.
		{params.MainnetChainConfig, 100, 1767884400, ID{Hash: checksumToBytes(0xa87e764d), Next: 1767884400}, nil},

		// Local is in Shanghai, remote announces Berlin + knowledge about Shanghai. Remote is simply out
		// of sync, accept.
		{params.MainnetChainConfig, 123456, 1767884401, ID{Hash: checksumToBytes(0xa87e764d), Next: 1767884400}, nil},

		// Local is in Berlin, remote announces Shanghai. Local is out of sync, accept.
		{params.MainnetChainConfig, 8, 0, ID{Hash: checksumToBytes(0xc5168552), Next: 0}, nil},

		// Local is in Shanghai. Remote announces Berlin but is not aware of further forks. Remote needs a
		// software update.
		{params.MainnetChainConfig, 100, 1767884400, ID{Hash: checksumToBytes(0xa87e764d), Next: 0}, ErrRemoteStale},

		// Local is in Berlin, and is aware of Shanghai. Remote announces Shanghai + 0xffffffff. Local
		// needs a software update, reject.
		{params.MainnetChainConfig, 8, 0, ID{Hash: checksumToBytes(checksumUpdate(0xc5168552, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		//----------------------
		// Timestamp based tests
		//----------------------

		// Local is on the last fork, remote announces the same. No future fork is announced.
		{params.MainnetChainConfig, 1000000, 1767884640, ID{Hash: checksumToBytes(0xe6abf589), Next: 0}, nil},

		// Local is on the last fork, remote announces the same. Remote also announces a next fork at
		// time 0xffffffff, but that is uncertain.
		{params.MainnetChainConfig, 1000000, 1767884640, ID{Hash: checksumToBytes(0xe6abf589), Next: math.MaxUint64}, nil},

		// Local is currently in Pascal only (so it's aware of Lorentz), remote announces also Pascal, but
		// it's not yet aware of Lorentz. In this case we don't know if Lorentz passed yet or not.
		{params.MainnetChainConfig, 1000000, 1767884630, ID{Hash: checksumToBytes(0xc1265f22), Next: 0}, nil},

		// Local is currently in Pascal only (so it's aware of Lorentz), remote announces also Pascal, and
		// it's also aware of Lorentz. We don't know if Lorentz passed yet (will pass) or not.
		{params.MainnetChainConfig, 1000000, 1767884630, ID{Hash: checksumToBytes(0xc1265f22), Next: 1767884640}, nil},

		// Local is exactly on Lorentz, remote announces Pascal + knowledge about Lorentz. Remote is
		// simply out of sync, accept.
		{params.MainnetChainConfig, 1000000, 1767884640, ID{Hash: checksumToBytes(0xc1265f22), Next: 1767884640}, nil},

		// Local is in Lorentz, remote announces Bohr + knowledge about Pascal. Remote is definitely out
		// of sync. It may or may not need the Lorentz update, we don't know yet.
		{params.MainnetChainConfig, 1000000, 1767884640, ID{Hash: checksumToBytes(0x9d7b15c4), Next: 1767884630}, nil},

		// Local is in Pascal, remote announces Lorentz. Local is out of sync, accept.
		{params.MainnetChainConfig, 1000000, 1767884630, ID{Hash: checksumToBytes(0xe6abf589), Next: 0}, nil},

		// Local is in Lorentz. Remote announces Pascal but is not aware of further forks. Remote needs a
		// software update.
		{params.MainnetChainConfig, 1000000, 1767884640, ID{Hash: checksumToBytes(0xc1265f22), Next: 0}, ErrRemoteStale},

		// Local is in Lorentz, and isn't aware of more forks. Remote announces Lorentz + 0xffffffff.
		// Local needs a software update, reject.
		{params.MainnetChainConfig, 1000000, 1767884640, ID{Hash: checksumToBytes(checksumUpdate(0xe6abf589, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is in Lorentz, remote is on a completely different chain.
		{params.MainnetChainConfig, 1000000, 1767884640, ID{Hash: checksumToBytes(0x12345678), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is in Lorentz, far in the future. Remote announces Gopherium (non existing fork) at some
		// future timestamp 8888888888, for itself, but past timestamp for local. Local is incompatible.
		{params.MainnetChainConfig, 88888888, 8888888888, ID{Hash: checksumToBytes(0xe6abf589), Next: 8888888888}, ErrLocalIncompatibleOrStale},

		// Local is in Pascal. Remote is also in Pascal, but announces Gopherium (non existing fork) at
		// timestamp 1767884630, before Lorentz. Local is incompatible.
		{params.MainnetChainConfig, 1000000, 1767884630, ID{Hash: checksumToBytes(0xc1265f22), Next: 1767884630}, ErrLocalIncompatibleOrStale},
	}
	genesis := core.DefaultGenesisBlock().ToBlock()
	for i, tt := range tests {
		filter := newFilter(tt.config, genesis, func() (uint64, uint64) { return tt.head, tt.time })
		if err := filter(tt.id); err != tt.err {
			t.Errorf("test %d: validation error mismatch: have %v, want %v", i, err, tt.err)
		}
	}
}

// Tests that IDs are properly RLP encoded (specifically important because we
// use uint32 to store the hash, but we need to encode it as [4]byte).
func TestEncoding(t *testing.T) {
	tests := []struct {
		id   ID
		want []byte
	}{
		{ID{Hash: checksumToBytes(0), Next: 0}, common.Hex2Bytes("c6840000000080")},
		{ID{Hash: checksumToBytes(0xdeadbeef), Next: 0xBADDCAFE}, common.Hex2Bytes("ca84deadbeef84baddcafe,")},
		{ID{Hash: checksumToBytes(math.MaxUint32), Next: math.MaxUint64}, common.Hex2Bytes("ce84ffffffff88ffffffffffffffff")},
	}
	for i, tt := range tests {
		have, err := rlp.EncodeToBytes(tt.id)
		if err != nil {
			t.Errorf("test %d: failed to encode forkid: %v", i, err)
			continue
		}
		if !bytes.Equal(have, tt.want) {
			t.Errorf("test %d: RLP mismatch: have %x, want %x", i, have, tt.want)
		}
	}
}

// Tests that time-based forks which are active at genesis are not included in
// forkid hash.
func TestTimeBasedForkInGenesis(t *testing.T) {
	var (
		time       = uint64(1690475657)
		genesis    = types.NewBlockWithHeader(&types.Header{Time: time})
		forkidHash = checksumToBytes(crc32.ChecksumIEEE(genesis.Hash().Bytes()))
		config     = func(shanghai, cancun uint64) *params.ChainConfig {
			return &params.ChainConfig{
				ChainID:                 big.NewInt(1337),
				HomesteadBlock:          big.NewInt(0),
				DAOForkBlock:            nil,
				DAOForkSupport:          true,
				EIP150Block:             big.NewInt(0),
				EIP155Block:             big.NewInt(0),
				EIP158Block:             big.NewInt(0),
				ByzantiumBlock:          big.NewInt(0),
				ConstantinopleBlock:     big.NewInt(0),
				PetersburgBlock:         big.NewInt(0),
				IstanbulBlock:           big.NewInt(0),
				MuirGlacierBlock:        big.NewInt(0),
				BerlinBlock:             big.NewInt(0),
				LondonBlock:             big.NewInt(0),
				TerminalTotalDifficulty: big.NewInt(0),
				MergeNetsplitBlock:      big.NewInt(0),
				ShanghaiTime:            &shanghai,
				CancunTime:              &cancun,
				Ethash:                  new(params.EthashConfig),
			}
		}
	)
	tests := []struct {
		config *params.ChainConfig
		want   ID
	}{
		// Shanghai active before genesis, skip
		{config(time-1, time+1), ID{Hash: forkidHash, Next: time + 1}},

		// Shanghai active at genesis, skip
		{config(time, time+1), ID{Hash: forkidHash, Next: time + 1}},

		// Shanghai not active, skip
		{config(time+1, time+2), ID{Hash: forkidHash, Next: time + 1}},
	}
	for _, tt := range tests {
		if have := NewID(tt.config, genesis, 0, time); have != tt.want {
			t.Fatalf("incorrect forkid hash: have %x, want %x", have, tt.want)
		}
	}
}
