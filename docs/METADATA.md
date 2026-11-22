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

## pane 退出观测（schema v2）

当产品实际观察到原 pane 退出时，list 中可有 pane_exit；history 在记录外层保留首次 pane_exit（pane_id、source、observed_at、exit_code）。退出码目前为 null；psmux 3.4.0 的 pane_dead_status 常量 0 不是真实退出码，observed_at 也不是进程退出时刻。历史 session.terminal_state 仍为 unknown。

这是终端 pane 的观测，不能证明 shell 内子 agent 的状态、某条任务成功或完整进程树已退出。断开、目标消失和显式 stop 不制造 pane_exit；以前已取得的观测继续保留。详见 [ADR-0010](adr/0010-pane-exit-observations.md)。

v1 数据库打开时事务性追加观测表并升级为 v2，原有身份与时间保留；冲突／失败不删库重建。旧 v1 程序会拒绝 v2，没有降级迁移。重要数据升级前保留副本；完整升级和备份／恢复验收仍待做。
