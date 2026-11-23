from pathlib import Path
import json
report='testing/results/2026-11-23-custom-groups.md'
adr='adr/0023-persistent-groups-and-membership.md'
p=Path('docs/adr/0023-persistent-groups-and-membership.md');s=p.read_text();s=s.replace('名称最多 50 code point／200 UTF-8 字节，目录最多 4096 字节','原始名称输入最多 4096 UTF-8 字节；清理后名称最多 50 code point／200 UTF-8 字节，目录最多 4096 字节');s=s.replace('用户仍可在启动表单覆盖目录。','用户仍可在启动表单覆盖或清空目录。组内启动前验证目标组；原 Run 和初始成员归属在一个元数据事务提交，外部进程创建仍不原子，失败报告实际身份且不自动重启／停止。');s=s.replace('不自动写偏好；组名变化不丢失折叠归属。','不自动写偏好；即使先前保存尚在进行，临时揭示也不排队追加保存。只有明确接受的偏好操作进入后续保存队列。组名变化不丢失折叠归属。');s+='\n## 实现与证据\n\n当前已接入持久树／成员版本、事务迁移与恢复、管理 g／M、组内启动目录、嵌套侧栏和 ui.toml v3。实际本机与 Windows 模型检查分别记于[执行报告](../testing/results/2026-11-23-custom-groups.md)；新桌面、来源画面对照、完整组编辑与整体接受仍开放。\n';p.write_text(s)
p=Path('docs/adr/README.md');s=p.read_text();needle='| [0022](0022-default-group-sidebar-navigation.md)';line=next(x for x in s.splitlines() if x.startswith(needle));s=s.replace(line,line+'\n| [0023](0023-persistent-groups-and-membership.md) | 稳定分组身份、嵌套创建／移动及 SQLite v4／ui.toml v3 | 已接受 | 已接入元数据与 Tab 侧栏；本机／Windows 模型及桌面范围见[报告](../testing/results/2026-11-23-custom-groups.md)，完整组编辑／体验接受仍开放 |');p.write_text(s)
p=Path('README.md');s=p.read_text();old='当前仅有默认组，自定义组创建和移动待实现。密度、视图模式和折叠共用 `ui.toml` v2；v1 只读加载，显式修改后升级。';new='管理 `g` 创建根组／子组（Tab 切名称／默认目录、↑↓ 切层级），`M` 移动所选会话；组内 `n` 预填默认目录，既有 Run 不执行 cd。空组和成员关系独立保存，移动不改变终端身份。密度、模式和稳定 Group ID 折叠共用 `ui.toml` v3；v1/v2 只读加载，显式修改后升级。';assert old in s;s=s.replace(old,new);s+='\n自定义组的 SQLite v4 迁移、备份／恢复与实际检查见[分组报告](docs/'+report+')及 [ADR-0023](docs/'+adr+')。组重命名／删除／重排、新桌面完整流程和上游画面对照仍未完成，不提升整体验收。\n';p.write_text(s)
p=Path('docs/ROADMAP.md');s=p.read_text();s=s.replace('下一步接入自定义组创建与移动，继续统一上游差异与最终资源／发布门槛；阶段状态不提升。','后续 [ADR-0023]('+adr+') 已接入自定义根组／子组、移动和持久成员关系，实际检查见[分组报告]('+report+')。下一步完成新桌面流程与来源画面对照，再补组重命名／删除／重排，继续最终资源／发布门槛；阶段状态不提升。');p.write_text(s)
p=Path('docs/DECISIONS.md');s=p.read_text();s+='\n2026-11-23：[ADR-0023]('+adr+') 增加稳定 Group ID 和逻辑 Session 成员版本，SQLite v4 与 ui.toml v3；g／M 和默认目录沿固定来源适配。组名／目录不作终端寻址，移动只改元数据，组内启动的 Run／成员在一个元数据事务提交；临时揭示不排队保存。实现与 Windows／桌面证据分开记录，完整组编辑和上游体验继续开放。\n';p.write_text(s)
p=Path('docs/ARCHITECTURE.md');s=p.read_text();s+='\n分组扩展见 [ADR-0023]('+adr+')：schema v4 追加作用域内稳定 `groups` 树和 `session_groups` 单调成员版本，当前逻辑归属与原 Run 启动快照分离。Manager 先记录原始发现，再在同一只读元数据事务投影名称／组树／成员；移动在独立事务核对完整原目标和预期成员版本。组内启动先校验组，再创建外部终端，最后原子提交 Run 与初始成员；失败不会假定进程未创建。侧栏用同一可见树做键盘、渲染和鼠标命中，Tab／终端输入身份不随移动改变。\n';p.write_text(s)
p=Path('docs/METADATA.md');s=p.read_text();s+='\n## 持久分组（schema v4）\n\n按 [ADR-0023]('+adr+')，当前正常打开会在一个事务中将 v1/v2/v3 升级至 v4；原名称、Run、停止／观测时间不变。旧程序拒绝新格式，没有降级迁移。离线 check／backup／restore 可只读处理 v1–v4，旧备份只在正常打开恢复目标时升级；输出文件仍不包含在数据库快照中。分组树的循环、缺失父组和防御边界无效时拒绝检查／备份，保留源库。\n\n当前 psmux Tab 管理焦点下 g 创建根组或子组；Tab／Shift+Tab 切换名称和默认目录，↑↓ 切换根组／子组，Enter 确认、Esc 取消。M 将所选 Session 移到稳定目标组，原 Run、后端目标及 cwd 不变。空组保留，重开仍有原组／成员关系；历史展示当前逻辑组，并不改写原启动证据。组内 n 创建的新终端预填默认目录，可覆盖或清空。提交错误要求返回、刷新核对，不自动重试可能已完成的操作。\n\n成员版本防止旧窗口／多 UI 的 ABA 覆盖；名称版本独立。每作用域最多 256 组（含默认组）、最多 16 层，原始名称 4096 字节、清理后 50 code point／200 字节、目录 4096 字节是防御边界，未作为资源预算验收。组重命名、删除／重排及全量迁移异常矩阵仍待后续实现／验证；当前范围见[分组报告]('+report+')。\n';p.write_text(s)
p=Path('docs/TECH_STACK.md');s=p.read_text();s=s.replace('完整组树／创建移动与视觉接受仍未完成','另已接入持久嵌套组与创建／移动；完整组编辑与视觉接受仍未完成');s=s.replace('v1/v2→v3 事务迁移和离线检查／非覆盖快照恢复已实现','后续 [v4 持久组／成员版本]('+adr+')、v1/v2/v3→v4 事务迁移及离线检查／非覆盖快照恢复已接入');p.write_text(s)
p=Path('docs/FEATURE_LEDGER.md');s=p.read_text();features={'FL-001','FL-003','FL-004','FL-005','FL-011','FL-012','FL-013','FL-018','FL-019'};lines=s.splitlines();
for i,line in enumerate(lines):
 if any(line.startswith('| '+f+' |') for f in features):
  lines[i]=line.replace(' / 未发布 |','；[持久组／创建移动]('+report+') / 未发布 |')
