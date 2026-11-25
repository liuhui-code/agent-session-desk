# 启动模板

2026-11-26 UI 更新（[ADR-0042](adr/0042-direct-create-terminal-host-and-focus.md)）：管理 `n` 直接使用最近一次成功的完整配置，首次读 `--config` 的 `templates.default`；未指定配置文件时使用 scoped 元数据旁的 `launch.toml`，缺失时生成一次系统 shell 默认模板。`N` 才明确打开编辑表单。最近成功记录独立为 `last-launch.toml`，包含明确 env 覆盖、按私有运行数据保管；不写回作者的默认配置。CLI start 成功也更新最近配置，失败／不确定结果不更新，不从列表的 Command 重建 args。

例如默认配置：

```toml
schema_version = 1
[templates.default]
name = "PowerShell"
executable = "powershell.exe"
args = ["-NoLogo", "-NoProfile", "-NoExit"]
cwd = 'C:\project'
```

用 `asd --config .\agents.toml` 打开 UI 即可使用此默认模板。已有 last-launch.toml 时仍优先最近记录；需要重新采用默认时先退出 UI，备份并移走该最近记录，不能覆盖正在使用的数据。

日期：2026-11-22。M1 首个只读配置切片，完整 M0/M1 验收未完成。合同见 [ADR-0007](adr/0007-explicit-launch-configuration.md)。

在明确的文件（例如 `agents.toml`）中配置自己的 CLI：

```toml
schema_version = 1

[templates.dev]
name = "开发"
executable = "你的程序.exe"
args = ["参数", "含 空格", ""]
cwd = "."

[templates.dev.env]
ASD_PROJECT = "示例"
```

```powershell
asd --config .\agents.toml start --template dev
asd --config .\agents.toml start --template dev --name 评审 --cwd C:\project --env ASD_PROJECT=review
```

后端路径、runtime-dir、namespace 等全局选项也必须放在 `start` 之前。原有 `asd start -- EXE ARG...` 保持可用。

- 相对模板 cwd 以配置文件目录为基准；不写 cwd 也使用该目录。显式 CLI cwd 以当前调用目录为基准。
- executable 相对路径及 PATH 按最终 cwd／env 解析；Windows 脚本仍需显式 shell，不能把整条命令塞进 executable。
- 不插值 `$HOME`、`${VAR}` 或 `~`；args 是独立字符串，不经 shell 拼接。Windows 路径可用 TOML 单引号字面字符串，例如 `'C:\tools\agent.exe'`。
- CLI 的 name/cwd/env 覆盖模板，重复 env 采用最后值。模板不能同时指定原始 executable。
- `schema_version`、字段名严格区分大小写；未知字段、重复键、错误类型、缺失／空文件、未来版本均失败。每次 start 读取，运行中的会话不随文件变化。
- 最多读取 1 MiB；错误不回显配置正文。配置中如有敏感 env，应自行限制文件访问并留在版本库之外。

配置解析与后端传输是两层边界。psmux 3.4.0/tmux 的已知参数改写形式仍明确拒绝，见 [开发指南](M0_DEVELOPMENT.md)。没有把 TOML 支持等同于任意 argv 都支持。

本次没有保存／迁移／热重载、主题文件或持久 Session 数据库。实际证据见[执行报告](testing/results/2026-11-22-m1-config-tokens.md)。UI 样式开发入口见 [ADR-0008](adr/0008-ui-design-tokens.md)。

2026-11-24 UI 补充：[ADR-0039](adr/0039-launch-dialog-environment-overrides.md) 在 Tab 新建表单提供临时环境覆盖（JSON KEY=value 字符串数组），不读取／写回此模板。空白继承、重复键和平台解析仍沿既有后端合同；取消或重开不保留表单。此次只补启动请求，不新增配置编辑／保存／迁移，实际范围见[报告](testing/results/2026-11-24-launch-environment.md)。
