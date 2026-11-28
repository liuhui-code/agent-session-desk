"""Create settings only in a fresh, package-owned portable Terminal directory."""
import json
from pathlib import Path

PLAIN = "{a81a7af0-f002-4a1e-982a-202611280001}"
WORKBENCH = "{a81a7af0-f002-4a1e-982a-202611280002}"


def write_settings(root: Path) -> str:
    settings = root / "settings"
    settings.mkdir()  # Refuse to overwrite settings from a used host.
    shell = r"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe"
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
                "colorScheme": "ASD Classic",
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
        "schemes": [{"name": "ASD Classic", "background": "#000000",
                     "foreground": "#BFBFBF", "cursorColor": "#BFBFBF",
                     "selectionBackground": "#4D4D4D"}],
    }
    (settings / "settings.json").write_text(json.dumps(doc, indent=2) + "\n", encoding="utf-8")
    (root / ".portable").touch(exist_ok=False)
    return WORKBENCH
