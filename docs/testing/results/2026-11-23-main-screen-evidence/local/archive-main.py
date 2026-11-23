"""Archive only allowlisted public synthetic main-screen evidence."""
import hashlib,json,re
from pathlib import Path
R=Path('/private/tmp/asd-main-parity-20261123');D=Path.cwd()/'docs/testing/results/2026-11-23-main-screen-evidence';D.mkdir(parents=True,exist_ok=True)
def sha(b):return hashlib.sha256(b).hexdigest()
entries=[]
def put(p,n):
    raw=p.read_bytes();enc='utf-16' if raw.startswith((b'\xff\xfe',b'\xfe\xff')) else 'utf-8-sig';s=raw.decode(enc)
    s=s.replace(str(Path.cwd()),'<REPO>').replace('/Users/'+'liuhui','<LOCAL_TEST_HOME>');s=re.sub('liu'+' xun yuan','<WINDOWS_TEST_USER>',s,flags=re.I);b=s.replace('\r\n','\n').encode('utf-8')
    dest=D/n;dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(b)
    entries.append({'path':str(n),'original_sha256':sha(raw),'archived_sha256':sha(b),'bytes':len(b),'transformation':enc+' to UTF-8/LF; known local/test paths redacted'})
for prefix in ('native','native-allmotion','native-final'):
    for p in sorted((R/prefix).rglob('*')):
        if not p.is_file():continue
        if p.suffix=='.png':
            if p.name not in ('main-nested-selected-120-public.png','main-divider-hover-120-public.png','main-divider-pressed-120-public.png'):continue
            raw=p.read_bytes();dest=D/prefix/p.relative_to(R/prefix);dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(raw);entries.append({'path':str(dest.relative_to(D)),'original_sha256':sha(raw),'archived_sha256':sha(raw),'bytes':len(raw),'transformation':'Unchanged inspected PNG'});continue
        assert p.suffix in ('.json','.log','.ps1','.toml');put(p,Path(prefix)/p.relative_to(R/prefix))
names='''nested-surface-red.log nested-surface-green.log divider-initial-check.log divider-red.log divider-final-red.log main-green.log main-green-first-fixture-failure.log ui-final-first-fixture-failure.log ui-regression.log ui-final.log go-test.log go-race.log go-vet.log go-mod-verify.log source-limits.log catalog-render.log catalog-check.log source-fingerprints.json binary-hashes.json product-build-info.txt source-execution-review.json upstream-source.json upstream-main-reference_test.go upstream-main-reference.json upstream-reference.log upstream-reference-isolated.log upstream-overlay.json product-main-reference_test.go product-reference-first-build.go.txt product-main-reference.json product-main-before.json product-overlay.json product-before-overlay.json product-reference.log product-before.log divider-red-overlay.json workspace_view-before.go workspace_view.go.before workspace_sidebar.go.before compare-main-roles.py role-comparison.json role-comparison.log native-run.log run-main-parity-tests.ps1 setup-main-parity-desktop.ps1 start-main-parity-setup.ps1 start-main-parity-desktop.ps1 desktop-driver-first.ps1 desktop-driver-guarded-backup-not-executed.ps1 main-parity-desktop-driver.ps1 main-parity-flow.ps1 inspect-label-driver-ready.ps1 preflight-main-parity.ps1 inspect-main-driver-start.ps1 inspect-main-driver-run2.ps1 resume-main-driver-preflight.ps1 preflight-main-input.ps1 probe-main-desktop.ps1 prepare-blocked-cleanup.ps1 cleanup-blocked-main.ps1 finalize-blocked-main.ps1 desktop-preflight.log desktop-setup.log desktop-start.log desktop-ready.log desktop-flow.log desktop-driver-start-inspection.json desktop-resume-preflight.log desktop-input-preflight.log desktop-start-run2.log desktop-ready-run2.log desktop-driver-run2-inspection.json desktop-availability.json blocked-cleanup-preparation.log blocked-cleanup.log blocked-finalize.log archive-main.py verify-main.py archive-local-failures.json mouse-mode-red.log mouse-mode-green.log mouse-mode-docs.json allmotion-source-delta.json source-fingerprints-before-allmotion.json binary-hashes-before-allmotion.json product-build-info-before-allmotion.txt go-test-before-allmotion.log go-race-before-allmotion.log go-vet-before-allmotion.log go-mod-verify-before-allmotion.log source-limits-before-allmotion.log main-green-before-allmotion.log run-main-allmotion-tests.ps1 native-allmotion-run.log mouse-probe-overlay.json workspace-mouse-probe.go mouse-trace-helper.go'''.split()
for n in names:put(R/n,Path('local')/(n+'.txt' if n.endswith('.go') else n))
for p in sorted((R/'desktop-final').glob('*.ps1')):put(p,Path('local/desktop-final')/p.name)
for n in ('desktop-preflight.log','desktop-setup.log','desktop-start.log','continuation-start.log','controller-replacement.log','native-controller-start.log','formal-native-flow.log','responsive-readonly.log','final-tail.log','cleanup-final.log','native-progress.json','responsive-progress.json','responsive-review-rejection.json','native-move-probe2.log','native-move-probe3.log','move-task-inspection.json','actual-build-hashes.json'):
    put(R/'desktop-final'/n,Path('local/desktop-final')/n)
overlay_inputs={}
for n in ('upstream-overlay.json','product-overlay.json','product-before-overlay.json','divider-red-overlay.json','mouse-probe-overlay.json'):
    for v in set(json.loads((R/n).read_text())['Replace'].values()):
        p=Path(v);target=Path('local')/(p.name+'.txt' if p.name.endswith('.go') else p.name)
        if p.parent!=R:target=Path('local/overlay-inputs')/(p.parent.name+'-'+p.name+'.txt');put(p,target)
        overlay_inputs[v]=str(target)
p=R/'overlay-inputs.json';p.write_text(json.dumps(overlay_inputs,indent=2)+'\n');put(p,Path('local/overlay-inputs.json'))
put(Path('/private/tmp/asd-upstream-compare-20261123/reference.sb'),Path('local/reference.sb'))
m={'base_commit':'20574bb','scope':'Nested card/divider role repair and synthetic main-screen audit; Final native UI and bounded 269-receipt desktop passed; earlier blocked set separate. First clipboard restoration unknown, final terminal bodies excluded.','native_zip_sha256':sha((R/'native-blocked-export.zip').read_bytes()),'final_public_zip_sha256':sha((R/'final-public-export.zip').read_bytes()),'privacy':'Final protocol/captures exclude terminal text/tree and non-fixture lines. Embedded original_json_sha256 values hash decoded UTF-8 JSON text; candidate 49-column PNG omitted for window-edge content. No no-interference claim.','artifacts':entries}
(D/'manifest.json').write_text(json.dumps(m,indent=2)+'\n');print(json.dumps({'artifacts':len(entries)}))
