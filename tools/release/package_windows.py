#!/usr/bin/env python3
"""Package verified inputs offline; never build, launch, download or publish."""
import argparse
import hashlib
import json
import re
import shutil
import stat
import zipfile
from pathlib import Path, PurePosixPath

from terminal_profile import write_settings

ROOT = Path(__file__).resolve().parents[2]
TERMINAL_SHA256 = "273d6f1502eca1b5bd95c43f5d4d7ec72cc503df340102f8659717e5bead3437"


def sha(path):
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(32768), b""):
            h.update(block)
    return h.hexdigest()


def put_json(path, doc):
    path.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def copy(src, dest):
    if src.is_symlink() or not src.is_file():
        raise ValueError("Expected a regular input file")
    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(src, dest)


def extract_terminal(archive, dest, expected=TERMINAL_SHA256):
    if expected is not None and sha(archive) != expected:
        raise ValueError("Official Terminal archive hash mismatch")
    with zipfile.ZipFile(archive) as z:
        entries, seen, total = [], set(), 0
        prefix = None
        for item in z.infolist():
            p = PurePosixPath(item.filename)
            if (p.is_absolute() or "\\" in item.filename or ":" in item.filename
                    or ".." in p.parts or stat.S_ISLNK(item.external_attr >> 16)):
                raise ValueError("Unsafe Terminal archive member")
            if not p.parts or (prefix is not None and p.parts[0] != prefix):
                raise ValueError("Terminal archive must have one root")
            prefix = p.parts[0]
            if item.is_dir():
                continue
            name = PurePosixPath(*p.parts[1:])
            if (not name.parts or name.parts[0].lower() in {"settings", ".portable"}
                    or str(name).lower() in seen or item.file_size > 64 << 20):
                raise ValueError("Non-pristine Terminal archive")
            seen.add(str(name).lower())
            total += item.file_size
            entries.append((item, name))
        if total > 384 << 20 or len(entries) > 2000:
            raise ValueError("Terminal archive exceeds bound")
        if not {"windowsterminal.exe", "notice.html"} <= seen:
            raise ValueError("Terminal executable or official notices missing")
        dest.mkdir()
        for item, name in entries:
            target = dest.joinpath(*name.parts)
            target.parent.mkdir(parents=True, exist_ok=True)
            with z.open(item) as src, target.open("xb") as out:
                shutil.copyfileobj(src, out, 32768)


def build_info(src, revision):
    text = src.read_text(encoding="utf-8")
    if ("vcs.revision=" + revision not in text or "vcs.modified=false" not in text
            or "GOOS=windows" not in text or "GOARCH=amd64" not in text):
        raise ValueError("Clean matching Windows build metadata required")
    return "portable-program:\n" + "\n".join(text.splitlines()[1:]) + "\n"


