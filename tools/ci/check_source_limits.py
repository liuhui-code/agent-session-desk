#!/usr/bin/env python3
"""Enforce the repository's 500 physical lines per production source file rule."""
from pathlib import Path
import sys


def violations(root: Path) -> list[tuple[str, int]]:
    result = []
    for source in sorted(root.rglob("*")):
        if source.suffix not in {".go", ".py", ".ps1"} or source.name.endswith("_test.go") or any(
                part in {".git", "vendor"} for part in source.relative_to(root).parts):
            continue
        count = len(source.read_text(encoding="utf-8").splitlines())
        if count > 500:
            result.append((str(source.relative_to(root)), count))
    return result


def main() -> int:
    root = Path(__file__).resolve().parents[2]
    failures = violations(root)
    for path, count in failures:
        print(f"{path}: {count} lines (maximum 500)", file=sys.stderr)
    if not failures:
        print("Production source size: all files <= 500 lines")
    return int(bool(failures))


if __name__ == "__main__":
    sys.exit(main())
