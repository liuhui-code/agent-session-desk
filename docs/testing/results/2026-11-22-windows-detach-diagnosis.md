# Windows 分离键诊断：输入法模式与实际按键投递

日期：2026-11-22。**WIN-UI-05 在“取消候选、切至英文、顺序按 Ctrl+b / d”的明确条件下完成回归。** 中文模式中的 d 会先进入微软拼音组合，不能据此认定 ASD 或 psmux 丢键。本轮没有修改后端、键位或自动切换用户输入法；修正了产品提示、操作指南和测试前提。

本报告接续 [Windows 桌面执行记录](2026-11-22-windows-desktop-acceptance.md)，保留其中首次 blocked 的历史。完整显示缩放/另一套 IME、真实目标 agent 等门槛仍开放，不构成完整 M0 或 Windows 支持声明。

## 可复现边界与根因

固定组合仍为 Windows 10 build 19045、官方便携 Windows Terminal 1.25.3131.0、PowerShell 7.7.0、官方 psmux 3.4.0；后端 EXE SHA-256 为 `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`。诊断使用独立 namespace `asd-detach-20261122` 和合成 ReadConsoleW 夹具，只向自有前台 Terminal 输入。

反馈循环以真实窗口状态为判据：attach 时显示夹具画面；分离后直接客户端返回 PowerShell 提示符，ASD 客户端返回管理器并显示“已返回管理器”。每次回归还核对 Session/Run 未变、夹具只启动一次，未用关闭客户端或重新启动 Run 冒充分离。

复现后按顺序检验三项假设：① IME/驱动截住 d；② psmux prefix 状态或时序；③ ASD 终端交接留下异常状态。实际结果：

1. ASD 下“Ctrl+b → 物理 VK_D”复现失败。绕过 ASD 直接 attach 同一个 psmux 后端仍失败，排除 ASD 独有的交接原因。
2. 直接客户端只启用私有、限量的 `PSMUX_INPUT_DEBUG`。日志收到 `Char('b') / CONTROL / Press` 并发出 `prefix-begin`，随后物理 d **没有产生后端键盘事件**。同一时刻截图可见微软拼音“的/都/对”候选，证明按键在到达 psmux 前用于 IME 组合。[候选现场](2026-11-22-detach-evidence/ime-blocked-physical-d.png)
3. 取消候选后，只发送 d，不重复 Ctrl+b，原 prefix 仍可完成分离。日志相隔约 44 秒后才收到 `Char('d') / 无修饰键 / Press`；此样本不支持“短暂等待导致 prefix 丢失”的猜测。
4. Shift 切到英文模式后，以物理 a 的直出画面检查模式，不能只看 HKL：中英文模式下 HKL 都是 `134481924`。相同 Ctrl+b / 物理 VK_D 在直接 psmux 下连续 3 轮成功，在原预览 ASD 下连续 3 轮成功。
5. 再在同一个 ASD attach 客户端做单变量对照：英文→中文模式，Ctrl+b / 物理 d 留在候选；只 Escape 取消候选、Shift 切英文，再按**物理 d**就返回管理器。没有重发 prefix、没有使用 Unicode d 替代，也没有重启客户端或 Run。[中文状态](2026-11-22-detach-evidence/managed-chinese-red.png)、[同客户端恢复](2026-11-22-detach-evidence/managed-english-green.png)、[结构化收据](2026-11-22-detach-evidence/mode-control.json)

因此，本轮失败的直接原因是**测试没有核验输入法模式，d 被中文组合截留**。此前不断重发 prefix、混用 Unicode 和虚拟键的样本不能作为可靠的分离回归；屏幕上的 `^Bd` 也不能证明 agent 接收或执行了指令。

