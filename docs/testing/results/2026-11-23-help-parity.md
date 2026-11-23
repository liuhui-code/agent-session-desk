# 整体差异审计与固定上游帮助还原

日期：2026-11-23；基于 `00576dd` 加本轮修改。新增[22 项整体差异清单](../../design/UI_PARITY_AUDIT.md)，按来源行为、实际产品、局部证据和剩余缺口逐项跟进；本轮接入 U13 帮助的键盘／视觉还原，完整键鼠／粘贴事件矩阵与其他 IME 仍开放，未把整体验收缩减为本组件。ADR-0014/0016、Feature Ledger 和 TC-UI-001 同步，依赖、数据库和 ui.toml 格式不变。

## 来源与实现

固定干净 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT。[16 份来源指纹](2026-11-23-help-parity-evidence/local/upstream-source.json)覆盖帮助、宽度适配、样式及清单所指主要界面文件。选择性迁入 help.go 的短／全列表、分区键列、导航／关闭和首尾页布局，以及 cellwidth.go 的对话框宽度适配。styles.go 仅组合 tokens；help_entries.go 的说明只列本产品已支持操作。

原管理 ? 是不可滚动的整页说明，窄／低窗口会裁掉命令。现在默认打开居中“五键入门”，Tab 切全部并重置滚动；完整页支持 j/k、↑↓、PgUp/PgDn、Ctrl+U/D、Home/End、g/G。其他键只关闭帮助，包括 q／Ctrl+C／n／F6／Ctrl+PgDn／Enter，不执行原动作；粘贴与鼠标由帮助消费。打开前沿既有 FIFO 排空并断开客户端，关闭时重新核对并连接原活动终端。后台元数据／事件仍处理，无额外终端历史、进程或偏好保存。

紫框／键列、accent 标题、cyan 分区、dim 斜体页脚、yellow 滚动提示沿固定来源。v1 Width 转为 v2 外宽；中文按单元格换行并计为实际行，短窗口严格裁切。F6、产品 Tab 和通用 CLI 文案替代来源布局／业务专用键，不显示未实现的 agent／账户／远程等入口。

隔离辅助测试实际执行未修改的来源 HelpOverlay，在 120×32、63×25、49×25 分别渲染短页、完整首／尾页，共九对来源／产品组件。Darwin sandbox 禁止个人目录及出站网络，overlay 隔离来源原测试／TestMain；原始输入、overlay 和路径映射归档，未运行上游完整应用或个人会话。来源实际观察确认 Tab 选择全部、End 到达末页、q 关闭。

[最终几何](2026-11-23-help-parity-evidence/local/help-geometry.json)中八对框的 x/y/宽/高一致：120 列短／完整外宽 58／82，63 列均 55，49 列均 41。49 列短帮助因中文换行少一行，产品 15／来源 16 行，垂直位置随居中差一行；如实保留，不填空行伪装同文案。首轮产品完整末页多一行空白，经来源实际执行发现后已修正为省略未使用下方提示行。几何一致不等于完整文字／像素或 Windows 上游接受。

## RED → GREEN 与视觉发现

- 首条行为测试在旧整页帮助失败；接入短／全卡片后通过。第一次接入误放到 Close 方法，导致编译失败；日志保留，改入 Update 后复验，不将其算成产品运行失败。
- 完整帮助最初无法滚到尾部，行为测试失败；接入来源导航、同一实际换行预算和边界夹取后通过。末页多空白行的独立回归先失败，按来源修正后通过。
- 增补帮助打开前已接受输入排空；关闭键只消费一次，粘贴／鼠标不新建、不换目标、不发输入；两页在 1–120 列、1–32 行的选定边界保持尺寸与无光标。增补不全部声称独立经历 RED。
- 第一轮桌面交互通过后，逐张看原 PNG 发现文字后的空白透出外层背景，形成条纹。逐单元格 SGR 测试确认行 10／列 42 为 default，预期 surface `(36,40,59)`，RED 保留。修复键列前缀／间距与每行尾空白的显式背景填充，复用既有 token 样式；按实际展开后的行计高。三个尺寸、短／完整首尾九个状态的每个内部单元格均复验通过。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 全通过，含缓存结果；源码行数与 catalog 检查通过。最终[201 份 Go／模块指纹](2026-11-23-help-parity-evidence/local/source-fingerprints.json)延续此前范围，包含两份历史证据里的参考 Go main，不将其当作产品运行模块。TC-UI-001 增加设计内容；339 项仍 designed/not_run、216 ready／123 needs_decision，不以本轮结果提升整条设计用例。

来源执行的首次自动权限审核超时，CreateProcess 未启动；按返回指示仅重试一次后成功。[记录](2026-11-23-help-parity-evidence/local/approval-retry.json)区分审核超时与产品／来源测试失败，没有绕过 Darwin 隔离限制。

## Windows 原生与实际桌面

