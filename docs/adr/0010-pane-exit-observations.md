# ADR-0010：保存 pane 退出观测，保持 agent 结果未知

- 日期：2026-11-22
- 状态：已接受（Accepted），用户授权继续 M1 退出证据切片
- 关联：FR-05、FR-10、FR-12；FL-005、FL-011、FL-013；细化 ADR-0002/0004/0009
- 实现与验证：[退出观测](../../internal/backend/exit.go)、[持久化](../../internal/store/exit.go)、[实际报告](../testing/results/2026-11-22-m1-pane-exit.md)；真实 agent 退出码和完整生命周期仍待做

## 能力边界

固定 psmux 3.4.0 源码 `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 `src/format.rs` 将 `pane_dead_status`、`pane_dead_signal` 和 `pane_dead_time` 返回为字面 `0`，不能用作准确退出码或退出时间。Context7 返回的 master 文档不能证明该稳定发行版已有新能力。

保留直接自定义 CLI 启动，不增加每个会话的包装常驻进程，不为取得退出码改变用户命令。当前只记录产品关联 pane 的退出观测；即使 pane 入口恰为 agent，也不据此推出某条任务已完成。显式 shell 中的子 agent 已退出而 shell 仍在时，终端仍可连接，agent 状态未知。

## 观测合同

后端通过同一次格式查询取得 `pane_id|pane_dead`。仅在 ID 与原记录一致且状态恰为 1，随后又能核对原 Session/Run/pane 关联时，返回 `pane_exit`。错 pane、格式不完整、截断、命令失败或关联变化均降级 unknown，不写退出证据。

观测包含 pane_id、来源 `psmux.pane_dead`／`tmux.pane_dead` 和产品观测时间 observed_at。**observed_at 不是进程退出时刻；exit_code 明确为 null。** 当前两个后端都不提供数值退出码合同。通用终端 API 不升级为 agent-accepted/completed、agent 就绪或原上下文恢复。

这些检查不构成跨后端原子快照、任意外部 respawn 或服务器替换下的完整进程身份保证。原 pane 消失、后台断开和 list 未列出仍是不确定性；stop_confirmed_at 继续只记录显式停止，两者不互相伪造。

## 存储与迁移

SQLite user_version 升至 **2**，application_id 不变。在 ADR-0009 的 v1 表之上追加 pane_exits，按 scope/run_id 一对一保存首次观测，并以复合外键关联 Run。不保存无限事件列表；重复发现核对相同 pane/source 后只读，保留首次时间。冲突或比来源更强的证据（含非 null 退出码）拒绝。

Session/Run 元数据和首次退出观测在同一事务提交。`history` 在外层返回可选 pane_exit，内层 session.terminal_state 仍为 unknown，不把历史观测恢复为当前在线／当前退出状态。不因读取历史自动创建进程。

空库直接初始化 v2。已识别的 v1 库在一个 BEGIN IMMEDIATE 事务中添加表并更新 user_version，保留身份、入库和停止时间；失败回滚，不删库重建。v2 必需字段继续校验；未来版本和外来库仍拒绝。旧 v1 程序会拒绝 v2 库，没有降级迁移；重要 v1 数据升级前应保留副本。完整升级、强杀、掉电和备份工具仍不是已验收能力。

## 验证

测试覆盖错目标／未知状态、拒绝伪造数值码、跨重开保持、首次时间／重复读取、事务冲突和 v1 迁移／失败保留。原生 Windows 用独立 fixture 的退出 0/7 对照产品的 null 退出码，并验证 shell 留存场景和 history。准确范围与失败轮次只以执行报告为准。

## 依据

- [固定 psmux format.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/format.rs)
- [固定 tmux 3.8 手册](https://github.com/tmux/tmux/blob/3.8/tmux.1)
- [ADR-0009](0009-session-run-metadata.md)、[M1 元数据报告](../testing/results/2026-11-22-m1-metadata.md)
