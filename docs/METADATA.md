# 会话元数据与历史

日期：2026-11-22。本页是 M1 最小切片，完整阶段尚未完成；合同见 [ADR-0009](adr/0009-session-run-metadata.md)。

启动仍使用自定义 CLI 或 TOML 模板。元数据自动记录 Session/Run 身份、名称、入口和后端关联；运行状态始终向实际后端核对。

```powershell
asd --state-dir C:\asd-data start --name 开发 -- 自定义程序.exe 参数
asd --state-dir C:\asd-data list --json
asd --state-dir C:\asd-data history --limit 20
asd --state-dir C:\asd-data stop 完整session_id
```

history 返回 JSON，即使后端程序缺失也可读取。它显示入库记录而非在线证明，terminal_state 为 unknown；stop_confirmed_at 只说明产品曾确认停止了该后端目标。它不提供终端日志、agent 退出码或模型上下文。list 只显示实际后台记录，历史不会自动变成可输入会话。

`--state-dir` 放在子命令之前，指定 metadata.sqlite3 的父目录。默认使用用户配置目录下 agent-session-desk；如果明确传入 `--runtime-dir` 且未设 state-dir，则在该 runtime 的 namespace 子目录存储，方便隔离。查询时使用相同 runtime-dir/namespace；同一个数据库内不同后台作用域互相隔离。目录搬迁和作用域合并尚无迁移功能。

启动报错若已包含 Session/Run ID 且提示终端已创建、元数据未提交，先用相同参数的 list 检查，避免重复 start。成功发现可以补录后台现有记录。数据库故障会明确报错；不要删除或覆盖故障文件来冒充恢复。重要记录应由用户保存在保留目录，当前未实现备份工具。

读取默认最近 100 项，limit 为 1–1000。读取内存和 SQLite 连接有边界，元数据磁盘保留/清理策略和完整资源预算仍待验证。库打开会初始化空库并短暂获取写锁校验 schema；当前不提供只读文件模式。SQLite 文件只用于本机；本次不承诺网络共享目录、完整升级迁移和掉电恢复。
