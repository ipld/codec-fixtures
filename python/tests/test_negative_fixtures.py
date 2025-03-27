from pathlib import Path
import json

import pytest
from ipld_dag_pb import encode, decode


NEGATIVE_FIXTURES_DIR = Path(__file__).parents[2] / "negative-fixtures/dag-pb"


def load_negative_encode_fixtures():
    """Load negative encode fixtures for dag-pb"""
    fixtures = []
    encode_dir = NEGATIVE_FIXTURES_DIR / "encode"

    for file in encode_dir.iterdir():
        with open(file, "r") as f:
            data = json.load(f)
            for fixture in data:
                fixtures.append((fixture["name"], fixture,))

    return fixtures


def load_negative_decode_fixtures():
    """Load negative decode fixtures for dag-pb"""
    fixtures = []
    decode_dir = NEGATIVE_FIXTURES_DIR / "decode"

    for file in decode_dir.iterdir():
        with open(file, "r") as f:
            data = json.load(f)
            for fixture in data:
                fixtures.append((fixture["name"], fixture))

    return fixtures


@pytest.mark.parametrize("name, fixture", load_negative_encode_fixtures())
def test_negative_encode(name, fixture):
    """Test that encoding invalid input produces the expected error"""
    dag_data = fixture.get("dag-json")

    with pytest.raises(Exception):
        encode(dag_data)


@pytest.mark.parametrize("name, fixture", load_negative_decode_fixtures())
def test_negative_decode(name, fixture):
    """Test that decoding invalid input produces the expected error"""
    hex_data = fixture.get("hex")
    bytes_data = bytes.fromhex(hex_data)

    with pytest.raises(Exception):
        decode(bytes_data)
