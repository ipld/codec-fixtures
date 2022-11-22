package codec_fixtures

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipld/go-ipld-prime/linking"
	"github.com/ipld/go-ipld-prime/multicodec"

	"github.com/ipfs/go-cid"
	_ "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/vulcanize/go-codec-dageth/header"
	"github.com/vulcanize/go-codec-dageth/log"
	"github.com/vulcanize/go-codec-dageth/log_trie"
	"github.com/vulcanize/go-codec-dageth/rct"
	"github.com/vulcanize/go-codec-dageth/rct_trie"
	account "github.com/vulcanize/go-codec-dageth/state_account"
	"github.com/vulcanize/go-codec-dageth/state_trie"
	"github.com/vulcanize/go-codec-dageth/storage_trie"
	"github.com/vulcanize/go-codec-dageth/tx"
	"github.com/vulcanize/go-codec-dageth/tx_trie"
	"github.com/vulcanize/go-codec-dageth/uncles"
)

type codecName = string

type codecFixture struct {
	codec codecName
	cid   cid.Cid
	value ipld.Node
}

type fixtureSet = map[codecName]codecFixture

var dagPbLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x70, // "dag-pb"
	MhType:   0x12, // "sha2-256"
	MhLength: 32,
}}
var dagCborLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x71, // "dag-cbor"
	MhType:   0x12, // "sha2-256"
	MhLength: 32,
}}
var dagJsonLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x0129, // "dag-json"
	MhType:   0x12,   // "sha2-256"
	MhLength: 32,
}}
var ethHeaderLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x90, // "eth-block"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethUnclesLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x91, // "eth-block-list"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethTxLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x93, // "eth-tx"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethTxTrieLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x92, // "eth-tx-trie"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethRctLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x95, // "eth-tx-receipt"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethRctTrieLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x94, // "eth-tx-receipt-trie"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethLogLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x9a, // "eth-receipt-log"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethLogTrieLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x99, // "eth-receipt-log-trie"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethStateTrieLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x96, // "eth-state-trie"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethStorageTrieLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x98, // "eth-storage-trie"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var ethStateAccountLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x97, // "eth-account-snapshot"
	MhType:   0x1b, // "keccak-256"
	MhLength: 32,
}}
var codecs = map[codecName]ipld.LinkPrototype{
	"dag-pb":               dagPbLp,
	"dag-cbor":             dagCborLp,
	"dag-json":             dagJsonLp,
	"eth-block":            ethHeaderLp,
	"eth-block-list":       ethUnclesLp,
	"eth-tx":               ethTxLp,
	"eth-tx-trie":          ethTxTrieLp,
	"eth-tx-receipt":       ethRctLp,
	"eth-tx-receipt-trie":  ethRctTrieLp,
	"eth-receipt-log":      ethLogLp,
	"eth-receipt-log-trie": ethLogTrieLp,
	"eth-account-snapshot": ethStateAccountLp,
	"eth-state-trie":       ethStateTrieLp,
	"eth-storage-trie":     ethStorageTrieLp,
}
var ethCodecs = map[codecName]ipld.LinkPrototype{
	"eth-block":            ethHeaderLp,
	"eth-block-list":       ethUnclesLp,
	"eth-tx":               ethTxLp,
	"eth-tx-trie":          ethTxTrieLp,
	"eth-tx-receipt":       ethRctLp,
	"eth-tx-receipt-trie":  ethRctTrieLp,
	"eth-receipt-log":      ethLogLp,
	"eth-receipt-log-trie": ethLogTrieLp,
	"eth-account-snapshot": ethStateAccountLp,
	"eth-state-trie":       ethStateTrieLp,
	"eth-storage-trie":     ethStorageTrieLp,
}
var defaultLinkSystem = cidlink.DefaultLinkSystem()

