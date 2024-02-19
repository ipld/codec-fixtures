mod utils;

use ipld_core::{cid::Cid, codec::Codec, ipld::Ipld};
use ipld_dagpb::{DagPbCodec, Error as DagPbError};
use multihash_codetable::{Code, MultihashDigest};
use serde_ipld_dagcbor::{codec::DagCborCodec, error::CodecError as DagCborError};
use serde_ipld_dagjson::{codec::DagJsonCodec, error::CodecError as DagJsonError};
use thiserror::Error;

/// Mapping between string identifiers and actual codecs.
#[allow(clippy::enum_variant_names)]
enum IpldCodec {
    DagCbor,
    DagJson,
    DagPb,
}

#[allow(clippy::enum_variant_names)]
#[derive(Debug, Error)]
enum Error {
    #[error("DAG-CBOR error: {0}")]
    DagCbor(#[from] DagCborError),
    #[error("DAG-JSON error: {0}")]
    DagJson(#[from] DagJsonError),
    #[error("DAG-PB error: {0}")]
    DagPb(#[from] DagPbError),
}

impl IpldCodec {
    /// Map codec strings to actual codecs.
    fn new(codec: &str) -> Self {
        match codec {
            "dag-cbor" => Self::DagCbor,
            "dag-json" => Self::DagJson,
            "dag-pb" => Self::DagPb,
            _ => panic!("Unknown codec"),
        }
    }

    /// Encode some IPLD object into bytes with the codec the enum represents.
    fn encode(&self, ipld: &Ipld) -> Result<Vec<u8>, Error> {
        match self {
            Self::DagCbor => Ok(DagCborCodec::encode_to_vec(ipld)?),
            Self::DagJson => Ok(DagJsonCodec::encode_to_vec(ipld)?),
            Self::DagPb => Ok(DagPbCodec::encode_to_vec(ipld)?),
        }
    }

    /// Decode some byte of the codec the enum represents into IPLD object.
    fn decode(&self, bytes: &[u8]) -> Result<Ipld, Error> {
        match self {
            Self::DagCbor => Ok(DagCborCodec::decode_from_slice(bytes)?),
            Self::DagJson => Ok(DagJsonCodec::decode_from_slice(bytes)?),
            Self::DagPb => Ok(DagPbCodec::decode_from_slice(bytes)?),
        }
    }
}

impl From<IpldCodec> for u64 {
    fn from(codec: IpldCodec) -> Self {
        match codec {
            IpldCodec::DagCbor => <DagCborCodec as Codec<Ipld>>::CODE,
            IpldCodec::DagJson => <DagJsonCodec as Codec<Ipld>>::CODE,
            IpldCodec::DagPb => <DagPbCodec as Codec<Ipld>>::CODE,
        }
    }
}

#[test]
fn codec_fixtures() {
    for dir in utils::fixture_directories("fixtures") {
        let fixture_name = dir
            .path()
            .file_name()
            .expect("Directory must have a name")
            .to_os_string()
            .to_str()
            .expect("Filename must be valid UTF-8")
            .to_string();
        println!("Testing fixture {}", fixture_name);
        let fixtures = utils::load_fixtures(&dir);
        for from_fixture in &fixtures {
            if utils::skip_test(&dir, &from_fixture.codec) {
                continue;
            }

            // Take a fixture of one codec and…
            let decoded: Ipld = IpldCodec::new(&from_fixture.codec)
                .decode(&from_fixture.bytes)
                .expect("Decoding must work");

            // …transcode it into any other fixture.
            for to_fixture in &fixtures {
                if utils::skip_test(&dir, &to_fixture.codec) {
                    continue;
                }
                let codec = IpldCodec::new(&to_fixture.codec);
                let data = codec.encode(&decoded).expect("Encoding must work");
                let digest = Code::Sha2_256.digest(&data);
                let cid = Cid::new_v1(codec.into(), digest);
                assert_eq!(
                    cid, to_fixture.cid,
                    "CIDs match for the data decoded from {} encoded as {}",
                    from_fixture.codec, to_fixture.codec
                );
            }
        }
    }
}

#[test]
fn negative_fixtures() {
    for codec_dir in utils::fixture_directories("negative-fixtures") {
        let codec_name = codec_dir
            .file_name()
            .to_str()
            .expect("Codec names are valid UTF-8")
            .to_string();
        let codec = IpldCodec::new(&codec_name);

        let encode_fixtures = utils::load_negative_fixtures(codec_dir.path(), "encode");
        for fixture in encode_fixtures {
            println!(
                "Testing negative encode fixture for {}: {}",
                codec_name, fixture.name
            );

            // The `fixture.dag_json` is already decoded into an `Ipld` object, so we can use it
            // directly to encode the fixtures.
            if codec.encode(&fixture.dag_json.unwrap()).is_ok() {
                panic!("did not error");
            }
        }

        let decode_fixtures = utils::load_negative_fixtures(codec_dir.path(), "decode");
        for fixture in decode_fixtures {
            println!(
                "Testing negative decode fixture for {}: {}",
                codec_name, fixture.name
            );
            if codec.decode(&fixture.hex.unwrap()).is_ok() {
                panic!("did not error");
            }
        }
    }
}
