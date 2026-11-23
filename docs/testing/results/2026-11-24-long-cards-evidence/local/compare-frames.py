from pathlib import Path
import json
out=Path(__file__).resolve().parent
source=json.loads((out/'upstream-frames.json').read_text());before=json.loads((out/'product-before.json').read_text());after=json.loads((out/'product-after.json').read_text())
assert len(source)==len(before)==len(after)==8
cases=[]
for s,b,a in zip(source,before,after):
 assert s['name']==b['name']==a['name']
 cases.append(dict(name=s['name'],title_before=b['plain'][0]==s['crop_padded_for_product_surface'],title_after=a['plain'][0]==s['crop_padded_for_product_surface'],metadata_before=b['plain'][1:]==s['metadata_rows_padded'],metadata_after=a['plain'][1:]==s['metadata_rows_padded'],source_final_pad_cells=s['width']-s['crop_width'],all_widths_exact=all(w==a['width'] for w in a['line_widths'])))
result={'scope':'inert text crop/wrap with product prefixes, ANSI stripped and explicit final-cell padding; not original full-frame/RGB equivalence','cases':cases,'summary':{k:sum(c[k] for c in cases) for k in ['title_before','title_after','metadata_before','metadata_after','all_widths_exact']}}
(out/'comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result['summary']))
