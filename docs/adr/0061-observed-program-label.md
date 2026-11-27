# ADR-0061：连接行显示当前程序

2026-11-27，用户要求。连接行不再复述启动shell。List 在原pane身份/dead查询追加 pane_current_command；仍核对pane身份，失败不保留旧标签。命令观察失败可回退原状态查询，不能因此改写启动身份。

CurrentProgram仅内存，json忽略；Command仍是不可变启动快照。只留96字节内basename，参数/控制串拒绝，去.exe；pwsh/powershell/cmd等shell及node包装名不展示。没有可靠品牌证据则只显示Connected，不凭屏幕/标题猜agent。不用于ready/完成/批准或自动输入。

Context7取master文档仅为线索，实际用本机固定bundle验证。改名fixture模拟codex/claude/opencode子进程，验证开始/退出及原Command保持；不是品牌Agent兼容。
