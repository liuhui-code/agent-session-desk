# Windows 10 SSH 操作验证与桌面验收阻断记录

执行日期：2026-11-22，最终核验时间 03:37:55 UTC。**仅完成 SSH 可观察的局部操作；完整 Windows Terminal 桌面验收仍为 blocked，M0 / FL-016 未验收。**

## 被测组合与证据

- Windows 10 家庭版，10.0.19045。SSH 位于 Session 0，explorer 位于 Session 2；当前用户 Appx 查询未发现 Windows Terminal，PATH 未发现 wt、PowerShell 7、Go 或 psmux。此查询不证明其他用户或其他位置均未安装。
- 用户确认目前只有 SSH。首轮未识别已有的交互计划任务通道，误判为没有桌面控制入口；后续已恢复并完成桌面校准，见文末更正。OpenSSH PTY / conhost 不等同于 Windows Terminal 桌面。
- 产品使用本项目 [v0.0.0-m0-preview.1](https://github.com/liuhui-code/agent-session-desk/releases/tag/v0.0.0-m0-preview.1) 的 Windows amd64 程序，manifest 源码为 `5c94b7d407a59be3f8fa476ea6c584aac67110bf`。这次没有重新构建源码。
- 测试 CLI 是 Windows 自带 PowerShell 5.1 的合成回显循环，通过明确的 `-EncodedCommand` 启动；不是实际 agent，也不是 `cmd/asd-fixture`。仅输入合成标记和“中文”，没有读取真实会话内容。
- 私有目录 `asd-acceptance-20261122`，namespace `asd-ssh-20261122`；所有调用保持相同 backend/runtime/namespace。两个会话已按精确 ID 停止，测试程序与文件保留用于复现；没有配置新的远程控制入口。

| 对象 | SHA-256（下载后及 Windows 解包后核验） |
| --- | --- |
| 产品 ZIP | `ca89f1923a5a1c38a46863c601121fbf45446e917e2e3c9231001d9d23628f1c` |
| 产品 EXE | `a810705c4322e9f774b61b9159f39340b7681ca4a275d2615b1f8c938fb79cca` |
| 官方 psmux 3.4.0 ZIP | `a66b6be7c13c6fa0dc3660a267b9e34eaa91a4eb7f07fbd55a0468018bd0f721` |
| psmux EXE | `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca` |

原始合成证据：[创建与列表](2026-11-22-ssh-evidence/setup.json)、[重开后快照与停止](2026-11-22-ssh-evidence/check-cleanup.json)、[PTY 字节记录](2026-11-22-ssh-evidence/pty-transcript.json)。PTY 记录含终端控制序列，应作为 JSON 数据检查，不直接回放到终端。

## 实际完成的 SSH 操作

| 局部检查 | 结果与依据 |
| --- | --- |
| WIN-UI-02 的 CLI 子步骤 | pass：创建“中文输入 A”“会话 B”，中文标题正确，两个 Session ID 与两个 Run ID 分别不同；没有以此宣称桌面执行通过 |
| 进入、输入与分离 | pass（SSH 范围）：Enter 进入 A，写入 `ASD_A_20261122`，收到回显；Ctrl+b、d 后管理器显示“已返回管理器”且继续接受管理键 |
| 切换与输入隔离 | pass（SSH 范围）：j、Enter 进入 B，写入 `ASD_B_20261122 中文`，收到回显；分离后最终 A 快照不含 B 标记，B 快照不含 A 标记 |
| 退出与重开 | pass（SSH 范围）：q 正常退出，另一个 SSH PTY 重开管理器，再 q 退出；后续列表与创建记录的 Session/Run 一致 |
| 精确停止 | pass（后端范围）：两个精确 Session ID 分别返回 `backend_session_stopped`，之后私有 namespace 列表为空；未进行完整进程树残留测量，不扩大为全机清理结论 |

输入中文是 UTF-8 字符传输，**未操作中文输入法的组合、候选、提交或取消**。未提供桌面截图/录屏，PTY 输出不能证明字体、光标或桌面焦点表现正确。

## 桌面执行表

对应 [WIN-UI 执行协议](../WINDOWS_INTERACTIVE_ACCEPTANCE.md)。以下状态针对原协议要求的完整桌面场景，局部成功保留在上表。

| 项目 | 状态 | 尚缺证据 |
| --- | --- | --- |
| WIN-UI-01 | blocked | Windows Terminal 120×32 / 60×18 布局与恢复观察 |
| WIN-UI-02 | blocked | CLI 子步骤已通过；原协议的 Windows Terminal 桌面前提未满足 |
| WIN-UI-03 | blocked | 桌面选中目标、IME 组合与提交及焦点 |
| WIN-UI-04 | blocked | IME 取消、多行 Unicode 粘贴边界 |
| WIN-UI-05 | blocked | SSH attach/detach 已执行；桌面光标和焦点未观察 |
| WIN-UI-06 | blocked | SSH 切换及输入隔离已执行；桌面画面与焦点未观察 |
| WIN-UI-07 | blocked | 桌面 resize、Ctrl+C 目标及管理器恢复 |
| WIN-UI-08 | blocked | SSH 重开身份已核验；桌面完整流程未执行 |
| WIN-UI-09 | blocked | 精确停止及列表为空已核验；原协议的桌面与无关进程不受影响观察未完成 |
| WIN-UI-10 | blocked | 备用字体、缩放和 IME 组合未测试 |

## 失败与重试

1. Windows 的 `.ps1 -File` 执行被现有执行策略拒绝。保留系统策略，以 PowerShell 命令模式执行相同测试源；没有修改执行策略或安全软件。
2. 首次远程 CMD 启动的带空格绝对路径被 shell 拆分。改为先进入私有测试目录、再以相对路径运行；这是远程调用构造错误，不据此判断产品参数运输失败。
3. 首轮快照断言误读顶层 `text`，实际产品 JSON 位于 `snapshot.text`；断言失败后 finally 已停止首轮两个测试会话。修正测试脚本后从空 namespace 重做创建、进入 A/B、回显、分离、退出、重开、断言及停止，最终证据对应第二轮身份。没有将首轮失败记为产品通过。
4. 一次重开调用未保留本地交互 stdin，无法发送 q；重新以持久 PTY 执行重开与 q，正常返回 0。该调用不是桌面验收证据。

复验使用的测试源保存在 [setup.ps1](2026-11-22-ssh-evidence/setup.ps1) 和 [check-cleanup.ps1](2026-11-22-ssh-evidence/check-cleanup.ps1)。setup 的解包针对全新目录，复验已解包目录时跳过两条 Expand-Archive；不得在已有活测试会话上重复 setup。JSON 中 PowerShell 5.1 的数组扩展属性可能表现为 `value/Count` 包装，按实际保存的结构读取。

涉及 FL-002/003/006/012/016 的局部证据；各条保持原实现状态。339 项 catalog 设计/执行基线不变。本次没有修改产品代码，未重新执行 Go 测试；文档 catalog 校验仅用于文档一致性。

## 后续更正：SSH 可桥接交互桌面

用户指出以前曾通过 SSH 进行桌面点击后，重新核查发现旧项目的 InteractiveToken 桌面任务仍在。之前将“SSH 在 Session 0、桌面在 Session 2”直接推导为无法操作桌面，证据不足。

2026-11-22 04:26:24–04:26:26 UTC，使用独立普通用户交互任务在 Session 2 创建自有测试窗口，真实鼠标点击、SendInput 输入 `asd-ui-20261122`，前台 PID 与测试进程 PID 同为 13232。控件读回同文，截图已实际查看；校准成功后窗口关闭，本次专用任务已删除。未启动旧 Zed 测试，未修改系统执行策略。

证据：[桌面校准结果](2026-11-22-ssh-evidence/desktop-probe-v2.json)与[自有测试窗口截图](2026-11-22-ssh-evidence/desktop-probe-v2.png)。原结果中的绝对截图路径在归档 JSON 中改为同目录文件名。

**已证明当前机器的桌面点击、文字输入和截图通道可用；尚未在本产品上执行桌面验收，也未验证窗口缩放或 IME。** 上表 blocked 不再表示“SSH 无法访问桌面”，而表示对应产品桌面场景仍未执行；必须继续准备 Windows Terminal 等前提并实际运行，不能将校准窗口成功计入产品通过数。

后续已实际执行产品桌面流程，见[独立桌面执行报告](2026-11-22-windows-desktop-acceptance.md)。本页前述 blocked 表保留为首轮历史，不代表后续执行未发生。
