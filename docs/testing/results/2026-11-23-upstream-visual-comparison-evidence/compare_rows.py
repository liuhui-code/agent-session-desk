"""Compare original terminal rows with UTF-8 reference replay UIA text."""
import json
import unicodedata
from pathlib import Path

root = Path(__file__).resolve().parent
results = []
for path in sorted((root / "utf8-replay/desktop/responses").glob("*.json")):
    receipt = json.loads(path.read_text())
    command = receipt["command"]
    if command.get("kind") != "capture":
        continue
    name = command["label"].removeprefix("reference-replay-").removesuffix("-public")
    expected = (root / "frames" / f"{name}.txt").read_text().splitlines()
    actual = receipt["state"]["text"].replace("\r\n", "\n").split("\n")[1:]

    def normalize(value):
        return unicodedata.normalize("NFC", value.rstrip(" "))

    differences = []
    for row in range(max(len(expected), len(actual))):
        source = normalize(expected[row]) if row < len(expected) else None
        rendered = normalize(actual[row]) if row < len(actual) else None
        if source != rendered:
            differences.append({"row": row, "source": source, "rendered": rendered})
    results.append({"frame": name, "expected_rows": len(expected),
                    "uia_rows": len(actual), "differences": differences})
assert len(results) == 8, "Expected exactly eight replay captures"
print(json.dumps(results, ensure_ascii=False, indent=2))
