# Windows 分离键诊断证据

对应 [诊断报告](../2026-11-22-windows-detach-diagnosis.md)。均为自有窗口、合成进程和明确的诊断 namespace。

- `responses/`、`done/`：实际命令及观察，不把截图/等待计成用户动作。起始复现和恢复失败亦保留。
- `psmux-input-debug.log`：固定后端直接 attach 的有界事件日志，只在该私有进程作用域启用；没有真实 agent 内容。UTC 收据与日志本地 UTC+8 时间对应。
- `direct-regression.json`：直接客户端后两轮；第一轮是 responses/0031–0034。`managed-regression.json`：旧预览三轮；`candidate-regression.json`：本轮提示候选三轮。
- `mode-control.json`：同 attach 内先失败、再只改输入模式完成同一个 prefix。`final-check.json`：Run/夹具未重启、Windows 候选进程及哈希。
- `candidate-build.json`：构建工作树的源码 SHA-256；`verification.json` 及 verify-*.log：实际 Go 检查输出；go-download.json 是官方固定工具链发行物的校验元数据。
- `*.ps1`、run-*.py、request.py：本轮测试辅助源，不是默认 CI 或产品实现；使用时必须更换私有目录、namespace、任务名与命令 ID。它们通过已授权的 windows-pc SSH 和 InteractiveToken 桥接，仅向已核验前台的私有 Terminal 发送输入。不应重放已结束会话的命令。
- 选定 PNG 各边裁掉 8 像素，移除外框阴影区域的无关桌面像素；未修改产品画面。JSON/日志中的测试用户主目录被替换为 `<USERPROFILE>`。原始 ZIP、原图及固定源码归档哈希见 derivation.json；最终归档哈希见 SHA256SUMS。

图形驱动从上一轮本仓库的 `2026-11-22-desktop-evidence/desktop-driver.ps1` 派生；只改用独立诊断输出目录，保留旧的事件机制以作对照。并未借改输入驱动隐藏原失败。psmux 固定源码只读用于核对日志和 prefix 实现，未迁入产品或此目录。

文本归档统一使用 LF，保证 Git 检出后的 SHA256SUMS 可复核；derivation.json 中原件哈希保持不变。
