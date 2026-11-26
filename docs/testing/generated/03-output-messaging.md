# 输出、投递、队列、控制 CLI 与回复

> 从 [JSON 源](../cases/03-output-messaging.json) 自动生成。请修改源并运行 `python tools/test_catalog.py render`。

**全部为测试设计，当前未执行产品或 Windows 测试。** ready 仅表示预期依据明确。

[总索引](../CASE_INDEX.md) · [覆盖表](../COVERAGE.md) · [字段约定](../CASE_FORMAT.md)

| ID | 场景 | 优先级 | 阶段 | 规范状态 |
| --- | --- | --- | --- | --- |
| [TC-READ-001](#tc-read-001) | 读取指定输出同时保留完整来源且不写入终端 | P1 | M3 | 依据明确 |
| [TC-READ-002](#tc-read-002) | 相似名称与重命名后仍按稳定身份读取 | P0 | M3 | 依据明确 |
| [TC-READ-003](#tc-read-003) | 其他会话的游标不得串读 | P0 | M3 | 依据明确 |
| [TC-READ-004](#tc-read-004) | 旧 Run 游标不能无提示跨入新 Run | P0 | M3 | 依据明确 |
| [TC-READ-005](#tc-read-005) | 无输出与范围末尾的空结果可以区分故障 | P1 | M3 | 依据明确 |
| [TC-READ-006](#tc-read-006) | 固定历史按页读取无无声遗漏或重复 | P1 | M3 | 依据明确 |
| [TC-READ-007](#tc-read-007) | 读取时新增输出的分页语义明确 | P1 | M3 | 待决策 |
| [TC-READ-008](#tc-read-008) | 历史轮转后旧游标返回可见缺口 | P1 | M3 | 依据明确 |
| [TC-READ-009](#tc-read-009) | TUI 和消费者关闭期间的未采集区间不伪造完整日志 | P1 | M3 | 依据明确 |
| [TC-READ-010](#tc-read-010) | 引用快照在源内容追加或清理后保持提交时正文 | P0 | M3 | 依据明确 |
| [TC-READ-011](#tc-read-011) | 原始 VT、屏幕快照与结构化历史不混称 | P1 | M3 | 依据明确 |
| [TC-READ-012](#tc-read-012) | 单条超长输出按字节预算限制而非仅按行数 | P1 | M3 | 待决策 |
| [TC-READ-013](#tc-read-013) | 分块边界上的 UTF-8 与组合字符不损坏正文 | P1 | M3 | 依据明确 |
| [TC-READ-014](#tc-read-014) | 畸形与越界游标不触发任意文件访问 | P0 | M3 | 依据明确 |
| [TC-READ-015](#tc-read-015) | 读取过程中来源退出仍保留真实 Run 与边界 | P0 | M3 | 依据明确 |
| [TC-SEND-001](#tc-send-001) | 未知 CLI 显式投递只升级到可证明状态 | P0 | M3 | 依据明确 |
| [TC-SEND-002](#tc-send-002) | 仅粘贴不额外发送提交键 | P1 | M3 | 依据明确 |
| [TC-SEND-003](#tc-send-003) | 粘贴并提交使用目标已声明的提交方式 | P1 | M3 | 依据明确 |
| [TC-SEND-004](#tc-send-004) | 中文多行与尾随换行的传输符合明确文本合同 | P1 | M3 | 待决策 |
| [TC-SEND-005](#tc-send-005) | Shell 元字符作为正文传输而不在管理器侧执行 | P0 | M3 | 依据明确 |
| [TC-SEND-006](#tc-send-006) | 并发临时文件和命名缓冲区不交叉污染 | P0 | M3 | 依据明确 |
| [TC-SEND-007](#tc-send-007) | 提交请求后切换 UI 不改变已绑定目标 | P0 | M3 | 依据明确 |
| [TC-SEND-008](#tc-send-008) | 空白正文与终端控制字节使用明确验证规则 | P1 | M3 | 待决策 |
| [TC-SEND-009](#tc-send-009) | 消息正文上限按编码字节验收 | P1 | M3 | 待决策 |
| [TC-SEND-010](#tc-send-010) | 相同幂等键与相同载荷重试只建立一个请求 | P0 | M3 | 依据明确 |
| [TC-SEND-011](#tc-send-011) | 相同幂等键但目标或载荷不同不得静默复用成功 | P0 | M3 | 待决策 |
| [TC-SEND-012](#tc-send-012) | 多个进程同时重试相同幂等请求仍只建立一项 | P0 | M3 | 依据明确 |
| [TC-SEND-013](#tc-send-013) | 入队持久化失败时不先写入目标 | P0 | M3 | 依据明确 |
| [TC-SEND-014](#tc-send-014) | 入队已提交但返回给发送者前断开仍可幂等找回 | P0 | M3 | 依据明确 |
| [TC-SEND-015](#tc-send-015) | 部分写入后通道失败不标为完整输入写入 | P0 | M3 | 依据明确 |
| [TC-SEND-016](#tc-send-016) | Host 写出后回执落盘前崩溃恢复为不确定而非自动重放 | P0 | M3 | 依据明确 |
| [TC-SEND-017](#tc-send-017) | 后端命令返回成功不等于 agent 接收与完成 | P0 | M3 | 依据明确 |
| [TC-SEND-018](#tc-send-018) | 消息入队后原指令文件变化不替换已保存正文 | P0 | M3 | 依据明确 |
| [TC-QUEUE-001](#tc-queue-001) | 无消费者时消息落盘但不会自动推进 | P0 | M3 | 依据明确 |
| [TC-QUEUE-002](#tc-queue-002) | busy 到 ready 后由共享 Host 在 UI 关闭时投递 | P0 | M3 | 依据明确 |
| [TC-QUEUE-003](#tc-queue-003) | 输出静默不能把未知 CLI 变成 ready | P0 | M3 | 依据明确 |
| [TC-QUEUE-004](#tc-queue-004) | approval 状态不被普通消息注入当作 ready | P0 | M3 | 依据明确 |
| [TC-QUEUE-005](#tc-queue-005) | 旧 Run 的 ready 信号不能释放新 Run 的消息 | P0 | M3 | 依据明确 |
| [TC-QUEUE-006](#tc-queue-006) | 已知较旧的 ready 事件不能覆盖更新的 busy | P0 | M3 | 依据明确 |
| [TC-QUEUE-007](#tc-queue-007) | attach 用户占用输入时默认暂缓自动投递 | P0 | M3 | 依据明确 |
| [TC-QUEUE-008](#tc-queue-008) | 无法确认外部 attach 占用时保持等待 | P0 | M3 | 依据明确 |
| [TC-QUEUE-009](#tc-queue-009) | 投递检查与用户取得输入之间的竞争不能交错正文 | P0 | M3 | 依据明确 |
| [TC-QUEUE-010](#tc-queue-010) | 无活跃 Run 时入队保持等待直到显式指定 | P0 | M3 | 依据明确 |
| [TC-QUEUE-011](#tc-queue-011) | 目标重启后旧队列暂停且原因可见 | P0 | M3 | 依据明确 |
| [TC-QUEUE-012](#tc-queue-012) | 后端 pane 或 PID 复用不能绕过逻辑 Run 校验 | P0 | M3 | 依据明确 |
| [TC-QUEUE-013](#tc-queue-013) | 多发送者正文与提交按同一投递顺序串行完成 | P0 | M3 | 依据明确 |
| [TC-QUEUE-014](#tc-queue-014) | 一个目标 busy 不应无故阻塞其他可投递目标 | P1 | M3 | 依据明确 |
| [TC-QUEUE-015](#tc-queue-015) | 重复 ready 通知不重复消费同一消息 | P0 | M3 | 依据明确 |
| [TC-QUEUE-016](#tc-queue-016) | 同时启动两个 Host 不产生两个有效消费者 | P0 | M3 | 依据明确 |
| [TC-QUEUE-017](#tc-queue-017) | 旧 Host 暂停后恢复不能与新所有者同时写入 | P0 | M3 | 待决策 |
| [TC-QUEUE-018](#tc-queue-018) | Host 在领取后首次写入前崩溃时根据证据恢复 | P0 | M3 | 依据明确 |
| [TC-QUEUE-019](#tc-queue-019) | 已持久化 input_written 的请求在 Host 重启后不重放 | P0 | M3 | 依据明确 |
| [TC-QUEUE-020](#tc-queue-020) | 取消与领取或写入竞争具有明确边界 | P0 | M3 | 待决策 |
| [TC-QUEUE-021](#tc-queue-021) | 就绪信号源断开后不沿用过时 ready | P0 | M3 | 依据明确 |
| [TC-QUEUE-022](#tc-queue-022) | 队列容量上下界与拒绝行为可核对 | P1 | M3 | 待决策 |
| [TC-QUEUE-023](#tc-queue-023) | 消息有效期与系统时钟变化按已定策略处理 | P1 | M3 | 待决策 |
| [TC-QUEUE-024](#tc-queue-024) | 快速 ready/busy 事件不会被绘制合并误当作持续可写 | P0 | M3 | 依据明确 |
| [TC-QUEUE-025](#tc-queue-025) | Host 正常关闭时保留已接受队列与未完成尝试 | P0 | M3 | 依据明确 |
| [TC-CTL-001](#tc-ctl-001) | 具备工具能力的 agent 实际调用 list/read/send/status | P0 | M3 | 依据明确 |
| [TC-CTL-002](#tc-ctl-002) | 列表返回稳定身份及可证明状态 | P1 | M3 | 依据明确 |
| [TC-CTL-003](#tc-ctl-003) | 消息状态查询不把等待时长当作完成 | P0 | M3 | 依据明确 |
| [TC-CTL-004](#tc-ctl-004) | 无效参数与命令错误具有稳定机器可读合同 | P1 | M3 | 待决策 |
| [TC-CTL-005](#tc-ctl-005) | 不存在的 Session/Run/消息 ID 不回落到当前会话 | P0 | M3 | 依据明确 |
| [TC-CTL-006](#tc-ctl-006) | 只读命令不改变 agent 输入或启动新 Run | P1 | M3 | 依据明确 |
| [TC-CTL-007](#tc-ctl-007) | IPC 或存储不可达时错误不伪装为业务成功 | P0 | M3 | 依据明确 |
| [TC-CTL-008](#tc-ctl-008) | 结构化输出中中文及换行不会破坏记录边界 | P1 | M3 | 待决策 |
| [TC-CTL-009](#tc-ctl-009) | 指令文件路径含空格中文及文件不可读时行为明确 | P1 | M3 | 依据明确 |
| [TC-CTL-010](#tc-ctl-010) | CLI 超时重试可以找回同一请求而非重复建立 | P0 | M3 | 依据明确 |
| [TC-CTL-011](#tc-ctl-011) | 普通终端文字不被管理器当成 agent 工具调用 | P1 | M3 | 依据明确 |
| [TC-REPLY-001](#tc-reply-001) | A 请求 B 后按 message_id 得到真实关联回复 | P0 | M3 | 依据明确 |
| [TC-REPLY-002](#tc-reply-002) | 回复到达、完成声明和用户验收分别表达 | P0 | M3 | 依据明确 |
| [TC-REPLY-003](#tc-reply-003) | 错误 Run 的回复不能完成当前请求 | P0 | M3 | 依据明确 |
| [TC-REPLY-004](#tc-reply-004) | 不存在的 reply_to 不误挂到最近请求 | P0 | M3 | 依据明确 |
| [TC-REPLY-005](#tc-reply-005) | 延迟旧 Run 回复不影响同 Session 的新运行 | P0 | M3 | 依据明确 |
| [TC-REPLY-006](#tc-reply-006) | 重复回复的去重语义由稳定回复标识决定 | P1 | M3 | 待决策 |
| [TC-REPLY-007](#tc-reply-007) | 多条进度与最终回复的聚合规则明确 | P1 | M3 | 待决策 |
| [TC-REPLY-008](#tc-reply-008) | 终端出现 done 或完成样式文本不能伪造关联回执 | P0 | M3 | 依据明确 |
| [TC-REPLY-009](#tc-reply-009) | 并发请求乱序返回仍逐项正确关联 | P0 | M3 | 依据明确 |
| [TC-REPLY-010](#tc-reply-010) | 相同消息编号的错误来源会话不能冒充回复方 | P0 | M3 | 待决策 |

### TC-READ-001

**读取指定输出同时保留完整来源且不写入终端**

- 功能：[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0045](../../adr/0045-explicit-message-outbox-and-owned-paste.md)
- 优先级 / 阶段 / 层次：P1 / M3 / windows_e2e
- 类型：functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- Session A/Run A1 的 fixture 已输出带递增序号的 20 行，另一个 Session B 输出不同标记
- 读取前保存 A、B 的输入字节计数

#### 操作步骤

1. 通过管理 CLI 读取 A 的指定范围
2. 记录响应及输入计数变化

#### 预期与不变量

- 响应携带 A、A1、快照时间、范围/游标与对应正文
- 无 B 的内容；A、B 的输入计数均不增加

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-002

**相似名称与重命名后仍按稳定身份读取**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B 使用符合当前命名规则的相似但不同名称，session_id 不同，分别输出 A-only、B-only

#### 操作步骤

1. 使用 A 的 session_id 读取
2. 重命名并重排 A、B 后用同一 ID 再读

#### 预期与不变量

- 两次均读取 A 的内容和对应 Run
- 名称和列表位置不参与消息或读取寻址

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-003

**其他会话的游标不得串读**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / contract
- 类型：negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B 均有输出，取得 A 的有效游标 C-A

#### 操作步骤

1. 对 B 发起携带 C-A 的读取请求

#### 预期与不变量

- 拒绝不匹配游标或返回明确的范围不匹配结果
- 不得把 C-A 的偏移静默解释为 B 的位置并声称连续读取

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-004

**旧 Run 游标不能无提示跨入新 Run**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0045](../../adr/0045-explicit-message-outbox-and-owned-paste.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A1 已产生游标 C1，A 已重新启动为 A2 且输出明显不同

#### 操作步骤

1. 用 C1 请求读取 A 的后续内容

#### 预期与不变量

- 结果明确指出旧 Run/当前 Run 不匹配，或只返回明确标注为 A1 的历史
- 不能把 A1 与 A2 内容拼接成连续的同一次运行

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-005

**无输出与范围末尾的空结果可以区分故障**

- 功能：[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A1 启动后不输出，B1 已输出且游标位于可读末尾

#### 操作步骤

1. 分别读取 A1 与 B1 末尾之后的内容
2. 随后让 A1 输出一行再读

#### 预期与不变量

- 空结果仍携带明确来源和当前可读范围，不伪造错误正文或其他会话缓存
- 新增一行可被后续读取，旧空结果不被当成终端退出

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-006

**固定历史按页读取无无声遗漏或重复**

- 功能：[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / contract
- 类型：functional、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A1 固定输出 37 条唯一序号后暂停输出；测试页大小设为 10

#### 操作步骤

1. 从首个有效游标开始按返回的下一游标读取至末尾
2. 按序号核对拼接结果

#### 预期与不变量

- 37 条可读记录完整且每条出现一次
- 最后不足一页时边界正确，没有死循环或游标倒退

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-007

**读取时新增输出的分页语义明确**

- 功能：[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已确定 live-tail 或固定快照分页合同
- A1 的序号输出可由屏障暂停，首次读取返回游标后再新增数据

#### 操作步骤

1. 读取第一页后释放新增 11 条记录
2. 继续分页并重复读取边界页

#### 预期与不变量

- 实现符合已记录的快照或 live-tail 合同，说明本次覆盖的截止范围
- 不得把并发追加误报为先前快照已经包含，也不静默漏掉仍宣称在范围内的数据

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-008

**历史轮转后旧游标返回可见缺口**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：recovery、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 按已固定的日志保留规则使 C-old 所指记录被轮转清理，保留更新的尾部

#### 操作步骤

1. 用 C-old 读取后续
2. 查看 CLI 与 TUI 的范围说明

#### 预期与不变量

- 返回明确的已丢失/已轮转边界与实际可读起点
- 不声称从 C-old 到最新内容连续完整；保留部分来源正确

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-009

**TUI 和消费者关闭期间的未采集区间不伪造完整日志**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 终端后台存活；后端保留窗口小于本次输出；关闭所有产品消费者

#### 操作步骤

1. fixture 在关闭期间输出超过后端保留窗口的序号数据
2. 重开管理器并读取该时段

#### 预期与不变量

- 返回实际保留内容并标记未采集/不可恢复范围
- 不把当前屏幕或尾部标记为该时段完整对话

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-010

**引用快照在源内容追加或清理后保持提交时正文**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已从 A1 读取片段 X，发送表单保留 X 的来源和内容

#### 操作步骤

1. 以 X 为引用创建发送至 B1 的请求
2. 让 A 输出新内容并清理或轮转 X 的原日志
3. 查看消息记录并投递

#### 预期与不变量

- 消息和实际投递引用仍为提交时的 X，不从新屏幕动态替换
- 来源仍指向原 A1 和范围；原日志清理不使已保存的引用正文丢失

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-011

**原始 VT、屏幕快照与结构化历史不混称**

- 功能：[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / windows_e2e
- 类型：functional、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- FX-VT 用回车覆盖同一行并切换屏幕，未声明结构化历史能力

#### 操作步骤

1. 分别使用可用的原始记录与屏幕读取入口
2. 查看读取类型及界面标签

#### 预期与不变量

- 返回内容明确表示原始流或屏幕状态，不伪装成逐条聊天消息
- 无适配器时不宣称得到模型完整上下文

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-012

**单条超长输出按字节预算限制而非仅按行数**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：boundary、performance
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 读响应预算 B_READ_BYTES、截断/分页合同已固定
- fixture 生成不含换行的可识别长记录

#### 操作步骤

1. 分别读取长度 B_READ_BYTES-1、B_READ_BYTES、B_READ_BYTES+1 的记录

#### 预期与不变量

- 响应与缓存遵守已批准的预算策略，超出部分以续游标或明确截断表示
- 不因只有一行而无界分配，也不静默丢尾后声称完整

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

#### 参数变体

- payload_bytes：`["B_READ_BYTES-1", "B_READ_BYTES", "B_READ_BYTES+1"]`


### TC-READ-013

**分块边界上的 UTF-8 与组合字符不损坏正文**

- 功能：[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / contract
- 类型：boundary、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)

#### 前置条件

- 已固定字节游标和文本解码合同；数据包含中文、emoji 与 e 加组合重音

#### 操作步骤

1. 使后端每次输出在多字节字符中间分块
2. 跨页读取并拼接正文，与 fixture 原始数据比较

#### 预期与不变量

- 合法 UTF-8 不因分块出现替换字符、重复或丢字
- 游标不会把同一个多字节序列的一部分误标为已消费完整字符

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-READ-014

**畸形与越界游标不触发任意文件访问**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / contract
- 类型：negative、boundary、security
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 建立含已知内容的隔离数据目录和不属于日志的 sentinel 文件

#### 操作步骤

1. 传入空/截断游标、负数偏移、超大整数、路径片段与其他 Run 标记
2. 监测访问记录和返回内容

#### 预期与不变量

- 无崩溃、整数回绕或越权解析到 sentinel
- 返回明确无效/超出范围结果，不随机读取其他文件或会话

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

#### 参数变体

- cursor_variant：`["空/截断编码", "负数", "超大整数", "../sentinel", "其他 Run 标记"]`


### TC-READ-015

**读取过程中来源退出仍保留真实 Run 与边界**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A1 输出有限数据；读取通过屏障停在捕获与返回之间

#### 操作步骤

1. 在屏障处结束 A1 并启动 A2
2. 释放读取并查看结果

#### 预期与不变量

- 只能返回可证明属于 A1 的内容或明确失败/部分结果
- 不得把 A2 的来源或正文拼进已开始的 A1 读取

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-001

**未知 CLI 显式投递只升级到可证明状态**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0045](../../adr/0045-explicit-message-outbox-and-owned-paste.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- FX-ECHO 无 ready、accepted、completed 能力且保持存活

#### 操作步骤

1. 创建指令但不执行显式投递
2. 检查队列与输入日志；再执行明确投递

#### 预期与不变量

- 显式动作前无输入写入，消息保持 queued/等待
- 成功写入后最多为 input_written；没有证据不能显示 accepted 或 completed

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-002

**仅粘贴不额外发送提交键**

- 功能：[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / windows_e2e
- 类型：functional、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- FX-ECHO 区分文本字节和提交事件，已固定 paste-only 合同
- 目标无用户输入占用

#### 操作步骤

1. 显式选择仅粘贴并发送一段不含换行的文本

#### 预期与不变量

- 文本按合同完整出现
- 提交事件为零，不自动附加 Enter 或其他提交序列

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-003

**粘贴并提交使用目标已声明的提交方式**

- 功能：[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / windows_e2e
- 类型：functional、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- FX-READY 声明提交方式；每次有效提交记录独立事件
- 确认输入控制已交接

#### 操作步骤

1. 发送包含多行正文的请求并选择提交
2. 核对正文和提交事件

#### 预期与不变量

- 正文在提交前完整到达，提交次数符合一次请求的合同
- 不得用猜测的 Enter 把不支持的目标误当作已接收

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-004

**中文多行与尾随换行的传输符合明确文本合同**

- 功能：[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / windows_e2e
- 类型：boundary、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-009](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- newline、尾随换行及 bracketed paste 处理规则已固定
- 正文为中文/英文混排、空行、emoji、末尾换行的 UTF-8 文件

#### 操作步骤

1. 分别选择仅粘贴与粘贴并提交
2. 对照文件和 fixture 接收的正文/提交事件

#### 预期与不变量

- 正文按已记录的换行合同处理，不丢空行、中文或尾部内容
- 尾随换行不会产生合同之外的隐式提交或多轮任务

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

#### 参数变体

- line_ending：`["LF", "CRLF", "末尾无换行", "末尾一个换行", "末尾多个换行"]`


### TC-SEND-005

**Shell 元字符作为正文传输而不在管理器侧执行**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / windows_e2e
- 类型：negative、security、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 隔离临时目录中有 sentinel；目标为只记录输入的 FX-ECHO
- 正文包含单/双引号、反引号、$()、&、|、;、%、反斜线与 PowerShell 字符串

#### 操作步骤

1. 通过文件/缓冲区发送正文
2. 核对输入记录、sentinel 和管理器子进程树

#### 预期与不变量

- 目标按约定收到原文；管理器没有把正文解释为 shell 命令
- sentinel 未变化且没有由正文诱发的额外命令进程

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-006

**并发临时文件和命名缓冲区不交叉污染**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 两个发送者给不同 Session 各发送带唯一标记的长正文
- 后端缓冲区准备可设置屏障

#### 操作步骤

1. 同时创建缓冲内容，交叉释放准备与投递屏障
2. 核对每个目标及投递后临时资源

#### 预期与不变量

- 每个目标只收到自己的完整正文，引用缓冲区不会被另一发送覆盖
- 已不再使用的临时资源按合同清理，不删除另一进行中请求的内容

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-007

**提交请求后切换 UI 不改变已绑定目标**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：state、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 表单明确选择 Session B/Run B1，另有 C1

#### 操作步骤

1. 提交并在后端写入前快速切换列表至 C
2. 释放投递屏障

#### 预期与不变量

- 消息记录与输入均仍指向 B1，C1 输入不增加
- 不能从当前选中 Tab 临时重新解析目标

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-008

**空白正文与终端控制字节使用明确验证规则**

- 功能：[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：boundary、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-MSG-VALIDATION](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已决定允许的空正文、空白正文、NUL/ESC 等字符及拒绝提示

#### 操作步骤

1. 依次尝试空文件、仅空白、仅换行、嵌入 NUL/ESC 的内容
2. 核对是否入队和目标实际收到内容

#### 预期与不变量

- 允许者按已公布规则传输，拒绝者在写入前明确报错
- 不能静默去除控制字节后把不同正文当作同一请求，或额外触发未请求的终端操作

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

#### 参数变体

- body_variant：`["空", "仅空格", "仅换行", "嵌入 NUL", "嵌入 ESC"]`


### TC-SEND-009

**消息正文上限按编码字节验收**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 消息预算 B_MSG_BYTES 和超过上限处理已固定
- 正文同时包括 ASCII 与多字节中文样本

#### 操作步骤

1. 分别提交 B_MSG_BYTES-1、B_MSG_BYTES、B_MSG_BYTES+1 字节内容

#### 预期与不变量

- 边界接受/拒绝符合已记录合同，不用字符数量代替实际内存/传输字节
- 超限请求不产生被截断却标为完整的投递

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

#### 参数变体

- payload_bytes：`["B_MSG_BYTES-1", "B_MSG_BYTES", "B_MSG_BYTES+1"]`


### TC-SEND-010

**相同幂等键与相同载荷重试只建立一个请求**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0045](../../adr/0045-explicit-message-outbox-and-owned-paste.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：functional、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 发送者固定幂等键 K、目标 B1 和正文 X；队列暂不消费

#### 操作步骤

1. 连续提交 K/X 三次
2. 查看持久化收件项和各次返回 ID

#### 预期与不变量

- 只有一个逻辑请求及一个 message_id，重试可定位同一请求
- 这只验证管理器请求去重，不声称任意 agent 端 exactly-once

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-011

**相同幂等键但目标或载荷不同不得静默复用成功**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-MSG-IDEMPOTENCY](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 幂等键作用域、载荷指纹及冲突响应合同已固定
- K 已对应 B1/正文 X

#### 操作步骤

1. 用 K 发送正文 Y，再用 K 改为目标 C1
2. 查询原请求和两目标输入

#### 预期与不变量

- 原请求的目标与正文不被悄悄覆盖
- 冲突结果可识别，不把 X 的成功回执冒充为 Y 或 C1 已执行

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-012

**多个进程同时重试相同幂等请求仍只建立一项**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 16 个独立管理 CLI 进程使用同一作用域的 K、相同目标和正文
- 所有进程在建立请求前等待同一屏障

#### 操作步骤

1. 同时释放屏障
2. 等待返回并查询数据库/管理 API

#### 预期与不变量

- 唯一逻辑收件项；返回同一请求或可识别可重试的竞争结果
- 不得依赖单个 TUI 内存锁实现去重

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

#### 参数变体

- concurrent_senders：`[2, 16]`


### TC-SEND-013

**入队持久化失败时不先写入目标**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0045](../../adr/0045-explicit-message-outbox-and-owned-paste.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 通过故障点使创建请求的持久化提交失败，目标可正常接收输入

#### 操作步骤

1. 提交一条新请求
2. 恢复存储后查询请求和目标输入日志

#### 预期与不变量

- 不返回 queued/成功保存的假回执
- 目标输入没有该正文；错误可定位，恢复后不会出现未记录的隐藏投递

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-014

**入队已提交但返回给发送者前断开仍可幂等找回**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 故障点位于请求持久化成功之后、返回 message_id 之前
- 使用固定幂等键 K

#### 操作步骤

1. 中断返回通道使调用者超时
2. 重试同一 K 与正文并查询状态

#### 预期与不变量

- 可定位原请求而不再建立第二项
- 调用超时本身不被解释为目标未执行或允许无条件重放

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-015

**部分写入后通道失败不标为完整输入写入**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0045](../../adr/0045-explicit-message-outbox-and-owned-paste.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 目标记录每个实际收到的字节；传输可在正文中途失败

#### 操作步骤

1. 发送有唯一前后标记的正文，在中途关闭写入通道
2. 恢复连接并查询状态但不执行人工重发

#### 预期与不变量

- 不报告完整 input_written/accepted/completed
- 保留部分写入或 unknown 的可核对证据；不自动重新发送整段造成重复前缀

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

#### 参数变体

- fault_position：`["首字节后", "UTF-8 分块边界", "最后一个正文块前"]`


### TC-SEND-016

**Host 写出后回执落盘前崩溃恢复为不确定而非自动重放**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0045](../../adr/0045-explicit-message-outbox-and-owned-paste.md)、[ADR-0048](../../adr/0048-exclusive-processing-claims-and-uncertain-recovery.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- Host 已保存投递尝试；故障点在目标收到完整正文后、持久化写入回执前
- fixture 输入日志独立于 Host 记录

#### 操作步骤

1. 在故障点强制结束 Host
2. 重启 Host 和 TUI，观察队列与输入日志

#### 预期与不变量

- 记录该尝试结果 unknown 或同等明确不确定状态
- 不因数据库仍有旧 queued/进行中值而自动再写；不会伪造 accepted/completed

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果

说明： ADR-0048 当前为显式单次领取、私有领取文件和结果封存；unknown 不自动接管。独立处理进程与 SQLite 局部证据见 2026-11-26-task-claims，不代替共享 Host／自动消费／真实 agent 全场景。


### TC-SEND-017

**后端命令返回成功不等于 agent 接收与完成**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 后端成功写入但 FX-ECHO 不发任何接受/完成事件

#### 操作步骤

1. 执行发送并等待多个状态轮询周期

#### 预期与不变量

- 最多显示有证据的 input_written
- 不能用退出码 0、输出出现或等待时间推定 accepted/completed

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-SEND-018

**消息入队后原指令文件变化不替换已保存正文**

- 功能：[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：state、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 提交时 instruction 文件内容为 X，目标暂不可投递

#### 操作步骤

1. 通过该文件建立请求后将文件改成 Y 并删除
2. 解除投递阻断

#### 预期与不变量

- 若请求成功入队，其已保存正文仍为 X
- 不能等到消费时再读取同一路径得到 Y 或声称已保存但正文缺失

#### 需要的证据

- 产品提交、Windows/终端/后端/fixture 版本与测试参数
- 操作事件和持久化记录前后快照
- 独立 fixture 输入日志及可核对的输出/状态结果


### TC-QUEUE-001

**无消费者时消息落盘但不会自动推进**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 终端 Run B1 保持运行；没有产品 Host 或 TUI 消费者

#### 操作步骤

1. 通过管理 CLI 建立待投递请求并退出 CLI
2. 让 B1 进入 ready，再打开 TUI 查看消息

#### 预期与不变量

- 无消费者期间请求仍保留且目标未收到自动输入
- 界面不能声称关闭期间已投递；启用消费者后才按当前条件处理

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-002

**busy 到 ready 后由共享 Host 在 UI 关闭时投递**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / windows_e2e
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- M3 共享 Host 运行；B1 发出可信 busy；TUI 已关闭

#### 操作步骤

1. 创建消息，确认 busy 期间无写入
2. 让 B1 发出匹配当前 Run 的 ready

#### 预期与不变量

- busy 期间请求等待；ready 且输入可用后写入正确 B1
- 投递与回执可以在没有 TUI 的情况下推进，不要求重新打开界面

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-003

**输出静默不能把未知 CLI 变成 ready**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 未知 CLI 无 ready 能力，Host 已运行，队列非空

#### 操作步骤

1. 使用可控时钟推进多个检测周期，期间不输出任何内容

#### 预期与不变量

- 工作状态仍为 unknown
- 自动输入计数为零；可以使用明确的人工投递入口

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-004

**approval 状态不被普通消息注入当作 ready**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 合作 agent 处于待批准状态，输入区对确认键敏感
- 存在尚未提交的普通任务消息

#### 操作步骤

1. 发出 approval 状态并进行多次队列轮询
2. 再发出新的可靠 ready

#### 预期与不变量

- approval 期间不会自动把任务正文或 Enter 当作批准输入
- 只有通过允许的就绪与占用检查后才投递

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-005

**旧 Run 的 ready 信号不能释放新 Run 的消息**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- B1 已结束，B2 活跃且 busy；保留一个延迟到达的 B1 ready 事件

#### 操作步骤

1. 向 B2 建立消息后送达 B1 ready
2. 随后送达 B2 的真实 ready

#### 预期与不变量

- B1 的事件不改变 B2 的就绪判断
- 只有 B2 的有效事件可触发针对 B2 的投递

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-006

**已知较旧的 ready 事件不能覆盖更新的 busy**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / contract
- 类型：concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 测试适配器提供可验证的事件代际/顺序证据，当前事件为 busy(n+1)

#### 操作步骤

1. 延迟送达 ready(n)
2. 读取状态并触发调度

#### 预期与不变量

- 已证明陈旧的 ready 不回退当前状态
- 不向 busy 的 Run 注入；事件排序证据在诊断记录中可核对

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-007

**attach 用户占用输入时默认暂缓自动投递**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-012](../COVERAGE.md#fl-012)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / windows_e2e
- 类型：functional、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- B1 已发出 ready；用户通过管理器 attach 并正在编辑未提交内容

#### 操作步骤

1. 向 B1 入队自动消息并等待调度
2. 明确交接或分离后重新确认可投递条件

#### 预期与不变量

- attach 占用期间消息不进入用户草稿
- 交接前后正文均不交错；恢复投递前重新校验 Run 和状态

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-008

**无法确认外部 attach 占用时保持等待**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-012](../COVERAGE.md#fl-012)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 存在管理器之外的 attach 客户端，后端占用信息不可确认
- B1 有可靠 ready 与待投递消息

#### 操作步骤

1. 使占用探测失败或结果不确定
2. 运行调度并查看等待原因

#### 预期与不变量

- 自动写入不以探测失败当作无人占用
- 界面/CLI 给出输入占用无法确认的等待原因

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-009

**投递检查与用户取得输入之间的竞争不能交错正文**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-012](../COVERAGE.md#fl-012)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-11](../COVERAGE.md#fr-11)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 调度器暂停在检查输入可用之后、开始写入之前
- 用户 attach 请求可通过同一测试屏障触发

#### 操作步骤

1. 释放用户取得输入与消息写入的竞争屏障，重复两种先后顺序

#### 预期与不变量

- 最终只由一个输入拥有者完成对应动作；另一方等待或明确失败
- 用户编辑文本与消息正文/提交不会相互插入，也不把竞争重试当作新消息

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-010

**无活跃 Run 时入队保持等待直到显式指定**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 目标 Session B 无活跃 Run，但允许保存待投递请求

#### 操作步骤

1. 建立消息后启动 B2
2. 观察自动消费者，再执行显式指定 B2 的操作

#### 预期与不变量

- B2 出现本身不自动承接未绑定运行的旧消息
- 显式指定前无输入；指定后保存可追溯的新目标依据并重新检查条件

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-011

**目标重启后旧队列暂停且原因可见**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 消息已绑定 B1，B1 仍 busy；B2 尚不存在

#### 操作步骤

1. 结束 B1 并启动 B2，令 B2 ready

#### 预期与不变量

- 消息仍指向原 B1 并显示 Run 变化的暂停原因
- 不会自动改写 target_run_id 或向 B2 投递

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-012

**后端 pane 或 PID 复用不能绕过逻辑 Run 校验**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / contract
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 消息绑定 B1，后端重建后将相同 pane 标识或 PID 分配给 B2/C1
- 测试能显式控制底层定位结果

#### 操作步骤

1. 恢复旧队列消费者并尝试解析目标

#### 预期与不变量

- 不能仅凭相同底层编号认定原 Run 仍活着
- 消息暂停或失败并说明身份不匹配；新运行没有收到旧正文

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-013

**多发送者正文与提交按同一投递顺序串行完成**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 2 或 16 个发送者给同一 B1 提交各含唯一编号和多行正文的消息
- B1 每次只在完成当前输入后发出新的 ready

#### 操作步骤

1. 同时建立请求并让 Host 消费全部
2. 按记录的实际投递顺序核对接收日志

#### 预期与不变量

- 每条正文及其提交构成完整单元，不出现 A 前半段/B 正文/A 后半段
- 未规定并发入队的全局先后时不武断要求按墙上时间排序，但实际选定顺序可追溯

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录

#### 参数变体

- senders：`[2, 16]`


### TC-QUEUE-014

**一个目标 busy 不应无故阻塞其他可投递目标**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：concurrency、functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- B1 持续 busy，C1 ready 且输入空闲；两个目标均有请求

#### 操作步骤

1. 保持 B1 busy 并让调度器处理两目标
2. 核对 C1 的实际接收及 B1 的等待状态

#### 预期与不变量

- C1 能完成属于自己的投递，B1 继续等待
- 不能通过向 B1 强行输入来解除阻塞；不要求未批准的毫秒 SLA

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-015

**重复 ready 通知不重复消费同一消息**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- B1 当前有一条可投递消息，适配器会重复发送同一事件标识

#### 操作步骤

1. 并发送达 10 次相同 ready 通知
2. 等待全部调度结束并查询尝试记录

#### 预期与不变量

- 仅有一次可证明的正常投递，不因重复通知新建或重放消息
- 如发生通道不确定性，保留 unknown 而不是重复写入直到成功

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-016

**同时启动两个 Host 不产生两个有效消费者**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0048](../../adr/0048-exclusive-processing-claims-and-uncertain-recovery.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 同一产品数据目录与 namespace 中已有待投递消息
- 两个 Host 实例可以同时启动

#### 操作步骤

1. 在所有权获取处同步启动两个 Host
2. 令目标 ready 并观察进程与投递

#### 预期与不变量

- 同一队列/目标仅有一个有效投递所有者；另一个退出或明确待命
- 不会发生双写，所有权失败不终止 agent

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录

说明： ADR-0048 当前为显式单次领取、私有领取文件和结果封存；unknown 不自动接管。独立处理进程与 SQLite 局部证据见 2026-11-26-task-claims，不代替共享 Host／自动消费／真实 agent 全场景。


### TC-QUEUE-017

**旧 Host 暂停后恢复不能与新所有者同时写入**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 根据已确定的单实例/恢复合同构造 Host H1 暂停、H2 获取所有权
- 两者均曾看到相同待投递请求

#### 操作步骤

1. 恢复 H1，并让 H1/H2 同时尝试写入

#### 预期与不变量

- 失去所有权者不能继续写入当前目标
- 所有权冲突可诊断，不能用两份独立进程内锁掩盖双消费

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-018

**Host 在领取后首次写入前崩溃时根据证据恢复**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 故障点位于持久化领取记录之后、第一次目标写入之前
- fixture 的独立输入日志可确认本次测试未写入

#### 操作步骤

1. 强制结束 Host 后重启
2. 检查恢复后的尝试、消息和实际输入

#### 预期与不变量

- 有产品侧可证明未写入的证据时才按恢复合同继续；否则保持不确定等待
- 不能仅靠测试夹具知道未写入就让产品无证据地宣称失败可重发；请求记录不丢失

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-019

**已持久化 input_written 的请求在 Host 重启后不重放**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 一条消息已有可核对 input_written 回执，但目标没有 accepted/completed 信号

#### 操作步骤

1. 结束并重启 Host 多次，再发 ready 通知
2. 观察目标输入计数和消息状态

#### 预期与不变量

- 消息不因缺少完成回执而重新进入普通自动发送队列
- 保持已有写入证据与未知后续结果，不自动升为 completed

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-020

**取消与领取或写入竞争具有明确边界**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-MSG-CANCEL](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已明确 queued/领取中/部分写入/已写入各阶段的取消合同与交互文案

#### 操作步骤

1. 分别在领取前、领取后写入前、部分写入后、完整写入后请求取消
2. 核对取消结果和目标实际输入

#### 预期与不变量

- 取消结果符合阶段合同，不把可能已经执行的请求描述为已撤回全部影响
- 确认取消成功且保证未写入的请求不得被后续消费者再次发送

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录

#### 参数变体

- cancel_phase：`["领取前", "领取后写入前", "部分写入后", "完整写入后"]`


### TC-QUEUE-021

**就绪信号源断开后不沿用过时 ready**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- B1 曾 ready；之后状态源失连，不能确认当前输入状态
- 有新待投递消息

#### 操作步骤

1. 中断状态源并继续运行调度
2. 恢复可信来源并发出当前 Run 的新状态

#### 预期与不变量

- 失去可靠状态期间不把旧 ready 永久当作可写许可
- 状态来源与更新时间可辨；恢复后按新证据判断

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-022

**队列容量上下界与拒绝行为可核对**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-015](../COVERAGE.md#fl-015)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：boundary、performance
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 持久化队列预算 B_QUEUE_ITEMS/B_QUEUE_BYTES 及已接受请求的保留规则已固定
- 消费者暂停；每条请求可计算字节大小

#### 操作步骤

1. 填充到预算以下、等于预算及超出预算
2. 恢复消费者并查询所有已接受请求

#### 预期与不变量

- 超限处理符合已公开规则；不能先返回 queued 再无提示丢弃已接受请求
- 积压不会把完整队列无限复制到 UI 内存

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录

#### 参数变体

- queue_size：`["B_QUEUE_ITEMS-1", "B_QUEUE_ITEMS", "B_QUEUE_ITEMS+1", "B_QUEUE_BYTES-1", "B_QUEUE_BYTES+1"]`


### TC-QUEUE-023

**消息有效期与系统时钟变化按已定策略处理**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：boundary、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-MSG-TTL](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已决定是否启用 TTL；若启用，固定时间来源、到期规则和状态说明

#### 操作步骤

1. 在到期前/恰好到期/到期后尝试消费
2. 模拟墙上时钟前跳和回拨，再重启 Host

#### 预期与不变量

- 符合已接受的 TTL 合同；未启用 TTL 时不得自行过期
- 已过期或暂停请求不因回拨而自动复活，不能因时钟变化重放已写入请求

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-024

**快速 ready/busy 事件不会被绘制合并误当作持续可写**

- 功能：[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、stability
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)

#### 前置条件

- 适配器按固定种子生成 100 次带顺序证据的状态变化
- 正文写入前能通过屏障控制最后一个有效状态

#### 操作步骤

1. 让调度与低频 UI 刷新同时处理事件
2. 分别让写入时最后状态为 busy 和 ready

#### 预期与不变量

- 控制事件的顺序与判断不因 UI 合并刷新而丢失
- busy 条件无写入；允许写入时目标和证据匹配，不以旧画面状态代替当前状态

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-QUEUE-025

**Host 正常关闭时保留已接受队列与未完成尝试**

- 功能：[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)、[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- Host 存在 queued、领取未写、已写未确认三种请求，可用屏障固定阶段

#### 操作步骤

1. 请求正常关闭 Host，再重新启动
2. 对照请求数量、每项状态与目标输入

#### 预期与不变量

- 已接受请求和不确定尝试不丢失；关闭不杀死终端 Run
- 恢复后不会将所有中间态统一改成可自动重发，已写内容不重复

#### 需要的证据

- 产品与依赖版本、目标 Session/Run、故障点或事件序列
- 持久化消息/投递尝试/状态依据的前后记录
- 独立 fixture 输入/提交日志，Host 与 attach 生命周期记录


### TC-CTL-001

**具备工具能力的 agent 实际调用 list/read/send/status**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-007](../COVERAGE.md#fl-007)、[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-06](../COVERAGE.md#fr-06)、[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / windows_e2e
- 类型：functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 合作 agent A 有真实命令执行工具及管理 CLI 使用说明
- B1 输出可识别内容，基础发送能力已可运行

#### 操作步骤

1. 让 A 依次执行列举、读取 B、发送请求和查询状态的实际管理命令
2. 在 A 的工具日志中检查输入输出

#### 预期与不变量

- 读取内容、message_id 和真实状态进入 A 的工具上下文
- 不能以 UI 上显示了结果或往终端粘贴命令文字代替实际工具调用

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-CTL-002

**列表返回稳定身份及可证明状态**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 创建符合命名规则、名称相似但不同的 A/B，会话分别为活 Run、已退出 Run；某终端保留 shell

#### 操作步骤

1. 调用 list，重命名/排序后再次调用
2. 对照进程证据与记录

#### 预期与不变量

- 稳定 session_id/run_id 可定位原对象
- shell 存活不能使已退出 agent 被报告为正在工作，未知工作状态明确表示 unknown

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-CTL-003

**消息状态查询不把等待时长当作完成**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)

#### 前置条件

- 准备 queued、input_written、unknown 三种真实证据状态，无 agent 完成事件
- 另有合作式 agent 已提供有效 completed 证据并持久化的请求，与无完成事件的三个请求分开

#### 操作步骤

1. 连续查询各 message_id 并推进测试时间
2. 关闭再开启 TUI 后再次查询
3. 让已完成请求的终端断开，再查询其历史消息和当前连接状态

#### 预期与不变量

- 返回与持久化证据一致的阶段和原因
- 不能因轮询次数、UI 访问或时间经过升级 accepted/completed
- 已保存的完成证据仍可查询；连接断开可以改变连接状态，不能把完成事实抹成 unknown 或重新投递

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

#### 参数变体

- evidence_variant：`["queued 无完成事件", "input_written 无完成事件", "unknown 无完成事件", "completed 落盘后断开"]`


### TC-CTL-004

**无效参数与命令错误具有稳定机器可读合同**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-CLI-CONTRACT](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定 CLI 子命令、参数、退出码、stdout/stderr 与错误分类合同

#### 操作步骤

1. 传入缺失目标、未知选项、错误参数类型与互斥参数组合
2. 由测试调用方解析结果

#### 预期与不变量

- 错误能与成功及空结果区分，且符合已记录的机器接口
- 未通过参数验证的请求不写目标、不创建伪成功消息

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

#### 参数变体

- invalid_arguments：`["缺失目标", "未知选项", "错误参数类型", "互斥参数同时提供"]`


### TC-CTL-005

**不存在的 Session/Run/消息 ID 不回落到当前会话**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TUI 当前选中 A1，测试提供不存在的 Session/Run/message_id

#### 操作步骤

1. 分别执行读取、发送和状态查询
2. 检查 A1 的输入与请求记录

#### 预期与不变量

- 明确表示目标缺失或不匹配，不使用当前会话作为隐式回退
- 不会返回另一个消息的状态或向 A1 写入

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-CTL-006

**只读命令不改变 agent 输入或启动新 Run**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-007](../COVERAGE.md#fl-007)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-06](../COVERAGE.md#fr-06)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 存在活 A1 与已退出 B1，记录两者 Run 集合和输入计数

#### 操作步骤

1. 多次执行 list/read/status，并读取 B1 保存历史

#### 预期与不变量

- 不会隐式启动 B2、提交提示或把查询文本发送给 A1
- 允许按需管理器/后端查询进程，不将其误判为新 agent Run

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-CTL-007

**IPC 或存储不可达时错误不伪装为业务成功**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 针对实现已选的命令路径设置 IPC 拒绝、连接中断或存储不可读故障
- 目标 agent 继续独立运行

#### 操作步骤

1. 执行读、发送、状态命令
2. 恢复控制面后检查是否存在请求

#### 预期与不变量

- 返回可识别的不可达或结果不确定状态，不能输出无依据成功
- 不自行重启 agent；结果可能已提交时提示用稳定幂等键/消息 ID 查询而不是盲目重发

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

#### 参数变体

- control_failure：`["连接被拒绝", "返回中途断开", "本地记录不可读"]`


### TC-CTL-008

**结构化输出中中文及换行不会破坏记录边界**

- 功能：[FL-010](../COVERAGE.md#fl-010)
- 需求：[FR-09](../COVERAGE.md#fr-09)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：compatibility、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-CLI-CONTRACT](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定机器输出格式、编码、版本标识和诊断输出通道
- 会话名称与输出含中文、引号、反斜线、换行和 emoji

#### 操作步骤

1. 调用 list/read/status 的机器接口
2. 使用独立解析器逐条解析并复原字符串

#### 预期与不变量

- 响应可按合同解析，正文与诊断不会混成无法识别的数据
- 转义后复原值准确，不能靠文本 grep 偶然得到正确字段

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-CTL-009

**指令文件路径含空格中文及文件不可读时行为明确**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / windows_e2e
- 类型：compatibility、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 测试目录包含带空格和中文的指令文件，另准备缺失/不可读文件
- 目标 FX-ECHO 无可靠就绪信号，记录收到的正文；提供已实现的显式投递入口

#### 操作步骤

1. 用实际 send 文件入口分别提交可读、缺失、不可读路径
2. 对可读文件核对消息编号、queued 状态与已保存正文；先确认目标尚未被写入，再显式投递到已确认的 Run

#### 预期与不变量

- 可读文件先持久化为对应消息编号；显式投递后正文按输入合同到达目标，回执只到可证明阶段
- 缺失/不可读情况在写入前明确失败，不把路径字符串当正文或发送空消息

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

#### 参数变体

- file_case：`["空格中文路径", "文件缺失", "文件不可读"]`


### TC-CTL-010

**CLI 超时重试可以找回同一请求而非重复建立**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-09](../COVERAGE.md#fr-09)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：recovery、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 请求创建已提交，响应在测试屏障处被阻塞；调用方保留幂等键 K

#### 操作步骤

1. 终止等待的 CLI，再用同一 K 重试并查询
2. 解除原响应屏障并检查请求集合

#### 预期与不变量

- 所有结果归于同一逻辑请求，发送者超时不等于请求不存在
- 不会因为两个 CLI 生命周期不同而绕过幂等范围

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-CTL-011

**普通终端文字不被管理器当成 agent 工具调用**

- 功能：[FL-010](../COVERAGE.md#fl-010)
- 需求：[FR-09](../COVERAGE.md#fr-09)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 目标为不具备工具调用能力的 FX-ECHO
- 另一个会话可被管理 CLI 控制

#### 操作步骤

1. 向 FX-ECHO 粘贴看似 list/read/send 的文字
2. 检查控制接口调用计数和另一个会话输入

#### 预期与不变量

- 文字仅作为普通终端内容处理；管理器不会自动代执行
- 界面不因此宣称该 agent 已具备主动协作能力

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-REPLY-001

**A 请求 B 后按 message_id 得到真实关联回复**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-010](../COVERAGE.md#fl-010)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-09](../COVERAGE.md#fr-09)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)、[ADR-0048](../../adr/0048-exclusive-processing-claims-and-uncertain-recovery.md)
- 优先级 / 阶段 / 层次：P0 / M3 / windows_e2e
- 类型：functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B 均为可实际调用工具的合作 agent，明确记录 A1/B1
- A 能查询请求 ID，B 能提交带 reply_to 的回复

#### 操作步骤

1. A 向 B 发出含唯一标记的请求
2. B 实际处理并返回带原 message_id 的结果
3. A 通过工具查询原请求的回复

#### 预期与不变量

- A 得到与原请求关联的正文及 B1 来源
- 输入/工具调用/回复三条证据可串联，不能用预填 UI 文本代替

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

说明：ADR-0047 当前切片使用原 Run 本地委派凭据、只读 inbox、每请求一份关联结果与离线查询；不验证真实模型 agent、进度聚合或 completed。实际局部执行见 2026-11-26-structured-replies 报告，目录仍为 designed/not_run。 ADR-0048 当前为显式单次领取、私有领取文件和结果封存；unknown 不自动接管。独立处理进程与 SQLite 局部证据见 2026-11-26-task-claims，不代替共享 Host／自动消费／真实 agent 全场景。


### TC-REPLY-002

**回复到达、完成声明和用户验收分别表达**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：state、functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- B 先返回普通说明，再发具有关联依据的完成声明；用户尚未验收
- 另有中间接收通知丢失、但最终完成事件来源/Run/请求关联均有效的变体

#### 操作步骤

1. 按顺序接收两类事件并查询状态
2. 查看界面和机器返回
3. 对该变体只交付有效完成事件，不补造丢失的 accepted 通知

#### 预期与不变量

- 普通回复不会自动等同于 completed
- 可信完成声明可以记录完成证据，但不显示用户已验收通过
- 有效完成证据可支持 completed，不强制所有中间通知逐一到达；同时保留实际证据来源而不伪造丢失事件

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

#### 参数变体

- event_delivery：`["普通回复再可信完成", "中间确认缺失但可信完成"]`


### TC-REPLY-003

**错误 Run 的回复不能完成当前请求**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)、[ADR-0048](../../adr/0048-exclusive-processing-claims-and-uncertain-recovery.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 请求 M 绑定 B1；B2 或其他 Run 返回带 M 编号的回复

#### 操作步骤

1. 提交来源 Run 不匹配的回复
2. 查询 M 的回复与状态

#### 预期与不变量

- 来源不匹配可识别，不能仅凭 message_id 字符串更新 M 为 accepted/completed
- B1 的请求证据和 B2 的活动不混合

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

说明：ADR-0047 当前切片使用原 Run 本地委派凭据、只读 inbox、每请求一份关联结果与离线查询；不验证真实模型 agent、进度聚合或 completed。实际局部执行见 2026-11-26-structured-replies 报告，目录仍为 designed/not_run。 ADR-0048 当前为显式单次领取、私有领取文件和结果封存；unknown 不自动接管。独立处理进程与 SQLite 局部证据见 2026-11-26-task-claims，不代替共享 Host／自动消费／真实 agent 全场景。


### TC-REPLY-004

**不存在的 reply_to 不误挂到最近请求**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 有多个正常请求，构造不存在/畸形 reply_to

#### 操作步骤

1. 提交包含该 reply_to 的结果
2. 查看正常请求集合和返回结果

#### 预期与不变量

- 关联失败可识别，不采用最近消息、当前 Tab 或相同名称猜测父请求
- 任何已有请求都不会被该回复标记完成

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

说明：ADR-0047 当前切片使用原 Run 本地委派凭据、只读 inbox、每请求一份关联结果与离线查询；不验证真实模型 agent、进度聚合或 completed。实际局部执行见 2026-11-26-structured-replies 报告，目录仍为 designed/not_run。


### TC-REPLY-005

**延迟旧 Run 回复不影响同 Session 的新运行**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：state、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- M 属于 A1 向 B1 的请求；B1 的可信回复被延迟
- A 或 B 已进入新 Run，旧请求记录仍保留

#### 操作步骤

1. 释放原 B1/M 的回复
2. 同时查询旧 M 与新 Run 的请求

#### 预期与不变量

- 旧回复保留明确原始来源/关联，不改挂到新 Run
- 当前运行不会收到隐式续接或旧任务完成通知冒充自己的结果

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

说明：ADR-0047 当前切片使用原 Run 本地委派凭据、只读 inbox、每请求一份关联结果与离线查询；不验证真实模型 agent、进度聚合或 completed。实际局部执行见 2026-11-26-structured-replies 报告，目录仍为 designed/not_run。


### TC-REPLY-006

**重复回复的去重语义由稳定回复标识决定**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)、[ADR-0048](../../adr/0048-exclusive-processing-claims-and-uncertain-recovery.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：recovery、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-REPLY-DEDUP](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定回复 ID、幂等范围及相同 ID 不同内容的冲突合同
- B 发出回复 R 后无法确认控制面是否已保存

#### 操作步骤

1. 用相同 ID/相同正文重试，再用相同 ID/不同正文重试
2. 核对关联结果和状态

#### 预期与不变量

- 同一回复重试符合去重合同，不伪造多个独立处理结果
- 内容冲突可识别，不覆盖已经保存的原回复而无说明

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

说明：ADR-0047 当前切片使用原 Run 本地委派凭据、只读 inbox、每请求一份关联结果与离线查询；不验证真实模型 agent、进度聚合或 completed。实际局部执行见 2026-11-26-structured-replies 报告，目录仍为 designed/not_run。 ADR-0048 当前为显式单次领取、私有领取文件和结果封存；unknown 不自动接管。独立处理进程与 SQLite 局部证据见 2026-11-26-task-claims，不代替共享 Host／自动消费／真实 agent 全场景。


### TC-REPLY-007

**多条进度与最终回复的聚合规则明确**

- 功能：[FL-014](../COVERAGE.md#fl-014)
- 需求：[FR-13](../COVERAGE.md#fr-13)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：functional、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-REPLY-DEDUP](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定一个请求是否允许多条回复、部分结果及最终标记的合同

#### 操作步骤

1. B 对 M 依次发送两个进度片段和一个最终结果
2. 重新打开客户端并查询 M

#### 预期与不变量

- 按既定规则保留关联与顺序，不把第一条进度误当最终完成
- 最终标记来自可验证事件；重开不丢失仍承诺保留的部分结果

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

说明：ADR-0047 当前切片使用原 Run 本地委派凭据、只读 inbox、每请求一份关联结果与离线查询；不验证真实模型 agent、进度聚合或 completed。实际局部执行见 2026-11-26-structured-replies 报告，目录仍为 designed/not_run。


### TC-REPLY-008

**终端出现 done 或完成样式文本不能伪造关联回执**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-008](../COVERAGE.md#fl-008)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-07](../COVERAGE.md#fr-07)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 未知 CLI 不具备可信回执能力，但可输出看似完成的文字或 JSON
- M 仅有 input_written 证据

#### 操作步骤

1. 让 CLI 输出 done、完成、completed 及含 M 字符串的普通正文
2. 查询 M 的状态

#### 预期与不变量

- 普通终端输出不被直接提升为可信关联完成事件
- M 仍保持有证据的阶段，输出可查看但不冒充结构化回执

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志


### TC-REPLY-009

**并发请求乱序返回仍逐项正确关联**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-010](../COVERAGE.md#fl-010)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-09](../COVERAGE.md#fr-09)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：concurrency、functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 向 B/C 建立 20 个不同请求，正文、目标 Run 与 ID 均可核对

#### 操作步骤

1. 按固定随机种子逆序和交错返回回复
2. A 并发查询各请求并比较期望映射

#### 预期与不变量

- 每个结果按 message_id/reply_to 与来源定位，不依赖返回顺序或列表位置
- 不存在跨请求正文、Run 或完成状态污染

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

#### 参数变体

- reply_order：`["逆序", "固定种子交错"]`


### TC-REPLY-010

**相同消息编号的错误来源会话不能冒充回复方**

- 功能：[FL-014](../COVERAGE.md#fl-014)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-13](../COVERAGE.md#fr-13)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)、[ADR-0048](../../adr/0048-exclusive-processing-claims-and-uncertain-recovery.md)
- 优先级 / 阶段 / 层次：P0 / M3 / integration
- 类型：negative、security
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[TBD-CLI-CONTRACT](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- M 的目标为 B1，C1 可调用自己的管理工具并知道 M 的文本编号
- 可信来源按实现的本机身份合同获取而非只信任显示名

#### 操作步骤

1. 由 C1 尝试给 M 提交声称来自 B1 的回复

#### 预期与不变量

- 来源不匹配被识别，不能仅凭正文中的 B1 字样认证来源
- M 不被伪造为 B 已接受/完成；本测试只覆盖同产品中的来源校验，不扩大为多用户权限平台

#### 需要的证据

- 产品/agent/依赖版本与实际使用命令（凭据脱敏）
- CLI stdout、stderr、退出结果及 agent 实际工具调用记录
- 关联 message_id、Session/Run、投递/回复证据与 fixture 独立日志

说明：ADR-0047 当前切片使用原 Run 本地委派凭据、只读 inbox、每请求一份关联结果与离线查询；不验证真实模型 agent、进度聚合或 completed。实际局部执行见 2026-11-26-structured-replies 报告，目录仍为 designed/not_run。 ADR-0048 当前为显式单次领取、私有领取文件和结果封存；unknown 不自动接管。独立处理进程与 SQLite 局部证据见 2026-11-26-task-claims，不代替共享 Host／自动消费／真实 agent 全场景。
