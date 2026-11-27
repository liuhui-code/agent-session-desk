# 2026-11-27 当前程序标签

ADR-0061。旧卡片用不可变Command，所以PowerShell内开CLI仍显示pwsh；改为原pane身份同时观测的CurrentProgram，仅内存不入JSON。安全basename、隐藏shell/node包装名，未识别不猜。一次查询扩展，无新增轮询。

RED：UI缺CurrentProgram/旧启动标签；GREEN覆盖独立标签、不可变Command及控制/参数拒绝。普通全套通过；原native后端helper需支持新增format，未放松产品目标校验。首轮loopback受沙箱限制，复验分列。

Win10原生：指定开发psmux，同一PowerShell启动3个重命名fixture（codex/claude/opencode），List观测名字及退出清除，原Command保持；通过。只证明进程观测，不是3品牌Agent兼容。无真实内容、全桌面或资源预算接受。
