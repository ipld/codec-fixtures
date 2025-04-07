
from typing import TypeAlias
from multiformats.varint import decode
import pytest
from pathlib import Path
from multiformats import CID
import ipld_car

FIXTURES_DIR = Path(__file__).parents[2] / "fixtures"
REPO_ROOT = FIXTURES_DIR.parent


def load_all_fixtures() -> list[ipld_car.Block]:
    """Load all non-negative fixtures CID and data"""
    fixture_blocks: list[ipld_car.Block] = []

    for dir in sorted(FIXTURES_DIR.iterdir()):
        if dir.is_dir():  # skip .gitattributes file
            for file in dir.iterdir():
                fixture_cid = CID.decode(file.stem)
                with file.open(mode="rb") as file_obj:
                    fixture_data = file_obj.read()
                fixture_blocks.append((fixture_cid, fixture_data,))

    return fixture_blocks


@pytest.fixture
def car_fixture_data() -> bytes:
    """Load CAR fixture data from CAR file"""
    with open(REPO_ROOT / "fixtures.car", mode="rb") as f:
        return f.read()


@pytest.mark.parametrize("fixture_block", load_all_fixtures())
def test_car_decode(fixture_block, car_fixture_data):
    decoded_roots, decoded_blocks = ipld_car.decode(car_fixture_data)
    assert decoded_roots == []
    assert len(decoded_blocks) == 273
    assert fixture_block in decoded_blocks


def test_car_encode(car_fixture_data):
    encoded_car = ipld_car.encode(roots=[], blocks=load_all_fixtures())
    assert encoded_car == car_fixture_data
