import hashlib
import json
from pathlib import Path
import re

ROOT = Path('/private/tmp/asd-group-rename-20261123')
DEST = Path.cwd() / 'docs/testing/results/2026-11-23-group-rename-evidence'
DEST.mkdir(parents=True, exist_ok=False)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def redact(text):
    text = text.replace(str(Path.cwd()), '<REPO>')
    text = text.replace('<LOCAL_TEST_HOME>', '<LOCAL_TEST_HOME>')
    return re.sub('<WINDOWS_TEST_USER>', '<WINDOWS_TEST_USER>', text, flags=re.I)


entries = []


def archive(source, relative):
    raw = source.read_bytes()
    data = raw
    transformation = 'unchanged PNG'
    if source.suffix != '.png':
        encoding = 'utf-16' if raw.startswith((b'\xff\xfe', b'\xfe\xff')) else 'utf-8-sig'
        text = raw.decode(encoding)
        data = redact(text).replace('\r\n', '\n').encode('utf-8')
        transformation = encoding + ' to UTF-8/LF; known test users and local paths redacted'
    target = DEST / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(data)
    entries.append({'path': str(relative), 'original_sha256': digest(raw),
                    'archived_sha256': digest(data), 'bytes': len(data),
                    'transformation': transformation})


for path in sorted((ROOT / 'native').rglob('*')):
    if path.is_file():
        assert path.suffix in ('.json', '.log', '.ps1', '.png', '.toml')
        archive(path, Path('native') / path.relative_to(ROOT / 'native'))

names = [
    'store-red.log', 'store-green.log', 'manager-red.log', 'ui-red.log',
    'rename-green.log', 'rename-regression.log', 'rename-regression-final.log',
    'ack-red.log', 'ack-green.log', 'go-test.log', 'go-race.log', 'go-vet.log',
    'go-mod-verify.log', 'catalog-check.log', 'source-limits.log',
    'upstream-reference.log', 'product-reference.log', 'upstream-source.json',
    'source-fingerprints.json', 'binary-hashes.json', 'upstream-rename-frames.json',
    'product-rename-frames.json', 'rename-geometry.json', 'native-run.log',
    'product-build-info.txt',
    'run-group-rename-tests.ps1', 'setup-group-rename-desktop.ps1',
    'start-group-rename-setup.ps1', 'start-group-rename-desktop.ps1',
    'group-rename-desktop-driver.ps1', 'group-rename-flow.ps1',
    'preflight-group-rename.ps1', 'desktop-preflight.log', 'desktop-setup.log',
    'desktop-start.log', 'desktop-flow.log', 'cleanup-export-group-rename-first.ps1',
    'cleanup-export-group-rename.ps1', 'cleanup-export-final.log',
    'inspect-group-rename-cleanup-first.ps1', 'harness-failures.json',
    'upstream-rename-reference_test.go', 'product-rename-reference_test.go',
    'archive-group-rename.py', 'verify-group-rename.py', 'verify-group-rename-first.py',
]
for name in names:
    relative = name + '.txt' if name.endswith('.go') else name
    archive(ROOT / name, Path('local') / relative)

# Keep reference harness inputs self-contained; remap placeholders when rerunning.
old = Path('/private/tmp/asd-groups-desktop-20261123')
for name in ['upstream-unused-tests.go', 'upstream-reference-main.go']:
    archive(old / name, Path('local') / (name + '.txt'))
for name in ['upstream-rename-overlay.json', 'product-rename-overlay.json']:
    archive(ROOT / name, Path('local') / name)
archive(Path('/private/tmp/asd-upstream-compare-20261123/reference.sb'), Path('local/reference.sb'))

manifest = {
    'base_commit': '5f2e310', 'product_source_state': 'base plus archived source fingerprints',
    'zip_sha256': digest((ROOT / 'native-export.zip').read_bytes()),
    'zip_bytes': (ROOT / 'native-export.zip').stat().st_size,
    'scope': 'Group rename slice; Windows native tests and two ordinary shell desktop flow; not full acceptance',
    'redaction': 'Known test usernames, repository and home paths only; original PNGs unchanged',
    'artifacts': entries,
}
(DEST / 'manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({'artifacts': len(entries), 'zip_sha256': manifest['zip_sha256']}))
