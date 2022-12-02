mod utils;

use libipld::{
    cid::Cid,
    codec::Codec,
    ipld::Ipld,
    multihash::{Code, MultihashDigest},
    IpldCodec,
};

/// Mapping between string identifiers and actual codecs.
struct Codecs;

// Currently the only codec that is implemented based on Serde is DAG-CBOR. Until others have been
// implemented, use the non-Serde based ones.
impl Codecs {
    /// Map codec strings to actual codecs.
    fn get(codec: &str) -> IpldCodec {
        match codec {
            "dag-json" => IpldCodec::DagJson,
            "dag-pb" => IpldCodec::DagPb,
            _ => panic!("Unknown codec"),
        }
    }
}

#[test]
fn codec_fixtures() {
    for dir in utils::fixture_directories("fixtures") {
        if utils::skip_test(&dir) {
            continue;
        }

        let fixture_name = dir
            .path()
            .file_stem()
            .expect("Directory must have a name")
            .to_os_string()
            .to_str()
            .expect("Filename must be valid UTF-8")
            .to_string();
        println!("Testing fixture {}", fixture_name);
        let fixtures = utils::load_fixtures(dir);
        for from_fixture in &fixtures {
            // Take a fixture of one codec and…
            let decoded: Ipld = match &from_fixture.codec[..] {
                "dag-cbor" => {
                    serde_ipld_dagcbor::from_slice(&from_fixture.bytes).expect("Decoding must work")
                }
                _ => Codecs::get(&from_fixture.codec)
                    .decode(&from_fixture.bytes)
                    .expect("Decoding must work"),
            };

            // …transcode it into any other fixture.
            for to_fixture in &fixtures {
                let (codec_code, data) = match &to_fixture.codec[..] {
                    "dag-cbor" => (
                        0x71,
                        serde_ipld_dagcbor::to_vec(&decoded).expect("Encoding must work"),
                    ),
                    _ => {
                        let codec = Codecs::get(&to_fixture.codec);
                        (
                            codec.into(),
                            codec.encode(&decoded).expect("Encoding must work"),
                        )
                    }
                };
                let digest = Code::Sha2_256.digest(&data);
                let cid = Cid::new_v1(codec_code, digest);
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

        let encode_fixtures = utils::load_negative_fixtures(codec_dir.path(), "encode");
        for fixture in encode_fixtures {
            println!(
                "Testing negative encode fixture for {}: {}",
                codec_name, fixture.name
            );
            let error = match &codec_name[..] {
                "dag-cbor" => serde_ipld_dagcbor::to_vec(&fixture.dag_json.unwrap()).is_err(),
                _ => Codecs::get(&codec_name)
                    .encode(&fixture.dag_json.unwrap())
                    .is_err(),
            };
            if !error {
                assert!(false, "Did not error")
            }
        }

        let decode_fixtures = utils::load_negative_fixtures(codec_dir.path(), "decode");
        for fixture in decode_fixtures {
            println!(
                "Testing negative decode fixture for {}: {}",
                codec_name, fixture.name
            );
            let error = match &codec_name[..] {
                "dag-cbor" => {
                    serde_ipld_dagcbor::from_slice::<Ipld>(&fixture.hex.unwrap()).is_err()
                }
                _ => Codecs::get(&codec_name)
                    .decode::<Ipld>(&fixture.hex.unwrap())
                    .is_err(),
            };
            if !error {
                assert!(false, "Did not error")
            }
        }
    }
}