def package(args):
    c = json.loads(args.contract.read_text(encoding="utf-8-sig"))
    if (c["protocol"] != 1 or c["platform"] != "windows/amd64"
            or c["product"] != "asd-windows-amd64.exe"
            or not re.fullmatch(r"v[0-9A-Za-z._-]{1,62}", c["tag"])
            or not re.fullmatch(r"[0-9a-f]{40}", c["revision"])
            or sha(args.program) != c["sha256"]
            or args.program.stat().st_size != c["size"]):
        raise ValueError("Verified native update contract required")
    product_info = build_info(args.build_info, c["revision"])
    launcher_info = build_info(args.launcher_build_info, c["revision"])
    expected_modules = {(row["module"], row["version"]) for row in
                        json.loads((args.licenses / "go-modules.json").read_text())}
    actual_modules = {tuple(line.split()[1:3]) for line in product_info.splitlines()
                      if line.split() and line.split()[0] == "dep"}
    if actual_modules != expected_modules:
        raise ValueError("Go module notice set differs from actual product")
    args.output.mkdir()  # Fresh output only; failures keep their evidence.
    stage = args.output / "package"
    stage.mkdir()
    for src, name in [(args.program, "asd.exe"), (args.launcher, "asd-launch.exe"),
                      (args.contract, "asd-update.json"),
                      (ROOT / "tools/release/asd.cmd", "asd.cmd"),
                      (ROOT / "tools/release/Start.cmd", "Start.cmd"),
                      (ROOT / "docs/WINDOWS_PORTABLE.md", "README.md"),
                      (ROOT / "THIRD_PARTY_NOTICES.md", "THIRD_PARTY_NOTICES.md")]:
        copy(src, stage / name)
    runtime = c["compatibility"]["runtime"]
    if set(runtime) != {"psmux.exe", "conpty.dll", "OpenConsole.exe"}:
        raise ValueError("Unexpected runtime file set")
    for name, expected in runtime.items():
        src = args.runtime / name
        if sha(src) != expected:
            raise ValueError("Pinned terminal runtime mismatch")
        copy(src, stage / "terminal-runtime" / name)
    for src in sorted(args.licenses.rglob("*")):
        if src.is_file():
            copy(src, stage / "licenses" / src.relative_to(args.licenses))
    host = stage / "terminal-host"
    extract_terminal(args.terminal_zip, host)
    profile = write_settings(host)
    copy(ROOT / "third_party/windows-terminal/LICENSE", stage / "licenses/windows-terminal/LICENSE")
    put_json(stage / "local-host.json", {
        "terminal": {"path": "terminal-host/WindowsTerminal.exe", "sha256": sha(host / "WindowsTerminal.exe")},
        "profileGuid": profile,
    })
    (stage / "build-info.txt").write_text(product_info, encoding="utf-8")
    (stage / "launcher-build-info.txt").write_text(launcher_info, encoding="utf-8")
    put_json(stage / "alpha.json", {
        "schema_version": 1, "version": c["tag"], "entrypoint": "asd.exe",
        "ui_source_commit": c["revision"], "namespace": "asd-release-alpha2",
        "files": {name: sha(stage / name) for name in ["asd.exe", "asd-launch.exe"]},
    })
    files = {p.relative_to(stage).as_posix(): sha(p)
             for p in sorted(stage.rglob("*")) if p.is_file()}
    put_json(stage / "manifest.json", {
        "schema_version": 1, "version": c["tag"], "platform": c["platform"],
        "product_source_commit": c["revision"], "signed": False,
        "terminal_archive_sha256": TERMINAL_SHA256,
        "acceptance": "Experimental prerelease; bounded Windows evidence only",
        "files": files,
    })
    archive = args.output / ("asd-windows-x64-" + c["tag"].removeprefix("v") + ".zip")
    with zipfile.ZipFile(archive, "x", zipfile.ZIP_DEFLATED, compresslevel=9) as z:
        for src in sorted(stage.rglob("*")):
            if src.is_file():
                info = zipfile.ZipInfo(src.relative_to(stage).as_posix(), (2026, 11, 28, 0, 0, 0))
                info.compress_type = zipfile.ZIP_DEFLATED
                info.external_attr = 0o100644 << 16
                z.writestr(info, src.read_bytes())
    for src, name in [(args.program, "asd-windows-amd64.exe"), (args.contract, "asd-update.json"),
                      (stage / "manifest.json", "manifest.json"), (args.source_zip, "psmux-source.zip")]:
        copy(src, args.output / name)
    assets = [archive] + [args.output / n for n in ["asd-windows-amd64.exe", "asd-update.json", "manifest.json", "psmux-source.zip"]]
    (args.output / "SHA256SUMS").write_text("".join(sha(p) + "  " + p.name + "\n" for p in assets), encoding="ascii")
    print(json.dumps({"files": len(files) + 1, "zip_bytes": archive.stat().st_size, "zip_sha256": sha(archive)}))


if __name__ == "__main__":
    p = argparse.ArgumentParser(description=__doc__)
    for name in ["program", "launcher", "contract", "runtime", "licenses", "terminal-zip", "source-zip",
                 "build-info", "launcher-build-info", "output"]:
        p.add_argument("--" + name, type=Path, required=True)
    package(p.parse_args())
