from pathlib import Path, PurePosixPath
import hashlib
import json
import zipfile

task = Path('/private/tmp/asd-rename-20261123')
dest = Path('<REPO>/docs/testing/results/2026-11-23-session-rename-evidence')
dest.mkdir(exist_ok=False)
entries = []
archives = []

def sha(data):
    return hashlib.sha256(data).hexdigest()

def save(data, name):
    path = dest / name
    path.parent.mkdir(parents=True, exist_ok=True)
    encoding = 'binary'
    archived = data
    if path.suffix in {'.log', '.json', '.ps1', '.py', '.go', '.mod', '.sum', '.sb'}:
        encoding = 'utf-16' if data[:2] in {b'\xff\xfe', b'\xfe\xff'} else 'utf-8-sig'
        text = data.decode(encoding)
        for before, after in [
            ('<WINDOWS_TEST_USER>', '<WINDOWS_TEST_USER>'),
            ('<REPO>', '<REPO>'),
            ('<LOCAL_TEST_HOME>', '<LOCAL_TEST_HOME>'),
        ]:
            text = text.replace(before, after)
        archived = text.replace('\r\n', '\n').encode('utf-8')
    path.write_bytes(archived)
    entries.append({'path': name, 'original_sha256': sha(data), 'archived_sha256': sha(archived), 'original_bytes': len(data), 'archived_bytes': len(archived), 'decoded_as': encoding})

for source, target in [
    ('native-model-tests-export.zip', 'native/model'),
    ('native-final-ui-tests-export.zip', 'native/ui-footer'),
    ('native-surface-ui-tests-export.zip', 'native/ui-surface'),
    ('native-complete-ui-tests-export.zip', 'native/ui-complete'),
    ('native-desktop-export.zip', 'native/full'),
    ('native-surface-export.zip', 'native/surface'),
]:
    raw = (task / source).read_bytes()
    archives.append({'name': source, 'sha256': sha(raw), 'bytes': len(raw)})
    with zipfile.ZipFile(task / source) as z:
        for info in z.infolist():
            name = info.filename.replace('\\', '/')
            if info.is_dir() or name.endswith('/'):
                continue
            p = PurePosixPath(name)
            assert not p.is_absolute() and '..' not in p.parts
            assert p.suffix in {'.json', '.log', '.png', '.ps1'}, name
            save(z.read(info), target + '/' + name)

for pattern in ('*.log', '*.ps1', 'source-fingerprints*.json', 'binary-hashes.json', 'upstream-source.json', 'upstream-rename-frames.json'):
    for path in sorted(task.glob(pattern)):
        save(path.read_bytes(), 'local/' + path.name)
for name in ('main.go', 'go.mod', 'go.sum'):
    save((task / 'reference' / name).read_bytes(), 'reference/' + name)
save(Path('/private/tmp/asd-upstream-compare-20261123/reference.sb').read_bytes(), 'reference/reference.sb')
save(Path(__file__).read_bytes(), 'archive-rename.py')

initial = task / 'native-desktop-export-initial.zip'
archives.append({'name': initial.name, 'sha256': sha(initial.read_bytes()), 'bytes': initial.stat().st_size, 'note': 'Original five-record batched UI audit retained in corrected full export; seven scalar identities are authoritative in final-audit-corrected.json.'})
(dest / 'manifest.json').write_text(json.dumps({
    'date': '2026-11-23',
    'base_commit': '6b14692',
    'scope': 'ADR-0021 persistent names, source component rendering, local checks, native Windows model/store suites and bounded ordinary-shell desktop flows; not release acceptance.',
    'archives': archives,
    'transformations': 'Decode UTF-8 BOM or UTF-16 BOM, normalize CRLF to LF, replace known test usernames/home/repository paths. PNG bytes unchanged. Backslash ZIP paths normalized; directories remain directories.',
    'exclusions': ['executables', 'runtime ports and authentication keys', 'SQLite databases', 'original clipboard contents', 'Terminal settings backups', 'shell history files', 'uninspected startup screenshots'],
    'files': entries,
}, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({'archived_files': len(entries), 'bytes': sum(v['archived_bytes'] for v in entries), 'archives': len(archives)}))
