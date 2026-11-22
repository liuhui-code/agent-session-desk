# 启动模板

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
