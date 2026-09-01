package codec_fixtures

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/multicodec"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

func TestFixtures(t *testing.T) {
	dirs, err := os.ReadDir("../fixtures/")
	if err != nil {
		t.Fatalf("failed to open fixtures dir: %v", err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		fixtureName := dir.Name()
		if reason, blacklisted := FixtureBlacklist[fixtureName]; blacklisted {
			fmt.Printf("Skipping fixture '%v': %v\n", fixtureName, reason)
			continue
		}
		t.Run(fixtureName, func(t *testing.T) {
			data, err := loadFixture(fixtureName)
			if err != nil {
				t.Fatalf("failed to load fixture: %v", err)
			}
			for fromCodec := range data {
				for toCodec := range data {
					msg := fmt.Sprintf("decode(%v)->encode(%v)", fromCodec, toCodec)
					verifyCid(t, msg, data[fromCodec].value, codecs[toCodec], data[toCodec].cid)
				}
			}
		})
	}
}

func verifyCid(t *testing.T, desc string, node ipld.Node, toEnc ipld.LinkPrototype, expected cid.Cid) {
	actual, err := nodeToCid(toEnc, node)
	if err != nil {
		t.Fatalf("failed to convert node to CID: %v", err)
	}
	if !expected.Equals(actual) {
		t.Fatalf("[%v] generated CID (%v) does not match expected (%v)", desc, expected.String(), actual.String())
	}
}

func TestNegatigeFixtures(t *testing.T) {
	dirs, err := os.ReadDir("../negative-fixtures/")
	if err != nil {
		t.Fatalf("failed to open negative fixtures dir: %v", err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		codecName := dir.Name()
		t.Run(codecName, func(t *testing.T) {
			t.Run("encode", func(t *testing.T) {
				files, err := os.ReadDir(filepath.Join("../negative-fixtures/", codecName, "encode"))
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						return // ignore missing
					}
					t.Fatalf("failed to open negative fixtures dir: %v", err)
				}
				for _, file := range files {
					if file.IsDir() {
						continue
					}
					fixtureData, err := os.ReadFile(filepath.Join("../negative-fixtures/", codecName, "encode", file.Name()))
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							return // ignore missing
						}
						t.Fatalf("failed to read fixture data: %v", err)
					}
					var fixtures []negativeFixtureEncode
					err = json.Unmarshal(fixtureData, &fixtures)
					if err != nil {
						t.Fatalf("failed to decode fixture data: %v", err)
					}
					for _, fixture := range fixtures {
						fixtureName := fmt.Sprintf("%s/encode/%s", codecName, fixture.Name)
						if reason, blacklisted := FixtureBlacklist[fixtureName]; blacklisted {
							fmt.Printf("Skipping fixture '%v': %v\n", fixtureName, reason)
							continue
						}
						t.Run(fixture.Name, testNegativeFixtureEncode(codecName, fixture))
					}
				}
			})

			t.Run("decode", func(t *testing.T) {
				files, err := os.ReadDir(filepath.Join("../negative-fixtures/", codecName, "decode"))
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						return // ignore missing
					}
					t.Fatalf("failed to open negative fixtures dir: %v", err)
				}
				for _, file := range files {
					if file.IsDir() {
						continue
					}
					fixtureData, err := os.ReadFile(filepath.Join("../negative-fixtures/", codecName, "decode", file.Name()))
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							return // ignore missing
						}
						t.Fatalf("failed to read fixture data: %v", err)
					}
					var fixtures []negativeFixtureDecode
					err = json.Unmarshal(fixtureData, &fixtures)
					if err != nil {
						t.Fatalf("failed to decode fixture data: %v", err)
					}
					for _, fixture := range fixtures {
						fixtureName := fmt.Sprintf("%s/decode/%s", codecName, fixture.Name)
						if reason, blacklisted := FixtureBlacklist[fixtureName]; blacklisted {
							fmt.Printf("Skipping fixture '%v': %v\n", fixtureName, reason)
							continue
						}
						t.Run(fixture.Name, testNegativeFixtureDecode(codecName, fixture))
					}
				}
			})
		})
	}
}

