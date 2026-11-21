# M0 原生 Windows 后端：执行、修复与复跑

日期：2026-11-21（Asia/Shanghai）。结果：**五项原生 Windows 进程／后端场景及六个边界子测试全部通过，native 步骤没有跳过；完整 Windows 桌面交互、资源与发布验收未完成。** 本报告记录固定官方 psmux 发行物在 GitHub Windows runner 上的实际执行，包含两次失败与修复后的成功轮次。

这批是独立实现测试，不是 339 项设计用例的整套执行。原 [M0 源码迁入报告](2026-11-21-m0-source-migration.md)保留当时 Windows `not_run` 的历史结果。

## 1. 被测提交与运行记录

| 轮次 | 被测提交 | GitHub Actions | Windows 原生后端结果 |
| --- | --- | --- | --- |
| 1 | `9d5061b19b50cc4cbbc90bd40926af347eb09239` | [44859255945](https://github.com/liuhui-code/agent-session-desk/actions/runs/44859255945) | 5 fail、0 skip；均停在版本探测，尚未创建测试会话 |
| 2 | `9bbacc42bf75b1abac42c763493d10df0892d4fc` | [44859681759](https://github.com/liuhui-code/agent-session-desk/actions/runs/44859681759) | 4 pass、1 fail、0 skip；6 个启动前拒绝子测试 pass |
| 3 | `b6f2cd2adf24a77d52244aeed1afd1c2ab5dad87` | [44860102659](https://github.com/liuhui-code/agent-session-desk/actions/runs/44860102659) | **5 pass、0 fail、0 skip；6 个边界子测试 pass** |

第 1 轮引入原生测试、独立 fixture 日志、固定发行物与 CI，以及原 pane 的精确定位修复。第 2 轮修正版本 banner 解析。第 3 轮只修正“原 pane 消失”的故障构造并加强存活检查，没有放宽 Preview/Attach 的拒绝断言。

三轮 Linux job 均通过；Windows 的普通 Go 测试、race、vet、模块校验、目录校验与原生构建也均通过。最终成功轮次的 Windows job 为 `136711294940`，Linux job 为 `136711294786`，native 命令实际 exit 0。各项结论以对应 [机器证据](2026-11-21-windows-evidence/)为准，失败不因后续复跑而删除。

## 2. 实际环境与依赖

| 项目 | 记录 |
| --- | --- |
| 系统 | Microsoft Windows Server 2025 Datacenter，`10.0.26100`，BuildNumber `26100`，X64 |
| GitHub 镜像 | `win25-vs2026`，ImageVersion `20261116.315.1`；`windows-2025` 标签本身不是不可变版本 |
| 执行 shell | PowerShell `7.6.10`，原生 Windows Go/psmux 进程 |
| Go / 构建 | Go `1.27.4 windows/amd64`；普通检查及 race 使用 CGO=1，产品与 fixture 构建使用 CGO=0 |
| UI 模块 | Bubble Tea `2.1.2`、Lip Gloss `2.1.1`、x/ansi `0.11.9`；仍未引入 Bubbles |
| psmux | 官方 `v3.4.0` Windows x64 包，对应稳定源码 `20aa1d7254ee93f64dafa059ef89341d9767470a` |
| 实际版本输出 | `tmux 3.4.0` 以及 `psmux 3.4.0 (20aa1d7 2026-10-13)`，原始输出随第二轮及后续证据保留 |
| Agent / 数据 | 与被测提交一起构建的 `asd-fixture.exe`；只使用测试生成的路径、文本与指定环境值，没有真实 agent 凭据或私人会话 |
| 终端 / 字体 / IME / DPI | 未执行桌面交互；这些字段不适用于本轮后端断言，不作为界面验收证据 |
| 资源环境 | 未采集 CPU 型号、总内存、全进程树或资源预算；不据此作性能结论 |

来源、下载 URL、归档大小及 Action commit 固定在 [windows-dependencies.json](../../../tools/ci/windows-dependencies.json)。本轮核对的 SHA-256：

- ZIP：`a66b6be7c13c6fa0dc3660a267b9e34eaa91a4eb7f07fbd55a0468018bd0f721`。
- `psmux.exe`：`bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`。

安装脚本验证 ZIP 大小、ZIP 哈希和 exe 哈希，native harness 再核对 exe 哈希。产品 `doctor` 自身仍只检查版本，不承担发行物来源认证。镜像包含 WSL；测试走原生 Windows 路径，但不能满足 `ENV-WIN-BASE` 的“没有安装 WSL”前提，也不等于 Windows 10/11 支持声明。

## 3. 原生场景与独立观察

真实后端测试位于 [integration_windows_test.go](../../../internal/backend/integration_windows_test.go)，fixture 位于 [cmd/asd-fixture](../../../cmd/asd-fixture/)。

| 原生测试 | 最终第三轮结果 | 实际验证的范围 |
| --- | --- | --- |
| `TestWindowsPsmuxNativeLaunchLifecycle` | pass | 含空格／中文 exe 与 cwd、空参数、引号、反斜杠、Unicode 和目标环境覆盖；重新发现同一 Session/Run；被动 preview 不改变尺寸；输入、正常退出与只读保留画面 |
| `TestWindowsPsmuxNativeIsolationAndRejectedLaunches` | pass | 同一后端中的两会话和另一私有 runtime/namespace 组合；6 项非法启动拒绝；停止精确对象后其余对象仍可观测 |
| `TestWindowsPsmuxNativeRapidExitRetainsScreen` | pass | fixture 快速非零退出，终端退出状态及保留画面；退出码来自 fixture 独立记录，不来自 psmux 占位字段 |
| `TestWindowsPsmuxNativeMissingOriginalPaneNeverFallsBack` | pass | 原 pane 非活动时仍读取原内容、拒绝进入其他活动 pane；随后确认原 pane 缺席且另一 pane 仍能接收输入，再验证 Preview/Attach 均拒绝替代目标 |
| `TestWindowsPsmuxNativeCLIProcessExitAndReconnect` | pass | `asd start` 进程退出后，fixture 精确 Windows handle 仍存活；新 `asd list/preview` 找回同一 Run；显式 stop 后观察该 fixture 退出 |

fixture 把 `started/input/exiting` 写入独立的新 JSONL 文件，拒绝覆盖旧文件；argv/cwd/env 的依据来自这个通道，不从被终端折行或重绘的屏幕反推。已知测试文本与环境值是合成数据。每轮的 namespace 和 Session/Run 随机 ID 记录在测试事件与 fixture 日志中；随机值仅用于隔离，不用作结果判定种子。

本轮没有执行交互式 attach 客户端。测试中的 `Attach()` 只准备和验证目标；测试用 `send-keys` 是 fixture 驱动，不表示产品已提供 M3 的跨会话 Send API、输入仲裁或 agent 接收／完成回执。

### 普通测试与 native 执行门槛

三轮 Windows 普通 `go test -json ./...` 每轮为 **43 个顶层 pass、140 个子测试 pass、5 个顶层 skip**。五个 skip 都是显式 opt-in 的 native 场景，此阶段尚未设置真实后端环境。随后专用 native 步骤设置固定 exe、fixture 和产品路径，单独运行五项场景；此步骤不允许任何 skip，且要求至少五个顶层 pass。第三轮实际为 **5 个顶层与 6 个子测试 pass、0 fail、0 skip**。

因此第 1、2 轮都被正确判定为失败，第 3 轮通过；没有用普通单元测试的成功覆盖真实后端的失败。6 个边界子测试分别为无效可执行文件、无效目录、短选项改写、独立分号、转义分号及环境值裁剪。目录工具仅验证 339 项资料结构，不能计入产品测试通过数。

`-race` 只运行普通 Go 套件；真实 psmux 的独立 native 步骤没有使用 `-race`。两类证据不能合并描述为“真实后端已通过 race 检查”。

## 4. 发现、原因与修复

### 4.1 官方版本 banner 被拒绝

第一轮的五项 native 场景均返回 `psmux version "unknown" is outside this M0 contract`。产品解析器只接受两个字段，未容纳发行版附带的构建来源。稳定源码 [cli.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/cli.rs) 的 `print_version` 调用 [types.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/types.rs) 的 `build_version_string`；第二轮原始输出确认实际 banner 带括号中的 commit 和日期。

修复提交 `9bbacc42` 支持源码定义的 provenance 格式和 CRLF，仍要求 psmux 自身版本精确为 `3.4.0`；仅 tmux 兼容行、其他版本、预发布版本和任意额外词不通过。相同 banner 加入本地子进程合同及版本回归；没有取消版本保护或固定哈希检查。

### 4.2 原 pane 精确定位与返回前复核

首轮 CI 前，固定源码审查发现“全局 session `-t` 加局部 pane `-t`”会让部分校验只看到第一个目标；capture 内部在目标缺失时还可能回退当前活动 pane。产品改用单个 `SESSION:.%N` 目标，并在返回文本前再次检查原 pane，无法确认时丢弃快照。单元回归覆盖 capture 后原 pane 消失的情况。

这是普通删除竞态的防错读措施，不是原子快照或外部 respawn／同名 server 替换的完整仲裁保证。具体稳定源码依据和实现说明见 [Windows 验证指南](../NATIVE_WINDOWS.md)。

### 4.3 删除故障的构造必须有独立证据

第二轮中，`kill-pane` 返回成功，但 `list-panes` 在 20 秒观察窗口内仍包含原 pane，因此测试失败。**实际观测只证明原 pane 对象仍可枚举；该轮未记录此时的 pane_dead，不能把源码解释写成直接观测。**

这里的 Go 测试与 CI 结果为 `fail`；因故障前提没有建立，删除后防错读这项产品验收应记为 `blocked（测试故障构造）`，其拒绝断言当轮未执行。不能据此宣布 Preview 存在错读，也不能算产品通过。

进一步检查 [pane.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/pane.rs) 与 [tree.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/tree.rs)：kill 路径会结束进程，但 `remove_node` 对窗口的单一 root Leaf 保留节点；结合本产品的 `remain-on-exit` 配置，这能解释“进程退出但对象仍在”。不能只根据 `main.rs` 发出的命令正文断言 target 丢失，因为 [session.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/session.rs) 还会传送独立 TARGET 行。

提交 `b6f2cd2a` 改为先查询原 pane 的真实 window ID，使用精确 `kill-window -t SESSION:@WINDOW` 删除仅含原 pane 的窗口。随后必须同时确认原 pane 缺席、另一 pane 存在，并从另一 fixture 的 JSONL 观察到新输入，才执行产品 Preview/Attach 的拒绝断言。产品本身没有新增 kill-pane/kill-window 操作；这是测试故障构造的修正。

第三轮实际记录原窗口 `@1`、原 pane `%1`、另一 pane `%2`；删除后确认 `%1` 缺席，并在另一 fixture 的日志观察到 `sentinel-survived-removal` 输入。之后 Preview 与 Attach 的拒绝断言均通过。该轮才形成“原对象消失后不转读／进入替代对象”的完整场景证据。

## 5. 证据保存与环境结束

每轮证据目录包含 runner 的 `summary.json`、前后源码指纹、完整 native Go 事件、普通测试的最终结果事件、非空命令输出和合成 fixture 记录。`evidence-index.json` 记录下载 artifact 的摘要、每个原文件的大小／SHA-256、保存方式与计数。

普通测试完整 stdout 的哈希保留，仓库只存其 pass/fail/skip 事件；零字节输出保存大小与哈希，不创建空文件。其他标为 `verbatim` 的内容原样保存。已确认每轮 25 个受测源码／模块文件的前后指纹相同，并逐个对照对应 Git 提交的 blob；文档提交不会被当成重新执行产品代码。

GitHub artifact 保留 14 天，具体到期时间在索引中；以上关键文本和合成证据进入仓库，结论不依赖临时 artifact 长期可下载。

测试只清理自己的随机 namespace 和精确后端名称；独立产品进程场景保留精确 process handle，stop 后核对目标 fixture 已退出，必要的清理也只针对该 handle。没有按进程名批量终止、修改用户全局配置或使用真实 agent 凭据。测试 harness 的清理结果不等于已测完整 psmux／ConPTY／任意 agent 子进程树，也不证明没有隐藏 warm 进程。

## 6. Feature Ledger 与剩余门槛

新增局部证据映射到 FL-001/002/004/005/006/013/016/019，完整说明见 [Feature Ledger](../../FEATURE_LEDGER.md)。FL-006 的实际证据是**管理命令进程结束后**独立进程重新发现同一 Run，不是桌面 TUI 关闭与 attach/detach 的实测替代。

所有原“实现中”条目保持不变，没有能力升级为已验收／已发布。339 项 JSON 设计用例和生成目录仍保持 `designed/not_run`；本轮独立测试不能直接把相似的整条 catalog 用例改为 pass。FL-020 人工介入提醒、M3 消息与 Host 仍未实现。

下一道 M0 门槛继续保留：

- Windows 11／Windows Terminal 的实际 UI、字体、IME、粘贴、Ctrl+C、焦点、resize、进入／分离／返回，以及至少一个真实目标 agent。
- 确认无 WSL 环境的完整流程；最低 OS／架构／shell 组合由结果确定。
- 0/1/5/10 会话全进程树、工作集／Private Bytes、句柄、响应及输出负载；先收集基线再确认资源预算。
- 后端繁忙／断连时列表完整性、外部 respawn、进程树退出和更广的 argv／shell 参数组合；成功的有限样本不扩大为所有 CLI 兼容承诺。
- 完整表单、持久配置、SQLite 历史、有界日志及后续阶段功能；这批后端测试不代替实现。

重现入口见 [原生 Windows 自动化指南](../NATIVE_WINDOWS.md)。本轮不发布预览安装包，M0 保持未完成。
