from pathlib import Path
import json
src=json.loads(Path('/private/tmp/asd-custom-groups-20261123/upstream-source.json').read_text())
hashes={x['path']:x['sha256'] for x in src['files']}
p=Path('third_party/reuse-manifest.json');x=json.loads(p.read_text())
items=[('internal/session/groups.go','internal/session/group.go',['group name sanitizer']),('internal/ui/group_dialog.go','internal/ui/group_dialog.go',['create name/default path fields','g context root/subgroup, Up/Down and Tab focus','move group list Enter/Esc']),('internal/ui/group_dialog.go','internal/ui/group_dialog_view.go',['14-cell field labels and responsive input widths','root/subgroup context and parent label','centered create/move layout and bounded hints']),('internal/ui/home.go','internal/ui/group_dialog.go',['g on header vs Session and M selected-Session context'])]
for source,dest,parts in items:
 existing=next((e for e in x['source_files'] if e.get('source_file')==source and e.get('destination_file')==dest),None)
 entry={'project':'agent-deck','repository':'https://github.com/asheshgoplani/agent-deck','commit':src['commit'],'source_file':source,'source_sha256':hashes[source],'destination_file':dest,'license':'MIT','parts':parts,'mode':'selectively_adapted','modifications':'ADR-0023: stable product Group IDs, scoped SQLite tree/membership and CAS; v2 real cursor/total widths, Chinese labels, token styles, bounded viewports and original Run/input fences. No filesystem path IDs, upstream remote/worktree/provider model or v1 production dependency.'}
 if existing:existing.update(entry)
 else:x['source_files'].append(entry)
for e in x['source_files']:
 if e.get('destination_file') in ['internal/ui/sidebar_navigation.go','internal/ui/sidebar_group_view.go']:
  e['modifications']='ADR-0022 default group extended by ADR-0023 to stable persisted nested product groups; source root hotkeys, carets, level indentation and token roles retained. Session/Run targets and native input remain independent; no inferred agent state or upstream remote model.'
p.write_text(json.dumps(x,ensure_ascii=False,indent=2)+'\n')
p=Path('docs/REUSE_MANIFEST.md');s=p.read_text();s+='\n2026-11-23 持久自定义组：ADR-0023 从同固定 Agent Deck 提交选择性改编 `internal/session/groups.go` 的名称清理，以及 `group_dialog.go`／`home.go` 的 g 上下文、根组／子组、名称／默认目录字段、Tab／↑↓ 和 M 选择目标组。目标为 `internal/session/group.go`、`internal/ui/group_dialog.go`、`group_dialog_view.go`，既有组标题／导航／卡片延伸为嵌套树。14 单元格标签、响应式字段、紫色框、accent 模式／选项和灰色占位均经 v2／tokens 适配。稳定 ID、SQLite v4、成员 CAS、事务快照、ui.toml v3 和输入排空是本产品连接规则；来源的路径补全／最近目录下拉、远程／worktree、组改名／删除／重排未随本切片迁入。来源 SHA／Git blob 见[报告](testing/results/2026-11-23-custom-groups.md)及机器清单；MIT 与原版权保持。\n';p.write_text(s)
p=Path('THIRD_PARTY_NOTICES.md');s=p.read_text();s+='\nADR-0023 继续选择性适配同固定 Agent Deck 的 `internal/session/groups.go` 名称清理及 `internal/ui/group_dialog.go`／`home.go` 分组创建、根组／子组上下文、默认目录字段与移动弹窗；组标题／卡片／导航沿既有来源扩展。保留 MIT 及原版权，逐文件 SHA 在复用清单。产品使用独立稳定身份、元数据和 v2 输入，不继承上游路径寻址／远程模型或 v1 发布依赖。\n';p.write_text(s)
p=Path('docs/M0_DEVELOPMENT.md');s=p.read_text();s+='\n### 自定义组创建与移动（psmux Tab）\n\n管理焦点 g 打开创建分组；组标题默认子组，会话行默认根组。Tab／Shift+Tab 在名称与可选默认目录间切换，↑↓ 切换根组／子组，Enter 确认，Esc 取消。M 为所选会话选择目标组；终端焦点的 g／M 仍原样交给程序。组内 n 预填默认目录，可覆盖或清空，只影响新 Run。1–9 选择根组，标题 Enter／→／Tab 折叠；空组保留。\n\nSQLite 当前 v4；正常打开会事务迁移 v1/v2/v3，旧二进制拒绝新 schema。独立 ui.toml v3 保存稳定 Group ID 折叠，旧 v1/v2 只读加载；Tab／搜索临时展开不自动落盘，包括先前偏好保存尚未返回的情况。完整合同与实际范围见 [ADR-0023](adr/0023-persistent-groups-and-membership.md)及[分组报告](testing/results/2026-11-23-custom-groups.md)。\n';p.write_text(s)
