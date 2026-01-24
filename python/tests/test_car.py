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
    """Load CAR fixture data from CAR file: `fixtures.car`"""
    with open(REPO_ROOT / "fixtures.car", mode="rb") as f:
        return f.read()


@pytest.mark.parametrize("fixture_block", load_all_fixtures())
def test_car_decode(fixture_block, car_fixture_data):
    decoded_roots, decoded_blocks = ipld_car.decode(car_fixture_data)
    assert decoded_roots == []
    assert len(decoded_blocks) == 273
    assert fixture_block in decoded_blocks


def test_car_encode(car_fixture_data):
    # encoded blocks from the fixtures dir into CAR
    fixture_blocks = load_all_fixtures()
    encoded_fixture_blocks_car = ipld_car.encode(roots=[], blocks=fixture_blocks)

    decoded_car_fixture_blocks_root, decoded_car_fixture_blocks = ipld_car.decode(
        encoded_fixture_blocks_car
    )
    decoded_car_fixture_file_root, decoded_car_fixture_file = ipld_car.decode(
        car_fixture_data
    )

    assert decoded_car_fixture_blocks_root == decoded_car_fixture_file_root

    # verify same blocks are present in encoded-then-decoded-fixture-blocks
    # and decoded car fixture file
    assert len(decoded_car_fixture_blocks) == len(decoded_car_fixture_file)
    assert (
        set(block[0] for block in decoded_car_fixture_blocks) ==
        set(block[0] for block in decoded_car_fixture_file)
    )

    # verify content by CID
    encoded_blocks_dict = {block[0]: block[1] for block in decoded_car_fixture_blocks}
    fixture_data_dict = {block[0]: block[1] for block in decoded_car_fixture_file}

    for cid, data in fixture_data_dict.items():
        assert cid in encoded_blocks_dict
        assert encoded_blocks_dict[cid] == data
