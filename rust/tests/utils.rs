use std::env;
use std::fs::{self, DirEntry};
use std::path::PathBuf;

use ipld_core::{cid::Cid, ipld::Ipld};

static FIXTURE_SKIPLIST: [(&str, &str, &str); 1] = [
    // This test is skipped as the current DAG-JSON decoder decodes such an integer into a float.
    (
        "dag-json",
        "int--11959030306112471732",
        "integer out of int64 range",
    ),
];

/// Contents of a single fixture.
#[derive(Debug)]
pub struct Fixture {
    pub codec: String,
    pub cid: Cid,
    pub bytes: Vec<u8>,
}

/// Contents of a single negative fixture.
#[derive(Debug)]
pub struct NegativeFixture {
    pub name: String,
    /// Negative decode fixtures use hex values.
    pub hex: Option<Vec<u8>>,
    /// Negative encode fixtures use DAG-JSON values.
    pub dag_json: Option<Ipld>,
}

/// Returns all fixtures from a directory.
pub fn load_fixtures(dir: &DirEntry) -> Vec<Fixture> {
    fs::read_dir(dir.path())
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
pub fn fixture_directories(name: &str) -> Vec<DirEntry> {
    let rust_dir = env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR must be set");
    let mut fixtures_dir = PathBuf::from(rust_dir);
    fixtures_dir.push(format!("../{}", name));

    // Only take directories, exclude files
    fs::read_dir(&fixtures_dir)
        .expect("Cannot open fixtures directory")
        .filter_map(Result::ok)
        .filter(|dir| dir.path().is_dir())
        .collect()
}

/// Returns true if a test fixture is on the skip list
pub fn skip_test(dir: &DirEntry, codec: &str) -> bool {
    for (skip_codec, name, reason) in FIXTURE_SKIPLIST {
        if codec == skip_codec
            && dir
                .path()
                .into_os_string()
                .to_str()
                .unwrap()
                .ends_with(name)
        {
            eprintln!("Skipping {} fixture '{}': {}", codec, name, reason);
            return true;
        }
    }
    false
}

/// Returns all test fixtures from the given directory.
///
/// `en_or_decode` specifies the directory name for encode/decode fixtures, hence should either be
/// `encode` or `decode`
pub fn load_negative_fixtures(mut dir: PathBuf, en_or_decode: &str) -> Vec<NegativeFixture> {
    dir.push(en_or_decode);
    if let Ok(read_dir) = fs::read_dir(&dir) {
        read_dir
            .filter_map(|file| {
                // Filter out invalid files.
                let file = file.ok()?;

                let path = file.path();
                let bytes = fs::read(path).expect("File must be able to be read");
                // Use DAG-JSON for parsing, so we don't need and extra JSON parser.
                let ipld: Ipld = serde_ipld_dagjson::from_slice(&bytes).expect("It's valid JSON");

                if let Ipld::List(list) = ipld {
                    let fixtures: Vec<_> = list
                        .iter()
                        .map(|fixture| {
                            if let Ok(Some(Ipld::String(name))) = fixture.get("name") {
                                let dag_json = fixture
                                    .get("dag-json")
                                    .expect("dag-json field exists")
                                    .cloned();
                                let hex = if let Ok(Some(Ipld::String(data))) = fixture.get("hex") {
                                    Some(hex::decode(data).unwrap())
                                } else {
                                    None
                                };
                                NegativeFixture {
                                    name: name.to_string(),
                                    hex,
                                    dag_json,
                                }
                            } else {
                                panic!("Negative fixture has no name");
                            }
                        })
                        .collect();
                    Some(fixtures)
                } else {
                    None
                }
            })
            .flatten()
            .collect()
    } else {
        Vec::new()
    }
}
