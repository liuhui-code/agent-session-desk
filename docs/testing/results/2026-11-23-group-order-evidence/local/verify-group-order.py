"""Verify archived evidence consistency; this does not execute the product."""
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import struct

REPO = Path.cwd()
EVIDENCE = REPO / 'docs/testing/results/2026-11-23-group-order-evidence'


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
assert len(fingerprints) == 185
for path, expected in fingerprints.items():
    assert sha((REPO / path).read_bytes()) == expected, path
source = read('local/upstream-source.json')
assert source['commit'] == '80494e527393b7ead00da30018e6ee7e60e9d7a0' and source['status'] == ''
source_order = {s['label']: [g['path'] for g in s['groups']] for s in read('local/upstream-order-snapshots.json')}
assert source_order == {
    'initial': ['my-sessions', 'first', 'first/review', 'first/implementation', 'second'],
    'root-up': ['my-sessions', 'second', 'first', 'first/review', 'first/implementation'],
    'root-down': ['my-sessions', 'first', 'first/review', 'first/implementation', 'second'],
    'child-up': ['my-sessions', 'first', 'first/implementation', 'first/review', 'second'],
}

counts = {'store': (58, 84), 'manager': (15, 17), 'ui': (145, 327)}
native = read('native/native-tests/outcome.json')
binaries = read('local/binary-hashes.json')
assert len(native['results']) == 3
for result in native['results']:
    package = result['package']
    assert (result['top_level_pass'], result['all_pass']) == counts[package]
    assert result['exit_code'] == result['fail'] == result['skip'] == result['remaining_exact_image'] == 0
    assert result['binary_sha256'].lower() == binaries['group-order-' + package + '.exe']
    text = (EVIDENCE / ('native/native-tests/' + package + '-tests.log')).read_text()
    assert len(re.findall(r'^--- PASS:', text, re.M)) == result['top_level_pass']
    assert len(re.findall(r'^\s*--- PASS:', text, re.M)) == result['all_pass']
    assert not re.search(r'^\s*--- (?:FAIL|SKIP):', text, re.M)

flow = read('native/flow-result.json')
assert flow['result'] == 'passed' and flow['commands'] == 160
responses = sorted((EVIDENCE / 'native/desktop/responses').glob('*.json'))
done = sorted((EVIDENCE / 'native/desktop/done').glob('*.json'))
assert [p.stem for p in responses] == [f'{i:03}' for i in range(1, 161)]
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
assert kinds == Counter(keys=36, state=87, text=23, capture=7, paste=3, resize=3, finish=1)
assert len(screenshots) == 7
resume = read('native/flow-resume.json')
assert resume['previous_completed_sequence'] == 87 and resume['replayed_mutations'] == 0
assert resume['original_driver_state'] == 'Running'
inspection = read('local/inspect-after-wait-failure.json')
assert inspection['task_state'] == 'Running' and inspection['last_response'] == '087.json'
assert inspection['failure']['sequence'] == 87 and inspection['command']['kind'] == 'state'

before = {v['session_id']: v for v in rows(flow['before'])}
after = {v['session_id']: v for v in rows(flow['after'])}
assert len(before) == 2 and before.keys() == after.keys()
for sid in before:
    for field in ['session_id', 'run_id', 'backend_id', 'command', 'title']:
        assert before[sid][field] == after[sid][field]
processes = rows(flow['before_processes'])
assert len(processes) == 2 and processes == rows(flow['after_processes'])
metadata = {v['label']: v['value'] for v in rows(read('native/metadata-observations.json'))}
assert len(metadata) == flow['metadata_snapshots'] == 18
initial = metadata['before-order']
original_groups = {g['ID']: g for g in rows(initial['Groups'])}
assert len(original_groups) == 5
a, b = flow['root_A']['ID'], flow['root_B']['ID']
review, implementation = flow['review']['ID'], flow['implementation']['ID']
root_steps = ['K-root-up', 'J-root-down', 'ctrl-up-root', 'ctrl-down-root',
              'shift-up-root', 'shift-down-root', 'plus-root-up', 'minus-root-down']
for index, label in enumerate(root_steps):
    value = metadata[label]
    groups = {g['ID']: g for g in rows(value['Groups'])}
    roots = sorted((g for g in groups.values() if g['ParentID'] == ''), key=lambda g: g['Order'])
    assert [g['ID'] for g in roots] == ['default', *([b, a] if index % 2 == 0 else [a, b])]
    assert groups[a]['Revision'] == groups[b]['Revision'] == index + 2
    assert groups[review] == original_groups[review] and groups[implementation] == original_groups[implementation]
    assert groups['default'] == original_groups['default']
    for id in [a, b]:
        for field in ['ID', 'ParentID', 'Name', 'DefaultDir', 'Scope']:
            assert groups[id][field] == original_groups[id][field]
    assert value['Members'] == initial['Members'] and value['Runs'] == initial['Runs']
