# 原生复制／搜索证据索引

结论见[实际报告](../2026-11-23-native-copy-search.md)。本目录不含二进制、数据库、认证 key 文件或原剪贴板内容。

- `initial/`：首轮复制成功但未画选区的构建，以及开发中的 RED/GREEN／诊断日志。部分文件名含 green 的中间尝试实际仍失败，须以内容为准；最终结果在 `final2/`。
- `before-resize-fix/`：接入选区后的中间构建、过严复制光标拒绝、Running 任务观察与只读协议探针。初次业务失败的结构化内容保存在 `native-inspect-failure.log` 内的 failure 字段；随后 flow-failure 文件被续接脚本更新，不冒称仍有其原字节副本。
- `final2/`：最终源码指纹、本地 Go 检查、三个最终 EXE 的原生测试／回执、实际续接、独立审计和清理。
- `final2/ui-tests-final2.log`：93 顶层／254 含子测试；`terminal-tests-final2.log`：20／38，均 0 FAIL／SKIP。它们是 Windows 上执行的 Go 套件；真实 psmux 另由桌面流程验证，不把 fake transport 合同当作 real-backend 场景。
- `final2/desktop/responses/` 与 `done/`：同一任务累计 150 个命令回执，含旧 UI 的 resize 失败、退出保护和新 UI 成功续接。ok 表示驱动动作／抓取成功，业务结果由独立断言核对。
- `final2/copy-metadata-probe.json`：只认证后 dump-state，没有 attach、resize 或输入；61×18 下原生复制光标仍为行 24／列 41。认证值从未输出。
- `final2/shells-before-ui-replacement.json`／`flow-result.json`：两个原 Run、PID／FILETIME／完整路径在替换前后相同。`replaced-ui-process.json` 记录最终 UI 自身身份。
- `final2/exit-guard-failure.json`／`native-inspect-exit.log`：物理 SendKeys q 未退出旧 UI，保护阻止了启动命令。成功脚本为 `resume-after-exit-guard.ps1`，从序号 99 继续，经观察只读管理页、Esc／Unicode q 和退出确认后才启动新 UI。
- `final2/final-audit.json`／`cleanup.json`：两个 shell 与最终 UI 登记身份已结束，指定映像无存活，任务／Session 清理，原终端配置恢复。
- `final2/desktop/clipboard-restored.json`：原剪贴板文本与格式名恢复；非文本内容未验证，原数据只在驱动内存，不进日志。expectClipboard 只允许合成行；摘要对返回实际字符串另做大小写敏感的 CRLF→LF 核对。
- `protocol-source.json`：固定上游、psmux 与实际锁定库 API 的文件哈希；Rust 仅参考协议。
- `manifest.json`：临时文件原字节／归档 SHA-256；PNG 未改，文字统一 UTF-8/LF 并替换测试用户、实际 home／仓库路径。清单、此索引和后续检查摘要不自包含哈希。
- `checks-summary.json`：独立核对最终套件数量、124 项源码指纹、150 回执、原身份集合、实际合成复制内容与清理。

复现需使用新隔离目录并先核对进程／任务。失败后从最新已确认状态和新序号续接，不直接重新执行全部输入；不确定输入禁止重放。恢复设置需校验前后 hash，停止目标需按登记 Session/Run／PID／FILETIME／路径，不按名称强杀。引用的新 UI 只替换视图，不重建 shell。

全部 339 项目录仍 designed/not_run；完整鼠标、CLI 主动复制、IME、真实 agent、连续历史／缩放与资源门槛保留。截图中的历史编号缺口没有补写或掩盖。