// create a test function an individual negative test fixture for encode
func testNegativeFixtureEncode(codecName string, fixture negativeFixtureEncode) func(t *testing.T) {
	return func(t *testing.T) {
		dagJsonDecoder, err := multicodec.DefaultRegistry.LookupDecoder(dagJsonLp.Codec)
		if err != nil {
			t.Fatalf("could not choose a dag-pb encoder: %v", err)
		}

		// construct the data model form to encode from the dag-json data in the fixture
		nb := basicnode.Prototype.Any.NewBuilder()
		byts, err := json.Marshal(fixture.DagJson)
		if err != nil {
			t.Fatalf("failed to encode dag-json fixture data")
		}
		dagJsonDecoder(nb, bytes.NewReader(byts))
		node := nb.Build()

		// look up encoder to test
		encoder, err := linkSystem.EncoderChooser(codecs[codecName])
		if err != nil {
			t.Fatalf("could not choose an encoder: %v", err)
		}

		// encode, should error
		var buf bytes.Buffer
		err = encoder(node, &buf)
		if err == nil {
			t.Errorf("should error on encode")
		} else if !strings.EqualFold(err.Error(), fixture.Error) {
			t.Logf("error mismatch: [%s] ~= [%s]", err.Error(), fixture.Error)
		}
	}
}

// create a test function an individual negative test fixture for decode
func testNegativeFixtureDecode(codecName string, fixture negativeFixtureDecode) func(t *testing.T) {
	return func(t *testing.T) {
		byts, err := hex.DecodeString(fixture.Hex)
		if err != nil {
			t.Fatal(err)
		}

		// look up decoder to test
		decoder, err := multicodec.DefaultRegistry.LookupDecoder(codecs[codecName].(cidlink.LinkPrototype).Codec)
		if err != nil {
			t.Fatalf("could not choose a dag-pb encoder: %v", err)
		}

		// decode, should error
		nb := basicnode.Prototype.Any.NewBuilder()
		err = decoder(nb, bytes.NewReader(byts))
		if err == nil {
			t.Errorf("should error on encode")
		} else if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(fixture.Error)) {
			t.Errorf("error mismatch: [%s] ~= [%s]", err.Error(), fixture.Error)
		}
	}
}

// TestNoncanonicalFixtures verifies blocks that decode successfully but are
// not in canonical form, so they can never round-trip byte-for-byte: decoding
// must succeed, the decoded node must equal the node decoded from the
// canonical bytes, and canonical re-encoding must produce the canonical CID.
//
// TODO: once codec implementations expose an opt-in encoder for alternate
// forms (e.g. the dag-pb Data-first field order proposed by IPIP-550,
// https://github.com/ipfs/specs/pull/550), promote these to full round-trip
// fixtures with a per-fixture encode hint.
func TestNoncanonicalFixtures(t *testing.T) {
	dirs, err := os.ReadDir("../noncanonical-fixtures/")
	if err != nil {
		t.Fatalf("failed to open noncanonical fixtures dir: %v", err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		codecName := codecName(dir.Name())
		t.Run(string(codecName), func(t *testing.T) {
			files, err := os.ReadDir(filepath.Join("../noncanonical-fixtures/", string(codecName), "decode"))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return
				}
				t.Fatalf("failed to open noncanonical fixtures dir: %v", err)
			}
			for _, file := range files {
				fixtureData, err := os.ReadFile(filepath.Join("../noncanonical-fixtures/", string(codecName), "decode", file.Name()))
				if err != nil {
					t.Fatalf("failed to read noncanonical fixture file: %v", err)
				}
				var fixtures []noncanonicalFixtureDecode
				if err := json.Unmarshal(fixtureData, &fixtures); err != nil {
					t.Fatalf("failed to parse noncanonical fixture file: %v", err)
				}
				for _, fixture := range fixtures {
					t.Run(fixture.Name, testNoncanonicalFixtureDecode(codecName, fixture))
				}
			}
		})
	}
}

func testNoncanonicalFixtureDecode(codecName codecName, fixture noncanonicalFixtureDecode) func(t *testing.T) {
	return func(t *testing.T) {
		byts, err := hex.DecodeString(fixture.Hex)
		if err != nil {
			t.Fatalf("failed to parse fixture hex: %v", err)
		}
		canonicalByts, err := hex.DecodeString(fixture.CanonicalHex)
		if err != nil {
			t.Fatalf("failed to parse fixture canonicalHex: %v", err)
		}
		expectedCid, err := cid.Decode(fixture.CanonicalCid)
		if err != nil {
			t.Fatalf("failed to parse fixture canonicalCid: %v", err)
		}

		node, err := decodeBytes(codecName, byts)
		if err != nil {
			t.Fatalf("failed to decode noncanonical block: %v", err)
		}
		canonicalNode, err := decodeBytes(codecName, canonicalByts)
		if err != nil {
			t.Fatalf("failed to decode canonical block: %v", err)
		}
		// Logical equality is verified through canonical form: both decodes
		// must re-encode to the same canonical CID. ipld.DeepEqual is not
		// used here because the two wire forms decode to maps whose entry
		// order differs, which DeepEqual treats as unequal.
		verifyCid(t, "reencode(noncanonical)", node, codecs[codecName], expectedCid)
		verifyCid(t, "reencode(canonical)", canonicalNode, codecs[codecName], expectedCid)
	}
}