assert metadata['pinned-and-boundary-noop']['Groups'] == metadata['minus-root-down']['Groups']
for label in ['child-K-up', 'backup-order', 'reopened-order']:
    value = metadata[label]
    groups = {g['ID']: g for g in rows(value['Groups'])}
    assert groups[a]['Order'] == 1 and groups[b]['Order'] == 2
    assert groups[a]['Revision'] == groups[b]['Revision'] == 9
    assert groups[review]['Order'] == 1 and groups[implementation]['Order'] == 0
    for id in [review, implementation]:
        assert groups[id]['Revision'] == 2 and groups[id]['ParentID'] == a
        for field in ['ID', 'ParentID', 'Name', 'DefaultDir', 'Scope']:
            assert groups[id][field] == original_groups[id][field]
    assert value['Members'] == initial['Members'] and value['Runs'] == initial['Runs']
assert metadata['child-K-up']['Groups'] == metadata['backup-order']['Groups'] == metadata['reopened-order']['Groups'] == flow['final_groups']
assert flow['root_swaps'] == 8 and flow['child_swaps'] == 1
inspected = {g['ID']: g for g in rows(inspection['metadata']['Groups'])}
assert inspected[a]['Revision'] == inspected[b]['Revision'] == 2
assert inspected[a]['Order'] == 2 and inspected[b]['Order'] == 1

observations = rows(read('native/native-observations.json'))
assert len(observations) == 10
for observation in observations:
    assert observation['complete'] is True and observation['numbered_ids'] == list(range(1, 41))
    text = observation['capture_text']
    assert [int(v) for v in re.findall(r'^history line (\d+)\s*$', text, re.M)] == list(range(1, 41))
    letter = observation['label'][0]
    owner = processes[0 if letter == 'A' else 1]
    sid = next(sid for sid, v in before.items() if v['title'] == 'Copy ' + letter)
    assert observation['session_id'] == sid and observation['run_id'] == before[sid]['run_id']
    outputs = re.findall(r'^ORDER_' + letter + r':(\d+)\s*$', text, re.M)
    if observation['label'] == letter + '-explicit-enter':
        assert outputs == [str(owner['pid'])]
    else:
        assert outputs == []
        lines = [line.strip() for line in text.splitlines() if line.strip()]
        expected = "> Write-Output ('ORDER_" + letter + ":'+$PID)" + (' # KJ+-' if letter == 'A' else '')
        assert lines[-1] == expected
preferences = rows(read('native/preferences.json'))
assert len(preferences) == flow['preference_snapshots'] == 2
for value, folds in zip(preferences, [[b], sorted([a, b])]):
    # Only this fixture's five scalar fields; not a general TOML parser.
    settings = {k: json.loads(v) for k, v in re.findall(r'(?m)^([a-z_]+) = (\"[^\"]*\"|[0-9]+|true|false)$', value['text'])}
    assert set(settings) == {'schema_version', 'sidebar_density', 'sidebar_mode', 'default_group_collapsed', 'collapsed_groups'}
    assert settings['schema_version'] == 3 and settings['sidebar_density'] == 'compact' and settings['sidebar_mode'] == 'grouped'
    assert settings['collapsed_groups'].split(',') == folds
assert (EVIDENCE / 'native/ui-final.toml').read_text() == preferences[-1]['text']

audit = read('native/final-audit.json')
assert audit['result'] == 'passed' and audit['private_images_alive'] == 0 and audit['remaining_sessions'] == '[]'
assert len(audit['stops']) == 2 and {(v['session_id'], v['run_id']) for v in audit['stops']} == {(v['session_id'], v['run_id']) for v in before.values()}
originals = rows(audit['original_processes'])
assert len(originals) == 5 and all(v['original_ended'] for v in originals)
assert audit['settings_sha256'].lower() == '1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
assert sorted(audit['tasks_absent']) == ['AsdGroupOrderDesktop20261123', 'AsdGroupOrderSetup20261123']
clip = audit['clipboard']
assert clip['restored'] and clip['text_restored'] and clip['format_names_restored']
assert clip['caps_before'] is clip['caps_after'] is True and clip['nontext_content_verified'] is False
result = {'result': 'passed', 'artifacts': len(paths), 'source_fingerprints': len(fingerprints),
          'native_packages': counts, 'actions': dict(kinds), 'screenshots': len(screenshots),
          'source_order_snapshots': len(source_order), 'root_swaps': 8, 'child_swaps': 1,
          'native_captures': len(observations), 'metadata_snapshots': len(metadata),
          'original_processes_ended': len(originals), 'scope': 'Bounded archived evidence consistency; no new product execution or full acceptance'}
(EVIDENCE / 'verification.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(result, ensure_ascii=False))
