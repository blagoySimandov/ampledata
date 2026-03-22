import json
import os
import pytest


def pytest_addoption(parser):
    parser.addoption(
        "--save-output",
        action="store",
        nargs="?",
        const="crawl_output.json",
        default=None,
        metavar="FILE",
        help="Save crawl results to FILE (default: crawl_output.json)",
    )


@pytest.fixture(autouse=True)
def save_crawl_output(request):
    output_file = request.config.getoption("--save-output")
    collected: dict[str, list[str]] = {}

    yield collected

    if output_file is None or not collected:
        return

    existing = {}
    if os.path.exists(output_file):
        with open(output_file) as f:
            existing = json.load(f)

    existing.update(collected)
    with open(output_file, "w") as f:
        json.dump(existing, f, indent=2)
