use std::env;
use std::fs::{self, DirEntry};
use std::path::PathBuf;

use libipld::{
    cid::Cid,
    codec::Codec,
    ipld::Ipld,
    multihash::{Code, MultihashDigest},
    IpldCodec,
};

static FIXTURE_SKIPLIST: [(&str, &str); 1] =
    [("int--11959030306112471732", "integer out of int64 range")];

/// Contents of a single fixture.
#[derive(Debug)]
struct Fixture {
    codec: String,
    cid: Cid,
    bytes: Vec<u8>,
}

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

/// Returns all fixtures from a directory.
fn load_fixture(dir: DirEntry) -> Vec<Fixture> {
    fs::read_dir(&dir.path())
        .unwrap()
        .filter_map(|file| {
            // Filter out invalid files.
            let file = file.ok()?;

            let path = file.path();
            let extension = path
                .extension()
                .expect("Filename must have an extension")
                .to_os_string()
                .into_string()
                .expect("Extension must be valid UTF-8");
            let cid = path
                .file_stem()
                .expect("Filename must have a name")
                .to_os_string()
                .into_string()
                .expect("Filename must be valid UTF-8");
            let bytes = fs::read(&path).expect("File must be able to be read");

            Some(Fixture {
                codec: extension,
                cid: Cid::try_from(cid.clone()).expect("Filename must be a valid Cid"),
                bytes,
            })
        })
        .collect()
}

/// Returns the paths to all directories that contain fixtures.
fn fixture_directories() -> Vec<DirEntry> {
    let rust_dir = env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR must be set");
    let mut fixtures_dir = PathBuf::from(rust_dir);
    fixtures_dir.push("../fixtures");

    // Only take directories, exclude files
    fs::read_dir(&fixtures_dir)
        .expect("Cannot open fixtures directory")
        .filter_map(Result::ok)
        .filter(|dir| dir.path().is_dir())
        .collect()
}

/// Returns true if a test fixture is on the skip list
fn skip_test(dir: &DirEntry) -> bool {
    for (name, reason) in FIXTURE_SKIPLIST {
        if dir
            .path()
            .into_os_string()
            .to_str()
            .unwrap()
            .ends_with(name)
        {
            eprintln!("Skipping fixture '{}': {}", name, reason);
            return true;
        }
    }
    false
}

#[test]
fn codec_fixtures() {
    for dir in fixture_directories() {
        if skip_test(&dir) {
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
        let fixtures = load_fixture(dir);
        for from_fixture in &fixtures {
            // Take a fixture of one codec and…
            let decoded: Ipld = Codecs::get(&from_fixture.codec)
                .decode(&from_fixture.bytes)
                .expect("Decoding must work");

            // …transcode it into any other fixture.
            for to_fixture in &fixtures {
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