微软说明，拼音的 Shift 可在中英文模式间切换，且具体键位可以自定义；产品不应假定所有输入法都使用相同切换键，见 [Microsoft Pinyin 官方说明](https://learn.microsoft.com/en-us/globalization/input/simplified-chinese-ime)。本轮以现场候选、英文直出与后端事件证明状态，官方说明仅用于解释默认操作。

Context7 返回的 psmux 文档来自 master，故没有用它推断固定发行物的行为。实际只读检查固定提交 `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 [client.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/client.rs)与 debug_log.rs：后端输入日志有数量上限并随 `PSMUX_DATA_DIR` 进入私有目录。固定源码为 MIT；本轮只读核查，未迁入 Rust 源码或替换后端。日志的墙钟为 Windows 本地时间（UTC+8），队列收据为 UTC。

## 最小修正与验证

- 管理器帮助页、连接提示及 CLI help 明确要求取消输入法候选、切英文，再按并松开 Ctrl+b、随后按 d。
- [开发指南](../../M0_DEVELOPMENT.md)和 [WIN-UI-05 执行协议](../WINDOWS_INTERACTIVE_ACCEPTANCE.md)同步输入法前提。没有更换默认分离键，没有在用户未提交候选时自动确认、注入或切换输入法。
- 没有改动 Session/Run、psmux 参数运输、终端进程托管或 parser 防护。测试脚本仍区分物理虚拟键与 KEYEVENTF_UNICODE，不能把两者混计。

| 执行对象 | 实际结果 |
| --- | --- |
| 原预览：直接 psmux，已核验英文模式 | 3/3 分离到外层 shell，真实后端日志含 Ctrl+b 与 d；第一轮见 responses/0031–0034，另两轮见 direct-regression.json |
| 原预览：ASD 管理器，同模式及事件路径 | 3/3 返回管理器，见 managed-regression.json |
| 同一个 ASD attach 内中文→取消→英文 | 对照通过；只补物理 d 完成原 prefix，见 mode-control.json |
| 提示修正版：Windows 桌面实际执行 | 3/3 返回管理器，见 candidate-regression.json；候选 EXE 哈希、实际进程路径与 source manifest 一致 |
| 新帮助页 | 120×32 与约 60×18 均实看可读，输入法前提和按键顺序可见；[窄窗口截图](2026-11-22-detach-evidence/candidate-help-narrow.png) |
| 源码检查 | Go 1.27.4：`go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 全部退出 0；catalog check --generated 通过 |
| Windows 构建与执行 | CGO=0 Windows amd64 交叉构建退出 0；随后复制到 Windows 实机运行、观察帮助页并执行上述 3 轮。未把交叉构建本身称作 Windows 测试 |

这是通过 OS SendInput / SendKeys 产生的虚拟键桌面自动化，不冒充人工按实体键盘或其他终端/IME 的验证。测试阶段的无关采集问题（本机 Python tarfile 的 filter 参数不兼容、生成候选回归脚本时变量名匹配错误）均在执行前发现并修正，没有改产品或计入通过轮次。

## 构建身份、清理与剩余范围

提示修正版基于 `5c94b7d407a59be3f8fa476ea6c584aac67110bf` 的工作树；本轮产品变化仅为三个 Go 文件的提示文案。候选 EXE SHA-256：`5085d052c5468ced425936409f935628d01b7a6c10e9fb3559869afc6a5088b9`。完整源码文件哈希和依赖锁文件见 [candidate-build.json](2026-11-22-detach-evidence/candidate-build.json)，实际 Windows 进程/EXE 核验见 [final-check.json](2026-11-22-detach-evidence/final-check.json)。未发布、未替换正式测试预览、未提交或推送。

诊断 Run 和原夹具保持同一身份，随后按精确 Session ID 停止；最终 namespace 空、原 PID/启动时间不再存活。专用交互任务已删除，私有 Terminal 窗口无残留，私有 Terminal 设置已恢复，直接客户端的诊断环境变量随作用域清除。见 cleanup.json 与 task-cleanup.json。没有改系统执行策略、输入法配置或无关任务。

本轮关闭的是 WIN-UI-05 在该组合和英文模式前提下的未决项。**中文候选仍会正常接管字母，本轮没有实现跨 IME 的全局分离键。** 系统显示缩放/另一套 IME、PowerShell Unicode 对照失败及其他 M0 门槛继续保留；339 项 catalog 的 designed/not_run 基线不变。

[证据目录](2026-11-22-detach-evidence/README.md)保留实际命令、输入事件、回归脚本、Go 输出和失败/恢复截图；文件哈希覆盖脱敏后的归档。没有把静态文档校验算作产品通过。
