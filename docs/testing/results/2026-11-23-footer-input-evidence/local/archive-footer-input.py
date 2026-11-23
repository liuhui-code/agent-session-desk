"""Archive allowlisted source, TDD and bounded Windows evidence."""
import hashlib
import json
from pathlib import Path
import re
ROOT=Path('/private/tmp/asd-footer-input-20261123')
DEST=Path.cwd()/'docs/testing/results/2026-11-23-footer-input-evidence'
DEST.mkdir(parents=True,exist_ok=True)
entries=[]
def sha(data):return hashlib.sha256(data).hexdigest()
def archive(source,relative):
    raw=source.read_bytes();data=raw;change='unchanged PNG'
    if source.suffix!='.png':
        encoding='utf-16' if raw.startswith((b'\xff\xfe',b'\xfe\xff')) else 'utf-8-sig'
        text=raw.decode(encoding).replace(str(Path.cwd()),'<REPO>')
        text=text.replace('/Users/'+'liuhui','<LOCAL_TEST_HOME>')
        text=re.sub('liu'+' xun yuan','<WINDOWS_TEST_USER>',text,flags=re.I)
        data=text.replace('\r\n','\n').encode('utf-8');change=encoding+' to UTF-8/LF; known local/test names/paths redacted'
    out=DEST/relative;out.parent.mkdir(parents=True,exist_ok=True);out.write_bytes(data)
    entries.append({'path':str(relative),'original_sha256':sha(raw),'archived_sha256':sha(data),'bytes':len(data),'transformation':change})
for prefix in ['native','native-full-flow']:
    for path in sorted((ROOT/prefix).rglob('*')):
        if path.is_file():
            assert path.suffix in ('.json','.log','.ps1','.png','.toml')
            archive(path,Path(prefix)/path.relative_to(ROOT/prefix))
names=[
 'paste-red.log','paste-green.log','footer-chips-red.log','footer-chips-green.log','footer-badge-red.log','footer-badge-green.log','footer-anchor-red.log',
 'ui-test-field-first-build.log','ui-regression.log','ui-final.log','target-red.log','target-green.log',
 'go-test.log','go-race.log','go-vet.log','go-mod-verify.log','source-limits.log','catalog-check.log',
 'upstream-footer-reference_test.go','product-footer-reference_test.go','upstream-overlay.json','product-overlay.json','upstream-reference.log','product-reference.log',
 'upstream-reference.json','product-reference.json','product-reference-before-target.json','upstream-source.json',
 'source-fingerprints.json','source-fingerprints-before-target.json','binary-hashes.json','binary-hashes-before-target.json','label-source-delta.json',
 'product-build-info.txt','product-build-info-before-target.txt','color-serialization.json','compare-footer-roles.py','compare-footer-roles-first.py','role-differences-first.json','role-comparison-first.log','role-comparison.json','role-comparison-before-target.json','role-comparison.log',
 'native-run.log','native-label-run.log','run-footer-input-tests.ps1','run-footer-label-tests.ps1',
 'setup-footer-input-desktop.ps1','start-footer-input-setup.ps1','start-footer-input-desktop.ps1','footer-input-desktop-driver.ps1','inspect-footer-driver-ready.ps1','preflight-footer-input.ps1',
 'footer-input-flow.ps1','footer-input-flow-fixed.ps1','footer-input-flow-resume.ps1','footer-input-flow-resume-fixed.ps1',
 'inspect-footer-wait.ps1','inspect-after-wait-failure.json','inspect-resume-shape.ps1','resume-shape.json','first-flow-failure-observed.json','resume-preflight-failure.json',
 'desktop-preflight.log','desktop-setup.log','desktop-start.log','desktop-ready.log','desktop-flow.log','desktop-resume.log','desktop-resume-fixed.log',
 'cleanup-export-footer-input.ps1','cleanup-export.log',
 'setup-footer-label-desktop.ps1','start-footer-label-setup.ps1','start-footer-label-desktop.ps1','footer-label-desktop-driver.ps1','inspect-label-driver-ready.ps1','preflight-footer-label.ps1','footer-label-flow.ps1',
 'desktop-label-preflight.log','desktop-label-setup.log','desktop-label-start.log','desktop-label-ready.log','desktop-label-flow.log','cleanup-export-footer-label.ps1','cleanup-label-export.log',
 'archive-footer-input.py','verify-footer-input.py',
]
for name in names:archive(ROOT/name,Path('local')/(name+'.txt' if name.endswith('.go') else name))
inputs={}
for overlay in ['upstream-overlay.json','product-overlay.json']:
    for value in sorted(set(json.loads((ROOT/overlay).read_text())['Replace'].values())):
        p=Path(value);relative=Path('local')/(p.name+'.txt') if p.parent==ROOT else Path('local/overlay-inputs')/(p.parent.name+'-'+p.name+'.txt')
        if p.parent!=ROOT and value not in inputs:archive(p,relative)
        inputs[value]=str(relative)
p=ROOT/'overlay-inputs.json';p.write_text(json.dumps(inputs,indent=2)+'\n');archive(p,Path('local/overlay-inputs.json'))
archive(Path('/private/tmp/asd-upstream-compare-20261123/reference.sb'),Path('local/reference.sb'))
manifest={'base_commit':'d3e1025','source_state':'base plus final archived Go/module fingerprints','scope':'Default footer and help paste parity; full input build and bounded final label repair are distinct evidence sets, not full UI/release acceptance','full_flow_zip_sha256':sha((ROOT/'native-full-export.zip').read_bytes()),'final_label_zip_sha256':sha((ROOT/'native-export.zip').read_bytes()),'artifacts':entries}
(DEST/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n');print(json.dumps({'artifacts':len(entries)}))
