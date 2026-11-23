"""Verify archived evidence consistency; this does not execute the product."""
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import struct

REPO = Path.cwd()
EVIDENCE = REPO / 'docs/testing/results/2026-11-23-group-removal-evidence'


def read(path):
    return json.loads((EVIDENCE / path).read_text())


def rows(value):
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
    assert Path(path).suffix not in ('.exe', '.sqlite3', '.key', '.port')
    if path.endswith('.png'):
        assert entry['original_sha256'] == entry['archived_sha256']
    else:
        assert not re.search('liu' + ' xun yuan|/Users/' + 'liuhui', data.decode('utf-8'), re.I), path
fingerprints = read('local/source-fingerprints.json')
assert len(fingerprints) == 196
for path, expected in fingerprints.items():
    assert sha((REPO / path).read_bytes()) == expected, path
source = read('local/upstream-source.json')
assert source['commit'] == '80494e527393b7ead00da30018e6ee7e60e9d7a0' and source['status'] == ''
assert read('local/upstream-removal-core.json') == {
    'default_refused': True, 'moved_ids': ['a', 'b'],
    'remaining_groups': ['my-sessions', 'other'], 'remaining_ids': ['a', 'b', 'c'],
}
geometry = read('local/removal-geometry.json')
assert len(geometry) == 9
upstream = read('local/upstream-removal-frames.json')
product = read('local/product-removal-frames.json')
assert len(upstream) == len(product) == 9
for g, u, p in zip(geometry, upstream, product):
    width = 41 if g['columns'] == 49 else 52
    assert g['review_fits'] and g['upstream']['width'] == g['product']['width'] == width
    assert g['upstream']['x'] == g['product']['x'] == (g['columns'] - width) // 2
    assert g['upstream']['height'] == (16 if g['mode'] == 'notice' else 19)
    expected_height = 15 if g['mode'] == 'notice' else (18 if g['columns'] == 49 else 17)
    assert g['product']['height'] == expected_height
    for frame in [u, p]:
        assert frame['columns'] == g['columns'] and frame['rows'] == g['rows']
        assert '\x1b[' in frame['content']

counts = {'store': (65, 99), 'manager': (16, 18), 'ui': (157, 341)}
native = read('native/native-tests/outcome.json')
binaries = read('local/binary-hashes.json')
assert len(native['results']) == 3
for result in native['results']:
    package = result['package']
    assert (result['top_level_pass'], result['all_pass']) == counts[package]
    assert result['exit_code'] == result['fail'] == result['skip'] == result['remaining_exact_image'] == 0
    assert result['binary_sha256'].lower() == binaries['group-removal-' + package + '.exe']
    text = (EVIDENCE / ('native/native-tests/' + package + '-tests.log')).read_text()
    assert len(re.findall(r'^--- PASS:', text, re.M)) == result['top_level_pass']
    assert len(re.findall(r'^\s*--- PASS:', text, re.M)) == result['all_pass']
    assert not re.search(r'^\s*--- (?:FAIL|SKIP):', text, re.M)

flow = read('native/flow-result.json')
assert flow['result'] == 'passed' and flow['commands'] == 146
responses = sorted((EVIDENCE / 'native/desktop/responses').glob('*.json'))
done = sorted((EVIDENCE / 'native/desktop/done').glob('*.json'))
assert [p.stem for p in responses] == [f'{i:03}' for i in range(1, 147)]
assert [p.name for p in responses] == [p.name for p in done]
kinds = Counter()
screenshots = []
for response_path, command_path in zip(responses, done):
    response = json.loads(response_path.read_text())
    assert response['ok'] is True and response['command'] == json.loads(command_path.read_text())
    kinds[response['command']['kind']] += 1
    if response['command']['kind'] == 'capture':
        name = response['command']['label'] + '.png'
        data = (EVIDENCE / 'native/desktop' / name).read_bytes()
        assert data.startswith(b'\x89PNG\r\n\x1a\n')
        assert list(struct.unpack('>II', data[16:24])) == response['state']['rect'][2:]
        screenshots.append(name)
assert kinds == Counter(keys=41, state=62, text=25, capture=10, paste=3, resize=4, finish=1)
assert len(screenshots) == 10
before = {v['session_id']: v for v in rows(flow['before'])}
after = {v['session_id']: v for v in rows(flow['after'])}
assert len(before) == 2 and before.keys() == after.keys()
for sid in before:
    for field in ['session_id', 'run_id', 'backend_id', 'command', 'title']:
        assert before[sid][field] == after[sid][field]
processes = rows(flow['before_processes'])
assert len(processes) == 2 and processes == rows(flow['after_processes'])
metadata = {v['label']: v['value'] for v in rows(read('native/metadata-observations.json'))}
assert len(metadata) == flow['metadata_snapshots'] == 12
initial = metadata['before-removal']
original_groups = {g['ID']: g for g in rows(initial['Groups'])}
assert len(original_groups) == 4
root, child, kept = [flow[key]['ID'] for key in ['removed_root', 'removed_child', 'retained_group']]
for label in ['default-enter-cancelled', 'n-cancelled', 'short-viewport-refused',
              'stale-preview-refused', 'repeated-y-refused']:
    value = metadata[label]
    groups = {g['ID']: g for g in rows(value['Groups'])}
    assert value['Members'] == initial['Members'] and value['Runs'] == initial['Runs']
    assert groups.keys() == original_groups.keys()
    for id in [root, child, 'default']:
        assert groups[id] == original_groups[id]
    expected = original_groups[kept].copy()
    if 'preview' in label or label == 'repeated-y-refused':
        expected.update(Name='保留 更新', Revision=2)
    assert groups[kept] == expected
