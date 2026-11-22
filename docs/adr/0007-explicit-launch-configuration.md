# ADR-0007：显式 TOML 启动模板

- 日期：2026-11-22
- 状态：已接受（Accepted），用户批准 M1 首个配置启动切片
- 关联：FR-01、FL-001、ADR-0003、ADR-0004、ADR-0005；TBD-CONFIG 部分细化
- 实现与验证：只读配置与 CLI 模板入口已实现；[实际报告](../testing/results/2026-11-22-m1-config-tokens.md)。不代表 M0 或 M1 整体完成

## 决定

沿用已接受的 TOML 方向，解析依赖固定 `github.com/BurntSushi/toml v1.6.0`，支持 TOML 1.1.0 语法，MIT，Copyright (c) 2013 TOML authors。通过 Go 模块使用，不复制解析器源码。依赖升级继续服从 ADR-0005。

入口为 `asd --config PATH start --template NAME`。只读明确路径，不搜索当前目录、不自动挑模板、不写回、不监听变更。每次 start 重新读取，已有 Run 不因文件变化而变更。

配置必须有整数 `schema_version = 1`，`templates` 按模板 ID 索引。模板字段仅有 `name`、`executable`、`args`（字符串数组）、`cwd`、`env`（字符串映射）；schema 字段严格区分大小写。模板 ID 和 env 名是数据。无文件返回读文件错误；空文档、缺失版本、未来版本、未知键、重复键、语法或类型错误都拒绝。全文件执行语法／类型／未知字段检查，选定模板再执行启动值验证。

`executable` 必须提供。`name` 缺省沿用入口文件名。相对 cwd 以配置文件所在目录为基准，缺省 cwd 为该目录；相对 executable/PATH 继续以最终 cwd 解析。配置路径按用户给出的路径处理，不额外解析符号链接的目标目录。不进行环境变量、波浪号或 shell 插值，args 保持空字符串、空格及 Unicode 参数边界。

CLI 显式 `--name`、`--cwd`、重复 `--env KEY=value` 覆盖模板；相对 CLI cwd 仍以调用者目录解析，同名 env 由最后值覆盖（Windows 不区分 env 名大小写）。不能同时提供模板和原始 executable；配置路径与模板选择必须成对。配置错误在初始化后端之前返回，已有 psmux/tmux 字面传输拒绝规则继续生效。配置解析成功不表示后端能够原样传递任意输入。

防御性读取上限为 1 MiB，超过即拒绝；这是实现边界，不是已测量的资源预算。解析错误只显示通用原因及可用的行／列，不回显源行、键名或 env 内容。OS 文件错误可包含用户指定的文件路径。

## 后果与后续门槛

配置可复用，直接 CLI 启动保持兼容。严格 schema 让拼写错误提前失败，增加字段需有明确兼容决定。本次不包含保存、迁移、热重载、主题文件、SQLite 或配置界面；TBD-CONFIG 的保存中断／旧版本恢复等问题保持 open。完整验收仍需真实目标 agent 和更多原生参数／脚本组合。

## 依据

- [精确依赖](../../go.mod)、[实现](../../internal/config/config.go)、[用法](../LAUNCH_CONFIG.md)
- [v1.6.0 发布与 TOML 1.1 支持](https://github.com/BurntSushi/toml/releases/tag/v1.6.0)
- [固定版本文档及许可](https://github.com/BurntSushi/toml/tree/v1.6.0)
