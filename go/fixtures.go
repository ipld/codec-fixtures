package codec_fixtures

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

type codecFixture struct {
	cid   cid.Cid
	value ipld.Node
}

type fixtureSet struct {
	dagcbor codecFixture
	dagjson codecFixture
}

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
var linkSystem = cidlink.DefaultLinkSystem()

func loadFixture(dir string) (fixtureSet, error) {
	files, err := ioutil.ReadDir("../fixtures/" + dir)
	fixtures := fixtureSet{}
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
		if ext == "dag-json" {
			err := dagjson.Decode(na, bytes.NewReader(byts))
			if err != nil {
				return fixtures, err
			}
			fixtures.dagjson = codecFixture{
				cid:   cid,
				value: na.Build(),
			}
		} else if ext == "dag-cbor" {
			err := dagcbor.Decode(na, bytes.NewReader(byts))
			if err != nil {
				return fixtures, err
			}
			fixtures.dagcbor = codecFixture{
				cid:   cid,
				value: na.Build(),
			}
		} else {
			fmt.Printf("unknown codec '%v' for fixture '%v'\n", ext, dir)
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