func setupEthLinkSystem() linking.LinkSystem {
	ethRegistry := multicodec.Registry{}
	ethRegistry.RegisterDecoder(0x90, header.Decode)
	ethRegistry.RegisterDecoder(0x91, uncles.Decode)
	ethRegistry.RegisterDecoder(0x92, tx_trie.Decode)
	ethRegistry.RegisterDecoder(0x93, tx.Decode)
	ethRegistry.RegisterDecoder(0x94, rct_trie.Decode)
	ethRegistry.RegisterDecoder(0x95, rct.Decode)
	ethRegistry.RegisterDecoder(0x96, state_trie.Decode)
	ethRegistry.RegisterDecoder(0x97, account.Decode)
	ethRegistry.RegisterDecoder(0x98, storage_trie.Decode)
	ethRegistry.RegisterDecoder(0x99, log_trie.Decode)
	ethRegistry.RegisterDecoder(0x9a, log.Decode)

	ethRegistry.RegisterEncoder(0x90, header.Encode)
	ethRegistry.RegisterEncoder(0x91, uncles.Encode)
	ethRegistry.RegisterEncoder(0x92, tx_trie.Encode)
	ethRegistry.RegisterEncoder(0x93, tx.Encode)
	ethRegistry.RegisterEncoder(0x94, rct_trie.Encode)
	ethRegistry.RegisterEncoder(0x95, rct.Encode)
	ethRegistry.RegisterEncoder(0x96, state_trie.Encode)
	ethRegistry.RegisterEncoder(0x97, account.Encode)
	ethRegistry.RegisterEncoder(0x98, storage_trie.Encode)
	ethRegistry.RegisterEncoder(0x99, log_trie.Encode)
	ethRegistry.RegisterEncoder(0x9a, log.Encode)

	return cidlink.LinkSystemUsingMulticodecRegistry(ethRegistry)
}

var rootFixturePath = "../fixtures/"
var rootKeccak256FixturePath = "../keccak256_fixtures/"

func loadFixture(rootPath, dir string, codecMap map[codecName]ipld.LinkPrototype, ls linking.LinkSystem) (fixtureSet, error) {
	files, err := os.ReadDir(rootPath + dir)
	fixtures := make(fixtureSet)
	if err != nil {
		return fixtures, err
	}
	for _, file := range files {
		if file.IsDir() {
			return fixtures, fmt.Errorf("%v is a directory", file.Name())
		}
		// fmt.Printf("Loading file %v\n", file.Name())
		ext := filepath.Ext(file.Name())
		cid, err := cid.Decode(strings.TrimSuffix(file.Name(), ext))
		if err != nil {
			return fixtures, err
		}
		byts, err := os.ReadFile(rootPath + dir + "/" + file.Name())
		if err != nil {
			return fixtures, err
		}
		ext = strings.TrimLeft(ext, ".")
		na := basicnode.Prototype.Any.NewBuilder()
		lp, ok := codecMap[ext]
		if !ok {
			fmt.Printf("unknown codec '%v' for fixture '%v'\n", ext, dir)
		}
		decoder, err := ls.DecoderChooser(lp.BuildLink(make([]byte, 32)))
		if err != nil {
			return fixtures, err
		}
		err = decoder(na, bytes.NewReader(byts))
		if err != nil {
			return fixtures, err
		}
		fixtures[ext] = codecFixture{
			codec: ext,
			cid:   cid,
			value: na.Build(),
		}
	}
	return fixtures, nil
}

func nodeToCid(ls linking.LinkSystem, lp ipld.LinkPrototype, node ipld.Node) (cid.Cid, error) {
	encoder, err := ls.EncoderChooser(lp)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("could not choose an encoder: %v", err)
	}
	hasher, err := ls.HasherChooser(lp)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("could not choose a hasher: %v", err)
	}
	err = encoder(node, hasher)
	if err != nil {
		return cid.Cid{}, err
	}
	lnk := lp.BuildLink(hasher.Sum(nil))
	cidLink, ok := lnk.(cidlink.Link)
	if !ok {
		return cid.Cid{}, err
	}
	return cidLink.Cid, nil
}
