# 状态、竞争与故障窗口验证模型

日期：2026-11-21。状态：测试设计，尚无产品执行结果。依据：[ADR-0004](../adr/0004-cli-agents-and-session-messaging.md)、[Feature Ledger](../FEATURE_LEDGER.md)；具体前提、步骤与证据见下列用例。

本文把“实际发生了什么”“产品能证明什么”“界面允许显示什么”分开，作为实现和故障注入时的判断依据。夹具的旁路日志可揭示缺陷，但不能替产品补上它并未持有的证据。

## 1. 消息状态以证据为准

| 状态或事实 | 最低证据 | 禁止推断 | 关联用例 |
| --- | --- | --- | --- |
| `queued` | 请求、目标与正文已成功持久化 | 不代表已写入或存在消费者 | [SEND-013](generated/03-output-messaging.md#tc-send-013)、[QUEUE-001](generated/03-output-messaging.md#tc-queue-001) |
| `input_written` | 目标通道写入的可核对证据，范围清楚 | 后端返回 0 不代表 agent 理解或任务完成 | [SEND-015](generated/03-output-messaging.md#tc-send-015)、[SEND-017](generated/03-output-messaging.md#tc-send-017) |
| `accepted` | 与正确请求、来源及 Run 关联的接收证据 | 打开界面、普通输出、静默不构成接收 | [SEND-001](generated/03-output-messaging.md#tc-send-001)、[REPLY-008](generated/03-output-messaging.md#tc-reply-008) |
| `completed` | 与请求关联的可信完成证据 | 不等于用户验收通过，也不要求所有中间通知都曾收到 | [REPLY-002](generated/03-output-messaging.md#tc-reply-002)、[REPLY-003](generated/03-output-messaging.md#tc-reply-003) |
| `failed / cancelled` | 明确失败或取消的阶段、范围和原因 | 部分写入后不能宣称全部影响已撤回 | [QUEUE-020](generated/03-output-messaging.md#tc-queue-020) |
| `unknown` | 结果无法确认，保留已有证据与缺失项 | 不能当成功，也不能当“肯定没执行，可以重发” | [SEND-016](generated/03-output-messaging.md#tc-send-016)、[CTL-007](generated/03-output-messaging.md#tc-ctl-007) |
| 暂停投递 | 存在 Run 变化、输入占用等阻断原因 | 暂停不等于删除请求或丢弃来源 | [QUEUE-007](generated/03-output-messaging.md#tc-queue-007)、[QUEUE-011](generated/03-output-messaging.md#tc-queue-011) |

这些状态不是必须逐级经过的总顺序。接收确认丢失，但随后取得可信、匹配请求的完成证据时，可以记录 `completed`；测试应丢弃中间通知验证这一点。反过来，已经持久化的接收/完成证据不能仅因终端断开而被抹掉。运输结果、agent 结果和用户验收应各自保留依据。

核心不变量：**状态不得强于证据；恢复不得弱化或抹掉已获得的事实；请求幂等不等于外部 CLI 执行 exactly-once。**

## 2. 身份与生命周期矩阵

| 变化 | 必须保持或重新核对 | 不允许的结果 | 关联用例 |
| --- | --- | --- | --- |
| 重命名、排序、切换视图 | 稳定 `session_id` 与已绑定目标 | 改投当前 Tab 或同名会话 | [READ-002](generated/03-output-messaging.md#tc-read-002)、[SEND-007](generated/03-output-messaging.md#tc-send-007) |
| 同 Session 从 Run B1 变成 B2 | 消息、游标、回复仍明确原 Run | 把旧正文、ready 或完成证据归给 B2 | [READ-004](generated/03-output-messaging.md#tc-read-004)、[QUEUE-005](generated/03-output-messaging.md#tc-queue-005)、[REPLY-005](generated/03-output-messaging.md#tc-reply-005) |
| 后端 pane/PID 被复用 | 产品 Run 身份和实际后端对象一致 | 仅凭底层编号相同投递旧消息 | [QUEUE-012](generated/03-output-messaging.md#tc-queue-012) |
| 入队时没有活跃 Run | 保留未绑定原因，显式指定目标 | 新 Run 出现就自动接收旧任务 | [QUEUE-010](generated/03-output-messaging.md#tc-queue-010) |
| 来源日志轮转或追加 | 已提交引用保留快照正文与原范围 | 引用内容随最新屏幕漂移 | [READ-008](generated/03-output-messaging.md#tc-read-008)、[READ-010](generated/03-output-messaging.md#tc-read-010) |

Session、Run、消息身份不能从显示时间推导；时钟回拨或相同时间戳不改变归属。见 [SOAK-007](generated/04-resilience-resources-release.md#tc-soak-007)。

## 3. 崩溃与恢复窗口

下表故障均使用命名屏障触发，避免靠随机 sleep 猜测时点。“未写入”必须是产品可证明的事实，不能只由测试控制器知道。

| 故障窗口 | 产品恢复时应核对 | 必须观察到的结果 | 关联用例 |
| --- | --- | --- | --- |
| 请求持久化提交前失败 | 是否已成功建立请求 | 不报告 queued，不先写入目标 | [SEND-013](generated/03-output-messaging.md#tc-send-013) |
| 入队成功、调用者未收到编号 | 幂等键与原请求是否已存在 | 重试找回同一请求；超时不代表未执行 | [SEND-014](generated/03-output-messaging.md#tc-send-014)、[CTL-010](generated/03-output-messaging.md#tc-ctl-010) |
| 入队/领取后、首次写入前崩溃 | 尝试记录、所有权、可证明的未写入依据 | 证据足够才按恢复合同继续；否则明确不确定，不能丢请求 | [QUEUE-018](generated/03-output-messaging.md#tc-queue-018) |
| 正文部分写入后断开 | 实际写入范围与缺失证据 | 不标完整 input_written，不自动重放整段 | [SEND-015](generated/03-output-messaging.md#tc-send-015) |
| 正文写完、回执落盘前崩溃 | 持久化尝试与可恢复证据 | 无充分证据时 unknown/暂停；不能因旧 queued 值再次写出 | [SEND-016](generated/03-output-messaging.md#tc-send-016) |
| input_written 回执已落盘后重启 | 已有写入事实及尚缺的 agent 证据 | 保留已写事实，不因缺少完成通知重新发送 | [QUEUE-019](generated/03-output-messaging.md#tc-queue-019) |
| 接收/完成证据已保存后终端断开 | 已有证据的请求、来源和 Run | 不抹掉历史证据；新连接状态与历史结果分开表达 | [REPLY-002](generated/03-output-messaging.md#tc-reply-002)、[CTL-003](generated/03-output-messaging.md#tc-ctl-003) |

同一窗口重复崩溃与重开，仍须保留原请求及尝试关系；不能换 message_id 隐藏重复投递。循环场景见 [SOAK-005](generated/04-resilience-resources-release.md#tc-soak-005)。

## 4. 并发与事件次序矩阵

| 竞争 | 必测次序或条件 | 判定不变量 | 关联用例 |
| --- | --- | --- | --- |
| ready 与 busy | ready(n) 延迟到 busy(n+1) 之后；重复 ready | 已证明陈旧的事件不覆盖当前状态；重复通知不重复写入 | [QUEUE-006](generated/03-output-messaging.md#tc-queue-006)、[QUEUE-015](generated/03-output-messaging.md#tc-queue-015) |
| ready 与 Run 更换 | B1 ready 在 B2 启动后到达 | 旧 Run 信号不释放新 Run 消息 | [QUEUE-005](generated/03-output-messaging.md#tc-queue-005) |
| attach 与写入 | 用户先占用、调度先检查、占用未知 | 默认等待或明确交接；正文和用户编辑不交错 | [QUEUE-007](generated/03-output-messaging.md#tc-queue-007)、[QUEUE-008](generated/03-output-messaging.md#tc-queue-008)、[QUEUE-009](generated/03-output-messaging.md#tc-queue-009) |
| 两个发送者 | 同目标不同正文；相同幂等键同时提交 | 正文与提交保持完整单元；同一请求只建立一项 | [QUEUE-013](generated/03-output-messaging.md#tc-queue-013)、[SEND-012](generated/03-output-messaging.md#tc-send-012) |
| 两个 Host | 同时启动；H1 暂停、H2 获权后 H1 恢复 | 同目标只有一个有效投递所有者；旧所有者不能恢复写入 | [QUEUE-016](generated/03-output-messaging.md#tc-queue-016)、[QUEUE-017](generated/03-output-messaging.md#tc-queue-017) |
| 取消与写入 | 领取前、领取后写前、部分写后、完整写后 | 取消范围真实，不把可能已执行说成全部撤回 | [QUEUE-020](generated/03-output-messaging.md#tc-queue-020) |
| 显示刷新与控制事件 | 高输出中快速 ready/busy、低频刷新 | 可合并显示帧，不能因此丢控制顺序 | [QUEUE-024](generated/03-output-messaging.md#tc-queue-024)、[PERF-007](generated/04-resilience-resources-release.md#tc-perf-007) |
| 多请求与乱序回复 | 正序、逆序、固定种子交错 | 按请求和来源关联，不按返回顺序猜测 | [REPLY-009](generated/03-output-messaging.md#tc-reply-009) |

并发发送没有天然的墙上时间总顺序。测试核对实现记录的仲裁顺序及每次正文/提交的完整性；公平性、过期、取消、所有权恢复等未决策略由 [BUDGETS_AND_DECISIONS.md](BUDGETS_AND_DECISIONS.md) 关闭后再作确定判定。

## 5. 最小可复现故障调度

1. 使用隔离数据目录和 namespace，固定版本、随机种子、消息正文、Session/Run 和时钟输入；建立无故障基线。
2. 在一个命名窗口设置屏障，记录产品持有的证据，再释放一个写入、状态事件、取消或退出动作。
3. 对竞争分别运行 A→B、B→A 及同屏障释放；每次从同一数据快照开始，不复用上一轮残留状态。
4. 只结束测试 Host/客户端或注入指定接口错误，保持独立 fixture 日志与控制器存活。
5. 重启并查询产品记录，对照独立接收字节、提交次数和回复关联；首次发现错投或非预期重放立即保存现场。
6. 只在单窗口结果可解释后扩大为固定种子的循环与积压场景；数量和资源判据沿用具体用例及已接受预算。

| 观察点 | 用途 |
| --- | --- |
| 产品持久化记录、投递尝试及状态依据 | 判断产品当时可以证明什么，区别已提交与只在内存中的状态 |
| 独立 agent 输入字节与提交事件 | 核对实际目标、正文、次数和是否交错，不与产品日志共用同一成功标志 |
| Host 所有权、进程退出及 attach 交接事件 | 复现并发先后和输入占用，识别旧所有者恢复写入 |
| UI/CLI 返回与工具调用上下文 | 核对展示语义及 agent 真正取得的信息，不能以截图单独证明执行 |

上述屏障、观察点与循环是待实现的测试合同，尚未构成自动化框架。边界数值与未决策略不在本文件中重新决定；未确定条件不得判为通过。

本模型提供设计覆盖与可复现切入点，不声称穷尽所有调度，也不是数学证明。新增故障路径应回填具体用例、证据和受影响 FL，而不能用“已有大量测试”替代验证。
