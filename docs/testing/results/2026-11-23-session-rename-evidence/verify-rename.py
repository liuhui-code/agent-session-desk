from pathlib import Path
import hashlib
import json
import re

root = Path.cwd()
assert (root / 'go.mod').exists(), 'Run from the repository root'
evidence = root / 'docs/testing/results/2026-11-23-session-rename-evidence'

def read(name):
    return json.loads((evidence / name).read_text())

def flat(value):
    if isinstance(value, list):
        return [item for child in value for item in flat(child)]
    if isinstance(value, dict) and 'value' in value and set(value) <= {'value', 'Count'}:
        return flat(value['value'])
    return [value]

manifest = read('manifest.json')
for item in manifest['files']:
    data = (evidence / item['path']).read_bytes()
    assert hashlib.sha256(data).hexdigest() == item['archived_sha256'], item['path']
    if item['path'].endswith('.png'):
        assert item['original_sha256'] == item['archived_sha256']
    assert not item['path'].endswith(('.exe', '.sqlite3', '.port', '.key', 'settings.before'))

fingerprints = read('local/source-fingerprints-complete.json')
for path, expected in fingerprints.items():
    assert hashlib.sha256((root / path).read_bytes()).hexdigest() == expected, path

model = read('native/model/native-test-results.json')
complete_ui = read('native/ui-complete/ui-tests-complete-result.json')
final_suites = [r for r in model if r['package'] != 'ui'] + [complete_ui]
for r in final_suites:
    assert r['exit_code'] == r['fail'] == r['skip'] == 0
assert sum(r['top_level_pass'] for r in final_suites) == 164
assert sum(r['all_pass'] for r in final_suites) == 353

def verify_flow(folder, result_file, count, original_count, prefix):
    flow = read(f'{folder}/{result_file}')
    assert flow['result'] == 'passed' and flow['commands'] == count
    before, after = flat(flow['before']), flat(flow['after'])
    assert len(before) == len(after) == original_count
    for b in before:
        a = next(a for a in after if a['session_id'] == b['session_id'])
        assert a['run_id'] == b['run_id'] and a['backend_id'] == b['backend_id']
    assert flat(flow['before_processes']) == flat(flow['after_processes'])
    outcomes = sorted((evidence / folder / 'desktop/responses').glob('*.json'))
    assert [p.name for p in outcomes] == [f'{n:03}.json' for n in range(1, count + 1)]
    receipts = [json.loads(p.read_text()) for p in outcomes]
    assert all(r['ok'] for r in receipts)
    observations = flat(read(f'{folder}/native-observations.json'))
    for o in observations:
        text = o['capture_text'].replace('\r\n', '\n')
        numbers = [int(n) for n in re.findall(r'^history line (\d+)\s*$', text, re.M)]
        assert numbers == list(range(1, 41)), o['label']
        assert o['complete'] and o['numbered_ids'] == numbers
        if 'explicit-enter' not in o['label']:
            tag = prefix if prefix == 'RENAME_STYLE' else ('RENAME_A' if o['label'].startswith('A-') else 'RENAME_B')
            draft = "Write-Output ('" + tag + ":'+$PID)"
            assert [line.rstrip() for line in text.splitlines() if line.strip()][-1] == '> ' + draft, o['label']
            assert not re.search(r'^' + tag + r':\d+\s*$', text, re.M)
    for tag, label in [('RENAME_STYLE', 'Copy A')] if prefix == 'RENAME_STYLE' else [('RENAME_A', 'Copy A'), ('RENAME_B', 'Copy B')]:
        pid = next(p['pid'] for p in flat(flow['before_processes']) if p['label'] == label)
        final = next(o for o in observations if 'explicit-enter' in o['label'] and re.search(r'^' + tag + ':' + str(pid) + r'\s*$', o['capture_text'], re.M))
        assert len(re.findall(r'^' + tag + ':', final['capture_text'], re.M)) == 1
        draft = "Write-Output ('" + tag + ":'+$PID)"
        assert sum(r['command'].get('text') == draft for r in receipts) == 1
    return flow, observations

full, full_obs = verify_flow('native/full', 'rename-result.json', 154, 3, 'RENAME')
surface, surface_obs = verify_flow('native/surface', 'style-result.json', 33, 1, 'RENAME_STYLE')
assert len(full_obs) == 13 and len(surface_obs) == 3
for folder, file, count in [('native/full', 'final-audit-corrected.json', 7), ('native/surface', 'final-audit.json', 3)]:
    audit = read(folder + '/' + file)
    assert audit['result'] == 'passed' and audit['private_images_alive'] == 0
    assert len(audit['original_processes']) == count
    assert all(p['original_ended'] and isinstance(p['pid'], int) and isinstance(p['filetime'], int) for p in audit['original_processes'])
    assert audit['clipboard']['text_restored'] and audit['clipboard']['format_names_restored']
    assert not audit['clipboard']['nontext_content_verified']
    assert audit['settings_sha256'].lower() == '1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'
    setup = read(folder + '/setup-outcome.json')
    assert setup['result'] == 'passed' and not setup['is_administrator'] and setup['session'] == 2
for g in surface['geometry']:
    assert g['width'] == {120: 46, 63: 46, 49: 41}[g['columns']]
    assert g['x'] == {120: 37, 63: 8, 49: 4}[g['columns']]
    assert g['bottom'] - g['top'] + 1 == 10

output = read('native/full/renamed-output-result.json')
assert output['result'] == 'passed'
record = flat(output['output'])[0]
b = next(b for b in flat(full['before']) if b['title'] == 'Copy B')
assert record['session_id'] == b['session_id'] and record['run_id'] == b['run_id']
assert not record['complete'] and record['source'] == 'screen_snapshot'
assert re.search(r'^RENAME_B:17392\s*$', record['text'], re.M)
assert len([a for a in flat(full['after']) if a['title'] == 'Copy A']) == 2

last = read('local/native-readonly-final-audit.log')
assert last['result'] == 'passed' and last['private_images_alive'] == 0 and len(last['tasks_absent']) == 4
assert len(flat(last['offline_history'])) == 3
assert all(h['session']['terminal_state'] == 'unknown' and h['stop_confirmed_at'] for h in flat(last['offline_history']))
assert any(h['session']['title'] == 'Copy A' and h['session'].get('title_revision') == 3 for h in flat(last['offline_history']))

summary = {'result': 'passed', 'manifest_files_verified': len(manifest['files']), 'current_go_module_fingerprints': len(fingerprints), 'native_top_level_pass': 164, 'native_pass_including_subtests': 353, 'native_fail': 0, 'native_skip': 0, 'full_flow_receipts': 154, 'full_flow_exact_pending_and_completed_captures': 13, 'surface_flow_receipts': 33, 'surface_exact_captures': 3, 'registered_original_processes_ended': [7, 3], 'final_private_images_alive': 0, 'catalog_execution_promoted': False, 'scope': 'These bounded checks do not prove real-agent, all IME, complete upstream visual or release acceptance.'}
(evidence / 'verification.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(summary, ensure_ascii=False))
