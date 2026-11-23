"""Verify this slice's archived receipts; does not execute product tests."""
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import struct

REPO = Path.cwd()
EVIDENCE = REPO / 'docs/testing/results/2026-11-23-group-rename-evidence'


def read(relative):
    return json.loads((EVIDENCE / relative).read_text())


def scalar_rows(value):
    if isinstance(value, dict) and set(value) == {'value', 'Count'}:
        assert value['Count'] == len(value['value'])
        return value['value']
    assert isinstance(value, list)
    return value


def sha(data):
    return hashlib.sha256(data).hexdigest()


manifest = read('manifest.json')
paths = set()
for entry in manifest['artifacts']:
    path = entry['path']
    assert path not in paths and '..' not in Path(path).parts
    paths.add(path)
    data = (EVIDENCE / path).read_bytes()
    assert sha(data) == entry['archived_sha256'] and len(data) == entry['bytes'], path
    assert Path(path).suffix not in ('.exe', '.sqlite3', '.key', '.port'), path
    if path.endswith('.png'):
        assert entry['original_sha256'] == entry['archived_sha256']
    else:
        text = data.decode('utf-8')
        assert not re.search('<WINDOWS_TEST_USER>|<LOCAL_TEST_HOME>', text, re.I), path

fingerprints = read('local/source-fingerprints.json')
for path, expected in fingerprints.items():
    assert sha((REPO / path).read_bytes()) == expected, path
assert len(fingerprints) == 179
source = read('local/upstream-source.json')
assert source['commit'] == '80494e527393b7ead00da30018e6ee7e60e9d7a0' and source['status'] == ''

native = read('native/native-tests/outcome.json')
expected_counts = {'store': (51, 77), 'manager': (14, 16), 'ui': (136, 310)}
binaries = read('local/binary-hashes.json')
for result in native['results']:
    package = result['package']
    assert (result['top_level_pass'], result['all_pass']) == expected_counts[package]
    assert result['exit_code'] == result['fail'] == result['skip'] == result['remaining_exact_image'] == 0
    assert result['binary_sha256'].lower() == binaries['group-rename-' + package + '.exe']
    log = (EVIDENCE / ('native/native-tests/' + package + '-tests.log')).read_text()
    assert len(re.findall(r'^--- PASS:', log, re.M)) == result['top_level_pass']
    assert len(re.findall(r'^\s*--- PASS:', log, re.M)) == result['all_pass']
    assert not re.search(r'^\s*--- (?:FAIL|SKIP):', log, re.M)
assert len(native['results']) == 3

flow = read('native/flow-result.json')
assert flow['result'] == 'passed' and flow['commands'] == 109
responses = sorted((EVIDENCE / 'native/desktop/responses').glob('*.json'))
done = sorted((EVIDENCE / 'native/desktop/done').glob('*.json'))
assert [p.stem for p in responses] == [f'{i:03}' for i in range(1, 110)]
assert [p.name for p in done] == [p.name for p in responses]
kinds = Counter()
screenshots = []
for path, command in zip(responses, done):
    response = json.loads(path.read_text())
    assert response['ok'] is True
    assert response['command'] == json.loads(command.read_text())
    kinds[response['command']['kind']] += 1
    if response['command']['kind'] == 'capture':
        name = response['command']['label'] + '.png'
        data = (EVIDENCE / 'native/desktop' / name).read_bytes()
        assert data.startswith(b'\x89PNG\r\n\x1a\n')
        size = struct.unpack('>II', data[16:24])
        assert list(size) == response['state']['rect'][2:]
        screenshots.append(name)
assert kinds == Counter(keys=33, state=41, text=17, capture=9, paste=5, finish=1, resize=3)
assert len(screenshots) == 9

before = {v['session_id']: v for v in scalar_rows(flow['before'])}
after = {v['session_id']: v for v in scalar_rows(flow['after'])}
assert len(before) == 2 and before.keys() == after.keys()
for sid in before:
    for field in ['session_id', 'run_id', 'backend_id', 'command', 'title']:
        assert before[sid][field] == after[sid][field]
processes = scalar_rows(flow['before_processes'])
assert len(processes) == 2
assert processes == scalar_rows(flow['after_processes'])
root_before, root_after = flow['root_before'], flow['root_after']
child_before, child_after = flow['child_before'], flow['child_after']
for original, named, name in [(root_before, root_after, '项目 甲 新版'), (child_before, child_after, '审核 二审')]:
    assert named['Name'] == name and named['Revision'] == original['Revision'] + 1
    for field in ['ID', 'ParentID', 'DefaultDir', 'Order', 'Scope']:
        assert named[field] == original[field]
