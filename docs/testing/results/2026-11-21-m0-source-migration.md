# M0 首批源码迁入：实际验证报告

日期：2026-11-21（Asia/Shanghai）。结果：**源码迁入、Linux 自动化与构建通过；M0 原生 Windows 验收未完成。**

本报告记录从文档基线进入首个实现切片的实际结果。不是 339 项设计用例的整套执行报告，也没有把 mock、交叉编译或上游能力描述当作 Windows 通过依据。

后续已执行的 Windows 进程／后端自动化见[独立原生报告](2026-11-21-m0-native-windows.md)。本报告保留首轮当时的 `not_run` / `blocked` 结果，不以后续成功覆盖历史。

## 1. 被测版本与证据

- 仓库基线父提交：`80288eaaa5c1da43cbaa1dd53517452d4e4e7e4f`。
- 被测代码：本报告同次提交中的 `cmd/`、`internal/`、`go.mod`、`go.sum`，共 23 个文件。
- 逐文件 SHA-256 及聚合算法：[source-fingerprints.json](2026-11-21-m0-evidence/source-fingerprints.json)。验证开始和结束时指纹一致。
- 聚合指纹：`0ef707a313522008fb5ce1025b6f9de4771f1486dc0fca3e769e4e66a60ff16c`。
- [逐命令退出码、输出、构建产物哈希和测试计数](2026-11-21-m0-evidence/verification.json)。命令记录中的 `<repo>` / `<toolchain>` 是执行目录的规范化标记。
- [Go 最终 pass/skip/fail 事件](2026-11-21-m0-evidence/go-test-events.jsonl)：从真实 `go test -json` 输出提取；原完整 stdout 的哈希也保留在验证 JSON 中。
- [工具链与 tmux 来源记录](2026-11-21-m0-evidence/toolchain.json)、[迁入来源和许可](../../REUSE_MANIFEST.md)。

开发构建发生在提交前，二进制 VCS 信息会指向父提交并带 `modified=true`；准确源码由上述指纹确定。二进制仅放在忽略的 `build/` 中，没有作为可发布兼容包入库或发行。

## 2. 环境

| 项目 | 实际环境 |
| --- | --- |
| 执行系统 | Linux x86_64，Debian 13，kernel 6.18.57 |
| Go | 官方 Go 1.27.4 Linux amd64；归档 SHA-256 与官方值核对一致 |
| UI 依赖 | Bubble Tea 2.1.2、Lip Gloss 2.1.1、x/ansi 0.11.9；没有 Bubbles |
| Go 检查模式 | `GOFLAGS=-mod=readonly`；验证期间模块与源码不变 |
| Unix 后台 | tmux 3.8，从官方发行源码在临时工具链目录编译；无系统安装 |
| Windows | 没有 Windows 主机；amd64 交叉编译，未运行 |
| psmux | 仅检查 v3.4.0 固定源码 `20aa1d7254ee93f64dafa059ef89341d9767470a`；没有在 Windows 执行 psmux |

本次没有安装新的产品常驻服务，没有修改用户全局后端配置，也没有用环境限制的绕过方式取得终端运行结果。

## 3. 检查结果

| 检查 | 结果 | 证明范围 |
| --- | --- | --- |
| `go test -count=1 -json ./...` | pass | Linux 下 4 个有测试包通过，命令合同使用假后端子进程 |
| `go test -race -count=1 ./...` | pass | 本次执行到的 Go 并发路径未报告 data race |
| `go vet ./...` | pass | 静态检查无报告 |
| `go mod verify` | pass | 当前缓存模块与校验记录一致 |
| Linux amd64、CGO=0：asd / fixture | pass | 开发二进制构建 |
| Windows amd64、CGO=0：asd / fixture | pass（仅交叉编译） | Windows 编译目标可构建，未执行 |
| Windows 的 backend / cli / session / ui 测试二进制 | pass（仅交叉编译） | 测试源码可针对 Windows 构建，未执行测试 |
| Linux 实际 asd help / version | pass | 入口、用法和开发版本可运行 |
| Linux 实际 asd doctor + tmux 3.8 | pass（版本探测） | 后端二进制可执行且版本符合合同；没有创建终端 |
| Linux 直接运行 fixture | pass | 空参数、空格、中文／引号／反斜杠、cwd、显式合成 env 与 stdin 回显；没有经过终端后台 |
| Linux 实际 asd list | blocked | 一次尝试遇到本地 Unix socket 权限限制，无法开始真实后端集成 |
| Windows 原生端到端、IME、attach/detach、生命周期 | not_run | 缺少原生 Windows 环境 |
| 内存／全进程树基线、长稳 | not_run | 本次没有对应真实后台负载，也没有接受数值预算 |

### Go 用例计数

| 包 | 顶层通过 | 顶层跳过 | 子测试通过 |
| --- | ---: | ---: | ---: |
| backend | 15 | 0 | 0 |
| cli | 4 | 0 | 0 |
| session | 5 | 1 | 0 |
| ui | 14 | 0 | 140 |
| 合计 | **38** | **1** | **140** |

`cmd/asd` 与 `cmd/asd-fixture` 没有测试文件，不计为跳过的测试。唯一真正跳过的用例为 `TestWindowsLaunchEnvironmentAndScriptBoundary`，理由是 Windows filesystem/PATH 语义需要 Windows 执行；它的测试二进制已经交叉编译，但没有运行。

