use std::env;
use std::fs::{self, DirEntry};
use std::path::PathBuf;

use libipld::{
    block::Block, cid::Cid, codec::Codec, ipld::Ipld, multihash::Code, store::DefaultParams,
    IpldCodec,
};

static FIXTURE_SKIPLIST: [(&str, &str); 16] = [
    ("int--11959030306112471732", "integer out of int64 range"),
    (
        "dagpb_11unnamedlinks+data",
        "DAG-PB isn't fully compatible yet",
    ),
    ("dagpb_1link", "DAG-PB isn't fully compatible yet"),
    ("dagpb_2link+data", "DAG-PB isn't fully compatible yet"),
    (
        "dagpb_4namedlinks+data",
        "DAG-PB isn't fully compatible yet",
    ),
    (
        "dagpb_7unnamedlinks+data",
        "DAG-PB isn't fully compatible yet",
    ),
    ("dagpb_Data_zero", "DAG-PB isn't fully compatible yet"),
    ("dagpb_empty", "DAG-PB isn't fully compatible yet"),
    ("dagpb_Links_Hash_some", "DAG-PB isn't fully compatible yet"),
    (
        "dagpb_Links_Hash_some_Name_some",
        "DAG-PB isn't fully compatible yet",
    ),
    (
        "dagpb_Links_Hash_some_Name_zero",
        "DAG-PB isn't fully compatible yet",
    ),
    (
        "dagpb_Links_Hash_some_Tsize_some",
        "DAG-PB isn't fully compatible yet",
    ),
    (
        "dagpb_Links_Hash_some_Tsize_zero",
        "DAG-PB isn't fully compatible yet",
    ),
    ("dagpb_simple_forms_2", "DAG-PB isn't fully compatible yet"),
    ("dagpb_simple_forms_3", "DAG-PB isn't fully compatible yet"),
    ("dagpb_simple_forms_4", "DAG-PB isn't fully compatible yet"),
];

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
            .expect("Filenome must be valid UTF-8")
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
                let block = Block::<DefaultParams>::encode(
                    Codecs::get(&to_fixture.codec),
                    Code::Sha2_256,
                    &decoded,
                )
                .expect("Encoding must work");
                assert_eq!(
                    block.cid(),
                    &to_fixture.cid,
                    "CIDs match for the data decoded from {} encoded as {}",
                    from_fixture.codec,
                    to_fixture.codec
                );
            }
        }
    }
}
