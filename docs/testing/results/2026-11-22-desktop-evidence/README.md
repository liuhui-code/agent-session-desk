# Windows 桌面合成测试证据

本目录对应 [执行报告](../2026-11-22-windows-desktop-acceptance.md)，不代表 M0 全部验收通过。

- `binary-hashes.json`：实际产品/后端/Terminal/PowerShell 的哈希；`native-reader-build.json`：本轮合成原生读取程序身份。
- `done/`、`responses/`：实际队列命令及其结果，含捕获时的终端 UIA 文本、前台进程、矩形及键盘布局。命令数不是用户动作数，截图与等待亦是命令；失败及未进入预期目标的输入不得计为产品通过。
- `fixture-*.jsonl`：仅合成测试输入的独立进程记录，保留 PowerShell Unicode 变化与重复粘贴污染的早期样本。
- `final-check.json`、`pair-final-check.json`：正式样本断言、身份和定向 A→B→A；`cleanup.json`、`task-cleanup.json`：精确清理与本次任务移除。
- PNG：选定自有窗口截图，逐张从上下左右各裁掉 8 像素，去除外框阴影附近的其他桌面内容；没有修改产品绘制或候选文字。原像素尺寸、原截图哈希、原始 ZIP 哈希见 `derivation.json`。
- JSON/JSONL 移除了 UTF-8 BOM，并将测试用户主目录替换为 `<USERPROFILE>`；保留时序、PID、Run、文本和失败值。`derivation.json` 中哈希对应处理前原件，`SHA256SUMS` 对应本目录最终文件。
- PS1 是本轮测试辅助程序的最终源码快照，非产品实现、非默认 CI，不会自动执行；不能直接重跑固定 namespace 或覆盖既有任务。早期驱动的矩形表达式、重复 paste 分支、数组包装等问题及复验在报告中说明，最终脚本不能倒充早期无错误运行。

所有客户端都使用私有 runtime/namespace。运行桌面驱动需要当前已登录用户的普通权限 InteractiveToken，逐次核对前台窗口；不需要修改系统执行策略。原始 ZIP 留在本次本机临时工作目录，仓库不收录真实会话、环境转储或桌面全屏。

官方依赖来源：https://github.com/microsoft/terminal/releases/tag/v1.25.3131.0 、https://github.com/PowerShell/PowerShell/releases/tag/v7.7.0 、https://github.com/psmux/psmux/releases/tag/v3.4.0 。这些只作为隔离验收工具，不更改产品发行依赖。

测试源保留了失败版本 `desktop-driver-v2.ps1`（重复 paste 分支）及最终版本 `desktop-driver.ps1`；最终启动器为 `start-driver-v3.ps1`，v2 启动器只保留失败/中间运行的真实配置。任何复现都应换用全新目录、任务名和 namespace，不应直接重放旧 done 命令。

文本归档统一使用 LF，保证 Git 检出后的 SHA256SUMS 可复核；derivation.json 中原件哈希保持不变。
