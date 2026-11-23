# 删除组结构、成员收纳与原终端保留

日期：2026-11-23；基于 `8e8170b` 加本轮修改，196 份 Go／模块指纹见[源清单](2026-11-23-group-removal-evidence/local/source-fingerprints.json)。管理焦点选中组标题按 d，先核对影响，默认取消；明确确认后把逻辑成员移入默认组，再删除该组及子组结构。原 Session/Run/Tab、终端、历史和输出记录保留。完整合同见 [ADR-0024](../../adr/0024-group-removal-and-member-preservation.md)，SQLite v4、ui.toml v3 和依赖不变；本轮不代表完整上游体验或发布接受。

## 来源、界面与必要适配

固定干净 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT。confirm_dialog.go／home_keys_dialogs.go 的默认 Cancel、红框／黄警告、按钮和键位，以及 home.go／session/groups.go 的默认组保护、子树删除与成员收纳语义，选择性迁入；[来源 SHA／Git blob](2026-11-23-group-removal-evidence/local/upstream-source.json)与机器复用清单同步。

隔离辅助测试实际执行未修改的来源 DeleteGroup：默认组拒删，a／b 移入默认组，a／b／c 均保留，剩余 default 与 other；见[来源结果](2026-11-23-group-removal-evidence/local/upstream-removal-core.json)。源码辅助测试通过 overlay 隔离原测试／TestMain，Darwin sandbox 禁止个人目录与出站网络；未运行完整上游应用或个人会话。overlay 与输入映射一并归档，可在重新固定源码后重映射路径。

来源和产品各执行 120×32、63×25、49×25 的取消焦点／删除焦点／默认组提示，共九对实际组件帧。外宽分别为 52／52／41，水平位置一致；[几何结果](2026-11-23-group-removal-evidence/local/removal-geometry.json)保留真实高度：来源删除 19 行，产品宽窗 17、窄窗 18 行；提示为 16／15 行。逐行检查确认标题、警告、详情、按钮间留白一致，高度差来自英文详情／键位提示的换行与中文本地化，不人为添加空行。此处是组件执行及单元格证据，不能冒充 Windows 上运行上游完整应用。

终端焦点的 d/y/n 仍传给原 CLI；管理组标题 d 打开预览，←/h 选删除，→/l/Tab 选取消，Enter 激活当前项，y/Y 确认，n/N/Esc 取消。默认组只显示不能删除的提示。预览未完成或完整提示放不下时不能确认；组成员包含历史逻辑 Session，不只数当前活终端。

预览印记覆盖完整作用域组树和受影响成员 ID／归属／版本；同数量替换、ABA、默认组名称或其他组树变化均拒绝旧预览。单个 immediate 事务重新核对，再移动成员并各加一版本、按子组优先删组，SQL／投影失败整体回滚。Manager 不调用终端停止／输入。成员流式扫描，不保存无界成员副本；上下文期限和既有组／层级限制是防御边界，不是资源预算通过。

删除成功仅清理对应折叠 ID；临时 Tab 展开不落盘。已接受偏好、活动写入和临时可见状态分开保存，迟到写入失败不能恢复已删折叠，未知／其他组偏好保留。抑制集合只保留当前已保存／活动写入引用，随回执回收；数据库与 TOML 不原子，保存失败明确报告，不能撤销或暗中重试删除。迟到发现／改名／排序结果也不能复活已删组。

## RED → GREEN 与本机检查

- Store 首轮因缺少删除接口失败，同时有测试误用 `StoppedAt` 字段；改为实际 `StopConfirmedAt` 后完成接口，真实数据库验证子树／成员收纳、历史成员、原 Run／输出及重开。原失败日志完整保留，不把测试字段错误算成产品缺陷。
- Manager 首轮缺少转发接口；接入后 nil 终端对象也可预览／删除，证明容器操作不依赖终端调用。UI 首轮管理 d 无行为，随后实现只读预览、默认取消和异步确认。
- 短窗口仍能确认的行为测试失败，修正为完整提示适配窗口后才提交；旧排序／改名回执能重建已删组的两个回归分别失败，按完整观察代次／稳定身份合并后通过。
- 偏好测试发现删除清理会保存另一 Tab 的临时展开，且旧写入失败会恢复已删折叠。将明确接受文档、活动写入与临时可见状态分开，过滤并回收已删引用后通过。
- 补充真实双连接竞争仅一笔提交、预览读锁、整棵组树变化、同数量替换／ABA、成员版本耗尽、投影失败／父组 DELETE 失败整体回滚、默认／跨作用域／伪造预览拒绝，以及 UI 加载取消、输入排空、忙时防重复和较新 Run／版本保留。这些增补不全部声称独立经历 RED。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 均通过，含缓存结果。source limits 和 catalog 检查另归档；TC-SESSION-008 增加设计范围，仍为全部 339 项 designed/not_run、216 ready／123 needs_decision，文档检查不计作产品执行。

## Windows 原生与桌面结果

Windows 10 build 19045，实际执行三个测试程序，计数不累加重复构建：

| 套件 | 顶层／含子测试 PASS | FAIL／SKIP |
| --- | --- | --- |
| Store | 65／99 | 0／0 |
| Manager | 16／18 | 0／0 |
| UI | 157／341 | 0／0 |

