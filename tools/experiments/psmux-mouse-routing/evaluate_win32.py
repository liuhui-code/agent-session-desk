"""Offline review of the pinned (5,4)->(9,4) native mouse fixture.

This checks complete recorded intervals, not OS delivery or hardware identity.
Exit 0 means the supplied trace matches; it is not product acceptance.
"""

import argparse
import json
from pathlib import Path
import sys


def evaluate(rows: list, intervals: list) -> dict:
    if not isinstance(rows, list) or not isinstance(intervals, list) or not rows or not intervals:
        raise ValueError("records and reviewed intervals must be nonempty")
    for index, row in enumerate(rows, 1):
        if not isinstance(row, dict) or any(type(row.get(key)) is not int for key in
               ("n", "x", "y", "buttons", "control", "flags")):
            raise ValueError("all record fields must be integers")
        if row["n"] != index:
            raise ValueError("record sequence has a gap, duplicate or reordering")
    previous = 0
    cases = []
    for part in intervals:
        keys = ("begin_exclusive", "end_inclusive", "button", "mask")
        if not isinstance(part, dict) or any(type(part.get(key)) is not int for key in keys):
            raise ValueError("interval fields must be integers")
        begin, end = part["begin_exclusive"], part["end_inclusive"]
        if begin != previous or not begin < end <= len(rows):
            raise ValueError("intervals must cover the full trace exactly once")
        if part["button"] not in (0, 1, 2) or part["mask"] not in (0, 8, 16):
            raise ValueError("unsupported fixture button/modifier")
        button = (1, 4, 2)[part["button"]]
        control = {0: 0, 8: 2, 16: 8}[part["mask"]]
        expected = [(5, 4, button, control, 0),
                    (9, 4, button, control, 1),
                    (9, 4, 0, control, 0)]
        actual = [tuple(row[key] for key in ("x", "y", "buttons", "control", "flags"))
                  for row in rows[begin:end]]
        # Positioning motions before press are not button transitions. Consume
        # only neutral motion at the fixture start; never filter after press.
        offset = 0
        while offset < len(actual) and actual[offset] == (5, 4, 0, 0, 1):
            offset += 1
        actual = actual[offset:]
        cases.append({"button": part["button"], "mask": part["mask"],
                      "strict_once": actual == expected,
                      "expected": expected, "actual": actual})
        previous = end
    if previous != len(rows):
        raise ValueError("unreviewed trailing records")
    return {"passed": sum(case["strict_once"] for case in cases),
            "total": len(cases), "cases": cases,
            "scope": "offline trace review; no new native execution"}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("packets", type=Path, help="native reader NDJSON")
    parser.add_argument("intervals", type=Path, help="complete interval JSON array")
    args = parser.parse_args()
    try:
        rows = [json.loads(line) for line in args.packets.read_text(encoding="utf-8-sig").splitlines()]
        intervals = json.loads(args.intervals.read_text(encoding="utf-8-sig"))
        result = evaluate(rows, intervals)
    except (OSError, ValueError, TypeError, KeyError) as error:
        print(f"invalid evidence: {error}", file=sys.stderr)
        return 2
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if result["passed"] == result["total"] else 1


if __name__ == "__main__":
    sys.exit(main())
