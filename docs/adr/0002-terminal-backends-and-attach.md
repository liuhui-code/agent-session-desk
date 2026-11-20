# ADR-0002：采用可替换终端后端和全屏 attach 交互

- 日期：2026-11-21
- 状态：已接受（Accepted）
- 实现与验证状态：上游接口和平台差异已核查；本项目未完成 Windows 编译或实机验证
- 关联需求：FR-01、FR-02、FR-03、FR-05、FR-06、FR-10、FR-11、FR-12
- 关联决定：[ADR-0001](0001-architecture-and-reuse.md)、[ADR-0003](0003-bounded-resources-and-background-work.md)

## 背景与约束

现有 UI 参考项目仍包含 Unix/tmux 假设。ConPTY 提供 Windows 子终端托管机制，但直接使用它还需要处理输入输出、屏幕状态、尺寸、关闭和进程生命周期。优先复用现有后台可以缩小首版范围。

## 决定

1. 定义统一的终端后端接口，承载创建、发现、读取预览、attach、输入、尺寸、退出检查和停止。
2. **Windows 使用 psmux**；macOS/Linux 使用 tmux 适配，分别按平台验证。Unix 支持需另行完成适配并验收。
3. psmux **v3.4.0 正式发布**作为首轮 Windows 验证基线，tag 对应提交 `20aa1d7254ee93f64dafa059ef89341d9767470a`。实际测试记录二进制来源与 SHA-256；开发分支的同名 Cargo 版本不视为相同能力。
4. 初始保持一个产品逻辑 Session 对应一个后端 session，记录稳定目标及 Run generation。后台默认每 session 一个 server 的成本进入内存报告。
5. 管理状态使用列表与只读预览；进入会话时，通过子进程执行/终端交接机制运行 psmux attach，分离后恢复管理界面。该路径须经 Windows 实测验证。
6. attach 期间管理器侧栏与 Tab 暂时隐藏。首版不承诺常驻侧栏中的完整可交互嵌入终端。
7. 生命周期由终端后台托管，不另建第二个 ConPTY 层。停止、重启、归档与关闭界面分别处理。

## 兼容合同

实际采用的命令和参数必须逐项验证：创建、会话/pane 定位、capture、buffer/paste、提交键、attach/detach、resize、停止、退出证据和环境传递。psmux 的 tmux 风格接口不是整个 tmux 行为等价的证明。

优先复用 Agent Manager 对 Windows 目标定位、命令拆分、参数转义和版本探测的处理。它对控制通道、光标等能力的降级策略仅代表该版本的经验；新基线是否需要相同降级应由测试决定。

被动预览只读取当前终端状态，不持续改变其他 attach 客户端正在使用的尺寸。终端切换、控制台恢复、Ctrl+C、错误退出和快捷键冲突是 M0 的必要验证。

## 生命周期边界

- psmux session/pane 存在、shell 存活和 agent 存活分别获取证据。
- 缺少 agent 级证据时显示“终端运行中 / agent 状态未知”，不伪造忙闲或任务完成。
- 退出管理器后，存活后端继续托管；重新连接验证原 Run 身份。
- agent 重启生成新 Run；恢复原对话仅在 agent 提供并通过验证的能力下执行。
- 电脑重启后不能接回原来的活进程。

## 暂缓选择

直接 ConPTY 引擎、完整内嵌终端、每工作区共享一个后端 session，以及 psmux 开发分支共享 server 均暂缓。共享模式不是本轮稳定 v3.4.0 的能力承诺。若实测每 session 后台成本不达标，再通过独立 ADR 评估映射调整及故障范围。

## 验证与重新评估

Windows 实机需覆盖自定义命令、含空格路径、中文/宽字符/IME、多行粘贴、全屏 TUI、颜色/光标、缩放、分离/重连和目标退出。最低 Windows 构建、架构、终端与 shell 支持范围按结果确定。

若某个关键 psmux 接口无法可靠完成，先固定问题、最小复现和版本；再选择已验证的修复版本、局部适配或替代后台。不得以 Linux 编译或框架文档解除 Windows 发布门槛。

## 依据

- [psmux v3.4.0 发布](https://github.com/psmux/psmux/releases/tag/v3.4.0)
- [psmux 稳定版架构](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/docs/architecture.md)
- [Agent Manager Windows 适配](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/internal/tmux/platform_windows.go)
- [ASMGR attach 流程](https://github.com/izll/agent-session-manager/blob/e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed/ui/handlers_session.go)
- [Microsoft 创建 pseudoconsole](https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session)
- [Microsoft ClosePseudoConsole](https://learn.microsoft.com/en-us/windows/console/closepseudoconsole)