140 个布局子测试来自 10 个窗口宽度 × 7 个高度 × 2 个帮助状态，覆盖 50／80 列布局切换附近、极小窗口、中文、组合字符和 emoji。它们使用依赖的 cell 宽度规则，不证明所有 Windows 字体的实际字形宽度。

## 4. 发现与处理

| 实际发现 | 处理与回归 |
| --- | --- |
| 首轮 UI 宽度测试出现 keycap emoji 的测量／截断不一致；部分宽度溢出 1–17 cells | 选择性迁入 Agent Deck 的 `cellTruncate` 修复；最终 140 个布局组合通过 |
| 稳定 psmux 会在 `--` 之前改写 `-x=value`，并把独立 `;` / `\;` 解释为后端命令分隔符 | 启动前明确拒绝受影响形式，不改写用户参数；测试核对非法输入不会创建后端会话 |
| psmux 环境赋值会 trim，变量名合同也有限制；tmux 的末尾分号同样可能改变命令边界 | 校验 env、argv、可执行路径和 cwd，并在最终路径解析后复验；环境错误不输出变量值 |
| Windows PATH 查询的大小写与重复 env 可能跟后端收到的覆盖顺序不一致 | 环境覆盖固定为 last-wins；Windows 按不区分大小写去重，专属文件系统用例保留为待 Windows 执行 |
| 选择变化、后端列表重排、Run 改变后，旧异步结果可能覆盖新视图 | Session/Run 身份与请求代数核对；取消旧预览并等其结束，局部回归通过 |
| 后台可能已有另一个 active pane，或原 pane 退出后仍保留 shell/终端记录 | capture 绑定初始 pane；attach 重验身份和 active pane，不主动切换焦点；退出屏幕只读，不声称 agent 可继续输入 |
| 创建后写元数据失败可能留下孤立后台 | 只清理本次新建的精确 ID；无法确认清理则明确报告不确定，假后端子进程测试覆盖该窗口 |

后端 guard 的源码依据与精确范围见[开发指南](../../M0_DEVELOPMENT.md)。这批发现没有被用来宣称所有其他参数组合已原生兼容。

## 5. 真实终端阻断

产品的实际 `asd list` 返回 exit 1，核心错误为：

```text
tmux backend command failed: exit status 1:
error connecting to .../tmux-1000/asd-verification (Operation not permitted)
```

此前独立 Python `AF_UNIX` bind 在 `/tmp` 与工作区也返回 EPERM。由此将真实 Unix 后端会话／attach 验证记为 **blocked（执行环境）**，不记为产品通过，也不借此断言产品本身无法运行。没有尝试多次重试或绕过执行环境限制。

由于没有实际 Windows 执行，也没有声明 psmux no-warm 配置已降低内存、agent 在 UI 关闭后已继续运行，或输入法／Ctrl+C／全屏返回已通过。

## 6. 与已有 Feature Ledger / 测试库的关系

本次产生局部代码证据的能力是 FL-001/002/003/004/005/006/012/013/015/016/019，统一保持“实现中”。没有功能被标成“已验收”，M0 也没有关闭。

| 已有测试设计关注点 | 本次局部证据 | 仍缺的证据 |
| --- | --- | --- |
| TC-LCH-001/002/003/005/009/012/013/024 | 参数、cwd/PATH、无效入口、空值／NUL 与假后端 argv 合同 | 真实 Windows 后台和 agent 接收、完整创建 UI |
| TC-UI-001/002/003/004/007/008/009/010/012 | 源码追溯、空列表、身份选择、迟到结果、v2 按键、尺寸与宽字符单元检查 | 原生字体／终端行为、用户视觉评审 |
| TC-TRM-001/002/003/004/021 与焦点场景 | attach 命令准备、不偷客户端、失败返回、预览控制序列清理的局部合同 | 真实全屏交接、终端恢复、输入／IME／resize |
| TC-WIN-004/005/006/007 | 版本字符串边界、私有配置／环境代码、Windows 交叉编译 | 官方 psmux 二进制来源核对与 Windows 实机验证 |
| 后台、生命周期、资源与发布场景 | 后端身份重验、有界缓存、取消和错误传播的局部检查 | 存储、后台生存、故障恢复、真实负载、发行物与原生验收 |

表中映射是对设计意图的部分证据，**没有把对应整条 TC 判为 pass**。339 项 JSON 测试设计保持原来的 `designed/not_run`，自动生成目录仍只计算设计覆盖。目录校验结果不计为 Go 产品测试或 Windows 通过率。

## 7. 下一道验收门槛

在真实 Windows Terminal 中，用已记录发行来源／SHA-256 的 psmux 3.4.0 和当前源码先完成 fixture 路径，再验证至少一个真实自定义 agent。检查参数与中文路径、输入法／粘贴／Ctrl+C、全屏进入／分离／返回、关闭 UI 后接回同一 Run，以及 0/1/5/10 会话进程树与输出负载。剩余存储和表单实现按 M0/M1 顺序补齐。

本次交付将后续工作建立在可追溯、可构建、带局部回归的源码上；Windows 支持和低内存目标继续以这些实际结果验收。
