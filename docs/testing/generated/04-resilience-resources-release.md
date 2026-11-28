# 资源、长稳、控制边界、发布与属性测试

> 从 [JSON 源](../cases/04-resilience-resources-release.json) 自动生成。请修改源并运行 `python tools/test_catalog.py render`。

**全部为测试设计，当前未执行产品或 Windows 测试。** ready 仅表示预期依据明确。

[总索引](../CASE_INDEX.md) · [覆盖表](../COVERAGE.md) · [字段约定](../CASE_FORMAT.md)

| ID | 场景 | 优先级 | 阶段 | 规范状态 |
| --- | --- | --- | --- | --- |
| [TC-PERF-001](#tc-perf-001) | 0/1/5/10 会话全进程树基线 | P1 | M0 | 依据明确 |
| [TC-PERF-002](#tc-perf-002) | 禁预热仅作用于产品后台 | P1 | M0 | 依据明确 |
| [TC-PERF-003](#tc-perf-003) | 仅刷新可见预览而非后台全部屏幕 | P1 | M2 | 依据明确 |
| [TC-PERF-004](#tc-perf-004) | 单会话持续输出达到缓存上限 | P1 | M1 | 待决策 |
| [TC-PERF-005](#tc-perf-005) | 十会话同时高频输出仍可操作 | P1 | M2 | 待决策 |
| [TC-PERF-006](#tc-perf-006) | 超长无换行输出不能绕过字节预算 | P1 | M1 | 待决策 |
| [TC-PERF-007](#tc-perf-007) | 消息和待处理事件上限施加背压 | P0 | M3 | 待决策 |
| [TC-PERF-008](#tc-perf-008) | 日志轮转的磁盘与句柄上限 | P1 | M2 | 待决策 |
| [TC-PERF-009](#tc-perf-009) | 反复 attach/detach 后客户端资源回落 | P1 | M2 | 待决策 |
| [TC-PERF-010](#tc-perf-010) | 会话反复创建停止不留下孤儿 | P1 | M2 | 待决策 |
| [TC-PERF-011](#tc-perf-011) | 大历史列表与输出分页保持有界 | P1 | M2 | 待决策 |
| [TC-PERF-012](#tc-perf-012) | 空闲后的 CPU 与定时器开销 | P1 | M0 | 待决策 |
| [TC-PERF-013](#tc-perf-013) | 窗口尺寸和屏幕缓存上限 | P1 | M1 | 待决策 |
| [TC-PERF-014](#tc-perf-014) | 退出 UI 释放自身缓存且后台仍活着 | P1 | M2 | 依据明确 |
| [TC-PERF-015](#tc-perf-015) | 慢 capture 与坏会话不阻塞所有管理操作 | P1 | M2 | 待决策 |
| [TC-PERF-016](#tc-perf-016) | 机器内存压力下错误与恢复可观察 | P1 | M4 | 待决策 |
| [TC-SOAK-001](#tc-soak-001) | 8 小时混合会话稳定性 | P1 | M4 | 待决策 |
| [TC-SOAK-002](#tc-soak-002) | 24 小时十会话扩展压力 | P2 | M4 | 待决策 |
| [TC-SOAK-003](#tc-soak-003) | 持续输出中高频 resize 与 CJK 输入 | P1 | M4 | 待决策 |
| [TC-SOAK-004](#tc-soak-004) | 上万请求的持续队列消费 | P1 | M3 | 待决策 |
| [TC-SOAK-005](#tc-soak-005) | 反复 Host 崩溃与重开不自动重放 | P0 | M3 | 依据明确 |
| [TC-SOAK-006](#tc-soak-006) | 锁屏与睡眠唤醒后状态重新核对 | P1 | M4 | 依据明确 |
| [TC-SOAK-007](#tc-soak-007) | 壁钟跳变不能把旧 Run 信号当作新证据 | P1 | M3 | 待决策 |
| [TC-SOAK-008](#tc-soak-008) | 反复磁盘耗尽与恢复 | P0 | M4 | 待决策 |
| [TC-SOAK-009](#tc-soak-009) | 长时间隐藏会话与输出缺口 | P1 | M2 | 待决策 |
| [TC-SOAK-010](#tc-soak-010) | 后台缓慢响应时连续人工操作 | P1 | M4 | 待决策 |
| [TC-SEC-001](#tc-sec-001) | 消息正文不会在管理器 shell 中执行 | P0 | M3 | 依据明确 |
| [TC-SEC-002](#tc-sec-002) | 终端预览控制序列没有非预期副作用 | P0 | M1 | 待决策 |
| [TC-SEC-003](#tc-sec-003) | 另一个 namespace 的会话不会被停止或清理 | P0 | M2 | 依据明确 |
| [TC-SEC-004](#tc-sec-004) | 诊断信息不会意外公开启动环境中的密钥 | P1 | M4 | 待决策 |
| [TC-SEC-005](#tc-sec-005) | 本机控制入口的调用范围可核对 | P0 | M3 | 待决策 |
| [TC-SEC-006](#tc-sec-006) | 错误和超长后端响应不能破坏管理器状态 | P0 | M1 | 待决策 |
| [TC-REL-001](#tc-rel-001) | 干净原生 Windows 安装后完成双会话流程 | P0 | M4 | 依据明确 |
| [TC-REL-002](#tc-rel-002) | 缺少或不兼容的 psmux 给出可行动错误 | P0 | M4 | 依据明确 |
| [TC-REL-003](#tc-rel-003) | 发行包与源码和依赖版本一致 | P0 | M4 | 依据明确 |
| [TC-REL-004](#tc-rel-004) | 稳定 tag 与同版本字符串的开发构建不混用 | P0 | M0 | 依据明确 |
| [TC-REL-005](#tc-rel-005) | 升级中断后的数据与运行状态可恢复 | P0 | M4 | 待决策 |
| [TC-REL-006](#tc-rel-006) | 旧版本读取较新数据的回退边界 | P0 | M4 | 待决策 |
| [TC-REL-007](#tc-rel-007) | 卸载与清理的作用范围 | P0 | M4 | 待决策 |
| [TC-REL-008](#tc-rel-008) | 安装和数据路径包含中文空格 | P1 | M4 | 依据明确 |
| [TC-REL-009](#tc-rel-009) | 纯本地测试 agent 可离线完成管理流程 | P1 | M4 | 依据明确 |
| [TC-REL-010](#tc-rel-010) | 许可证与迁入来源材料完整 | P0 | M0 | 待决策 |
| [TC-REL-011](#tc-rel-011) | Charm v2 迁入后构建基线可复现 | P0 | M0 | 依据明确 |
| [TC-REL-012](#tc-rel-012) | 平台声明只覆盖真实验证组合 | P0 | M4 | 依据明确 |
| [TC-REL-013](#tc-rel-013) | 低内存宣传与测量范围一致 | P0 | M4 | 待决策 |
| [TC-REL-014](#tc-rel-014) | 发布依赖升级触发相应回归 | P1 | M4 | 依据明确 |
| [TC-REL-015](#tc-rel-015) | 诊断和恢复说明可由新用户完成 | P1 | M4 | 依据明确 |
| [TC-REL-016](#tc-rel-016) | 相同包多次启动与配置目录选择 | P0 | M4 | 待决策 |
| [TC-DEFER-001](#tc-defer-001) | Tab 内终端与管理导航的焦点合同 | P1 | M2 | 待决策 |
| [TC-DEFER-002](#tc-defer-002) | Tab 内终端资源与 VT 完整性 | P1 | M2 | 待决策 |
| [TC-PROP-001](#tc-prop-001) | 随机生命周期操作保持 Session/Run 不变量 | P1 | M2 | 待决策 |
| [TC-PROP-002](#tc-prop-002) | 随机参数编码保持直接启动边界 | P1 | M1 | 待决策 |
| [TC-PROP-003](#tc-prop-003) | 分片方式不改变有效 UTF-8/VT 的最终屏幕 | P1 | M1 | 依据明确 |
| [TC-PROP-004](#tc-prop-004) | 随机消息事件序列不夸大证据或改变目标 | P1 | M3 | 依据明确 |
| [TC-PROP-005](#tc-prop-005) | 配置和元数据往返保持身份与参数 | P1 | M2 | 待决策 |
| [TC-PROP-006](#tc-prop-006) | 随机异常终端字节有界且不产生管理操作 | P1 | M1 | 待决策 |
| [TC-PROP-007](#tc-prop-007) | 历史分页组合与固定快照范围一致 | P1 | M3 | 待决策 |
| [TC-PROP-008](#tc-prop-008) | 并发发送调度不能把正文和提交交错 | P1 | M3 | 依据明确 |

### TC-PERF-001

**0/1/5/10 会话全进程树基线**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-016](../COVERAGE.md#fl-016)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-02](../COVERAGE.md#fr-02)、[FR-05](../COVERAGE.md#fr-05)、[FR-07](../COVERAGE.md#fr-07)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M0 / windows_e2e
- 类型：performance、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 独占的测试数据目录与产品 namespace；FX-LOAD 使用固定种子
- 固定 Windows、终端尺寸、字体、shell、后台和构建版本

#### 操作步骤

1. 分别启动 0、1、5、10 会话；每档稳定 5 分钟后采样 15 分钟
2. 依次测空闲、仅当前会话输出、全部会话输出
3. 用相同 shell/agent 负载直接运行作对照

#### 预期与不变量

- 报告包含 TUI/Host、psmux server/client、shell/控制台、agent 各组与整机数据
- 记录工作集、私有提交、CPU、句柄、线程和进程数；不得只报管理器一个进程
- 未固定预算时结果只作为基线，不判定低内存达标

#### 需要的证据

- CSV 时序与采样间隔
- 进程树及 PID/启动时间
- 构建/配置/负载种子清单

#### 参数变体

- sessions：`[0, 1, 5, 10]`
- load：`["idle", "selected_only", "all"]`


### TC-PERF-002

**禁预热仅作用于产品后台**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M0 / windows_e2e
- 类型：performance、state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 存在独立的用户 psmux namespace 和一个产品测试 namespace

#### 操作步骤

1. 记录用户后台配置和进程
2. 以产品私有目录、PSMUX_NO_WARM=1 与 warm off 启动/退出 5 个测试 session
3. 观察初始化后和闲置 15 分钟的进程树

#### 预期与不变量

- 产品 namespace 不出现未配置的备用 shell 池或待命 server
- 用户后台配置、会话和环境未被改变
- 每个剩余进程均有明确归属和生命周期依据；留存终端/shell 不证明 agent 仍存活，不能仅按是否关联活 agent 判泄漏

#### 需要的证据

- 前后配置差异
- 父子进程及 namespace 映射
- psmux 实际命令和版本


### TC-PERF-003

**仅刷新可见预览而非后台全部屏幕**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-003](../COVERAGE.md#fl-003)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：performance、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 10 个会话都有相同输出速率；有 capture 调用计数和可见会话标记

#### 操作步骤

1. 固定选择 A 运行 10 分钟
2. 切到 B 再运行 10 分钟
3. 进入 attach 或退出 TUI 后记录本产品预览采集；窗口最小化仅记观察值，不假设产品能感知终端窗口可见性

#### 预期与不变量

- 管理界面显示完整预览时读取集中于当前选择；其他会话仅有已定义的必要状态检查
- 切换后捕获对象正确变化，旧预览订阅释放
- 管理器退出后不遗留自身预览采集循环

#### 需要的证据

- 带 session/run 的调用计数
- 可见性事件与 CPU 时序


### TC-PERF-004

**单会话持续输出达到缓存上限**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / windows_e2e
- 类型：performance、boundary、stability
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定 B_CACHE_BYTES/B_SCREEN_CELLS 和截断策略
- 输出夹具含每块递增序号与校验值

#### 操作步骤

1. 以固定速率输出 30 分钟直至超过缓存预算多个数量级
2. 定期读取屏幕和尾部
3. 停止输出并等待正常后台清理

#### 预期与不变量

- 缓存容量维持已记录限额，不随总输出量增长
- 保留部分的字节/序号可核对，丢失区间有明确标记
- 回收后测量按预算判定，不把 GC 后暂时下降当作无泄漏证据

#### 需要的证据

- 缓存计数与私有内存时序
- 保留片段及 gap 标识


### TC-PERF-005

**十会话同时高频输出仍可操作**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-004](../COVERAGE.md#fl-004)
- 需求：[FR-03](../COVERAGE.md#fr-03)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：performance、concurrency、stability
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定输入响应预算 B_INPUT_P95_MS 与采样方法
- 10 个测试会话，每个使用独立序号域

#### 操作步骤

1. 所有会话持续输出 30 分钟
2. 按固定脚本搜索、切换、attach、detach，并向已占用输入的目标输入唯一标记

#### 预期与不变量

- 每个标记只进入选定 Run，切换不产生新的 agent Run
- 输入响应和队列积压有完整观测；按已固定预算判定
- 预览刷新可合并，命令事件不得因显示丢帧而丢失

#### 需要的证据

- 输入目标字节日志
- 延迟样本与积压时序
- 进程资源曲线


### TC-PERF-006

**超长无换行输出不能绕过字节预算**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / windows_e2e
- 类型：boundary、performance、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定 B_CACHE_BYTES；有超长行、全宽字符和不完整 UTF-8 的测试数据

#### 操作步骤

1. 输出长度 B-1、B、B+1、10B 的单行
2. 分别在缓存、预览和读取接口观察结果

#### 预期与不变量

- 限额按约定的字节/屏幕对象成本执行，不能只按行数计数
- 显示与截断规则可核对；无崩溃、无无界分配
- 非法或分割字符按输入合同处理，不伪造完整原文

#### 需要的证据

- 峰值私有内存
- 返回范围和截断信息
- 原始输入校验值

#### 参数变体

- length：`["B_CACHE_BYTES-1", "B_CACHE_BYTES", "B_CACHE_BYTES+1", "10*B_CACHE_BYTES"]`


### TC-PERF-007

**消息和待处理事件上限施加背压**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M3 / windows_e2e
- 类型：boundary、concurrency、performance
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[O-012](../../DECISIONS.md)、[TBD-BACKPRESSURE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定 B_MSG_BYTES/B_QUEUE_ITEMS/B_EVENT_BACKLOG 及满载策略
- ready fixture 先保持 busy

#### 操作步骤

1. 按上限附近尺寸和数量并发入队，再制造终端输出风暴
2. 逐渐使目标 ready，并停止生产者

#### 预期与不变量

- 可拒绝、等待或降速的具体行为与已记录策略一致
- 已确认持久化的请求不能静默丢失；控制事件不与显示帧一起丢弃
- 内存/数据库增长在配置范围内，解除压力后无需重启即可继续处理

#### 需要的证据

- 入队响应与持久化记录对照
- 队列长度/内存曲线
- 丢弃显示帧与控制事件统计

#### 参数变体

- count：`["B_QUEUE_ITEMS-1", "B_QUEUE_ITEMS", "B_QUEUE_ITEMS+1"]`


### TC-PERF-008

**日志轮转的磁盘与句柄上限**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：boundary、performance、stability
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定 B_LOG_TOTAL 与轮转/保留策略；只使用测试目录

#### 操作步骤

1. 持续输出至发生至少 20 次轮转
2. 在轮转前后并发读取旧游标和最新尾部
3. 停止会话并检查文件句柄与空间

#### 预期与不变量

- 保留大小符合预算及已说明的轮转瞬时余量
- 删除/轮转片段返回 gap，不用新文件相同偏移冒充旧输出
- 停止后不遗留本应释放的句柄，仍运行会话不受影响

#### 需要的证据

- 文件清单/尺寸时序
- 句柄归属
- 游标读取结果


### TC-PERF-009

**反复 attach/detach 后客户端资源回落**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-002](../COVERAGE.md#fl-002)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-05](../COVERAGE.md#fr-05)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：stability、performance、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-009](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 测试 agent 持续输出自身 Run 标识
- 已固定 detach 路径和资源回落容差

#### 操作步骤

1. 对同一 Run 执行 1000 次 attach/detach，分批采样
2. 最后保持管理模式 15 分钟

#### 预期与不变量

- agent 启动次数和 run_id 不变；允许每次创建 attach 客户端
- 已退出 attach 客户端和句柄被释放，不形成单调增长的残留集合
- 终端状态恢复且仍可输入，资源按已固定容差判断

#### 需要的证据

- 每次客户端 PID/退出事件
- Run 连续标记
- 批次资源时序


### TC-PERF-010

**会话反复创建停止不留下孤儿**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-004](../COVERAGE.md#fl-004)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：stability、recovery、performance
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-LIFECYCLE](../BUDGETS_AND_DECISIONS.md)、[TBD-ARCHIVE-DELETE](../BUDGETS_AND_DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 所有测试进程均带本次测试标记
- 明确停止语义；存档和保留设置已固定

#### 操作步骤

1. 连续 100 轮创建 10 个测试会话、使用、停止并按策略收纳
2. 每轮核对用户 namespace 与产品进程树

#### 预期与不变量

- 没有上一轮应结束却仍存活的 agent/子进程/attach 客户端
- session_id/run_id 不被复用到新逻辑对象
- 资源增长能归因于受保留策略限制的记录，而非活进程泄漏

#### 需要的证据

- 轮次清单与进程退出证据
- ID 集合
- 元数据和日志大小曲线


### TC-PERF-011

**大历史列表与输出分页保持有界**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-004](../COVERAGE.md#fl-004)、[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：performance、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-008](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 仅在隔离数据集生成历史 Session/Run 与轮转输出；活动会话保持 5 个
- 已固定分页与搜索返回边界

#### 操作步骤

1. 分别加载 1000 和 100000 条历史记录
2. 打开列表、翻页、搜索、读取有限历史片段
3. 返回活动会话并输入标记

#### 预期与不变量

- 只读取与当前查询/页面有关的数据，不把全部输出复制到每个视图
- 结果对应正确 Session/Run，超过返回范围时边界明确
- 按预算记录峰值内存和延迟，不能将大数据集不可用静默显示为空

#### 需要的证据

- 查询/页边界记录
- 响应时序
- 活动目标输入日志

#### 参数变体

- historical_records：`[1000, 100000]`


### TC-PERF-012

**空闲后的 CPU 与定时器开销**

- 功能：[FL-015](../COVERAGE.md#fl-015)
- 需求：—
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M0 / windows_e2e
- 类型：performance、stability
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 相同输出负载先运行 5 分钟，后停止输出
- 已固定空闲 CPU 采样窗口与预算

#### 操作步骤

1. 在 0、1、5、10 会话下分别空闲 30 分钟
2. 记录刷新、capture、timer 唤醒和进程 CPU

#### 预期与不变量

- 无输出和交互时没有无条件忙循环
- 必要的后台状态检查与刷新策略可解释
- 按空闲预算判定；仅观察到屏幕不变不能代替 CPU 数据

#### 需要的证据

- CPU/唤醒/调用计数时序
- 配置与窗口尺寸

#### 参数变体

- sessions：`[0, 1, 5, 10]`


### TC-PERF-013

**窗口尺寸和屏幕缓存上限**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-002](../COVERAGE.md#fl-002)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / windows_e2e
- 类型：boundary、performance、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-004](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定可接受行列范围及 B_SCREEN_CELLS
- 终端夹具可报告实际尺寸

#### 操作步骤

1. 向后端和管理器输入最小值、边界值、零/负数模拟值及过大尺寸
2. 循环从宽窗口缩到窄窗口

#### 预期与不变量

- 无效尺寸得到明确拒绝/约定夹取，不发生整数溢出或巨量分配
- 有效尺寸保持显示与光标一致
- 被动预览不持续改写真实 attach 客户端的尺寸

#### 需要的证据

- resize 请求和结果
- 缓存规模
- 光标/屏幕快照


### TC-PERF-014

**退出 UI 释放自身缓存且后台仍活着**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：performance、recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 5 个 Run 有连续序号输出，队列消费者情况明确

#### 操作步骤

1. 采样管理器打开时的进程树与内存
2. 正常退出 TUI，再以崩溃退出重复一轮
3. 观察 15 分钟后重新打开

#### 预期与不变量

- UI 进程退出且其资源释放；后台 Run 按真实存活状态保留
- 重开接回原 Run，不把新开的 UI 算作新任务
- 不把 UI 消失后的全机内存波动误判为终端后台泄漏

#### 需要的证据

- 进程生命周期
- Run 输出序号
- 前后资源分组


### TC-PERF-015

**慢 capture 与坏会话不阻塞所有管理操作**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-013](../COVERAGE.md#fl-013)、[FL-004](../COVERAGE.md#fl-004)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：performance、negative、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-004](../../DECISIONS.md)、[O-012](../../DECISIONS.md)、[TBD-BACKEND-TIMEOUT](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 一个会话的 capture 在故障屏障挂起，另有健康会话
- 已固定后台调用超时及取消规则

#### 操作步骤

1. 反复选中慢会话后切到健康会话
2. 同时请求停止一个独立测试 Run
3. 释放故障屏障

#### 预期与不变量

- 过期 capture 不能覆盖新选中的预览
- 健康对象的操作可继续，失败/超时可定位到原请求
- 未完成调用和重试数量受限，不因连续选择创建无界工作队列

#### 需要的证据

- 调用关联 ID 与时间线
- 预览目标记录
- 积压和线程曲线


### TC-PERF-016

**机器内存压力下错误与恢复可观察**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M4 / windows_e2e
- 类型：performance、recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-012](../../DECISIONS.md)、[TBD-BACKPRESSURE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 仅在资源受限测试 VM 进行；已固定压力停止条件
- 预先保存可恢复数据快照

#### 操作步骤

1. 运行 5 会话并逐级施加 VM 内存压力
2. 在受限条件下读取/切换/入队；解除压力

#### 预期与不变量

- 无错误目标、虚假成功或无条件重放
- 失败能归因，持久化已确认的记录仍可核对
- 性能数据与压力水平同时报告；退出或拒绝不得导致用户其他会话被清理

#### 需要的证据

- VM 压力配置
- 操作回执与数据快照
- 进程事件/资源曲线


### TC-SOAK-001

**8 小时混合会话稳定性**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M4 / soak
- 类型：stability、recovery、performance
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- ENV-WIN-BASE 已通过短程核心流程
- 固定 5 个会话、负载种子、预算与日志轮转设置

#### 操作步骤

1. 持续 8 小时交替 idle、burst、全屏 TUI、attach/detach 和有限输入
2. 每分钟保存分组资源与关键计数
3. 结束后对全部 Run 核对来源、退出和最后输出

#### 预期与不变量

- 全程无不可恢复卡死、错目标、虚假回执和无解释数据缺口
- 句柄、进程、队列与内存遵守已固定预算，窗口比较不能忽略渐进增长
- 日志收集器自身开销单独记录，不掩盖产品成本

#### 需要的证据

- 8 小时时序
- 事件与输入校验
- 环境/预算/种子清单

说明：8 小时是测试轮次设计，不是产品寿命上限或已完成测量。


### TC-SOAK-002

**24 小时十会话扩展压力**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P2 / M4 / soak
- 类型：stability、performance
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已通过 8 小时基础稳定性轮次；隔离机器有资源保护阈值
- 固定 10 会话与预算

#### 操作步骤

1. 持续 24 小时轮换单会话输出与全部输出
2. 每小时执行一次 UI 退出重开和历史分页
3. 记录任何采集中断

#### 预期与不变量

- 后台 Run 身份连续，重开不复制 agent
- 长周期资源趋势在记录的预算内；不能只比较首尾两个数
- 超出测试保护阈值时终止并保留失败证据，不删除失败轮次

#### 需要的证据

- 长时资源曲线与斜率分析
- 重连身份
- 中止原因与残留清理记录

说明：扩展轮次；是否列为每次发布门槛由发布风险决定，不替代基础 8 小时轮次。


### TC-SOAK-003

**持续输出中高频 resize 与 CJK 输入**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-002](../COVERAGE.md#fl-002)、[FL-012](../COVERAGE.md#fl-012)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M4 / soak
- 类型：stability、compatibility、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-004](../../DECISIONS.md)、[O-009](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 固定可接受尺寸范围和输入交接规则
- FX-VT 屏幕具有位置标记，FX-ECHO 记录已提交输入

#### 操作步骤

1. 连续 2 小时随机但可复现地 resize，执行至少 1000 次尺寸变化
2. 按固定间隔输入中英文混合多行并 attach/detach

#### 预期与不变量

- 输入顺序、目标、提交次数与脚本一致
- CJK 宽度、光标和终端返回状态没有累积漂移
- 无资源持续增长或事件积压超过预算

#### 需要的证据

- 尺寸种子及时间线
- 输入字节/文本对照
- 终端录屏与资源


### TC-SOAK-004

**上万请求的持续队列消费**

- 功能：[FL-015](../COVERAGE.md#fl-015)、[FL-009](../COVERAGE.md#fl-009)、[FL-014](../COVERAGE.md#fl-014)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-13](../COVERAGE.md#fr-13)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / soak
- 类型：stability、concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[O-012](../../DECISIONS.md)、[TBD-BACKPRESSURE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 合作式 FX-READY 明确上报接受/完成，正文带唯一标记
- 消息容量、仲裁和消费策略已固定

#### 操作步骤

1. 以固定种子在 A/B/C 间生成 10000 个请求及关联回复
2. 穿插 busy/ready、UI 关闭重开和受控短暂停顿
3. 等待所有可确认请求进入可解释状态

#### 预期与不变量

- 消息编号、来源 Run、reply_to 一致，无串线
- 正常无故障路径下每个请求有唯一管理器投递尝试与可核对正文
- 队列积压受限；unknown/paused 不被统计为成功，剩余请求有原因

#### 需要的证据

- 独立 agent 接收日志
- 持久化消息/尝试记录
- 积压和资源时序


### TC-SOAK-005

**反复 Host 崩溃与重开不自动重放**

- 功能：[FL-013](../COVERAGE.md#fl-013)、[FL-009](../COVERAGE.md#fl-009)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / soak
- 类型：stability、recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 只操作测试 Host；存在确定性投递故障屏障和独立输入证据

#### 操作步骤

1. 循环 30 次在持久化前、写入前、部分写入、写完回执前终止 Host
2. 每轮重开并观察恢复
3. 对无法证明未写入的请求只由测试人员显式处理

#### 预期与不变量

- 状态以已有证据恢复，可能写出时保持 unknown/暂停而不自动重放
- 原始请求与尝试历史保留，不能通过新 message_id 掩盖重复
- 多轮恢复没有积累多余消费者或无限资源

#### 需要的证据

- 故障点时间线
- 数据库快照
- 目标实际接收内容及次数


### TC-SOAK-006

**锁屏与睡眠唤醒后状态重新核对**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)、[FL-016](../COVERAGE.md#fl-016)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-02](../COVERAGE.md#fr-02)、[FR-05](../COVERAGE.md#fr-05)、[FR-07](../COVERAGE.md#fr-07)、[FR-11](../COVERAGE.md#fr-11)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M4 / soak
- 类型：stability、recovery、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 仅在测试 Windows VM/专机；保存所有 Run 身份和待投递消息
- 就绪信号与超时策略明确

#### 操作步骤

1. 完成 20 轮锁屏/解锁和睡眠/唤醒
2. 每次恢复后读取状态并继续一次显式交互
3. 记录后台被系统结束的轮次

#### 预期与不变量

- 活着的 Run 可重新接回；被系统结束的对象如实退出/未知，不冒充续接
- 旧 ready 不能绕过重新核对目标 Run 与输入占用
- 不保证系统强制结束后活进程还能恢复；队列保留真实状态

#### 需要的证据

- 系统电源事件
- Run/后台实际存活证据
- 投递记录


### TC-SOAK-007

**壁钟跳变不能把旧 Run 信号当作新证据**

- 功能：[FL-013](../COVERAGE.md#fl-013)、[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：stability、boundary、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-MSG-TTL](../BUDGETS_AND_DECISIONS.md)、[TBD-SIGNAL-FRESHNESS](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 用 FX-CLOCK 控制测试时钟，不改宿主全局时钟
- 信号有效期与排序依据待固定

#### 操作步骤

1. 在 busy/ready/入队之间令壁钟向前和向后跳变
2. 同时保持不同 Run 的事件编号
3. 重复触发延迟事件

#### 预期与不变量

- 不因显示时间相同或倒退而混淆 Run、请求和回执
- 超时/TTL 按最终记录的时钟政策执行，未定前不判通过
- 缺乏可靠就绪证据时继续等待，不以时间跳变制造 ready

#### 需要的证据

- 虚拟时间与事件序列
- 信号来源/Run/尝试记录


### TC-SOAK-008

**反复磁盘耗尽与恢复**

- 功能：[FL-013](../COVERAGE.md#fl-013)、[FL-015](../COVERAGE.md#fl-015)、[FL-009](../COVERAGE.md#fl-009)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M4 / soak
- 类型：stability、recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-008](../../DECISIONS.md)、[TBD-BACKPRESSURE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 仅使用有限容量临时虚拟磁盘；绝不填满宿主磁盘
- 数据库与日志路径、满载策略已记录

#### 操作步骤

1. 运行持续输出和排队负载
2. 执行 10 轮写入空间耗尽、恢复可用空间、重新读取/发送
3. 保留每次报错和用户可见回执

#### 预期与不变量

- 未落盘的请求不显示 queued 成功
- 已确认记录仍可核对，不能通过静默清库恢复
- 空间恢复后按明确恢复流程工作，无自动重放不确定请求

#### 需要的证据

- 磁盘容量曲线
- 存储错误与回执
- 前后完整性/消息证据


### TC-SOAK-009

**长时间隐藏会话与输出缺口**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-015](../COVERAGE.md#fl-015)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / soak
- 类型：stability、state、performance
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 5 个会话输出独立序号，非可见预览不刷新
- 后端历史保留量已固定

#### 操作步骤

1. 连续 8 小时只查看 A，其余后台持续运行
2. 最后逐个切换读取当前屏幕与可保留尾部

#### 预期与不变量

- 后台终端存活与应用层记录连续性分别报告
- 超出后端保留范围时返回明确 gap，不能伪造完整历史
- 切换后不会为了补齐历史无界加载；预览仍属于正确 Run

#### 需要的证据

- 原始负载序号
- 后端保留配置
- 读取范围/gap 与内存


### TC-SOAK-010

**后台缓慢响应时连续人工操作**

- 功能：[FL-013](../COVERAGE.md#fl-013)、[FL-012](../COVERAGE.md#fl-012)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-11](../COVERAGE.md#fr-11)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M4 / manual
- 类型：stability、usability、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-009](../../DECISIONS.md)、[O-012](../../DECISIONS.md)、[TBD-BACKEND-TIMEOUT](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 在测试 backend wrapper 注入有界延迟；界面仍有健康会话
- 操作响应预算与取消语义已固定

#### 操作步骤

1. 连续 30 分钟执行创建、取消表单、搜索、切换、attach/detach
2. 穿插读和显式发送；每项使用独立标记

#### 预期与不变量

- 输入不会因延迟响应覆盖而串入错误表单或 Run
- 过期任务有明确收尾，不积累无法取消的后台调用
- 可见等待/错误与真实请求对应，按已记录预算判定响应

#### 需要的证据

- 操作录屏与请求关联日志
- 未完成调用数
- 目标输入证据


### TC-SEC-001

**消息正文不会在管理器 shell 中执行**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-008](../COVERAGE.md#fl-008)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-07](../COVERAGE.md#fr-07)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / windows_e2e
- 类型：security、negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 使用无副作用的文本标记与只位于临时目录的哨兵
- 目标 FX-ECHO 逐字节记录输入

#### 操作步骤

1. 通过文件/缓冲区发送引号、分号、管道、反引号、美元括号、百分号、换行和非 ASCII 文本
2. 查看管理器进程与测试哨兵

#### 预期与不变量

- 正文按选定传输合同进入目标，不被管理器当作命令拼接执行
- 临时哨兵未因正文产生额外动作
- 只声明管理器传输安全；目标 agent 自己解释指令不属于隔离保证

#### 需要的证据

- 正文校验值与目标输入
- 管理器子进程列表
- 哨兵检查


### TC-SEC-002

**终端预览控制序列没有非预期副作用**

- 功能：[FL-002](../COVERAGE.md#fl-002)、[FL-007](../COVERAGE.md#fl-007)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-06](../COVERAGE.md#fr-06)、[FR-11](../COVERAGE.md#fr-11)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M1 / windows_e2e
- 类型：security、negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-TERMINAL-CONTROL](../BUDGETS_AND_DECISIONS.md)、[O-004](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 隔离测试终端；保存测试剪贴板/标题状态
- 预览对 OSC/CSI/DCS 的支持和拒绝规则待固定

#### 操作步骤

1. FX-VT 输出完整/分片/未终结控制序列及超长参数
2. 只查看被动预览，再按需进入实际 attach
3. 记录剪贴板、标题、输入与渲染

#### 预期与不变量

- 管理器的被动预览只执行已声明的显示行为，不把输出重新解释为管理命令
- 不产生非预期粘贴、发送、外部程序启动或无界解析
- 真实 attach 的终端能力与预览策略分开验收，不假定完全等价

#### 需要的证据

- 原始控制字节
- 外部状态前后对照
- 屏幕和资源记录


### TC-SEC-003

**另一个 namespace 的会话不会被停止或清理**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)、[FL-016](../COVERAGE.md#fl-016)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-02](../COVERAGE.md#fr-02)、[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)、[FR-07](../COVERAGE.md#fr-07)、[FR-11](../COVERAGE.md#fr-11)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：security、negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 同一测试用户存在产品A、产品B与用户自己的后台 namespace
- 所有目标有独立事件记录

#### 操作步骤

1. 在 A 执行发现、失败恢复、停止和退出清理
2. 制造与 B 相同的可见名称和底层短编号

#### 预期与不变量

- A 只操作已归属自己的完整后端目标
- B 与用户会话的进程、配置和记录未被更改
- 归属不明时停止破坏性操作并给出可定位状态

#### 需要的证据

- 完整目标定位参数
- 多 namespace 事件日志
- 目录差异


### TC-SEC-004

**诊断信息不会意外公开启动环境中的密钥**

- 功能：[FL-013](../COVERAGE.md#fl-013)、[FL-017](../COVERAGE.md#fl-017)
- 需求：[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0068](../../adr/0068-bounded-workbench-diagnostics.md)
- 优先级 / 阶段 / 层次：P1 / M4 / windows_e2e
- 类型：security、release、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-DIAGNOSTICS](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- 只使用假的 token、口令、用户名和测试路径
- 诊断导出范围、脱敏字段和用户选择规则已定义

#### 操作步骤

1. 在启动失败、后台报错和消息失败场景生成诊断信息
2. 检查默认 UI 错误、诊断文件和可分享模板

#### 预期与不变量

- 默认可分享诊断不泄露约定的敏感启动配置字段
- 必要的故障定位字段仍存在，脱敏不会改写持久化真实配置
- 终端正文可能含用户数据，不能宣称自动完全脱敏；分享需按既定范围处理

#### 需要的证据

- 包含哨兵的输入与导出对照
- 脱敏字段规则
- 错误定位记录


### TC-SEC-005

**本机控制入口的调用范围可核对**

- 功能：[FL-010](../COVERAGE.md#fl-010)、[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-09](../COVERAGE.md#fr-09)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M3 / contract
- 类型：security、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[TBD-IPC-TRUST](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- IPC 和调用者访问范围尚待 O-008/TBD-IPC-TRUST 明确
- 仅在隔离测试用户/进程中验证，不访问他人真实数据

#### 操作步骤

1. 以允许与不允许的测试调用者读取/发送/停止
2. 尝试连接错误 namespace、过期 endpoint 和伪造来源字段

#### 预期与不变量

- 结果符合已记录的本机访问模型，不能仅相信请求自报的来源标签
- 拒绝路径不修改目标或暴露不在允许范围内的内容
- 本条不要求首版实现远程多租户或 agent 沙箱

#### 需要的证据

- 调用者/endpoint/权限矩阵
- 操作结果与目标副作用


### TC-SEC-006

**错误和超长后端响应不能破坏管理器状态**

- 功能：[FL-013](../COVERAGE.md#fl-013)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M1 / contract
- 类型：security、negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-004](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- backend wrapper 可返回畸形 UTF-8、控制字符、超长字段和部分行
- 后端响应/错误文本大小上限待固定

#### 操作步骤

1. 在 list/capture/status 的返回边界逐字节截断
2. 注入不完整标识、重复字段和超长值
3. 恢复正常响应

#### 预期与不变量

- 不可解析结果明确为错误/未知，不转换为错误对象的有效 ID
- 错误显示不会形成新的输入命令
- 解析、日志和重试内存有界，正常目标仍可管理

#### 需要的证据

- 注入字节与解析结果
- 错误目标/副作用检查
- 资源曲线


### TC-REL-001

**干净原生 Windows 安装后完成双会话流程**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-016](../COVERAGE.md#fl-016)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-02](../COVERAGE.md#fr-02)、[FR-05](../COVERAGE.md#fr-05)、[FR-07](../COVERAGE.md#fr-07)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)
- 优先级 / 阶段 / 层次：P0 / M4 / release
- 类型：release、functional、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 全新测试 Windows 环境，无 WSL 和开发工具链
- 发布包、后台二进制与支持矩阵已固定

#### 操作步骤

1. 按发布说明安装/解压并配置 FX-ECHO 与 FX-READY
2. 创建两会话、切换、attach 输入、读输出、发送并查询关联回复
3. 退出界面后重连，再停止测试 Run

#### 预期与不变量

- 仅靠发布说明与声明的运行依赖即可完成流程
- 身份、输入与消息阶段均有实际证据，不能用截图代替
- 安装说明未隐含开发机的 PATH、配置或数据文件

#### 需要的证据

- 全过程录屏
- 发布包/依赖哈希
- 输入/消息/Run 证据


### TC-REL-002

**缺少或不兼容的 psmux 给出可行动错误**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-013](../COVERAGE.md#fl-013)、[FL-019](../COVERAGE.md#fl-019)
- 需求：[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)、[ADR-0070](../../adr/0070-explicit-portable-update-and-rollback.md)
- 优先级 / 阶段 / 层次：P0 / M4 / release
- 类型：release、negative、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 干净测试安装；分别准备无后台、声明范围外版本和无法执行的二进制

#### 操作步骤

1. 启动产品并创建会话
2. 按说明补齐正确后台后再次操作

#### 预期与不变量

- 问题指向具体缺失/版本/执行失败，不显示伪启动成功
- 不因缺依赖清理用户已有会话或配置
- 恢复后可继续，实际探测版本与文件来源写入诊断

#### 需要的证据

- 错误截图/退出结果
- 依赖探测日志
- 修复前后记录


### TC-REL-003

**发行包与源码和依赖版本一致**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-019](../COVERAGE.md#fl-019)
- 需求：—
- 决策：[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)、[ADR-0070](../../adr/0070-explicit-portable-update-and-rollback.md)
- 优先级 / 阶段 / 层次：P0 / M4 / release
- 类型：release、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- 已生成候选发布物和构建清单

#### 操作步骤

1. 核对产品 commit、工具链、Go 模块锁定信息与后台来源
2. 对实际包计算校验值
3. 比较文档声明与二进制可观察版本

#### 预期与不变量

- 发行物能追溯到明确源码与依赖
- 文档所列版本与实际采用一致，目标版本表不能代替构建清单
- 校验值覆盖实际分发文件，不能沿用上一版值

#### 需要的证据

- 构建清单
- 包文件哈希
- 源码与依赖锁定文件链接


### TC-REL-004

**稳定 tag 与同版本字符串的开发构建不混用**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-019](../COVERAGE.md#fl-019)
- 需求：—
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)、[ADR-0070](../../adr/0070-explicit-portable-update-and-rollback.md)
- 优先级 / 阶段 / 层次：P0 / M0 / manual
- 类型：release、compatibility、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 准备已核验正式版清单与仅用于对照的开发构建信息

#### 操作步骤

1. 比对 tag/commit/二进制哈希与 Cargo/版本字符串
2. 检索测试配置是否开启 shared server 实验选项

#### 预期与不变量

- 不能只因版本字符串相同而接受为相同验证基线
- M0/发布证据明确对应正式版；实验 shared mode 不进入首版通过结论
- 替换依赖后相关能力必须重验

#### 需要的证据

- 来源清单与配置差异
- 受影响用例列表


### TC-REL-005

**升级中断后的数据与运行状态可恢复**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-013](../COVERAGE.md#fl-013)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M4 / release
- 类型：release、recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[TBD-UPGRADE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 仅使用待迁移数据库副本和测试 Run
- 已固定 schema 版本、备份与升级恢复方案

#### 操作步骤

1. 在数据迁移开始、写入中和提交边界中断安装/迁移
2. 重新启动新版本并检查恢复
3. 查看仍存活 Run 与未决消息

#### 预期与不变量

- 不静默丢失已确认记录或生成重复 Run
- 失败保留明确原版本/恢复路径，不能把半迁移数据当成功
- 存活终端与数据库版本变化分别核对

#### 需要的证据

- 迁移前后数据哈希/校验
- 故障点与恢复日志
- Run/消息记录


### TC-REL-006

**旧版本读取较新数据的回退边界**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M4 / release
- 类型：release、negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[TBD-UPGRADE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 存在新旧版本测试包和较新 schema 的测试副本
- 兼容/回退策略已固定

#### 操作步骤

1. 用旧版本打开新数据
2. 按发布回退说明恢复
3. 对照未被打开的只读副本

#### 预期与不变量

- 不支持的格式明确拒绝或走已验证兼容路径，不能自动清库降级
- 回退不会将旧消息送入不匹配的 Run
- 文档与实际可回退范围一致

#### 需要的证据

- 版本/schema 清单
- 副本前后差异
- 退出与恢复记录


### TC-REL-007

**卸载与清理的作用范围**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M4 / release
- 类型：release、negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-PACKAGING](../BUDGETS_AND_DECISIONS.md)、[TBD-ARCHIVE-DELETE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 仅在测试 VM；产品及用户 psmux 各有一组会话
- 已决定卸载是否保留配置、消息和日志，以及活 Run 处理

#### 操作步骤

1. 执行默认卸载，再执行可选清理流程
2. 记录所有受影响文件和进程

#### 预期与不变量

- 只处理用户选择且产品拥有的对象
- 保留/删除行为符合说明，不能隐式删除用户工作目录或其他后台
- 活 Run 的处置明确，不留下无法解释的孤儿或误杀

#### 需要的证据

- 卸载选项与文件清单差异
- 进程树
- 重新安装后的数据结果


### TC-REL-008

**安装和数据路径包含中文空格**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-001](../COVERAGE.md#fl-001)、[FL-016](../COVERAGE.md#fl-016)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-02](../COVERAGE.md#fr-02)、[FR-05](../COVERAGE.md#fr-05)、[FR-07](../COVERAGE.md#fr-07)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M4 / release
- 类型：release、compatibility、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 测试用户和目录包含中文、空格、括号；依赖路径不同于产品路径

#### 操作步骤

1. 安装到测试目录
2. 启动、自定义 cwd、发送文件正文、退出重开
3. 核对路径参数和数据归属

#### 预期与不变量

- 可执行入口与参数按明确边界传递
- 数据写入声明的产品目录，不因切换 cwd 漂移
- 不在错误目录生成重复配置或空数据库

#### 需要的证据

- 实际 argv/cwd
- 目录清单
- 重开后身份与消息记录


### TC-REL-009

**纯本地测试 agent 可离线完成管理流程**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-001](../COVERAGE.md#fl-001)、[FL-010](../COVERAGE.md#fl-010)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-09](../COVERAGE.md#fr-09)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M4 / release
- 类型：release、functional、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 只用本地确定性夹具；测试 VM 禁止外部网络

#### 操作步骤

1. 配置并启动两个本地 agent
2. 切换、读取、显式投递、后台重连和控制 CLI 调用

#### 预期与不变量

- 核心本地管理不依赖模型账户、远程服务或在线注册
- 与 agent 自己联网失败分别表达，不能混淆管理器可用性
- 不把本条推广为任何真实 agent 都能离线推理

#### 需要的证据

- 网络禁用证明
- 本地操作结果
- 依赖清单


### TC-REL-010

**许可证与迁入来源材料完整**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-019](../COVERAGE.md#fl-019)
- 需求：—
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)
- 优先级 / 阶段 / 层次：P0 / M0 / manual
- 类型：release、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-010](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- 已确定项目许可证和实际迁入文件清单

#### 操作步骤

1. 逐个核对上游提交、原路径、本地路径和修改说明
2. 检查必要版权头、许可证和 NOTICE 是否进入源码与发行包

#### 预期与不变量

- 每个迁入文件都可追溯，所需归属材料存在
- 未决许可证不能被技术选型 Accepted 状态跳过
- 未实际复制的项目只列参考，不能混入已复用清单

#### 需要的证据

- 来源清单
- 版权/NOTICE 文件
- 包内容清单


### TC-REL-011

**Charm v2 迁入后构建基线可复现**

- 功能：[FL-019](../COVERAGE.md#fl-019)、[FL-003](../COVERAGE.md#fl-003)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)
- 优先级 / 阶段 / 层次：P0 / M0 / release
- 类型：release、compatibility、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已迁入选定 UI 切片并生成锁定文件
- 固定受支持构建工具链

#### 操作步骤

1. 从干净源码构建，不依赖开发机未跟踪文件
2. 检查实际使用的 UI import/事件/视图接口
3. 运行列表、表单、attach 返回的 Windows 回归

#### 预期与不变量

- 选定 UI 使用已记录稳定 v2 基线，不能仅改版本号而混用旧事件合同
- 无未经记录的双主版本 UI 运行路径
- 构建成功不替代 Windows 交互回归

#### 需要的证据

- 构建依赖清单
- 来源到适配映射
- Windows 回归结果


### TC-REL-012

**平台声明只覆盖真实验证组合**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-016](../COVERAGE.md#fl-016)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-02](../COVERAGE.md#fr-02)、[FR-05](../COVERAGE.md#fr-05)、[FR-07](../COVERAGE.md#fr-07)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)
- 优先级 / 阶段 / 层次：P0 / M4 / manual
- 类型：release、compatibility、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- 支持矩阵与所有适用测试报告已生成

#### 操作步骤

1. 逐项比较 README/发布说明中的 OS/架构/终端/shell/agent 与运行记录
2. 检查 failed/blocked/not_run 的处理

#### 预期与不变量

- 没有用 Linux 编译、截图或上游支持声明替代 Windows 实机结果
- 不支持和实验组合明确列出
- 没有运行的组合不能因同一语言或后台接口而标通过

#### 需要的证据

- 发布声明对照表
- 环境清单与报告链接


### TC-REL-013

**低内存宣传与测量范围一致**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-015](../COVERAGE.md#fl-015)
- 需求：—
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M4 / manual
- 类型：release、performance、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- 有候选发布物的 0/1/5/10 会话测量及预算

#### 操作步骤

1. 审查性能结论是否包含后台、shell/控制台与 agent 分组
2. 比较直接运行同 agent 的对照
3. 核对采样窗口、日志保留、预热设置和负载

#### 预期与不变量

- 结论注明场景和测量范围，不用空会话 TUI 数字代表完整多会话总成本
- 不将工作集简单相加冒充无共享重复的系统物理内存
- 预算未固定或测试未完成时不发布达标结论

#### 需要的证据

- 完整测量报告
- 宣传数字来源
- 预算审批/记录


### TC-REL-014

**发布依赖升级触发相应回归**

- 功能：[FL-019](../COVERAGE.md#fl-019)、[FL-017](../COVERAGE.md#fl-017)
- 需求：—
- 决策：[ADR-0005](../../adr/0005-dependency-lifecycle-and-upgrades.md)
- 优先级 / 阶段 / 层次：P1 / M4 / release
- 类型：release、compatibility、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 替换某个 UI/后台/存储依赖的测试分支
- 新旧构建均可追溯

#### 操作步骤

1. 按依赖影响选择启动、输入、CJK、attach 返回、持久化和消息状态用例
2. 在声明支持的 Windows 组合执行
3. 对照失败与行为差异

#### 预期与不变量

- 升级有明确来源和受影响功能证据
- 没有因只是补丁版本而跳过与风险相关的回归
- 改变架构/依赖代际时按 ADR 记录，失败不沿用旧通过报告

#### 需要的证据

- 变更依赖清单
- 选中用例与结果
- 行为差异说明


### TC-REL-015

**诊断和恢复说明可由新用户完成**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M4 / manual
- 类型：release、usability、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 干净测试用户，只提供发布文档
- 预先准备缺命令、错误 cwd、后台退出和 unknown 消息四类故障

#### 操作步骤

1. 执行说明中的定位与恢复步骤
2. 记录是否需要作者未写出的知识或本机文件

#### 预期与不变量

- 每类问题可定位到配置、后台、Run 或投递证据
- unknown 的恢复说明不诱导无条件重发
- 能区分查看历史、接回、重启和原生 resume

#### 需要的证据

- 新用户操作记录
- 缺漏说明列表
- 恢复后的实际状态


### TC-REL-016

**相同包多次启动与配置目录选择**

- 功能：[FL-017](../COVERAGE.md#fl-017)、[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)、[FL-001](../COVERAGE.md#fl-001)、[FL-016](../COVERAGE.md#fl-016)、[FL-020](../COVERAGE.md#fl-020)
- 需求：[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)、[FR-01](../COVERAGE.md#fr-01)、[FR-14](../COVERAGE.md#fr-14)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0055](../../adr/0055-scoped-workbench-instance-and-owned-window.md)
- 优先级 / 阶段 / 层次：P0 / M4 / release
- 类型：release、concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[TBD-INSTANCE-SCOPE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 产品实例/数据目录/Host 单实例策略已固定
- 只操作测试目录

#### 操作步骤

1. 同时启动两个 TUI 和多个管理 CLI 指向同一数据目录
2. 再启动指向另一数据目录的实例
3. 观察 Host 与队列消费者
4. 在 ADR-0055 Windows 工作台子范围，对同一用户／交互登录会话／backend/runtime/metadata 同时启动五次并顺序重复启动，记录 owner、端点及猫的实际数量；管理 CLI 仍独立工作。
5. 在不同作用域的两个专属 WT 窗口之间切换，从另一受控窗口点猫／重复启动，核对原 HWND、Session/Run 和未提交草稿；最小化后重试。
6. 正常退出及明确终止测试自己的 TUI／实例 helper 后重新启动；分别制造初始化中、端点超时和取消，核对不误报在线／成功且不因此另建工作台。

#### 预期与不变量

- 同一作用域不会因竞争产生未协调的双消费者
- 不同目录/namespace 不混淆 Session/Run 与配置
- 并发允许或拒绝按已记录策略执行，错误明确且无误投
- Windows 工作台参与本协议的同作用域只有一个 UI／猫；明确 token 绑定而非前台猜测，失败提示且不注入终端输入。
- UI 退出释放实例占用，仍存活的原 Run／草稿不因唤起或重开改变；旧版、不支持的多原生 Tab／混合权限仍明确分列。

#### 需要的证据

- 实例/endpoint/目录映射
- 消费者日志
- 消息与输入证据

说明：ADR-0055 固定的是 Windows UI 单实例与纯窗口唤起，不固定 M3 Host／双消费者的完整发行合同；原 needs_decision／设计执行状态保持，局部结果另存报告。


### TC-DEFER-001

**Tab 内终端与管理导航的焦点合同**

- 功能：[FL-018](../COVERAGE.md#fl-018)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)、[ADR-0016](../../adr/0016-tab-workspaces-and-input-ownership.md)、[ADR-0018](../../adr/0018-native-copy-search-and-host-clipboard.md)、[ADR-0028](../../adr/0028-explicit-visible-pane-copy.md)、[ADR-0029](../../adr/0029-native-copy-pointer-gestures.md)、[ADR-0030](../../adr/0030-direct-terminal-drag-selection.md)、[ADR-0031](../../adr/0031-program-left-mouse-routing.md)
- 优先级 / 阶段 / 层次：P1 / M2 / manual
- 类型：usability、compatibility、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-009](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- ADR-0014 已接受，Tab 内终端实现与 Windows 屏幕／输入合同已可运行；本条仍为 designed/not_run。
- 记录产品、固定上游版本、Windows／终端／shell／agent／psmux 哈希和快捷键配置；使用隔离合成数据。

#### 操作步骤

1. 打开普通 shell A 与 agent B 的 Tab，分别使用各自终端输入区，保留未提交文本。
2. 在 Tab 导航、按需表单与子终端间切换焦点，执行 A→B→A、重命名、排序、关闭／重新打开视图。
3. 覆盖 IME、鼠标、Tab 键、方向键、Enter、Ctrl+C 和保留管理快捷键；最后显式停止目标。
4. 用鼠标拖动可见 Tab，并在管理焦点用 [／] 移动活动 Tab；在终端验证同名字符；输入排空／切换期间再排序，释放／其他输入／resize 后不沿用旧手势。
5. 管理 v 进入后端选择／搜索，验证空格／整行／Enter、/／? 和取消；观察中文长提示、原生选区／命中／复制光标，核对外层终端剪贴板中仅合成内容。
6. 管理 V 复制所选 B、活动 Tab 保持 A；穿插原生复制事件、切 Tab、原 Run 替换、重复同代次旧完成及超限／C1／损坏文字；观察实际 OSC52 请求与子程序输入。
7. 原生复制模式内左键按下／拖动／释放，越过 Tab／侧栏边缘；穿插切 Tab、resize、画面尺寸变小再恢复、模式退出、搜索提示及重复释放，核对原 pane 与原队列。
8. 未声明 WantsMouse 的实时终端单击／微动及直接拖选；延迟原生模式帧，观察退出再恢复，穿插外部模式／鼠标声明变化和切 Tab；对比有／无交接能力连接。
9. 声明 WantsMouse 的程序左按／拖／释放，越过 Tab／框，压力填满等待队列；穿插键／粘贴／滚轮／切 Tab／缩放与外部模式／屏幕增长后恢复，核对原程序和新 Run 独立输入。

#### 预期与不变量

- 每个输入只被当前消费者处理，后台会话不收到键盘输入；不新增日常聊天输入框。
- 终端操作时 Tab 导航仍可达，无需先 detach；切换保留原 Run、屏幕和未提交输入。
- 关闭 Tab 只关闭视图，重开接回存活 Run；显式停止单独操作。普通 shell 命令返回后仍可输入。
- 旧全屏 attach 证据不能代替本能力；缺少真实 agent 时保留该样本待测。
- 排序只改 View 顺序，活动连接、Session/Run 与切换目标不变，不自动提交；不同标题长度不让指针下槽位反复变化，边缘不循环。当前内存顺序与持久保存的实现状态分别记录。
- 原草稿不因搜索／选择提交；复制事件与屏幕合并分离并有界，跨 Tab 在途提交串行，旧事件／重复完成不重开读取器。仅已发 OSC52 不标为系统已保存，不读取或回传用户剪贴板。
- 新鲜捕获只读原 pane 当前页，不用旧预览、历史或 manager 装饰；空／超限／截断／损坏文字不提交，C1 控制沿源规范化。V 终端焦点保持原文字；两个复制来源最多单 host 提交，独立令牌拒绝旧回执，排队目标重核，未发送旧代次不重放。请求输出不是系统剪贴板已保存，原生桌面另验。
- 仅已聚焦原生内容开始无修饰左键手势；内容零基坐标，经原 FIFO，不伪造键盘。手势越界不转管理操作；模式／视图／尺寸失效清除，恢复不续接，不合成释放或自动复制。CopyMode 观察不是锁，外部并发模式解释与实体鼠标／系统剪贴板独立记录；正常终端直接拖选与程序鼠标不以本地通过认定已验收。
- 单击／同格微动不进入模式、不发输入；首次实际运动带原锚点交接原 pane，原 FIFO 持续且不向新 Run 续接。写入接受不冒充已进入 CopyMode，观察退出或原锚点失效后不恢复；缺少能力不降级为文字／按键，原有原生模式指针仍可用。首次运动冻结与源客户端不同，真实后端／实体鼠标、程序鼠标和系统剪贴板分别验证。
- 左键只到原 pane 内容，重复事件不重发，不改为选区／文字／按键。程序按住期间保留一个释放位置，主动操作前在原 FIFO 释放一次；外部模式／几何失效不转投且结果未知。CopyMode 仍选区优先且取消不合成复制；VT／Win32、其他按钮／修饰／悬停及实体鼠标、系统剪贴板分别实测。

#### 需要的证据

- Windows Tab 切换与焦点录屏、输入接收独立记录、Session/Run 对照。
- 固定上游视觉对照、构建身份和实际失败／限制。

说明：编号 TC-DEFER-001 为历史稳定 ID；ADR-0014 已将其纳入首版，不再表示后置。具体快捷键与技术路径按 O-009 验证。


### TC-DEFER-002

**Tab 内终端资源与 VT 完整性**

- 功能：[FL-018](../COVERAGE.md#fl-018)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-11](../COVERAGE.md#fr-11)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：performance、compatibility、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-009](../../DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- ADR-0014 已接受，Tab 内屏幕模型与输入路径可运行；记录支持序列、资源口径及未定预算。本条仍为 designed/not_run。

#### 操作步骤

1. 用同一普通 shell／CLI 和 agent 分别对照直接终端、辅助全屏 attach 与 Tab 内路径；固定尺寸、输出负载和版本。
2. 覆盖 alternate screen、宽字符、resize、滚动和高输出
3. 测量屏幕模型重复缓存
4. 覆盖 0/1/5/10 会话和活动／后台 Tab，检查切换、关闭／重开、持续输出后的有界缓存。

#### 预期与不变量

- 已声明 VT 与输入行为可复现
- 内存和响应成本单独列明，不沿用当前全屏模式数据
- 未支持序列按声明处理，不造成其他会话状态污染
- 未定预算只报告测量与待判定；旧 attach 资源数据不能作为 Tab 路径结果。

#### 需要的证据

- 两路径对照记录
- 资源曲线与缓存对象
- 屏幕/输入证据

说明：TC-DEFER-002 保留原 ID，ADR-0014 将能力转为首版；终端集成合同与数值预算仍待决定。


### TC-PROP-001

**随机生命周期操作保持 Session/Run 不变量**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：property、state、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-LIFECYCLE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 生成创建、重命名、排序、关闭视图、attach、停止、重启和重连序列
2. 在每步后比较参考身份映射与实际记录；对失败序列做缩减

#### 预期与不变量

- 视图/名字变化不替换稳定身份；真正重启产生新 Run
- 没有对错误 Run 操作；终端、shell、agent 的存活证据不混为一体
- 未决停止/重启策略固定后才加入相应转移，不能让参考模型擅自作产品决定

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。


### TC-PROP-002

**随机参数编码保持直接启动边界**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-016](../COVERAGE.md#fl-016)
- 需求：[FR-01](../COVERAGE.md#fr-01)
- 决策：[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：property、fuzz、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-004](../../DECISIONS.md)、[TBD-CONFIG](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 生成受支持 Unicode 参数、空字符串、引号、空格和末尾反斜杠组合
2. 经配置保存、加载和直接启动交给 FX-ECHO
3. 将非法字符与超过已定长度的输入另组验证拒绝

#### 预期与不变量

- 受支持参数的个数、顺序和值与生成器预期一致
- 不由内容触发隐式 shell 解释
- 非法输入按明确规则拒绝且不产生目标进程；不把拒绝测试混入合法往返断言

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。


### TC-PROP-003

**分片方式不改变有效 UTF-8/VT 的最终屏幕**

- 功能：[FL-002](../COVERAGE.md#fl-002)、[FL-007](../COVERAGE.md#fl-007)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-06](../COVERAGE.md#fr-06)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：property、fuzz、boundary、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 选择有限且已支持的确定性 VT 字节流
2. 分别整块、逐字节、随机分块送入解析/屏幕边界
3. 在相同终端尺寸和相同已完成序列边界比较结果

#### 预期与不变量

- 同一有效字节流最终的字符、光标和属性状态一致
- 不因读取分片而插入多余字符或丢失已完成控制序列
- 查询响应或时间相关序列单独建模，不要求所有副作用逐帧完全一致

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。


### TC-PROP-004

**随机消息事件序列不夸大证据或改变目标**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：property、state、recovery、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 生成入队、重复请求、ready/busy、Run 更换、写入故障、确认与重启的受控序列
2. 对每个 message_id 保留独立来源证据和实际输入观察
3. 缩减违反不变量的轨迹

#### 预期与不变量

- 没有证据的请求不升级 accepted/completed
- 目标 Run 变化不会令旧请求进入新上下文
- 可能写出的未知请求不会因重试自动重放；已持久化证据不会仅因断开被抹掉

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。


### TC-PROP-005

**配置和元数据往返保持身份与参数**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：property、fuzz、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)、[TBD-CONFIG](../BUDGETS_AND_DECISIONS.md)、[TBD-UPGRADE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 生成符合已定 schema 的配置和 Session/Run 记录
2. 保存、关闭、重新加载并比较语义值；执行支持版本之间迁移
3. 另组生成错误类型、重复键和损坏内容

#### 预期与不变量

- 合法往返保持对外身份、参数边界、引用关系和已声明字符处理
- 不把损坏或未来格式默认为空成功配置
- 原始故障样本与恢复结果保留，不能因数据不可恢复而静默重建掩盖问题

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。


### TC-PROP-006

**随机异常终端字节有界且不产生管理操作**

- 功能：[FL-002](../COVERAGE.md#fl-002)、[FL-013](../COVERAGE.md#fl-013)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-02](../COVERAGE.md#fr-02)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：fuzz、negative、security、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-TERMINAL-CONTROL](../BUDGETS_AND_DECISIONS.md)、[O-012](../../DECISIONS.md)
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-VT](../FIXTURES_AND_FAULTS.md#fx-vt)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 按固定种子变异控制序列长度、参数、终止符与 UTF-8 字节
2. 限制每个样本输入量和运行时长，观察解析与预览；对超时样本留档
3. 发送已声明终止/复位后重新执行正常管理操作

#### 预期与不变量

- 异常样本不触发管理侧命令、发送或无关文件改动
- 资源受已固定解析/缓存限额约束，超时可归因并可恢复
- 对未终止序列不硬性要求内部普通文本立即显示

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。


### TC-PROP-007

**历史分页组合与固定快照范围一致**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：property、boundary、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 生成固定 Run 的有限不可变输出范围和已知缺口
2. 用多个合法页大小遍历相同范围，再与一次受限读取的参考结果比较
3. 随机替换为其他 Run 的游标或令一页记录轮转

#### 预期与不变量

- 同一固定范围的保留内容在分页后不重复、不串线；实际缺口明确表示
- 错误 Run 游标不读取另一个 Run 的内容
- 轮转后结果声明失效/缺口，不能用新文件同偏移补成假连续

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。


### TC-PROP-008

**并发发送调度不能把正文和提交交错**

- 功能：[FL-008](../COVERAGE.md#fl-008)、[FL-009](../COVERAGE.md#fl-009)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-07](../COVERAGE.md#fr-07)、[FR-08](../COVERAGE.md#fr-08)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M3 / integration
- 类型：property、concurrency、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已实现对应模块与确定性夹具；当前仅设计，未执行
- 使用隔离数据与固定种子；先有可人工审查的参考模型，模型不能直接调用被测实现计算预期

#### 操作步骤

1. 生成两个发送者、用户 attach 和两个竞争 Host 的屏障排列
2. 每个正文带不同长度和独立校验标记；按产品允许的调度写入
3. 对失败排列缩减并保存可重复屏障序列

#### 预期与不变量

- 每个实际投递的正文与其提交组成完整操作，不与另一正文混拼
- 输入被占用或所有权不明时按规则等待；旧所有者恢复不能继续双写
- 仅要求符合已声明的顺序与不变量，不偷偷要求所有独立请求按壁钟全局 FIFO

#### 需要的证据

- 随机种子、生成器版本、完整失败输入与最小化后的复现
- 参考模型结果、被测结果与独立观察证据
- 发现真实问题后在 Windows 对应路径增加确定性回归；模拟通过不代替原生验收

#### 参数变体

- campaign：`["固定回归种子", "新增探索种子；失败后保存"]`

说明：属性/模糊测试是计划；探索次数和超时在测试轮次记录，不计为额外通过用例。