[原生记录](2026-11-23-group-removal-evidence/native/native-tests/outcome.json)保留退出码、日志和三个精确测试映像无残留。Store 为真实 SQLite／文件，UI 服务 fixture 为模型合同；完整后端原生场景本轮未重跑。

实际桌面产品 SHA-256 `cf766c160dba91f7af3d63b1ec5ad2827eb67b015ef74e590e81176b342c4aad`，基于 8e8170b 加本轮未提交源码构建，指纹与本轮最终 Go 代码相同，不称为提交后重建。使用 ADR-0019 显式开发 bundle、Windows Terminal 1.25.3131.0 和 PowerShell 7.7.0，普通用户桌面 Session 2；不提升官方默认依赖或完整 Windows 兼容结论。

146 条连续回执均成功：41 keys、62 state、25 Unicode text、10 capture、3 paste、4 resize、1 finish；只读轮询不算额外能力。本次完整流程无失败重放。执行前修正了重开后选择默认组标题可能让 Enter 折叠组的脚本假设，改为搜索稳定原目标再进入；没有用额外 Enter 冒险提交草稿。

- A 保留未提交命令并追加字面 d/y/n，B 保留独立草稿；创建“项目 甲／审核／保留”，分别将原 A／B 移入子组／根组。两个逻辑成员及原 Run 均独立记录。
- 默认组提示拒删；Enter 默认取消、n 取消后组／成员／Run 和偏好不改。← 聚焦删除、Tab 回取消；120／63／49 列完整审阅，短高度显示放大提示并拒绝 y。
- 新预览后用独立进程调用真实 Store.RenameGroup，把无关“保留”改为“保留 更新”、版本 1→2；旧预览 y 明确拒绝，重复 y 不重试。辅助程序仅是本轮测试夹具，不是新增产品 CLI 或终端注入。
- 重新读取预览后明确确认，组树由四组变两组，两成员归 default、版本各由 1→2；其他组改名保留，原 Run 记录不变。数据库备份和 UI 重开保留同样结果。
- 三份偏好快照证明仅删根／子组折叠，保留其他组与预置未知 ID；三次读取同一显式 screen_snapshot，record/Session/Run/backend/text/sequence 全相同，complete 仍 false，不升级成连续日志。
- 12 次原生 capture 均保留编号 1–40，各一次。10 次未提交观察的最后非空行与完整草稿逐字一致；最后明确 Enter 后，各仅一次 `REMOVE_A:19332`／`REMOVE_B:12024`，对应原 PID。退出 UI 和重新连接未替换原进程、Run 或草稿。

十张未经编辑的 PNG 已逐张检查：[默认组保护](2026-11-23-group-removal-evidence/native/desktop/removal-protected-default-120-public.png)、[默认取消](2026-11-23-group-removal-evidence/native/desktop/removal-cancel-default-120-public.png)、[删除焦点](2026-11-23-group-removal-evidence/native/desktop/removal-delete-focused-120-public.png)、[Tab 回取消](2026-11-23-group-removal-evidence/native/desktop/removal-tab-cancel-120-public.png)、[63 列](2026-11-23-group-removal-evidence/native/desktop/removal-review-63-public.png)、[49 列](2026-11-23-group-removal-evidence/native/desktop/removal-review-49-public.png)、[短窗拒绝](2026-11-23-group-removal-evidence/native/desktop/removal-too-short-public.png)、[旧预览拒绝](2026-11-23-group-removal-evidence/native/desktop/removal-stale-preview-error-120-public.png)、[成员收纳](2026-11-23-group-removal-evidence/native/desktop/removal-collected-members-120-public.png)、[重开](2026-11-23-group-removal-evidence/native/desktop/removal-reopened-120-public.png)。失败框会保留后端英文错误细节，此处未宣称完整文案／视觉接受。

## 清理、归档与下一步

核对两条原 Session/Run/backend 后显式停止，2 shell＋2 UI＋1 Terminal 五个原 PID／FILETIME／路径确认结束，已知私有映像为零，自有两项任务不存在。专用 settings.json 恢复 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。剪贴板文本及格式名恢复，非文本内容等价未验证；Caps Lock 前后 true，未修改。

[证据清单](2026-11-23-group-removal-evidence/manifest.json)保留 ZIP、原／归档哈希、UTF-16/BOM 转换与已知用户名／路径脱敏，PNG 字节不改。[独立核验](2026-11-23-group-removal-evidence/verification.json)检查源指纹、原生计数、连续回执、组／成员变化、原草稿、输出记录、偏好及清理；不产生新产品执行结论。运行密钥、数据库／备份、真实历史、剪贴板内容和二进制未提交。

下一步先建立固定上游的整体差异清单，集中比较 Tab 日常流程中的布局／比例／颜色／间距、键盘和鼠标；再补会话手动排序等组织能力。长目录缩放锚定、其余 IME／焦点／鼠标矩阵、真实 agent、资源预算／长稳、无 WSL 环境及发布验收继续开放。Unicode d/y/n 注入只证明实际字符路由，不替代物理大小写或 IME 候选验收。
