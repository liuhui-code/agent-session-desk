# 2026-11-27 配置与 Agent 接入说明

文档更新，无产品代码或配置合同变更。新增 [CONFIGURATION.md](../../CONFIGURATION.md)，按 metadataPath／portableOptions／withQuickLaunch／各配置读取器核对目录、scope、便携回退、最近启动优先、名称来源、脚本入口和可选 Agent 消息／提醒接入。补四份不自动安装的 TOML 示例；README、接手索引、API、启动与工作区说明互链，ADR-0042／Feature Ledger 同步。

实际验证：四份独立示例分别由 config.LoadLaunch、LoadUI、keybindings.Load、LoadWorkspace 解析；总览内两份启动示例由 LoadLaunch 解析，六项均通过。[检查源码与原始输出](2026-11-27-configuration-guide-evidence/)保留。检查器临时放在 tools/tmp_config_doccheck，执行后移除；复现时可把归档的 Go 文本临时放回相同位置，在仓库根目录 go run ./tools/tmp_config_doccheck。

同时实际确认 LoadLaunch 拒绝含 workspaces 顶层表的配置，LoadWorkspace 接受其独立文件；文档明确拆分及 UI n／W 混用限制，本次没有修改读取器或声称该缺口已修复。

本地链接检查覆盖 README、总览、接手索引、API、启动与工作区六份文档，均可找到目标；diff 空白及产品源码行数检查通过。只做解析和文档检查，未运行全仓套件、未启动 shell／Agent／后端、未改日常 Windows 配置。路径为 Windows 占位值，解析通过不是程序存在、品牌兼容或原生 Windows 执行证据。
