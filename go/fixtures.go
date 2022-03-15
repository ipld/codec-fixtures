package codec_fixtures

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	_ "github.com/alexjg/go-dag-jose/dagjose"
	"github.com/ipfs/go-cid"
	_ "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
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
var dagJoseLp = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,
	Codec:    0x85, // "dag-jose"
	MhType:   0x12, // "sha2-256"
	MhLength: 32,
}}
var codecs = map[codecName]ipld.LinkPrototype{
	"dag-pb":   dagPbLp,
	"dag-cbor": dagCborLp,
	"dag-json": dagJsonLp,
	"dag-jose": dagJoseLp,
}
var linkSystem = cidlink.DefaultLinkSystem()

func loadFixture(dir string) (fixtureSet, error) {
	files, err := ioutil.ReadDir("../fixtures/" + dir)
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
		byts, err := ioutil.ReadFile("../fixtures/" + dir + "/" + file.Name())
		if err != nil {
			return fixtures, err
		}
		ext = strings.TrimLeft(ext, ".")
		na := basicnode.Prototype.Any.NewBuilder()
		lp, ok := codecs[ext]
		if !ok {
			fmt.Printf("unknown codec '%v' for fixture '%v'\n", ext, dir)
		}
		decoder, err := linkSystem.DecoderChooser(lp.BuildLink(make([]byte, 32)))
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

func nodeToCid(lp ipld.LinkPrototype, node ipld.Node) (cid.Cid, error) {
	encoder, err := linkSystem.EncoderChooser(lp)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("could not choose an encoder: %v", err)
	}
	hasher, err := linkSystem.HasherChooser(lp)
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
