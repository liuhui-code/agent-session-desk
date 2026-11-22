# 2026-11-22：原生终端传输技术切片

本轮落实 [ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md) 的终端技术验证，新增 [ADR-0015](../../adr/0015-pinned-psmux-native-terminal-client.md)。**已实现后端持久终端适配，并通过两项限定 Windows 原生测试；尚未连接 TUI Tab。** 上游视觉还原、Tab 导航、创建表单、焦点、键鼠／IME 和新路径资源测量仍未完成，目标保持开放。

## 代码与上游核对

- [internal/terminal](../../../internal/terminal/) 接入原生屏幕快照，保留尺寸、光标、样式段和 alternate-screen／mouse 声明；持续排空读取但只保留最新一帧。支持指定 pane 的字面 UTF-8、受限特殊键、无隐含回车的粘贴、resize 和仅关闭客户端。
- [Backend.OpenTerminal](../../../internal/backend/terminal.go) 复用既有版本、Session/Run、原 pane 与活动 pane 校验，读取私有有界端点文件并连接 loopback。首帧及后续帧身份不匹配即关闭；不启动 agent、shell、server 或额外 ConPTY，不自动重放不确定输入。
- 当前 TUI 和 CLI 使用入口未改变，新适配器已由原生测试调用，尚未成为默认 Tab 页面。现有全屏 attach 和被动预览行为保留；Unix Tab 适配未实现。
- 新增[上游还原源码基线](../../design/UPSTREAM_UI_BASELINE.md)，核对固定 Agent Deck 的 35/65 双栏、18% 紧凑侧栏、50/80 列断点、面板／终端框和 Tokyo Night 颜色。仅完成来源核对，没有把静态文档记为视觉实现。

固定 psmux 来源为 `20aa1d7254ee93f64dafa059ef89341d9767470a`，Agent Deck 来源为 `80494e527393b7ead00da30018e6ee7e60e9d7a0`；源码文件哈希、本轮父提交 `47ad08a`、Go 1.27.4 darwin/arm64 工具链及所有被测 Go 文件指纹保存在[manifest](2026-11-22-terminal-transport-evidence/source-manifest.json)。本次未新增 Go 依赖，也未复制 psmux Rust 实现；读取其固定内部协议以编写独立 Go 适配器。

## 真实 Windows 环境与结果

通过已配置 SSH 在 Windows 10 build **19045** 执行真实进程；未执行桌面驱动。psmux EXE SHA-256 固定为 `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`，报告版本 3.4.0。普通 shell 为便携 PowerShell 7.7.0，实际 FileVersion `7.7.0.500`、EXE SHA-256 `f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336`；fixture SHA-256 为 `f9a30f0dae7bd6a9fffe7deedf8a194405d67d651aad0b99cc20296e8cb5d9d6`。[环境收据](2026-11-22-terminal-transport-evidence/windows-environment.json)另记录探测工具和测试 EXE。

最终 Windows 测试 EXE SHA-256 为 `2561d5c72ba545c9b98a17763a1a80553b4455dd38bc4dfce53c030015d281b8`。run5 于 **2026-11-22 15:57:12 UTC** 结束，退出码 0，两项顶层测试通过、无失败／跳过：[日志](2026-11-22-terminal-transport-evidence/windows-run5/tests.log)、[收据](2026-11-22-terminal-transport-evidence/windows-run5/result.json)。

| 实际执行 | 观察与适用边界 |
| --- | --- |
| `TestWindowsPsmuxNativeTerminalScreenInputResizeAndReconnect` | 真实 fixture 原始屏幕 80×20，粘贴中文／emoji／组合字符／分号／反斜杠不隐含 Enter，显式 Enter 后独立 JSONL 精确收到；resize 至 92×24；关闭连接、新 Backend 发现同一 Session/Run、重新连接再写入唯一标记。对应原始 fixture 记录保存在 run5/native 内 |
| `TestWindowsPsmuxNativeTerminalOrdinaryShellKeepsPromptAndDraft` | 普通 PowerShell 无 agent 协议；连续两次执行命令，以独立文件验证结果和中文／emoji。第三条命令未提交时只出现在终端屏幕、不产生文件；断开并核对原 Run 后重连，单次 Enter 执行原草稿，文件内容精确匹配 |

