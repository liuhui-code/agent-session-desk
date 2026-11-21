# 状态、竞争与故障窗口验证模型

日期：2026-11-21。状态：测试设计，完整矩阵尚未执行；独立实现测试的局部结果另存于 [results/](results/)，不代表本矩阵整体通过。依据：[ADR-0004](../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0006](../adr/0006-human-attention-and-notifications.md)、[Feature Ledger](../FEATURE_LEDGER.md)；具体前提、步骤与证据见下列用例。

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

## 6. 人工介入、已读与通知是不同事实

FL-020 的设计场景见[人工介入提醒用例](generated/05-attention-notifications.md)。下表使用概念状态描述不变量，不预先规定数据库枚举或外部协议字段。

| 事实 | 所需依据 | 允许的展示或操作 | 不能推断 |
| --- | --- | --- | --- |
| 待批准 / 待回答 | 经验证来源、对应 Session/Run 与请求的可靠事件 | 显示原因、加入待处理入口、按偏好提醒 | 不等于 ready；不得把普通队列正文写入审批输入 |
| 疑似需处理 | 可追溯的启发式匹配及适用规则 | 单独标记疑似，用户可查看 | 不能提升为确定审批、已完成或可自动投递 |
| 状态未知 / 连接中断 | 当前缺乏可靠证据或事件通道断开 | 保留已有事项与依据，标记待确认 | 无输出、CPU 低、shell 存活、丢失连接都不证明问题解决 |
| 已查看 | 用户在产品中明确查看对应事项 | 更新已查看标记，保留待处理原因 | 不等于 Agent 已获得回答或请求结束 |
| 已稍后提醒 / 已隐藏通知 | 用户选择延后或隐藏的记录 | 改变提醒调度或呈现 | 不改变 Agent 请求生命周期；已隐藏不是已批准 |
| 已解决 / 已撤销 | 来源可信、与该请求匹配且时序有效的处理或撤销证据 | 结束对应待处理事项，保留依据 | 通用 busy、另一请求的回复、普通输出不能替代该请求证据 |
| 旧审批 / 提问失效 | 已确认对应 Run 退出或被替换 | 标记失效；旧通知只可定位历史 | 不能把答案投向新 Run，不能声称原请求已获批准 |
| 异常待检查 | 已确认的异常事实，关联相应 Run | 即使 Run 已结束也保留独立待检查事项 | 不因原 Run 退出而自动清除此事项 |
| 用户标记已检查 | 明确针对待检查事项的人工操作 | 结束该检查事项，保留人工处理来源 | 不伪造 Agent 接收、批准或任务成功证据 |
| 通知出口已接受 | 可核对的通知接口返回与尝试记录 | 记录发送阶段；应用内事项仍按自身状态存在 | 不证明横幅可见、用户已读、请求已解决 |

请求身份与事件身份分开：同一请求的重复投递或更新不新建多个待办；不同请求即使标题、正文和时间相同，仍独立处理。请求与显示名称、Tab 顺序、墙上时间无关。事件乱序、同版本冲突、快照完整性和重连规则由 `TBD-ATTENTION-EVENTS` 固定；不能用“最后收到的事件”覆盖已有更强证据。

## 7. 人工介入提醒的竞争与恢复窗口

| 竞争或故障 | 必测时序 | 保持的不变量 |
| --- | --- | --- |
| 查看与解决 | 先查看后解决、先解决后迟到查看、只关闭 OS 横幅 | 已读维度不修改请求事实；迟到查看不复活已结束事项 |
| 创建与解决事件乱序 | 解决已提交后收到旧创建；重复旧事件 | 保留版本/请求关联，不产生新的待办或提醒风暴 |
| 同文不同请求 | 同 Run 两个独立请求；处理其中一个 | 不按正文去重；处理一个不清除另一个 |
| 点击与 Run 退出/替换 | `F-ATTENTION-ACTIVATE` 后退出或启动新 Run | 最终重新核对，只定位有效事项或历史；不批准、提交按键或启动 Agent |
| 退出与异常记录 | Run 结束使原审批失效，同时形成异常待检查事项 | 两种事项分别保留语义；人工“已检查”不提升 Agent 执行状态 |
| attach 与另一会话提醒 | A 正在编辑，B 请求批准；同时 resize/IME 输入 | A 的输入、焦点、草稿保持；提醒不插入子终端文本，不自动跳转 |
| 提醒与消息队列 | B 在审批菜单，待发送普通任务同时就绪 | 人工介入状态不充当 ready；默认继续阻止普通任务注入 |
| 事项提交后崩溃 | `F-ATTENTION-AFTER-COMMIT` 处结束测试客户端或 Host | 重开能查到同一持久事项；事件重送不再建立一份 |
| 通知已发送但结果未记 | `F-NOTIFY-AFTER-SHOW` 后崩溃/丢回执 | 应用待办不丢；发送结果保留不确定性，不保证 OS 横幅 exactly-once，不无限重试 |
| 断连与快照对账 | `F-ATTENTION-RECONCILE` 时收到不完整/旧快照 | 不将快照中未出现的事项直接认定已解决，不复活旧 Run 请求 |
| 多 TUI/Host 与重复激活 | 同作用域两接收者；旧所有者恢复；通知双击 | 按有效所有权协调通知，避免重复通知风暴；激活不导致多个 Agent Run |
| UI 退出与消费者停止 | UI 存活但 attach、UI 退出 Host 存活、两者皆停分别测试 | 每种模式有明确事件接收者；无消费者时不声称持续监测，恢复只按实际证据对账 |
| 系统通知失败或被静音 | 禁用通知、免打扰、通知接口超时/拒绝 | 应用待处理记录可查询，通知传递状态不改变 Agent 请求状态 |
| 稍后提醒与时钟变化 | 延后后处理请求；延后期间重启/时钟回拨 | 已解决事项不再次作为待办提醒；具体时间边界按已定策略，不自行发明时长 |

通知发送、Agent 输入和用户验收分别观察。通知恢复可能出现无法证明是否已展示的窗口；此时不能为追求“必达”无界重发，也不能为了去重丢失应用内事项。重试、节流、人工隐藏、保留期限及容量边界见 `TBD-ATTENTION-POLICY`，Windows 注册/激活与终端铃声能力见 `TBD-WINDOWS-NOTIFY`。

人工介入的属性轨迹可组合重复、乱序、重命名、切换、已读、稍后提醒、断连、解决和 Run 替换，逐步与独立请求模型对照。状态空间和负载由具体用例记录；这仍是待执行的测试设计，不构成产品正确性证明。
