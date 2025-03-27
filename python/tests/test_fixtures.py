from pathlib import Path

from ipld_dag_pb import encode, decode, code
from multiformats import CID, multihash
import pytest


FIXTURES_DIR = Path(__file__).parents[2] / "fixtures"
REPO_ROOT = FIXTURES_DIR.parent
CODECS = {
    "dag-pb": {"name": "dag-pb", "code": code, "encode": encode, "decode": decode}
}


def get_dagpb_fixtures() -> list[str]:
    """Get all dag-pb fixtures directories"""
    fixtures = []
    for item in FIXTURES_DIR.iterdir():
        # get only subdirectories with atleast dag-pb file in it
        if item.is_dir() and any(f.suffix == '.dag-pb' for f in item.iterdir()):
            fixtures.append(item.name)
    return fixtures


def load_fixtures(fixture_name: str) -> dict[str, dict[str, str | bytes]]:
    """Load a fixture directory and return the data for each codec"""
    fixture_dir = FIXTURES_DIR / fixture_name
    data = {}

    for file in fixture_dir.iterdir():
        ext = file.suffix[1:]  # remove the leading dot
        if ext not in CODECS:
            continue

        cid_str = file.stem
        with open(file, 'rb') as f:
            bytes_data = f.read()

        data[ext] = {"cid": cid_str, "bytes": bytes_data}
    return data


def bytes_to_cid(data: bytes, codec_code: int = code) -> str:
    """Convert bytes to a CID using the sha2-256 hash function"""
    mh = multihash.digest(data, "sha2-256")
    return str(CID(base="base32", version=1, codec=codec_code, digest=mh))


@pytest.mark.parametrize("fixture_name", get_dagpb_fixtures())
def test_fixture(fixture_name):
    """Test a fixture by decoding and re-encoding it in different formats"""
    data = load_fixtures(fixture_name)

    # Skip if no dag-pb format is available for this fixture
    if "dag-pb" not in data:
        pytest.skip(f"No dag-pb format available for fixture {fixture_name}")

    for from_codec_name, from_data in data.items():
        from_codec = CODECS.get(from_codec_name)

        if not from_codec: continue

        decoded = from_codec["decode"](from_data["bytes"])

        for to_codec_name, to_codec in CODECS.items():
            if to_codec_name not in data:
                continue

            encoded = to_codec["encode"](decoded)

            actual_cid = bytes_to_cid(codec_code=to_codec["code"], data=encoded)
            expected_cid = data[to_codec_name]["cid"]

            assert actual_cid == expected_cid, (
                f"CID mismatch for {fixture_name}: "
                f"decode({from_codec_name})->encode({to_codec_name})\n"
                f"Expected: {expected_cid}\n"
                f"Actual: {actual_cid}"
            )
