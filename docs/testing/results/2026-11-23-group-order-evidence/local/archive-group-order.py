"""Archive this run's allowlisted inputs; original runtime files stay private."""
import hashlib
import json
from pathlib import Path
import re

ROOT = Path('/private/tmp/asd-group-order-20261123')
DEST = Path.cwd() / 'docs/testing/results/2026-11-23-group-order-evidence'
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
        data = redact(raw.decode(encoding)).replace('\r\n', '\n').encode('utf-8')
        transformation = encoding + ' to UTF-8/LF; known local/test usernames and paths redacted'
    out = DEST / relative
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_bytes(data)
    entries.append({'path': str(relative), 'original_sha256': digest(raw),
                    'archived_sha256': digest(data), 'bytes': len(data),
                    'transformation': transformation})


for path in sorted((ROOT / 'native').rglob('*')):
    if path.is_file():
        assert path.suffix in ('.json', '.log', '.ps1', '.png', '.toml')
        archive(path, Path('native') / path.relative_to(ROOT / 'native'))

names = [
    'store-red.log', 'store-green.log', 'ties-red.log', 'ties-green.log',
    'manager-red.log', 'ui-red.log', 'ui-behavior-red.log', 'vertical-green.log',
    'failure-red.log', 'failure-green.log', 'quit-red.log', 'quit-green.log',
    'order-regression.log', 'order-regression-final.log', 'order-final.log',
    'go-test.log', 'go-race.log', 'go-vet.log', 'go-mod-verify.log',
    'catalog-check.log', 'source-limits.log', 'product-build-info.txt',
    'upstream-reference.log', 'upstream-source.json', 'source-fingerprints.json',
    'binary-hashes.json', 'upstream-order-snapshots.json',
    'upstream-order-reference_test.go', 'upstream-order-overlay.json',
    'upstream-unused-tests.go', 'upstream-reference-main.go',
    'native-run.log', 'run-group-order-tests.ps1', 'setup-group-order-desktop.ps1',
    'start-group-order-setup.ps1', 'start-group-order-desktop.ps1',
    'group-order-desktop-driver.ps1', 'group-order-flow-first.ps1',
    'group-order-flow-fixed.ps1', 'group-order-flow-resume.ps1',
    'inspect-group-order.ps1', 'inspect-after-wait-failure.json',
    'preflight-group-order.ps1', 'desktop-preflight.log', 'desktop-setup.log',
    'desktop-start.log', 'desktop-flow.log', 'desktop-resume.log',
    'cleanup-export-group-order.ps1', 'cleanup-export.log',
    'archive-group-order.py', 'verify-group-order.py',
    'verify-group-order-first.py', 'verification-first-failure.json',
]
for name in names:
    relative = name + '.txt' if name.endswith('.go') else name
    archive(ROOT / name, Path('local') / relative)
archive(Path('/private/tmp/asd-upstream-compare-20261123/reference.sb'), Path('local/reference.sb'))

manifest = {'base_commit': 'ffd4d79',
            'product_source_state': 'base plus archived Go/module fingerprints',
            'zip_sha256': digest((ROOT / 'native-export.zip').read_bytes()),
            'zip_bytes': (ROOT / 'native-export.zip').stat().st_size,
            'scope': 'Sibling group order, native test executables and two shell desktop flow; not full acceptance',
            'redaction': 'Known local/test usernames and paths only; original PNGs unchanged',
            'artifacts': entries}
(DEST / 'manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({'artifacts': len(entries), 'zip_sha256': manifest['zip_sha256']}))
