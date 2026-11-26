# ADR-0049：原生 Agent 授权／选择界面的待处理提醒

- 日期：2026-11-26
- 状态：已接受；用户确认提醒对象是 Agent 终端中已有的授权／选择界面
- 关联：FR-14、FR-07/09/12；FL-020、FL-008/010/014/019
- 细化 ADR-0006 的第一切片；补充 ADR-0047/0048，不替换原生输入或已有任务归属合同

## 决定

用户优先需要发现 Agent 等待决策，并返回原工作现场。任务列表／取消／unknown 人工处置先登记遗留。保留 Agent 自己的 form；产品没有统一批准表单、批准命令或自动选择器。查看提醒不解决请求；跳转不发送批准、选项、回车或队列正文。

可信合作式来源使用原 Run 的 `agent-bind` 委派文件，通过 `attention-event` 上报 approval／question 的 pending、resolved 或 withdrawn。沿用本地 CLI、现有同用户能力委派与 SQLite，不新增依赖、网络监听器、常驻 worker 或每会话 notifier。文件持有者可代表该 Run 上报，不等于经过操作系统认证的 agent PID；来源身份边界沿用 ADR-0047。普通终端文本、静默与 shell 存活不能建立确定提醒。

事项有独立 attention_id、原 Session/Run 和递增 version；事件身份为 attention_id + version，与事项身份分离。同版本同正文幂等；同版本不同正文拒绝；旧版本忽略并返回最新记录。结束状态保留墓碑，不允许更高版本复活 pending，新的交互须新 ID。先收到结束事件也保存墓碑。新的请求正文清除已查看状态；查看记录与事件状态分开保存。断连不解决事项；明确记录 Run／pane 退出后使其 invalidated。

UI 在已有顶部留白显示待处理数量、对应会话行增加标记；不增加标题描述区、不压缩终端、不主动切焦点。`Alt+N` 或点击提醒优先定位未查看项，全部查看后轮转。按原输入归属先排空已接收写入；等待时新增按键／粘贴不发送。跳转服务重新核对事项版本、pending 和完整原 Run，旧目标／过期结果拒绝，连接仍走后端原身份核验。只持久记录 viewed，保留 pending，选择仍在原生终端输入。断连时标记为最后观察，拒绝凭旧观察直接跳转。

本轮通过 Context7 核对 [Claude Code 官方 hooks](https://code.claude.com/docs/en/hooks) 的 PermissionRequest／Notification。所列 PermissionRequest 输入不含可直接依赖的 tool_use_id，不能据此杜撰稳定交互 ID 和关闭关联；hook 的决定输出还能改变授权，因此本轮不自动安装实际 hook、不输出 allow／deny。先用保留原菜单的受控 CLI 验证通用合同，真实适配另补稳定关联和原生行为证据，不宣称 Claude Code 已接入。

同库事务中，pending 事项阻止向该 Run 的新自动终端投递和新结构化处理领取；已有领取的续期／结果元数据不受影响，普通用户输入照常。此约束不撤回已在途写入，不保证来源尚未上报或不遵守协议时仍能隔离授权菜单。

## 有界存储与迁移

SQLite v9 新增 attention；v1–v8 事务升级，ui.toml 不变。每数据库全部 scope 合计最多 256 个事项（包含墓碑），每事项只有最新事件与查看时间；事件 JSON 最大 16 KiB，标题 256 UTF-8 字节，详情 8 KiB，数据库文档 24 KiB。上限是防御边界，非测得的资源／时延预算。满额明确拒绝新事项，不静默删除墓碑、丢掉未解决事项或假装上报成功；清理／归档后续另定合同。快照包含事项和查看状态，不提供进程存活或外部交互状态恢复保证。

## 第一切片边界

当前没有真实目标 Agent，以原生终端中的可控 CLI 授权菜单和独立产品工具进程验证；通用合作式事件接口不等于任何品牌自动接入。接入者须证明稳定请求 ID、版本和关闭关联，不能凭屏幕文字推断已解决。

经 Context7 查阅的 [Claude Code hooks 官方参考](https://code.claude.com/docs/en/hooks)中，PermissionRequest 输入没有可直接假定为稳定交互 ID 的 tool_use_id；Notification 的 permission_prompt 也不足以证明精确关闭关联。本轮不生成未经验证的 Claude hook、不输出 allow/deny 决策，不声称 Claude 兼容。

桌面 Toast／声音、系统通知激活、稍后提醒／通知关闭、异常退出检查事项、辅助全屏 attach 的可见提醒、UI 退出后的持续监测和真实 Agent 适配仍按 ADR-0006 开放。显式 CLI 生产者可在 UI 不运行时保存事件，下一次打开读回；没有运行的生产者／消费者时不承诺监测。接口见 [ATTENTION.md](../ATTENTION.md)，实际检查单列于[执行报告](../testing/results/2026-11-26-native-decision-attention.md)，不提升完整 FL-020 或发布接受。
