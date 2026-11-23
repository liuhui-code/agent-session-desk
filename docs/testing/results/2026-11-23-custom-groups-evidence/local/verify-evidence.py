from pathlib import Path
import hashlib,json,re
root=Path('docs/testing/results/2026-11-23-custom-groups-evidence')
x=json.loads((root/'manifest.json').read_text())
expected={r['path']for r in x['files']}
actual={p.relative_to(root).as_posix()for p in root.rglob('*')if p.is_file()}-{'manifest.json','README.md'}
assert expected==actual,{'missing':sorted(expected-actual),'extra':sorted(actual-expected)}
for row in x['files']:
 data=(root/row['path']).read_bytes()
 assert hashlib.sha256(data).hexdigest()==row['sha256'],row['path']
 text=data.decode('utf-8')
 assert str(Path.home()) not in text,row['path']
 assert not re.search(r'C:[\\]Users[\\](?!<WINDOWS_TEST_USER>)[^\\\s]+',text),row['path']
source=json.loads((root/'local/source-after.json').read_text())
for row in source['files']:assert hashlib.sha256(Path(row['path']).read_bytes()).hexdigest()==row['sha256'],row['path']
results=json.loads((root/'native/final/results.json').read_text())
top=allpass=0
for result in results:
 assert result['exit_code']==result['fail']==result['skip']==0,result
 data=(root/'native/final'/(result['package']+'-tests.log')).read_text()
 assert len(re.findall(r'^--- PASS:',data,re.M))==result['top_level_pass']
 assert len(re.findall(r'^\s*--- PASS:',data,re.M))==result['all_pass']
 top+=result['top_level_pass'];allpass+=result['all_pass']
assert top==214 and allpass==413,(top,allpass)
assert json.loads((root/'native/final/environment.json').read_text())['remaining_test_images']==0
red=json.loads((root/'native/observer-red/result.json').read_text())
assert red['exit_code']==1 and red['sharing_violation'] and red['remaining_exact_image']==0
text=(root/'native/observer-green/observer-green.log').read_text()
for name in ['TestOutputObserverDoesNotPreventPublishedRecordRotation','TestOutputPublicationNeverExposesIncompleteOrExcessFiles']:
 assert len(re.findall(r'^--- PASS: '+name+r' ',text,re.M))==20,name
assert json.loads((root/'native/observer-green/result.json').read_text())['exit_code']==0
assert all(r['exit_code']==0 for r in json.loads((root/'local/check-exits.json').read_text()))
print(json.dumps({'artifact_count':len(expected),'source_files':len(source['files']),'native_top_level_pass':top,'native_all_pass':allpass,'observer_green_repetitions':40,'result':'verified'}))
