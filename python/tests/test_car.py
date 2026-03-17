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
                fixture_cid = file.stem
                with file.open(mode="rb") as file_obj:
                    fixture_data = file_obj.read()
                fixture_blocks.append((CID.decode(fixture_cid), fixture_data,))

    return fixture_blocks


@pytest.fixture
def car_fixture_data() -> bytes:
    """Load CAR fixture data from CAR file: `fixtures.car`"""
    with open(REPO_ROOT / "fixtures.car", mode="rb") as f:
        return f.read()


def test_car_decode(car_fixture_data):
    decoded_roots, decoded_blocks = ipld_car.decode(car_fixture_data)
    assert decoded_roots == []
    assert (
        sorted([(block[0].encode(), block[1],) for block in load_all_fixtures()]) ==
        sorted([(block[0].encode(), block[1],) for block in decoded_blocks])
    )


def test_car_encode(car_fixture_data):
    # encoded blocks from the fixtures dir into CAR
    fixture_blocks = sorted(load_all_fixtures(), key=lambda block: block[0].encode())
    encoded_fixture_blocks_car = ipld_car.encode(roots=[], blocks=fixture_blocks)
    assert encoded_fixture_blocks_car == car_fixture_data
    # verify same blocks are present in encoded-then-decoded-fixture-blocks
    # and decoded car fixture file