Windows 10 build 19045 的 UI 原生测试分三轮记录，不累加计数：首轮 162／346、末页修正后 163／347、最终背景修正后 **164 顶层／348 含子测试**；每轮 FAIL／SKIP 均 0，精确测试映像无残留。最终[原生结果](2026-11-23-help-parity-evidence/native/native-tests/outcome.json)与先前两轮分别保留；服务 fixture 证明模型合同，未重跑完整后端原生套件或 Store／Manager 套件，不冒充真实故障注入。

最终桌面产品 SHA-256 `97460bd184430f3e9c0064c68fd951598dea4834228f5b4eb4f77e1fd3e51132`；[构建清单](2026-11-23-help-parity-evidence/local/binary-hashes.json)、build-info 与精确源指纹保留。它是 00576dd 加最终未提交源码构建，不称为提交后重建。沿 ADR-0019 显式开发 bundle，Windows Terminal 1.25.3131.0、Consolas 12、PowerShell 7.7.0、普通用户桌面 Session 2；不提升官方默认依赖或完整 Windows 兼容结论。

背景修复前与修复后均分别执行完整双 shell 流程，命名空间、产品映像和根目录独立，不重放旧动作或复用已停止进程。每轮：

- A 在终端输入未提交命令并追加字面 `?nq`，B 保留独立草稿；切回管理焦点打开帮助。
- 短／全切换，首／尾页，j/k、方向、分页、Ctrl+U/D、g/G 均经真实 UI 操作与显示核对。120／63／49 列和低高度分别截图；末页无“下方还有”，继续向下不改变页面。
- 帮助中的粘贴不进入 CLI；n、Ctrl+PgDn、Enter、Ctrl+C、q 只关闭帮助。关闭后明确 F6，实际画面仍为 A 草稿；发现列表仍为原两条，没有额外启动。鼠标消费另由模型测试覆盖，不冒充本轮桌面点击。
- UI 退出后原终端和草稿保留；重开再次从短帮助进入，偏好文件前后字节及哈希相同。13 次原生 capture 各完整保留编号 1–40，各一次；11 个未提交观察的最后非空行与原草稿逐字一致。最后明确 Enter 后，两个唯一输出各出现一次，PID 对应各自原 shell。

修复前 118 条动作（32 keys、50 state、17 text、10 capture、4 paste、4 resize、1 finish）全部成功；修复后 119 条，仅 state 轮询变为 51，其余种类相同。最终两条输出分别为 `HELP_A:12716`／`HELP_B:6640`，对应新隔离轮次的原 PID；见[最终流程](2026-11-23-help-parity-evidence/native/flow-result.json)及[独立核验](2026-11-23-help-parity-evidence/verification.json)。未把修复前通过当成最终背景通过，两轮完整回执和修复前原图均归档。

修复后十张未经编辑 PNG 逐张检查：[120 短页](2026-11-23-help-parity-evidence/native/desktop/help-short-120-public.png)、[120 首页](2026-11-23-help-parity-evidence/native/desktop/help-full-first-120-public.png)、[120 尾页](2026-11-23-help-parity-evidence/native/desktop/help-full-last-120-public.png)、[63 首页](2026-11-23-help-parity-evidence/native/desktop/help-full-first-63-public.png)、[63 尾页](2026-11-23-help-parity-evidence/native/desktop/help-full-last-63-public.png)、[49 首页](2026-11-23-help-parity-evidence/native/desktop/help-full-first-49-public.png)、[49 尾页](2026-11-23-help-parity-evidence/native/desktop/help-full-last-49-public.png)、[49 短页](2026-11-23-help-parity-evidence/native/desktop/help-short-49-public.png)、[低高度](2026-11-23-help-parity-evidence/native/desktop/help-full-short-height-public.png)、[重开](2026-11-23-help-parity-evidence/native/desktop/help-reopened-short-120-public.png)。

## 清理与剩余范围

两轮均在核对原 Session/Run/backend 后显式停止，各 2 shell＋2 UI＋1 Terminal 五个原 PID／FILETIME／路径确认结束，已知私有映像为 0，自有任务不存在；不是全机器清理声明。专用 settings.json 恢复 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。剪贴板文本与格式名恢复，非文本内容等价未验证；Caps Lock 前后 true。原 namespace／设置／剪贴板不混算为新产品能力。

[清单](2026-11-23-help-parity-evidence/manifest.json)保存两份导出 ZIP 哈希及各文件原／归档哈希，已知用户名／路径脱敏，UTF-16/BOM 转 UTF-8/LF；PNG 字节不改。运行密钥、数据库／备份、真实用户历史、剪贴板内容和二进制不提交。核验检查最终精确源、三轮原生计数、连续回执、草稿／输入归属、偏好、九组几何、修复前背景缺口／最终内部背景及身份清理，不产生额外产品执行或整体验收。

下一项为 U14 页脚键／说明层级及 U01–U06 主界面同尺寸复查；同类表单背景也应逐单元格核对。会话行排序、停止 UI、比例保存、鼠标选择／可见页复制、长字段锚定、真实 agent、其他 IME／宿主矩阵和资源／发布门槛仍按清单开放。普通 shell／Unicode 注入不代替真实 agent 或物理键／IME 接受。
