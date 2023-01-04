# IPLD Codec Fixtures

[![Test against latest libraries](https://github.com/ipld/codec-fixtures/actions/workflows/cron.yml/badge.svg)](https://github.com/ipld/codec-fixtures/actions/workflows/cron.yml)

This repository contains fixtures for standard IPLD codecs. It is used to verify the correctness, compatibility and limitations of IPLD implementations.

## What?

The [fixtures](./fixtures/) directory contains a suite of test data, where each subdirectory comprises an encoded IPLD block in the formats that are supported for that data. A file containing the binary encoded form of that block has the name `<CID>.<codec-name>`, where the `CID` is the CIDv1 using a SHA2-256 multihash of the block for that codec. The `codec-name` is the standard codec name as found in the [multicodec table](https://github.com/multiformats/multicodec/blob/master/table.csv).

Implementations are expected to be able to:

1. Read and decode the IPLD block from these files
2. Re-encode the IPLD block using the supported codecs
3. Compare the CID of the re-encoded block to the expected CID as per the filename for the given codec

Since the block is encoded in different forms, by re-encoding each decoded form into the different codecs, we are able to test the correctness of the decoding as well as the encoding process. Where the CIDs do not match, there has been a problem either in decoding or encoding the data. If the same error occurs for the data loaded from differently encoded blocks, the error is likely to be with the encoding process. If the error only occurs when re-encoding from a single encoded form of the block then the error may be with the decoding process.

## Adding fixtures

The [_fixtures_src](./_fixtures_src/) directory contains the source of each of fixtures contained in the [fixtures](./fixtures/) directory. Each file in [_fixtures_src](./_fixtures_src/) contains an encoded form of a block using one of the supported codecs. The name of the file is `<fixture-name>.<codec-name>`. The [js/make-fixtures.js](./js/make-fixtures.js) program (run with `node js/make-fixtures.js`) is used to generate the fixtures in [fixtures](./fixtures/) for each of the source files.

Fixture generation uses the JavaScript stack for generating data, but this is not a requirement. If you would like to add fixtures and would like to create them manually, or add an alternative mechanism for generating fixtures from source then please do so.

## Negative Fixtures

Separately to the _positive_ fixtures, there is a [negative-fixtures](./negative-fixtures/) directory that contains failure conditions for codecs. These don't require a compile step, so are not included in the _fixtures_src directory but are edited directly.

The structure for negative fixtures subdirectories is: `negative-fixtures/<codec name>/{encode,decode}/<fixtures>.json`. Where `<codec name>` is the canonical name of the codec being tested. `encode` and `decode` are fixtures that test the encode and decode paths of a codec, and `<fixture>` is an arbitrary name containing one or more fixture within the JSON structure.

Test runners will load as manu fixture files as exist in these subdirectories and extract and run the cases defined there.

### Encode

Negative fixtures for an encode phase involve defining a data model form that should cause an error when passed through an `Encode()` for that codec. The data model form is instantiated by running the fixture data through an existing IPLD codec's decoder (mostly DAG-JSON data embedded within the JSON fixture document) and then attempting to pass that data model form into the codec and inspecting the error. Error messages may or may not be matched in some way, depending on the complexity of the implementation—it is more important that a failure occur than the error is exact.

### Decode

Negative fixtures for a decode phase involve loading a block's bytes from a hex form from the fixture data and passing those bytes through a `Decode()` for that codec and inspecting the error. Error messages may or may not be matched in some way, depending on the complexity of the implementation—it is more important that a failure occur than the error is exact.

## Implementations & Codecs

### Go

Fixtures are tested against the [go-ipld-prime](https://github.com/ipld/go-ipld-prime) stack:

* DAG-JSON: [github.com/ipld/go-ipld-prime/codec/dagjson](https://pkg.go.dev/github.com/ipld/go-ipld-prime/codec/dagjson)
* DAG-CBOR: [github.com/ipld/go-ipld-prime/codec/dagcbor](https://pkg.go.dev/github.com/ipld/go-ipld-prime/codec/dagcbor)
* DAG-PB:  [github.com/ipld/go-codec-dagpb](https://pkg.go.dev/github.com/ipld/go-codec-dagpb)

### JavaScript

Fixtures are tested against the [js-multiformats](https://github.com/multiformats/js-multiformats) stack:

* DAG-CBOR: [@ipld/dag-cbor](https://github.com/ipld/js-dag-cbor)
* DAG-JSON: [@ipld/dag-json](https://github.com/ipld/js-dag-cbor)
* DAG-PB: [@ipld/dag-pb](https://github.com/ipld/js-dag-pb)

### Rust

Fixtures are tested against the [libipld](https://github.com/ipld/libipld) stack:

* DAG-CBOR: [libipld-cbor](https://crates.io/crates/libipld-cbor)
* DAG-JSON: [libipld-json](https://crates.io/crates/libipld-json)

## Running tests

### JavaScript

```
make testjs
```

Or, in the [js](./js/) directory, run:

```
npm install
npm test
```

## Go

```
make testgo
```

Or, in the [go](./go/) directory, run:

```
go test
```

## Rust

```
make testrust
```

Or in the [rust](./rust/) directory, run:

```
cargo test -- --nocapture
```

## Generating testmark output for ipld.io

Each codec tested here has a corresponding file in the https://github.com/ipld/ipld repository which generates the https://ipld.io website containing the fixture data in [testmark](https://github.com/warpfork/go-testmark) format. The filename per codec is is `specs/codecs/<CODEC>/fixtures/cross-codec/index.md`.

The [js/make-testmark.js](js/make-testmark.js) program can be used to update those files when the data is updated here. Run it with `node js/make-testmark.js <path/to/ipld/ipld/repository>`.

## License

Licensed under either of

 * Apache 2.0, ([LICENSE-APACHE](LICENSE-APACHE) / http://www.apache.org/licenses/LICENSE-2.0)
 * MIT ([LICENSE-MIT](LICENSE-MIT) / http://opensource.org/licenses/MIT)

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in the work by you, as defined in the Apache-2.0 license, shall be dual licensed as above, without any additional terms or conditions.