s='\n'.join(lines)+'\n';marker='## 4.';idx=s.find(marker);note='2026-11-23 持久分组：FL-001/003/004/005/011/012/013/018/019 增加 [ADR-0023]('+adr+') 的稳定组树、成员版本 CAS、SQLite v4、g／M、组内新建目录和 ui.toml v3。移动不重启或写终端，空组／当前逻辑归属可重开恢复，旧迁移与快照保留原 Run 证据。实际本机／Windows 模型和桌面范围见[报告]('+report+')；组重命名／删除／重排、来源画面对照、真实 agent 与完整接受仍开放，各项保持实现中／未发布，339 项设计基线不提升。\n\n';assert idx>=0;s=s[:idx]+note+s[idx:];p.write_text(s)
for name in ['01-launch-terminal-ui.json','02-sessions-recovery-storage.json']:
 p=Path('docs/testing/cases')/name;x=json.loads(p.read_text())
 for c in x['cases']:
  if c['id'] in ['TC-UI-002','TC-UI-004','TC-UI-005','TC-SESSION-008','TC-STORE-013']:
   if 'ADR-0023' not in c['adrs']:c['adrs'].append('ADR-0023')
  if c['id']=='TC-UI-002':c['steps'].append('零会话用 g 创建根组／子组并重开；取消后核对无 Run，空组仍保留。')
  if c['id']=='TC-UI-004':c['steps'].append('创建嵌套组，在 g 名称／目录及 M 移动列表中缩放、切字段、滚动选择；核对真实光标和末项可见。')
  if c['id']=='TC-UI-005':c['steps'].append('保存折叠尚未返回时切 Tab 揭示嵌套目标，核对没有新增偏好写入；分组提交失败后返回刷新核对，禁止自动重试。')
  if c['id']=='TC-SESSION-008':
   c['preconditions'].append('按 ADR-0023 建立稳定 Group ID、父组和可选默认目录；原 Run 与原 cwd 已独立记录。')
   c['steps'].insert(0,'用 g 创建根组和子组，Tab 切名称／目录、↑↓ 切层级；同父同清理名同目录返回原组，不同目录拒绝。')
   c['steps'].append('用两个 UI 或连接从同一成员版本并发移动，另执行甲→乙→甲后提交原旧窗口；退出重开并检查名称、组树、成员及原 Run。')
   c['expected'].extend(['成员关系归属逻辑 Session，移动使用完整原 Session/Run/Backend 和预期成员版本；ABA 或过期窗口拒绝覆盖。','组内新建可预填／覆盖默认目录；Run 与初始归属一并提交，外部终端创建不假定原子。','组重排仍属完整用例的后续能力；当前创建／移动的局部模型结果不能宣称整条通过。'])
 p.write_text(json.dumps(x,ensure_ascii=False,indent=2)+'\n')
