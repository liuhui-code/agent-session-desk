# 组重命名、稳定归属与原终端输入

日期：2026-11-23；基于 `5f2e310` 加本轮修改，179 份 Go／模块源指纹见[清单](2026-11-23-group-rename-evidence/local/source-fingerprints.json)。管理焦点选中组标题后 r 已接入持久重命名；默认组、根组和子组均可改显示名。**本轮是组编辑切片，删除／排序、完整上游体验和发布仍未完成。** 决定补充于 [ADR-0023](../../adr/0023-persistent-groups-and-membership.md)，没有新 schema 或依赖升级。

## 来源与实现范围

来源为干净的固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT。`internal/ui/group_dialog.go` SHA-256 为 `0951f9efb6ed7d2cb6cf940fc476bb457f934cfb95bcc45f80c5e9cd6a003676`；ShowRename／GroupDialogRename 的仅名称、预填、CursorEnd 和重置隐藏目录焦点，以及 home.go 的组标题 r 是本轮依据。其余源码／Git blob／许可指纹见[来源记录](2026-11-23-group-rename-evidence/local/upstream-source.json)。

产品组名沿既有清理与长度规则，完整 Group 快照及版本作为 SQLite immediate 事务内的预期值：同级重名和过期／ABA 编辑拒绝，同值不增加版本，版本耗尽拒绝变化。仅写名称与组版本；Group ID、父组、顺序、目录、子组和成员版本不改。展示路径按稳定祖先 ID 动态计算，避免沿上游路径身份改写全部子组／成员／折叠键。默认组保留 default 身份与位置。SQLite 仍 v4，ui.toml 仍 v3。

弹窗沿原组操作路径等待已接受的终端输入结束，再断开客户端。Tab 不切入不可见目录，Enter 保存、Esc 取消；提交中不可丢弃或重复提交，错误保留且不自动重试。结果绑定原弹窗，失效发现不得回退已提交元数据；回执仅合并不低于已观测版本的组数据。管理改名不打开其他 Tab、不重启 Run、不注入命令。

组／会话重命名共用六行内容和既有 tokens／样式，没有新增独立设计数值。中文文案、稳定 ID、v2 真实光标和极窄窗口裁切属于明确适配。

## RED → GREEN 与本机回归

- 首条真实 SQLite 测试因缺少 RenameGroup 接口编译失败，接入后验证改名跨关闭／重开保留，子组、成员和原 Run 相同。
- Manager 无终端对象的调用先因缺少接口失败，接入后通过；该操作不依赖后端可执行文件或存活发现。
- 实际 Model 组标题 r 原来没有弹窗，新增行为失败；接入后验证仅名称、末尾光标、折叠、原 Tab／草稿和旧发现拒绝。
- 较迟保存回执会覆盖已观测到的更新组名，行为测试失败；按组版本合并后通过。失败日志与复验均保留。
- 另覆盖真实双连接并发仅一个成功、ABA、同级名称碰撞、作用域、伪造快照、非法输入／版本耗尽、取消、输入排空、忙时保护、失败不重试、旧弹窗结果和 120×32 至 1×1／0×0 边界。这些增补不全部声称经历独立 RED。
- 首轮新增取消断言误把合法 `size:73,25` 当作输入，失败日志保留；改为分开核对 resize 和输入，未修改产品以满足错误断言。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 均通过，含缓存结果；source limits 与 catalog 检查通过。用例 TC-SESSION-008 补充组改名设计，仍保持完整 339 项 designed/not_run、216 ready／123 needs_decision。catalog 工具只验证／生成文档。

## Windows 原生执行

Windows 10 build 19045，实际执行交叉构建后的三个测试程序；计数不累加重复轮次。

| 套件 | 顶层／含子测试 PASS | FAIL／SKIP |
| --- | --- | --- |
| Store | 51／77 | 0／0 |
| Manager | 14／16 | 0／0 |
| UI | 136／310 | 0／0 |

三项原生执行的日志、退出码和无残留测试映像见[结果](2026-11-23-group-rename-evidence/native/native-tests/outcome.json)。Store 使用真实 SQLite／文件，UI 服务 fixture 仅证明模型合同，不作为真实后端交互证据。完整后端原生场景并未在本轮重新执行。

产品 SHA-256 为 `af784f84610f61c46934f479f0dc641264a09f7d9dba44ccb10a2466293a94a6`；三个测试程序及产品哈希见[构建清单](2026-11-23-group-rename-evidence/local/binary-hashes.json)。产品为基于 5f2e310 的未提交源码构建，源码与本轮最终 Go 文件一致，不冒充提交后重新构建。使用 ADR-0019 的显式开发 bundle，官方默认依赖与完整发布门槛不因此改变。