for label in ['removed-structure', 'backup-removed-structure', 'reopened']:
    value = metadata[label]
    groups = {g['ID']: g for g in rows(value['Groups'])}
    assert groups.keys() == {'default', kept} and groups['default'] == original_groups['default']
    assert groups[kept] == flow['retained_group']
    assert value['Runs'] == initial['Runs']
    for member, old in zip(rows(value['Members']), rows(initial['Members'])):
        expected = old.copy()
        expected.update(GroupID='default', Revision=old['Revision'] + 1)
        assert member == expected
assert flow['final_groups'] == metadata['reopened']['Groups']
assert flow['final_members'] == metadata['reopened']['Members']
assert flow['independent_edit']['group_id'] == kept and flow['independent_edit']['revision'] == 2

observations = rows(read('native/native-observations.json'))
assert len(observations) == 12
for observation in observations:
    assert observation['complete'] is True and observation['numbered_ids'] == list(range(1, 41))
    text = observation['capture_text']
    assert [int(v) for v in re.findall(r'^history line (\d+)\s*$', text, re.M)] == list(range(1, 41))
    letter = observation['label'][0]
    owner = processes[0 if letter == 'A' else 1]
    sid = next(sid for sid, v in before.items() if v['title'] == 'Copy ' + letter)
    assert observation['session_id'] == sid and observation['run_id'] == before[sid]['run_id']
    outputs = re.findall(r'^REMOVE_' + letter + r':(\d+)\s*$', text, re.M)
    if observation['label'] == letter + '-explicit-enter':
        assert outputs == [str(owner['pid'])]
    else:
        assert outputs == []
        lines = [line.strip() for line in text.splitlines() if line.strip()]
        expected = "> Write-Output ('REMOVE_" + letter + ":'+$PID)" + (' # dyn' if letter == 'A' else '')
        assert lines[-1] == expected
preferences = rows(read('native/preferences.json'))
assert len(preferences) == flow['preference_snapshots'] == 3
for index, value in enumerate(preferences):
    # This fixture's five scalar fields only; not a general TOML parser.
    settings = {k: json.loads(v) for k, v in re.findall(r'(?m)^([a-z_]+) = (\"[^\"]*\"|[0-9]+|true|false)$', value['text'])}
    assert set(settings) == {'schema_version', 'sidebar_density', 'sidebar_mode', 'default_group_collapsed', 'collapsed_groups'}
    assert settings['schema_version'] == 3 and settings['sidebar_density'] == 'compact'
    assert settings['sidebar_mode'] == 'grouped' and settings['default_group_collapsed'] is False
    folds = [kept, flow['unknown_fold']] + ([root, child] if index == 0 else [])
    assert settings['collapsed_groups'].split(',') == sorted(folds)
assert (EVIDENCE / 'native/ui-final.toml').read_text() == preferences[-1]['text']
outputs = rows(read('native/output-observations.json'))
assert len(outputs) == flow['output_observations'] == 3
record = outputs[0]['record']
assert record['complete'] is False and record['source'] == 'screen_snapshot'
assert record['sequence'] == 1 and record['session_id'] in before
assert record['run_id'] == before[record['session_id']]['run_id']
assert all(v['record'] == record for v in outputs)
assert record['text'].endswith("> Write-Output ('REMOVE_A:'+$PID) # dyn\n")

audit = read('native/final-audit.json')
assert audit['result'] == 'passed' and audit['private_images_alive'] == 0 and audit['remaining_sessions'] == '[]'
assert len(audit['stops']) == 2
assert {(v['session_id'], v['run_id']) for v in audit['stops']} == {(v['session_id'], v['run_id']) for v in before.values()}
originals = rows(audit['original_processes'])
assert len(originals) == 5 and all(v['original_ended'] for v in originals)
assert audit['settings_sha256'].lower() == '1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
assert sorted(audit['tasks_absent']) == ['AsdGroupRemovalDesktop20261123', 'AsdGroupRemovalSetup20261123']
clip = audit['clipboard']
assert clip['restored'] and clip['text_restored'] and clip['format_names_restored']
assert clip['caps_before'] is clip['caps_after'] is True and clip['nontext_content_verified'] is False
result = {
    'result': 'passed', 'artifacts': len(paths), 'source_fingerprints': len(fingerprints),
    'native_packages': counts, 'actions': dict(kinds), 'screenshots': len(screenshots),
    'component_pairs': len(geometry), 'native_captures': len(observations),
    'metadata_snapshots': len(metadata), 'saved_output_observations': len(outputs),
    'original_processes_ended': len(originals),
    'scope': 'Bounded evidence consistency; no new product execution or full acceptance',
}
(EVIDENCE / 'verification.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(result, ensure_ascii=False))
