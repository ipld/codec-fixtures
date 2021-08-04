package codec_fixtures

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
)

func TestCodecs(t *testing.T) {
	dirs, err := ioutil.ReadDir("../fixtures/")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range dirs {
		fixtureName := dir.Name()
		if !dir.IsDir() {
			t.Fatalf("%v is not a directory", fixtureName)
		}
		if reason, blacklisted := FixtureBlacklist[fixtureName]; blacklisted {
			fmt.Printf("Skipping fixture '%v': %v\n", fixtureName, reason)
			continue
		}
		t.Run(fixtureName, func(t *testing.T) {
			data, err := loadFixture(fixtureName)
			if err != nil {
				t.Fatal(err)
			}
			verifyCid(t, "decode(dag-json)->encode(dag-json)", data.dagjson.value, dagJsonLp, data.dagjson.cid)
			verifyCid(t, "decode(dag-json)->encode(dag-cbor)", data.dagjson.value, dagCborLp, data.dagcbor.cid)
			verifyCid(t, "decode(dag-cbor)->encode(dag-json)", data.dagcbor.value, dagJsonLp, data.dagjson.cid)
			verifyCid(t, "decode(dag-cbor)->encode(dag-cbor)", data.dagcbor.value, dagCborLp, data.dagcbor.cid)
		})
	}
}

func verifyCid(t *testing.T, desc string, node ipld.Node, toEnc ipld.LinkPrototype, expected cid.Cid) {
	actual, err := nodeToCid(toEnc, node)
	if err != nil {
		t.Fatal(err)
	}
	if !expected.Equals(actual) {
		t.Fatalf("[%v] generated CID (%v) does not match expected (%v)", desc, expected.String(), actual.String())
	}
}
