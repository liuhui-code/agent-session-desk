# 2026-11-26 外置 Markdown 打开：局部执行

关联 [ADR-0052](../../adr/0052-open-external-markdown-documents.md)、[使用说明](../../EXTERNAL_MARKDOWN.md)及 FL-012/016/018。代码实现 Ctrl+单击引用、原会话路径解析、Windows 默认关联调用；本轮不替换日常 exe，不改变数据库／用户文件关联或读取真实会话内容。

## 源码、产物与 TDD

以 main `0f7a0d1a9ea7e627acacea4e1f281f216843393a` 加本轮显式 Go 文件覆盖，冻结 405 个源码／资源文件；排除原先未提交的 internal/backend/terminal.go 与 terminal_failure_test.go。Windows 构建全部从冻结目录执行。[run1 指纹](2026-11-26-external-markdown-evidence/run1/source-fingerprints.json)与[最终 run2 指纹](2026-11-26-external-markdown-evidence/run2/source-fingerprints.json)分别保留，后者修改含糊引用拒绝行为。最终 asd-markdown.exe SHA-256：`f844e69a9b7329143eacd37c3931f213794d812a87b43905f9d6d9cd83b1331f`，不是公开 Alpha 或日常 exe。

本地 Go 1.27.4 darwin/amd64，[精确版本](2026-11-26-external-markdown-evidence/go-version.txt)。初始新增测试因缺少 At／Resolve／UI 接口形成[RED](2026-11-26-external-markdown-evidence/red.txt)；同批测试的 fakeService 字段误用、macOS /var→/private/var 规范路径期望及环境变量不允许 NUL 的脚手架错误分别留在[first setup](2026-11-26-external-markdown-evidence/first-green-setup-failures.txt)和[NUL setup](2026-11-26-external-markdown-evidence/nul-env-setup-failure.txt)，修正测试前置条件后复验。Python 3.9 不支持 extractall(filter=...) 的归档准备失败后，改为仅解包经过路径检查的本仓库 git archive；初次缺少 go.mod 的构建没有生成可计入证据的产物。

新增两个歧义回归例先证明旧候选会误把 `C:\work space\plan.md` 的后半段、嵌套链接中的 `plan.md` 识别为相对文件：[实际 RED](2026-11-26-external-markdown-evidence/ambiguous-red.txt)。随后拒绝无明确边界的路径与复杂链接，并复验[最终 clean GREEN](2026-11-26-external-markdown-evidence/clean-green.txt)、[相关 race](2026-11-26-external-markdown-evidence/race.txt)、[vet](2026-11-26-external-markdown-evidence/vet.txt)及[模块验证](2026-11-26-external-markdown-evidence/modules.txt)。没有运行全量 UI／稳定性专项。

## Windows 原生结果

Windows 10 19045，明确使用现有本机 psmux 开发候选 `ff9d8893…fb7b16b`，不替换官方基线；复用已归档的 fixture `6ed5ed77…c73388f`，未把旧 fixture 计为本轮新建实现。

| 轮次 | document | backend | manager | UI | 失败／跳过 |
| --- | ---: | ---: | ---: | ---: | --- |
| run1 | 3 | 2 | 1 | 2 | 0／0 |
| run2 最终 | 3 | 2 | 1 | 2 | 0／0 |

每轮 8 个顶层测试，重复执行不相加为 16 个独立行为。摘要：[run1](2026-11-26-external-markdown-evidence/run1/summary.json)、[最终 run2](2026-11-26-external-markdown-evidence/run2/summary.json)；四包原始 stdout、驱动和各自二进制 SHA 均在对应目录。

- 文本测试覆盖普通／引号／基本 Markdown 链接、中文与 emoji 前缀单元格、URI、行列／锚点和可执行扩展拒绝；路径测试使用实际临时普通文件，不以“文件名看起来合法”代替存在性。
- UI 模型验证一次 Ctrl+点击只请求原目标文档，点击／释放没有写入终端、选区或程序鼠标手势；切会话后旧解析结果拒绝。这里使用内存终端及观察型 opener，不是实体鼠标。
- 真实 psmux 启动受控 CLI，原 pane 查询的绝对目录与子进程自己报告的 cwd 一致；篡改 Run 后拒绝。没有验证 PowerShell provider cd、shell OSC 目录集成或历史输出产生时的目录，不将静态 cwd 通过扩大为这些场景通过。
- Windows 原生执行 STA COM 初始化与 SHELLEXECUTEINFO x64 112-byte ABI；观察 open 动作、独立 lpFile、NULL lpParameters、默认 class 与 NOASYNC/NO_UI，取消请求不调用 Shell。ShellExecuteEx 调用本身由测试观察器替代，**未实际启动默认阅读器**，不能算默认应用关联或阅读器渲染已验收。

本轮没有计划任务、桌面／剪贴板 API 或强制结束产品。native Harness 正常清理自己创建的 fixture；最终脚本匹配四个明确测试路径的剩余进程为 0。该审计不覆盖所有后端进程。没有运行实际 Windows Terminal Ctrl+单击、个人编辑器、日常快捷方式或真实 Agent。

## 旧用例失败及剩余边界

扩大到相关程序／列表鼠标子集时，[候选日志](2026-11-26-external-markdown-evidence/related-ui.txt)有 5 个旧顶层失败。在无本轮改动的同一 main 基线目录执行相同旧用例，[baseline 日志](2026-11-26-external-markdown-evidence/baseline-related-ui.txt)得到完全相同 5 项：

- TestWorkspaceListMouseUsesVisibleRowsAndIgnoresIndicators
- TestWorkspaceListDoubleClickWhileConnectingKeepsManagementFocus
- TestProgramMouseKeepsOneReleaseSlotUnderMotionPressure
- TestProgramMouseExternalModeChangesDoNotReleaseIntoAnotherMode
- TestProgramMouseExternalScreenGrowthThenRestoreCannotResumeGesture

后三项已在 STAB-006 记录文字断言失配；前两项另记 STAB-007，不在文档功能切片中修改。保留失败，不宣称整个相关鼠标子集通过，也不认定它们由本轮引入或已经解决。

默认阅读器真正打开、实体 Ctrl+单击、动态／远程 cwd、换行长路径和复杂 Markdown 链接仍开放。路径规范化与 OS 打开之间不是文件系统原子快照；行列／锚点不保证默认应用定位。生产、测试与驱动代码均 ≤500 物理行，catalog 仍 339 项 designed/not_run；文档校验不提高产品接受状态。SQLite v9、ui.toml v7、依赖与用户日常入口保持本轮之前的状态。

实现与 ADR／证据提交：`66fc042b6bddaac519b2f321ed7900ab040f1359`（本地 main）。本补充记录提交号并合并重复决策／STAB-007 行，没有推送或替换日常 exe。
