"""Archive allowlisted help evidence; never copy runtime databases or credentials."""
import hashlib
import json
from pathlib import Path
import re
ROOT=Path('/private/tmp/asd-help-parity-20261123')
DEST=Path.cwd()/'docs/testing/results/2026-11-23-help-parity-evidence'
DEST.mkdir(parents=True,exist_ok=True)
entries=[]
def sha(data):
    return hashlib.sha256(data).hexdigest()
def archive(source,relative):
    raw=source.read_bytes();data=raw;transformation='unchanged PNG'
    if source.suffix!='.png':
        encoding='utf-16' if raw.startswith((b'\xff\xfe',b'\xfe\xff')) else 'utf-8-sig'
        text=raw.decode(encoding).replace(str(Path.cwd()),'<REPO>')
        text=text.replace('/Users/'+'liuhui','<LOCAL_TEST_HOME>')
        text=re.sub('liu'+' xun yuan','<WINDOWS_TEST_USER>',text,flags=re.I)
        data=text.replace('\r\n','\n').encode('utf-8')
        transformation=encoding+' to UTF-8/LF; known local/test usernames and paths redacted'
    out=DEST/relative;out.parent.mkdir(parents=True,exist_ok=True);out.write_bytes(data)
    entries.append({'path':str(relative),'original_sha256':sha(raw),'archived_sha256':sha(data),'bytes':len(data),'transformation':transformation})
for prefix in ['native','native-before-surface']:
    for path in sorted((ROOT/prefix).rglob('*')):
        if path.is_file():
            assert path.suffix in ('.json','.log','.ps1','.png','.toml')
            archive(path,Path(prefix)/path.relative_to(ROOT/prefix))
names=[
 'help-red.log','help-green.log','help-first-build.log','scroll-red.log','scroll-green.log',
 'help-regression.log','ui-before-tail.log','tail-row-red.log','tail-row-green.log',
 'surface-red.log','surface-green.log','source-limits.log','catalog-check.log',
 'go-test.log','go-race.log','go-vet.log','go-mod-verify.log',
 'go-test-before-surface.log','go-race-before-surface.log','go-vet-before-surface.log','go-mod-verify-before-surface.log',
 'source-fingerprints.json','source-fingerprints-first.json','source-fingerprints-before-surface.json',
 'binary-hashes.json','binary-hashes-first.json','binary-hashes-before-surface.json',
 'product-build-info.txt','upstream-source.json','approval-retry.json',
 'upstream-reference.log','product-reference.log','upstream-help-frames.json',
 'product-help-frames.json','product-help-frames-first.json','product-help-frames-before-surface.json',
 'help-geometry.json','help-geometry-first.json','help-geometry-before-surface.json','compare-help-geometry.py',
 'upstream-help-reference_test.go','product-help-reference_test.go',
 'upstream-help-overlay.json','product-help-overlay.json',
 'native-run.log','native-run-final.log','native-run-style.log',
 'setup-help-parity-desktop.ps1','start-help-parity-setup.ps1','start-help-parity-desktop.ps1',
 'help-parity-desktop-driver.ps1','help-parity-flow.ps1','inspect-driver-ready.ps1',
 'preflight-help-parity.ps1','run-help-parity-tests.ps1','cleanup-export-help-parity.ps1',
 'desktop-preflight.log','desktop-setup.log','desktop-start.log','desktop-ready.log','desktop-flow.log','cleanup-export.log',
 'setup-help-parity-style-desktop.ps1','start-help-parity-style-setup.ps1','start-help-parity-style-desktop.ps1',
 'help-parity-style-desktop-driver.ps1','help-parity-style-flow.ps1','inspect-style-driver-ready.ps1',
 'preflight-help-parity-style.ps1','run-help-parity-style-tests.ps1','cleanup-export-help-parity-style.ps1',
 'desktop-style-preflight.log','desktop-style-setup.log','desktop-style-start.log','desktop-style-ready.log','desktop-style-flow.log','cleanup-style-export.log',
 'archive-help-parity.py','verify-help-parity.py',
]
for name in names:
    archive(ROOT/name,Path('local')/(name+'.txt' if name.endswith('.go') else name))
inputs={}
for overlay in ['upstream-help-overlay.json','product-help-overlay.json']:
    for value in sorted(set(json.loads((ROOT/overlay).read_text())['Replace'].values())):
        path=Path(value)
        relative=Path('local')/(path.name+'.txt') if path.parent==ROOT else Path('local/overlay-inputs')/(path.parent.name+'-'+path.name+'.txt')
        if path.parent!=ROOT and value not in inputs:archive(path,relative)
        inputs[value]=str(relative)
p=ROOT/'overlay-inputs.json';p.write_text(json.dumps(inputs,indent=2)+'\n');archive(p,Path('local/overlay-inputs.json'))
archive(Path('/private/tmp/asd-upstream-compare-20261123/reference.sb'),Path('local/reference.sb'))
manifest={'base_commit':'00576dd','product_source_state':'base plus archived final Go/module fingerprints',
 'scope':'Fixed-source UI difference audit and help parity; two actual desktop flows, not overall acceptance',
 'zip_sha256':sha((ROOT/'native-export.zip').read_bytes()),'zip_before_surface_sha256':sha((ROOT/'native-before-surface-export.zip').read_bytes()),
 'redaction':'Known local/test usernames and paths only; PNG bytes unchanged','artifacts':entries}
(DEST/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'artifacts':len(entries),'final_zip_sha256':manifest['zip_sha256']}))