assert child_after['ParentID'] == root_after['ID']
assert flow['default_after']['ID'] == 'default' and flow['default_after']['Name'] == '终端'
assert flow['default_after']['Revision'] == 2 and flow['default_after']['Order'] == 0
assert flow['default_after']['ParentID'] == ''

metadata_rows = scalar_rows(read('native/metadata-observations.json'))
metadata = {v['label']: v['value'] for v in metadata_rows}
assert len(metadata) == flow['metadata_snapshots'] == 7
initial = metadata['before-rename']
for label in ['root-renamed', 'child-renamed', 'default-renamed', 'reopen-cancelled']:
    value = metadata[label]
    assert value['Version'] == 4 and value['Members'] == initial['Members'] and value['Runs'] == initial['Runs']
assert metadata['reopen-cancelled']['Groups'] == metadata['default-renamed']['Groups']
# The observer emits a singleton as an object rather than an array.
if isinstance(initial['Members'], dict) and 'SessionID' in initial['Members']:
    members = [initial['Members']]
else:
    members = scalar_rows(initial['Members'])
assert len(members) == 1 and members[0]['GroupID'] == child_after['ID'] and members[0]['Revision'] == 1

observations = scalar_rows(read('native/native-observations.json'))
assert len(observations) == 10
for observation in observations:
    assert observation['complete'] is True and observation['numbered_ids'] == list(range(1, 41))
    text = observation['capture_text']
    ids = [int(v) for v in re.findall(r'^history line (\d+)\s*$', text, re.M)]
    assert ids == list(range(1, 41))
    letter = observation['label'][0]
    owner = processes[0 if letter == 'A' else 1]
    sid = next(sid for sid, v in before.items() if v['title'] == 'Copy ' + letter)
    assert observation['session_id'] == sid and observation['run_id'] == before[sid]['run_id']
    output = re.findall(r'^RENAME_' + letter + r':(\d+)\s*$', text, re.M)
    if observation['label'] == letter + '-explicit-enter':
        assert output == [str(owner['pid'])]
    else:
        assert output == []
        lines = [line.strip() for line in text.splitlines() if line.strip()]
        assert lines[-1] == "> Write-Output ('RENAME_" + letter + ":'+$PID)"

preferences = scalar_rows(read('native/preferences.json'))
assert len(preferences) == flow['preference_snapshots'] == 2
assert preferences[0]['text'] == preferences[1]['text']
assert f'collapsed_groups = "{root_after["ID"]}"' in preferences[0]['text']
geometry = read('local/rename-geometry.json')
assert len(geometry) == 3
for row, width in zip(geometry, [46, 46, 41]):
    assert row['same_rectangle'] is True and row['upstream'] == row['product']
    assert row['product']['width'] == width and row['product']['height'] == 10
    assert 0 <= row['cursor']['X'] < row['columns'] and 0 <= row['cursor']['Y'] < row['rows']

audit = read('native/final-audit.json')
assert audit['result'] == 'passed' and audit['private_images_alive'] == 0
assert audit['remaining_sessions'] == '[]' and len(audit['stops']) == 2
assert {(v['session_id'], v['run_id']) for v in audit['stops']} == {(v['session_id'], v['run_id']) for v in before.values()}
originals = scalar_rows(audit['original_processes'])
assert len(originals) == 5 and all(v['original_ended'] for v in originals)
assert audit['settings_sha256'].lower() == '1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
assert sorted(audit['tasks_absent']) == ['AsdGroupRenameDesktop20261123', 'AsdGroupRenameSetup20261123']
clip = audit['clipboard']
assert clip['restored'] and clip['text_restored'] and clip['format_names_restored']
assert clip['caps_before'] is clip['caps_after'] is True and clip['nontext_content_verified'] is False

result = {'result': 'passed', 'artifacts': len(paths), 'source_fingerprints': len(fingerprints),
          'native_packages': expected_counts, 'actions': dict(kinds), 'screenshots': len(screenshots),
          'native_captures': len(observations), 'metadata_snapshots': len(metadata),
          'original_processes_ended': len(originals), 'geometry_pairs': len(geometry),
          'scope': 'Archived evidence consistency, exact source fingerprints and this bounded slice only; no new product execution'}
(EVIDENCE / 'verification.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(result, ensure_ascii=False))
