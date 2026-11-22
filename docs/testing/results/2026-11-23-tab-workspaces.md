# 2026-11-23：首个 Tab 终端工作区切片

结论：Windows psmux 默认工作区已从列表／全屏 attach 接到产品 Tab，活动 Tab 有真实终端输入与光标；普通 shell 的草稿切换、关闭重开、UI 重启、连续命令、中文／多行粘贴、创建表单和限定缩放有局部原生结果。**完整上游体验、IME 组合、子程序鼠标／滚动、Tab 排序和资源验收仍未完成。** 本报告不是 M0/M1/M2 全部通过或发布兼容声明。

对应 [ADR-0016](../../adr/0016-tab-workspaces-and-input-ownership.md)、[ADR-0015](../../adr/0015-pinned-psmux-native-terminal-client.md)、FL-001/002/003/004/005/006/012/015/016/018/019。339 项 catalog 仍为 designed/not_run；以下实际结果来自独立实现测试和桌面驱动，不将其换算为整条 catalog 场景通过。

## 构建、来源与环境

- 基础提交 `6acf32e`，本轮工作树源码及测试逐文件 SHA-256 见[源码指纹](2026-11-23-tab-workspace-evidence/source-manifest.json)。最终提交可通过 Git 中本报告的引入提交定位，不能把基础提交当作已含新 UI。
- Go 1.27.4、Bubble Tea 2.1.2、Lip Gloss 2.1.1、Bubbles 2.2.2 textinput。Bubbles 新增的 runewidth 0.0.31、uniseg `v0.4.8-0.20261110175422-3a370411ca06`、clipboard 0.1.4 精确锁定；uniseg 是预发布传递依赖，未宣称全部稳定。表单只接收终端粘贴事件，应用不主动读取剪贴板。
- 固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`：圆角终端布局、35%／18% 侧栏、Tokyo Night 色板及 newdialog 居中／焦点结构选择性改编。文件头、MIT、机器清单和 REUSE_MANIFEST 已同步。尚无同尺寸上游运行画面对比或用户视觉接受。
- 实际执行：Windows 10 build 19045，普通用户桌面 session 2，便携 Windows Terminal 1.25.3131.0，Consolas 12，初始 120×32 单元格／1113×657 窗口，窄窗口 540×360。原生 PowerShell 7.7.0。日期为本地 2026-11-23；原始 UTC 记录落在 2026-11-22 16:33–16:55。
- psmux 3.4.0 EXE SHA-256 `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`；只操作本轮私有 runtime-dir／namespace／state-dir，不扫描或控制全局用户会话。
- 最终 UI 应用 SHA-256 `bf5bb1a7e5c4b68f905456ea0efb93cf17d4a7faac95dd4f1128698bd7afc09d`；UI 测试 EXE `d40a7fdda080623f054689a8f355f66d7df3167ee3348a1e39976fef5046d162`；后端 run3 测试 EXE `139b3d717bc998201d4f5bfd8aeb9e55d81f61334b05feac8f37e23b06a23ea7`。后端 run3 与最终构建的 backend/terminal 生产代码相同，之后仅修改 UI 角色色／帮助和 CLI 使用说明；各阶段二进制不可混称同一构建。

## 实现边界

一个活动原生客户端，屏幕只保留最新帧；Session/Run 与临时 View ID 分开。切换与关闭只断开客户端，不启动或停止 Run。首次发现现有终端后打开 Tabs；关闭后轮询不自动重开，列表 Enter 可回到原 Run；本轮没有保存 Tab 顺序／关闭状态或复制 shell 草稿。

终端焦点下文字、已支持特殊键、粘贴进入 FIFO，`q`／`Ctrl+C` 属于目标程序。F6 切管理焦点，Ctrl+PgUp/PgDn 切 Tab；管理模式 n 创建、x 关闭视图、h 历史、a 辅助 attach、q 退出。管理视图转换、关闭和正常退出等待已接受输入写完；连接中关闭也需先排空原目标缓冲。失败不自动重放，不从通道写入推断 agent 接受或完成。

创建表单使用名称、executable、cwd、JSON 字符串数组 args，保留含空格／中文路径和参数边界。未提供 env 编辑或 TOML 保存。启动返回 Session/Run 身份后仍报错会禁止同表单重试，防止元数据失败引发重复进程。Unix tmux 能力声明返回不支持内嵌传输，继续旧列表／attach 入口。

## TDD 与本地检查

关键 RED 与 GREEN 均在 [tdd](2026-11-23-tab-workspace-evidence/tdd/)；有些实现尝试先遇到编译／依赖问题，保留原日志，不将失败文件名中的 green 当作通过结果。

| 行为 | 失败依据 | 通过依据 |
| --- | --- | --- |
| 受验证屏幕字符属性与尺寸 | red-01 | green-01 |
| Tab／真实光标、q／Ctrl+C 输入归属 | red-02/03 | green-02/03 |
| v2 边框总尺寸，最后单元格与光标 | red-04 | green-04 |
| 稳定 Run 的 A→B→A、切换时旧输入排空／新输入只到新 Tab | red-05/06 | green-05/06 |
| 输入字节上限、管理焦点和正常退出排空 | red-07/08/09 | 相应 green 及最终全量测试 |
| Manager 只转发原目标、关闭／重开和关闭前排空 | red-10/11/12 | 相应 green 及最终全量测试 |
| 通用创建表单／字面参数 | red-13；首次 green-13 为缺失 clipboard go.sum，未通过 | green-13-boundaries、green-16-all 与最终测试 |
| Unicode／多行粘贴仅一个原生 Paste 操作 | red-14 | green-14 |
| 打开历史前排空旧输入 | red-15 | green-16-all；中间 green-15 曾失败，不算通过 |
| 仅活动可见终端轮询兜底、鼠标 Tab 命中 | red-16/17 | green-16-all／green-17-all |
| 连接尚未完成时关闭不能丢缓冲输入 | red-19 | green-19 |
| 原生右边界待换行光标 | red-20 | green-20 与最终 Windows 复验 |

创建验证／取消、1×1 至 120×32 表单边界、启动结果不确定时拒绝重复创建也包含在最终实现测试。它们是合同验证，不单独虚构一个不存在的 RED。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、生产源码 500 行检查与 catalog `check --generated` 均通过；日志见[本地检查](2026-11-23-tab-workspace-evidence/logs/)。这些检查本身不是 Windows 桌面或完整产品验收。

## Windows 实际结果

最终 UI 测试 EXE：**42 个顶层测试通过，计入子测试为 182 条 PASS，0 FAIL、0 SKIP**，见[最终原生日志](2026-11-23-tab-workspace-evidence/windows/go-ui-final.log)。不能把子测试数写成顶层测试数。

固定后端／fixture／PowerShell 实际执行的两项终端合同再次通过，0 FAIL、0 SKIP，见[后端原生日志](2026-11-23-tab-workspace-evidence/windows/native-backend-run3/tests.log)。这两项不包含桌面 IME 或真实 agent。

| 桌面流程 | 独立观察／实际结果 | 边界 |
| --- | --- | --- |
| 两个普通 shell，A→B→A | 各自中文／emoji 命令草稿恢复；显式 Enter 前两份输出文件均不存在，提交后分别为 `中文A 😀`／`中文B 😀` | 此终端键位、PowerShell；不证明 agent 草稿恢复 |
| 连续命令 | B 再执行第二条命令，独立文件为 `second command`，提示符继续可用 | 同一 shell Run，不推断子 agent 生命周期 |
| UI 退出／再开 | B 的 `draft survives UI restart` 未提交草稿仍在；后端 Session/Run 未换 | Tabs 自身的排序／关闭状态未持久化 |
| 关 Tab／列表 Enter 重开 | B 的 `close reopen draft` 恢复，同一 Run，没有再启动 shell | 关闭视图与停止 Run 分离 |
| 新建表单 | 名称“普通终端 C”、含空格程序路径、中文 cwd、JSON args 提交，第三个普通 shell 可交互 | 本轮不编辑 env 或保存配置 |
| 窄窗口 | 可见 Tab／终端保留；子 PowerShell 通过独立文件报告实际终端宽度 54 列 | 响应式辅助侧栏隐藏；完整窄屏管理可发现性仍待改进 |
| 修复后缩放／焦点 | run3 的 5 轮宽窄＋管理／终端切换，共 20 个观察状态，无断开；随后独立文件为 `still interactive` | 限定短测；不是长时间 resize／输出压力或资源预算 |
| 最终多行粘贴 | Here-string 含中文两行；粘贴后文件不存在，显式 Enter 后文件正文为 `line 1 中文\nline 2` | 不追加 Enter；其他程序可自行解释粘贴中的换行 |
| 最终弹窗／帮助／退出 | 弹窗打开取消、帮助返回、管理 q 正常退出；最终清理后会话为空 | 子程序鼠标与滚动未做桌面验收；Tab 点击仅局部 Go 命中测试 |

桌面请求与返回按编号压缩保存为 [第一轮事件](2026-11-23-tab-workspace-evidence/windows/desktop/events.json.gz)、[第二轮事件](2026-11-23-tab-workspace-evidence/windows/desktop2/events.json.gz)。原始请求含合成输入；归一化的文字快照不能用于逐格宽度判断。所有输入前核对自有 WindowsTerminal PID／FILETIME／路径及前台身份；失配则拒绝，不向用户其他窗口输入。

保留三张未编辑的最终专用窗口截图：

- [Tab 与真实终端光标](2026-11-23-tab-workspace-evidence/screens/final-clean-terminal.png)
- [按需创建弹窗](2026-11-23-tab-workspace-evidence/screens/final-empty-dialog.png)
- [当前键位帮助](2026-11-23-tab-workspace-evidence/screens/final-help.png)

截图是当前切片的实际画面，不是上游对照或用户接受。其余含本机路径的截图留在本地测试目录，未修改图像后冒充原始证据，也未入库。

## 失败、诊断与修正

1. Windows PowerShell 首次将未加引号的 `-test.v` 解析成错误参数，原生测试根本未执行。run2 起改为引号字面参数。桌面驱动首次沿用旧输出目录，随后改为新目录／任务并重新启动；不能把这些准备失败算作产品通过。
2. 首轮驱动把 Ctrl+V 当作统一终端粘贴，A/B 文件未生成。三种快捷键同窗对照显示 Ctrl+V 未输入正文，Ctrl+Shift+V 与 Shift+Insert 才粘贴；更正驱动后全流程重跑通过。应用继续接收原生 PasteMsg，不擅自吞掉 Ctrl+V 去读取剪贴板。对照任务首轮还因前台失配拒绝，旧 error 文件仍保留，不能据其存在认定后续每次都失败。
3. 初次尺寸观察用了 .NET 相对文件路径，与 PowerShell 当前 location 的路径语义不同，预期位置没有文件。改为 PowerShell Set-Content 后独立观测到 54 列；保留首次失败。
4. 首个 UI 构建在宽窄／焦点切换中曾报 `terminal frame exceeds screen bounds`。第一次探针筛选数组错误；第二次记录集合作用域错误，显示 0 帧，均不能视为未复现。修正采集器后，空闲探针保存 58 帧无异常；带长行输入的后续探针保存 **251 帧，其中 1 帧 cols=54、rows=9、cursor_col=54、cursor_row=8、rows_v2=9**，证实旧列上界检查误拒绝右边界状态。探针原意是有限输出，但第一次命令拼接到残留草稿导致语法错误，不能把该轮写成成功执行输出工作负载。原生帧仍有效，是本次修复的依据。
5. 仅对 `cursor_col == cols` 将显示光标夹到最后一格；`cursor_col > cols` 和超界行仍拒绝，新增右边界 RED→GREEN、保留身份／控制字符／行溢出等拒绝测试。捕获帧[摘要](2026-11-23-tab-workspace-evidence/windows/resize-probe-summary-run4.json)和[归一化压缩帧](2026-11-23-tab-workspace-evidence/windows/resize-probe-frames-run4.json.gz)保留尺寸与 run width；原始文件 SHA-256 见归档清单。本轮 run3 才是实际受控输出后 20 个桌面状态与后续输入的通过复验。
6. 真实 IME 尝试只看到 `NIHAO`，没有候选；相同输入在外层 PowerShell 也没有形成候选。扫描码探针沿用了已关闭的第一轮 PID，被身份检查拒绝，**没有执行该对照**。未取得组合提交／取消证据，不能标 IME 通过，也不能据此断言产品 IME 坏了。unicode 文本／剪贴板通过不替代真实 IME。

## 清理与复现

本轮三个测试 Session 均通过完整 ID 与原 Run／命令核对后显式停止；最终 `list --json` 为 `[]`，私有 Windows Terminal 进程数 0；四个本轮任务已移除。便携终端 settings 恢复为原始备份，SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。没有删除其他任务或用户终端，详情见[清理日志](2026-11-23-tab-workspace-evidence/logs/cleanup.log)。测试生成的数据／便携依赖保留在验证机隔离目录，未作为运行数据提交。

脚本位于 [scripts](2026-11-23-tab-workspace-evidence/scripts/)。它们是此次执行原件；重新执行必须使用新隔离 root／任务并重新创建会话、窗口身份和备份，不能复用归档 PID、FILETIME 或设置哈希。原始日志中的 BOM／UTF-16／CRLF、用户 profile 与仓库路径统一归一化；映射、原始 SHA-256 和归档 SHA-256 分别见 [archive-manifest](2026-11-23-tab-workspace-evidence/archive-manifest.json) 与 [checksums](2026-11-23-tab-workspace-evidence/checksums.json)。未归档鉴权 key、SQLite 运行库、完整用户环境或测试 EXE。

## 下一切片仍需完成

- 对照固定上游的同尺寸运行画面，逐项核对背景／表单行填充、层级、密度和间距；当前弹窗仍有行背景细节待修，不宣称视觉还原完成。
- 恢复 `<`／`>` 比例调整、分隔符拖动、辅助列表密度／搜索；补标准 Tab 的关闭／新建可发现性、排序和溢出导航。
- 真正进入 IME 候选，验证提交／取消、光标定位和焦点切换；验证 Tab 点击、子程序鼠标和 scrollback，不把管理鼠标命中测试扩成终端鼠标支持。
- 新 Tab 路径的 0/1/5/10 会话资源、有限输出、关闭重开和完整观察覆盖；旧列表／attach 测量不可复用为新路径预算。
- 退出／错误 Run 的屏幕保留与明确只读状态、更多按键组合、alternate screen 和真实目标 agent。用户当前无目标 agent，仍明确未验证；不增加专用协议作为基础门槛。

这些缺口仍由 active goal 与 Feature Ledger 跟踪，未标整体验收或目标完成。
