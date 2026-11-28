"""Create settings only in a fresh, package-owned portable Terminal directory."""
import json
from pathlib import Path

PLAIN = "{a81a7af0-f002-4a1e-982a-202611280001}"
WORKBENCH = "{a81a7af0-f002-4a1e-982a-202611280002}"


def write_settings(root: Path) -> str:
    settings = root / "settings"
    settings.mkdir()  # Refuse to overwrite settings from a used host.
    shell = r"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe"
    # Reuse the complete, already recorded host palette, not a partial scheme.
    reference = Path(__file__).resolve().parents[2] / "docs/testing/results/2026-11-25-minimal-workspace-evidence/local/iterm-reference.json"
    scheme = json.loads(reference.read_text(encoding="utf-8"))["scheme"]
    doc = {
        "$schema": "https://aka.ms/terminal-profiles-schema",
        "language": "en-US",
        "defaultProfile": PLAIN,
        "initialCols": 120,
        "initialRows": 32,
        "confirmCloseAllTabs": True,
        "newTabMenu": [{"type": "profile", "profile": PLAIN}],
        "profiles": {
            "defaults": {
                "font": {"face": "Consolas", "size": 10},
                "colorScheme": scheme["name"],
                "startingDirectory": "%USERPROFILE%",
            },
            "list": [
                {"guid": PLAIN, "name": "Terminal", "tabTitle": "Terminal",
                 "commandline": f'"{shell}" -NoLogo -NoProfile',
                 "suppressApplicationTitle": True, "hidden": False},
                {"guid": WORKBENCH, "name": "Agent Session Desk",
                 "commandline": f'"{shell}" -NoLogo -NoProfile', "hidden": True},
            ],
        },
        "schemes": [scheme],
    }
    (settings / "settings.json").write_text(json.dumps(doc, indent=2) + "\n", encoding="utf-8")
    (root / ".portable").touch(exist_ok=False)
    return WORKBENCH
