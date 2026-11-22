# 有界屏幕输出记录

日期：2026-11-22。M1 实现切片，完整阶段状态见[清单](testing/M1_ACCEPTANCE.md)，合同见 [ADR-0013](adr/0013-bounded-screen-output-records.md)。

启动仍使用你自己的命令和工作目录，可写入[显式 TOML 模板](LAUNCH_CONFIG.md)。先从 list 取得完整 session_id，再显式保存画面：

```powershell
asd --config .\agents.toml start --template dev
asd list --json
asd record 完整session_id
asd history
asd output 完整run_id
```

相同 state-dir、runtime-dir 和 namespace 标识同一作用域；全局选项放在子命令之前。output 不依赖后端程序存在，停止后仍可读取保留的记录；它按 run_id 查询，不接受列表序号或名称。record 只捕获后台仍能核对的原 pane，不重启历史 Run，也不发送输入。

结果是 JSON。source=screen_snapshot、complete=false 始终保留；truncated=false 仅表示本次捕获未报告裁剪，不表示完整历史。程序、管理命令和 TUI 未采集的期间可能有缺口。终端输出中的 ANSI 保留为字符串数据并由 JSON 转义，不直接在管理器中执行。保存输出不等于恢复 agent 对话或模型上下文。

每次最多 200 行，文本最多 64 KiB；每 Run 保留最近两份，同一数据库所有作用域共最多 32 份，旧捕获会轮转。文件工程上限与完整资源预算分别计算；元数据保留量、多个 state-dir 和文件系统实际分配不在此数值中。

输出放在 metadata.sqlite3 旁的 output-records。该目录由产品标记管理，请勿在其中放置其他文件或改为 junction／symlink。未知目录、异常文件、链接或不可写错误会明确失败，不停止原 CLI；遇到“已发布但清理失败”，先按 record_id 检查已有记录。暂存不是正式输出；正常读者跳过未发布的暂存，后续有锁写入清理产品保留名字空间的遗留文件。

`metadata backup / restore` 仅处理 SQLite 元数据，不包含这些输出文件、配置或活进程。完整归档／恢复包与连续日志另行实施。原生局部验证及失败记录见[执行报告](testing/results/2026-11-22-m1-output-records.md)。
