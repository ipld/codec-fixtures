from pathlib import Path
import json

import pytest
from ipld_dag_pb import encode, decode, code
from multiformats import CID, multihash


NONCANONICAL_FIXTURES_DIR = Path(__file__).parents[2] / "noncanonical-fixtures/dag-pb"

# TODO: once codec implementations expose an opt-in encoder for alternate
# forms (e.g. the dag-pb Data-first field order proposed by IPIP-550,
# https://github.com/ipfs/specs/pull/550), promote these to full round-trip
# fixtures with a per-fixture encode hint.


def bytes_to_cid(data: bytes) -> str:
    """Convert bytes to a dag-pb CIDv1 using the sha2-256 hash function"""
    mh = multihash.digest(data, "sha2-256")
    return str(CID(base="base32", version=1, codec=code, digest=mh))


def load_noncanonical_decode_fixtures():
    """Load non-canonical decode fixtures for dag-pb"""
    fixtures = []
    decode_dir = NONCANONICAL_FIXTURES_DIR / "decode"
    if not decode_dir.is_dir():
        return fixtures

    for file in decode_dir.iterdir():
        with open(file, "r") as f:
            for fixture in json.load(f):
                fixtures.append((fixture["name"], fixture))

    return fixtures


@pytest.mark.parametrize("name, fixture", load_noncanonical_decode_fixtures())
def test_noncanonical_decode(name, fixture):
    """Non-canonical blocks decode successfully, equal the canonical decode,
    and canonical re-encoding produces the canonical CID"""
    value = decode(bytes.fromhex(fixture["hex"]))
    canonical_value = decode(bytes.fromhex(fixture["canonicalHex"]))
    assert value == canonical_value

    reencoded = bytes(encode(value))
    assert bytes_to_cid(reencoded) == fixture["canonicalCid"]