原生测试使用现有 native harness：临时 runtime、独立随机 namespace、精确 Session 清理并确认后台登记消失；没有按进程名批量停止程序。此处的清理证据不等于完整进程树生命周期测量。普通 shell 文件值由运行断言核对，临时目录随后由测试框架清理；fixture 独立 JSONL 已归档。

手工技术探测先于产品适配：第二轮脚本直接验证固定发行物的持久屏幕、十六进制输入、独立命令执行文件、resize、断开后原 Run 与原始清理 JSON `[]`，保存 [13 帧原生数据](2026-11-22-terminal-transport-evidence/wire-frames.json)及[结果](2026-11-22-terminal-transport-evidence/wire-probe2.log)。该脚本不是产品实现，也不替代上述产品代码测试。

## 失败与 TDD

逐行为记录在 [tdd/](2026-11-22-terminal-transport-evidence/tdd/)：

1. 原生屏幕读取、字面 UTF-8、慢视图合并、resize、粘贴／显式键分别 RED→GREEN。慢视图用 100 帧证明生产者不被 View 阻塞、产品没有积压全部历史。
2. Backend.OpenTerminal 的 RED 是 Windows 测试交叉编译时缺少接口；它只是构建失败证据，未计为 Windows 运行。首次实际 Windows 产品测试 run1 在读取屏幕时失败 `invalid terminal cell`，[原始失败](2026-11-22-terminal-transport-evidence/windows-run1/tests.log)保留。
3. 实际样式段 width 可覆盖整行（80、92 等），表示整段宽度；最初实现错误地只接受 1／2。补充整行样式段 RED，修复为逐行累计宽度校验后 GREEN。没有放开无限宽度、控制字符、错误 Session/pane 或越界光标。
4. run2 对修复后的首项原生测试通过；run3 增加普通 shell 场景，两项通过；run4 加入连接前版本探测、鉴权 key 及时释放和键名长度边界；run5 为补充无尺寸／输入副作用的 Refresh 接口后的最终源码复验。Refresh 另有局部协议 RED→GREEN，UI 需按可见性限制刷新节奏。所有轮次的日志／构建哈希均保留。
5. 手工探测第一轮的剩余计数字段为 1，未作为清理通过证据；第二轮改为原始 JSON `[]` 断言。归档首次因 PowerShell UTF-16 日志读取失败，随后按 BOM 解码并核对 JSON；未据该采集问题修改产品结果。

另有身份变化、split、危险单元格控制字符、行宽和光标越界拒绝的局部回归。本地共有八项 terminal 顶层测试（含六项拒绝子测试）；这些合成协议测试不证明外部 Windows 行为，原生测试单独记账。

## 检查与完整性

以下本地检查均通过，原始日志存于同一证据目录：

- `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`。
- `python3 -B tools/test_catalog.py render` 后 `check --generated`：339 个 designed/not_run 设计用例，20 FL、14 FR；新增 ADR 不改变产品执行数 0。
- `python3 -B tools/ci/check_source_limits.py`：生产源码物理总行数均不超过 500；`git diff --check`、来源指纹与归档校验另见核对日志。

Windows 测试 EXE 由当前源代码交叉构建，再上传 Windows 实际执行；交叉编译本身不计原生运行。源码、TDD、测试 JSONL 和构建哈希随本次实现提交保存，可用 `git log -- internal/backend/terminal.go` 定位准确提交。

归档只包含合成终端数据，不包含 auth key、端口文件或二进制。UTF-16／BOM／换行按文本规范化，用户目录与仓库绝对路径替换为占位符；manifest 保留规范化前文件哈希，[SHA256SUMS](2026-11-22-terminal-transport-evidence/SHA256SUMS)核对归档字节。

## 尚未证明

未实现／执行新的 Tab 页面、上游视觉评审、TUI 新建与配置表单、管理／终端焦点、鼠标和系统 IME、A→B→A Tab 切换、窗口 DPI、alternate-screen 桌面交互及新资源基线。只有 shell 与 fixture，没有真实目标 agent；未证明所有 CLI／shell、无 WSL、Unix Tab 或跨 psmux 版本协议兼容。内部协议仍有外部 server/pane 替换竞争边界，未知输入结果不得重放。

FL-018 仅因已有终端后端代码转为实现中，其他相关能力保持实现中、未发布；M0/M1/M2 整体验收和用户提出的完整界面目标均未完成。下一切片接入 TUI Tab、真实光标和有序输入，再做原生桌面与资源验证。
