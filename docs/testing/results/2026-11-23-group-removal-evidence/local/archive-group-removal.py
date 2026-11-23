"""Archive allowlisted evidence; never copy runtime databases or credentials."""
import hashlib
import json
from pathlib import Path
import re

ROOT = Path('/private/tmp/asd-group-removal-20261123')
DEST = Path.cwd() / 'docs/testing/results/2026-11-23-group-removal-evidence'
DEST.mkdir(parents=True, exist_ok=True)
entries = []


def sha(data):
    return hashlib.sha256(data).hexdigest()


def archive(source, relative):
    raw = source.read_bytes()
    data = raw
    transformation = 'unchanged PNG'
    if source.suffix != '.png':
        encoding = 'utf-16' if raw.startswith((b'\xff\xfe', b'\xfe\xff')) else 'utf-8-sig'
        text = raw.decode(encoding).replace(str(Path.cwd()), '<REPO>')
        text = text.replace('/Users/' + 'liuhui', '<LOCAL_TEST_HOME>')
        text = re.sub('liu' + ' xun yuan', '<WINDOWS_TEST_USER>', text, flags=re.I)
        data = text.replace('\r\n', '\n').encode('utf-8')
        transformation = encoding + ' to UTF-8/LF; known local/test usernames and paths redacted'
    output = DEST / relative
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_bytes(data)
    entries.append({'path': str(relative), 'original_sha256': sha(raw),
                    'archived_sha256': sha(data), 'bytes': len(data),
                    'transformation': transformation})


for path in sorted((ROOT / 'native').rglob('*')):
    if path.is_file():
        assert path.suffix in ('.json', '.log', '.ps1', '.png', '.toml')
        archive(path, Path('native') / path.relative_to(ROOT / 'native'))

names = [
    'store-red.log', 'store-green.log', 'manager-red.log', 'ui-red.log',
    'vertical-green.log', 'viewport-red.log', 'viewport-green.log',
    'stale-order-red.log', 'stale-order-green.log', 'stale-rename-red.log',
    'stale-rename-green.log', 'preferences-red.log', 'preferences-green.log',
    'store-regression.log', 'ui-regression.log', 'ui-final.log',
    'go-test.log', 'go-race.log', 'go-vet.log', 'go-mod-verify.log',
    'source-limits.log', 'catalog-check.log', 'product-build-info.txt',
    'source-fingerprints.json', 'binary-hashes.json', 'upstream-source.json',
    'upstream-reference.log', 'upstream-core-reference.log', 'product-reference.log',
    'upstream-removal-frames.json', 'product-removal-frames.json',
    'upstream-removal-core.json', 'removal-geometry.json',
    'upstream-removal-reference_test.go', 'upstream-removal-core-reference_test.go',
    'product-removal-reference_test.go', 'upstream-removal-overlay.json',
    'upstream-removal-core-overlay.json', 'product-removal-overlay.json',
    'edit-group-helper.go', 'edit-group-overlay.json', 'native-run.log',
    'run-group-removal-tests.ps1', 'setup-group-removal-desktop.ps1',
    'start-group-removal-setup.ps1', 'start-group-removal-desktop.ps1',
    'group-removal-desktop-driver.ps1', 'group-removal-flow.ps1',
    'preflight-group-removal.ps1', 'desktop-preflight.log', 'desktop-preflight-final.log',
    'desktop-setup.log', 'desktop-start.log', 'desktop-ready.log',
    'inspect-driver-ready.ps1', 'desktop-flow.log',
    'cleanup-export-group-removal.ps1', 'cleanup-export.log',
    'archive-group-removal.py', 'verify-group-removal.py',
]
for name in names:
    relative = name + '.txt' if name.endswith('.go') else name
    archive(ROOT / name, Path('local') / relative)

inputs = {}
for overlay in ['upstream-removal-overlay.json', 'upstream-removal-core-overlay.json',
                'product-removal-overlay.json', 'edit-group-overlay.json']:
    for value in sorted(set(json.loads((ROOT / overlay).read_text())['Replace'].values())):
        path = Path(value)
        if path.parent == ROOT:
            relative = Path('local') / (path.name + '.txt')
        else:
            relative = Path('local/overlay-inputs') / (path.parent.name + '-' + path.name + '.txt')
            if value not in inputs:
                archive(path, relative)
        inputs[value] = str(relative)
input_file = ROOT / 'overlay-inputs.json'
input_file.write_text(json.dumps(inputs, indent=2) + '\n')
archive(input_file, Path('local/overlay-inputs.json'))
archive(Path('/private/tmp/asd-upstream-compare-20261123/reference.sb'), Path('local/reference.sb'))
manifest = {
    'base_commit': '8e8170b', 'product_source_state': 'base plus archived Go/module fingerprints',
    'zip_sha256': sha((ROOT / 'native-export.zip').read_bytes()),
    'zip_bytes': (ROOT / 'native-export.zip').stat().st_size,
    'scope': 'Group removal, native tests and two-shell desktop; not full acceptance',
    'redaction': 'Known local/test usernames and paths only; original PNG bytes unchanged',
    'artifacts': entries,
}
(DEST / 'manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({'artifacts': len(entries), 'zip_sha256': manifest['zip_sha256']}))
