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
		}
		if !strings.EqualFold(err.Error(), fixture.Error) {
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
