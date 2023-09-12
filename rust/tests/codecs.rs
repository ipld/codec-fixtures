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

impl Codecs {
    /// Map codec strings to actual codecs.
    fn get(codec: &str) -> IpldCodec {
        match codec {
            "dag-cbor" => IpldCodec::DagCbor,
            "dag-json" => IpldCodec::DagJson,
            "dag-pb" => IpldCodec::DagPb,
            _ => panic!("Unknown codec"),
        }
    }
}

#[test]
fn codec_fixtures() {
    for dir in utils::fixture_directories("fixtures") {
        let fixture_name = dir
            .path()
            .file_stem()
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
            let decoded: Ipld = Codecs::get(&from_fixture.codec)
                .decode(&from_fixture.bytes)
                .expect("Decoding must work");

            // …transcode it into any other fixture.
            for to_fixture in &fixtures {
                if utils::skip_test(&dir, &to_fixture.codec) {
                    continue;
                }

                let codec = Codecs::get(&to_fixture.codec);
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
        let codec = Codecs::get(&codec_name);

        let encode_fixtures = utils::load_negative_fixtures(codec_dir.path(), "encode");
        for fixture in encode_fixtures {
            println!(
                "Testing negative encode fixture for {}: {}",
                codec_name, fixture.name
            );
            match codec.encode(&fixture.dag_json.unwrap()) {
                Ok(_) => assert!(false, "did not error"),
                Err(_) => assert!(true),
            }
        }

        let decode_fixtures = utils::load_negative_fixtures(codec_dir.path(), "decode");
        for fixture in decode_fixtures {
            println!(
                "Testing negative decode fixture for {}: {}",
                codec_name, fixture.name
            );
            match codec.decode::<Ipld>(&fixture.hex.unwrap()) {
                Ok(_) => assert!(false, "did not error"),
                Err(_) => assert!(true),
            }
        }
    }
}