## 实际来源帧与桌面

未修改的固定上游组件在禁止个人目录与出站网络的 Darwin 沙箱中实际渲染 ShowRename。辅助 TestMain 只执行测试，没有运行原 TestMain 或完整上游会话程序。产品调用实际组弹窗渲染，三对外框矩形完全一致：

| 窗口 | 来源／产品外宽 × 高 | 来源／产品左上角 |
| --- | --- | --- |
| 120×32 | 46×10 | 37,11 |
| 63×25 | 46×10 | 8,7 |
| 49×25 | 41×10 | 4,7 |

[几何结果](2026-11-23-group-rename-evidence/local/rename-geometry.json)与原始 ANSI 帧保留。矩形一致不代表所有像素相同，也不等于完整上游 Windows 应用或用户视觉接受。

实际产品运行于普通用户桌面 Session 2，Windows Terminal 1.25.3131.0、PowerShell 7.7.0。109 条连续回执全部成功：33 keys、41 state、17 Unicode text、9 capture、5 paste、3 resize、1 finish。三次窗口调整包括 Windows 最小宽度退化至 49 列；PNG 未编辑。

- 两条普通 shell 先各保存未提交命令，再创建根／子组，将 A 移入子组。折叠根组后改名，Tab 后追加文本仍编辑名称；原 Group ID、成员版本、Run 元数据与折叠文件哈希不变。
- 分别改子组和默认组；非法多行粘贴拒绝，取消不保存。UI 退出／重开后名称、折叠及两条原 Session/Run/Backend 保留，编辑后取消仍不改元数据。
- 10 次独立后端 capture 均保留编号 1–40，各一次。8 次未提交阶段的最后非空行与完整草稿逐字相同；最终各明确 Enter 后只产生一次 `RENAME_A:14184`／`RENAME_B:17344`，对应两条原 shell PID。进程 FILETIME／完整路径前后相同。
- 7 次只读元数据快照表明，改名和重开／取消期间原成员版本与 Run 记录不变；组名版本分别增加一次。桌面流程未施加双 UI 并发冲突，该合同由实际双 SQLite 连接测试证明。

可直接查看[120 列预填](2026-11-23-group-rename-evidence/native/desktop/rename-root-prefilled-120-public.png)、[63 列编辑](2026-11-23-group-rename-evidence/native/desktop/rename-root-edited-63-public.png)、[49 列编辑](2026-11-23-group-rename-evidence/native/desktop/rename-root-edited-49-public.png)、[子组改名及原草稿](2026-11-23-group-rename-evidence/native/desktop/renamed-child-tree-120-public.png)、[改名折叠后重开](2026-11-23-group-rename-evidence/native/desktop/renamed-folded-reopen-120-public.png)。全部九张 PNG 已逐张检查。

## 清理、归档与剩余范围

原两条 Session/Run 经完整身份核对后显式停止，2 shell＋2 UI＋1 Terminal 共五个登记的原 PID／FILETIME／路径均结束，已知私有映像为 0，两项自有任务不存在。专用 settings.json 恢复原 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`；剪贴板文本及格式名恢复，非文本内容等价未验证，Caps Lock 前后均为 true。

清理夹具失败单列：首次沿用三会话数量断言，在停止前退出；只读诊断又把 ConvertFrom-Json 外包一层数组，误报非标量目标。最终清理独立按两条标量完整身份重验，正常完成；没有重放桌面输入或重复执行已停止会话。归档器曾连同 verifier 的隐私匹配字面量一起脱敏，导致归档版误拒绝合法占位符；拆开匹配片段后直接运行归档版复验通过。原脚本与[故障说明](2026-11-23-group-rename-evidence/local/harness-failures.json)保留，不算产品失败。

[证据清单](2026-11-23-group-rename-evidence/manifest.json)记录原始 ZIP、每项原／归档哈希、UTF-16／BOM 解码、已知测试用户名／路径脱敏；PNG 按原字节保存。数据库、运行密钥、剪贴板原文、设置备份、历史文件和程序未提交。[独立核验](2026-11-23-group-rename-evidence/verification.json)验证源指纹、原生计数、连续回执、精确草稿／PID、编号历史、组元数据、几何及清理，不产生额外产品执行或完整验收结论。

继续推进组删除／收纳、同级排序和来源差异清单。长目录缩放锚定、其他焦点／IME 预编辑、普通 CLI 全屏／鼠标矩阵、真实 agent、最终资源预算／长稳、WSL 不存在环境和发布接受仍开放。本轮 Unicode 文本事件不证明物理大小写组合或 IME 候选行为；两条 PowerShell 不替代真实 agent 验收。
