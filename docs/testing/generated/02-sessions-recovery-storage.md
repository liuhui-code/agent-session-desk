# 会话、后台、恢复与存储

> 从 [JSON 源](../cases/02-sessions-recovery-storage.json) 自动生成。请修改源并运行 `python tools/test_catalog.py render`。

**全部为测试设计，当前未执行产品或 Windows 测试。** ready 仅表示预期依据明确。

[总索引](../CASE_INDEX.md) · [覆盖表](../COVERAGE.md) · [字段约定](../CASE_FORMAT.md)

| ID | 场景 | 优先级 | 阶段 | 规范状态 |
| --- | --- | --- | --- | --- |
| [TC-SESSION-001](#tc-session-001) | 空会话集合中的浏览与退出不创建运行 | P1 | M2 | 依据明确 |
| [TC-SESSION-002](#tc-session-002) | 相同启动模板创建不同会话身份 | P0 | M2 | 依据明确 |
| [TC-SESSION-003](#tc-session-003) | 重命名不改变身份与运行 | P1 | M2 | 依据明确 |
| [TC-SESSION-004](#tc-session-004) | 重复名称按已决规则处理而不覆盖会话 | P1 | M2 | 待决策 |
| [TC-SESSION-005](#tc-session-005) | 名称边界与特殊字符不会进入后台命令 | P1 | M2 | 待决策 |
| [TC-SESSION-006](#tc-session-006) | 搜索规则与结果寻址分离 | P1 | M2 | 依据明确 |
| [TC-SESSION-007](#tc-session-007) | 排序过滤变化不把选中目标换成同行号会话 | P0 | M2 | 依据明确 |
| [TC-SESSION-008](#tc-session-008) | 移动分组不改变运行与历史归属 | P1 | M2 | 依据明确 |
| [TC-SESSION-009](#tc-session-009) | 关闭视图再打开仍指向原会话 | P0 | M2 | 依据明确 |
| [TC-SESSION-010](#tc-session-010) | 连续切换仅允许按需创建 attach 客户端 | P0 | M2 | 依据明确 |
| [TC-SESSION-011](#tc-session-011) | 已退出会话归档保留可识别的历史 | P1 | M2 | 依据明确 |
| [TC-SESSION-012](#tc-session-012) | 运行中会话归档遵守明确处置策略 | P1 | M2 | 待决策 |
| [TC-SESSION-013](#tc-session-013) | 删除范围明确且不波及其他记录 | P2 | M2 | 待决策 |
| [TC-SESSION-014](#tc-session-014) | 并发重命名与过期列表操作不误作用于替代会话 | P0 | M2 | 依据明确 |
| [TC-LIFE-001](#tc-life-001) | detach 后原任务继续运行 | P0 | M0 | 依据明确 |
| [TC-LIFE-002](#tc-life-002) | 正常退出管理器保留终端后台 | P0 | M2 | 依据明确 |
| [TC-LIFE-003](#tc-life-003) | TUI 被强制结束不连带终止托管任务 | P0 | M2 | 依据明确 |
| [TC-LIFE-004](#tc-life-004) | attach 客户端异常退出不等于 agent 退出 | P0 | M2 | 依据明确 |
| [TC-LIFE-005](#tc-life-005) | agent 正常退出后终端不再被视为活跃 agent | P1 | M1 | 依据明确 |
| [TC-LIFE-006](#tc-life-006) | 非零退出保留失败原因与既有输出 | P1 | M1 | 依据明确 |
| [TC-LIFE-007](#tc-life-007) | agent 退出而 shell 留存时区分两层存活 | P0 | M1 | 依据明确 |
| [TC-LIFE-008](#tc-life-008) | 后端 session 存在但缺少 agent 证据时保持未知 | P0 | M2 | 依据明确 |
| [TC-LIFE-009](#tc-life-009) | 明确重启生成新 Run 并保留历史边界 | P0 | M2 | 依据明确 |
| [TC-LIFE-010](#tc-life-010) | PID 复用不能让旧 Run 关联或停止新进程 | P0 | M2 | 依据明确 |
| [TC-LIFE-011](#tc-life-011) | 后端 pane 标识复用不继承旧 Run | P0 | M2 | 依据明确 |
| [TC-LIFE-012](#tc-life-012) | 同 cwd 同命令的会话重连保持各自身份 | P0 | M2 | 依据明确 |
| [TC-LIFE-013](#tc-life-013) | 正常停止只结束明确选中的 Run | P0 | M2 | 依据明确 |
| [TC-LIFE-014](#tc-life-014) | 拒绝退出与子进程的停止升级策略 | P0 | M2 | 待决策 |
| [TC-LIFE-015](#tc-life-015) | 停止旧 Run 与新 Run 启动竞态不误杀新运行 | P0 | M2 | 依据明确 |
| [TC-LIFE-016](#tc-life-016) | 重复停止不会操作另一 Run | P1 | M2 | 依据明确 |
| [TC-LIFE-017](#tc-life-017) | 并发重启按已决规则保持单一当前 Run | P0 | M2 | 待决策 |
| [TC-LIFE-018](#tc-life-018) | 单会话后端 server 故障不错误清理其他会话 | P0 | M2 | 依据明确 |
| [TC-STORE-001](#tc-store-001) | 首次初始化不产生虚假历史或覆盖其他目录 | P1 | M1 | 依据明确 |
| [TC-STORE-002](#tc-store-002) | 元数据 Unicode 与特殊字符持久化往返 | P1 | M1 | 依据明确 |
| [TC-STORE-003](#tc-store-003) | 语法损坏的 TOML 不静默替换为可执行配置 | P1 | M1 | 依据明确 |
| [TC-STORE-004](#tc-store-004) | 缺失或空配置按明确初始化策略处理 | P1 | M1 | 待决策 |
| [TC-STORE-005](#tc-store-005) | 未知配置键和更高 schema 版本有可预测结果 | P1 | M1 | 待决策 |
| [TC-STORE-006](#tc-store-006) | 配置保存中断不留下无提示的半份有效配置 | P0 | M1 | 待决策 |
| [TC-STORE-007](#tc-store-007) | 配置只读或文件被占用时不显示保存成功 | P1 | M1 | 依据明确 |
| [TC-STORE-008](#tc-store-008) | 数据库只读时元数据变更不能伪装提交 | P0 | M1 | 依据明确 |
| [TC-STORE-009](#tc-store-009) | 数据库容量不足时保留已提交记录 | P0 | M1 | 依据明确 |
| [TC-STORE-010](#tc-store-010) | 数据库 busy 与锁等待按已定有界策略处理 | P1 | M2 | 待决策 |
| [TC-STORE-011](#tc-store-011) | 多个入口并发写不同会话不丢失已提交更新 | P0 | M2 | 依据明确 |
| [TC-STORE-012](#tc-store-012) | 损坏数据库不被静默重建成空成功状态 | P0 | M2 | 依据明确 |
| [TC-STORE-013](#tc-store-013) | 旧产品读取未来数据库版本不会破坏数据 | P0 | M4 | 待决策 |
| [TC-STORE-014](#tc-store-014) | 数据库正常迁移保持身份和引用关系 | P0 | M4 | 待决策 |
| [TC-STORE-015](#tc-store-015) | 数据库迁移中断按已定恢复合同处理 | P0 | M4 | 待决策 |
| [TC-STORE-016](#tc-store-016) | 产品回退时兼容限制明确且不逆向破坏数据 | P1 | M4 | 待决策 |
| [TC-STORE-017](#tc-store-017) | 日志轮转边界按字节预算正确保留范围 | P1 | M2 | 待决策 |
| [TC-STORE-018](#tc-store-018) | 读取与轮转并发不拼接不同 Run 的文件 | P0 | M2 | 待决策 |
| [TC-STORE-019](#tc-store-019) | 日志磁盘满或不可写时明确记录缺口 | P1 | M2 | 待决策 |
| [TC-STORE-020](#tc-store-020) | 无自有消费者期间只承诺后端保留的历史 | P0 | M2 | 依据明确 |
| [TC-STORE-021](#tc-store-021) | 清理日志不跟随外部路径删除无关数据 | P0 | M2 | 待决策 |
| [TC-STORE-022](#tc-store-022) | 配置变更只按已定时机影响后续运行 | P1 | M2 | 待决策 |
| [TC-REC-001](#tc-rec-001) | 创建开始记录后后端创建失败可以辨认 | P1 | M1 | 依据明确 |
| [TC-REC-002](#tc-rec-002) | 后端已创建但元数据提交前崩溃的恢复 | P0 | M1 | 待决策 |
| [TC-REC-003](#tc-rec-003) | 创建意图已落盘但进程未启动时崩溃的恢复 | P1 | M1 | 待决策 |
| [TC-REC-004](#tc-rec-004) | 外部 psmux 会话与名称碰撞不会被误领养或清理 | P0 | M2 | 依据明确 |
| [TC-REC-005](#tc-rec-005) | 停止命令返回丢失时依据实际状态恢复 | P0 | M2 | 依据明确 |
| [TC-REC-006](#tc-rec-006) | 界面关闭期间 agent 退出后重开不自动续跑 | P0 | M2 | 依据明确 |
| [TC-REC-007](#tc-rec-007) | 重启测试机器后历史记录不伪装成可接回活进程 | P0 | M2 | 依据明确 |
| [TC-REC-008](#tc-rec-008) | 持久化期间虚拟机断电后的恢复边界 | P2 | M4 | 待决策 |
| [TC-REC-009](#tc-rec-009) | 睡眠唤醒后的会话状态根据证据重新核对 | P1 | M2 | 依据明确 |
| [TC-REC-010](#tc-rec-010) | 历史 cwd 不可用时明确重启失败而不切换到其他目录 | P1 | M1 | 依据明确 |
| [TC-REC-011](#tc-rec-011) | 系统墙钟跳变不改变 Run 身份或事件因果 | P1 | M2 | 依据明确 |
| [TC-REC-012](#tc-rec-012) | 不支持或失败的原生 resume 不伪装成历史恢复成功 | P1 | M2 | 依据明确 |

### TC-SESSION-001

**空会话集合中的浏览与退出不创建运行**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：functional、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 隔离数据目录没有 Session/Run；后台没有属于本产品的 session。

#### 操作步骤

1. 打开管理器，依次执行空列表搜索、取消和刷新。
2. 退出并重新打开，读取会话与后台进程清单。

#### 预期与不变量

- 空态与无搜索结果可区分，未创建虚假 Session 或 Run。
- 浏览、刷新及退出不启动 agent，后台无新增托管对象。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-002

**相同启动模板创建不同会话身份**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：functional
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 准备同一启动模板和相同 cwd；夹具在每次启动输出唯一 nonce。

#### 操作步骤

1. 用该模板明确创建会话 A、B。
2. 分别记录 session_id、run_id、后端目标与启动 nonce。
3. 在 A、B 间切换并读取各自输出。

#### 预期与不变量

- A、B 的 Session 和 Run 身份不同，不能按模板名或 cwd 合并。
- 两份输出的 nonce 与目标一致，切换不产生第三次 agent 启动。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-003

**重命名不改变身份与运行**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0021](../../adr/0021-persistent-session-display-names.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 正在运行；已记录其 session_id/run_id/启动 nonce。

#### 操作步骤

1. 管理焦点 r 打开原名称预填的弹窗，改为中文名称；取消一次，再明确 Enter 保存并再次改名。
2. 按原 session_id 查询并进入会话。
3. 退出管理器后重新打开核对记录。

#### 预期与不变量

- Session/Run 与启动 nonce 保持不变，名称修改持久化。
- 旧名称不会成为仍可误指向其他会话的寻址身份。
- 重命名不向 agent 输入额外字符。
- 列表、Tab、搜索及离线历史显示当前名称；原启动快照保持不变，Ctrl+R 刷新或 UI 重开不回退名称。
- 选中后台会话改名不激活该 Tab；弹窗等待已接受写入排空，确认字符不进入原终端。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-004

**重复名称按已决规则处理而不覆盖会话**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0021](../../adr/0021-persistent-session-display-names.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：boundary、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-NAMING](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- ADR-0021 已允许显示名重复并规定短身份标记；创建的完整名称合同仍由 TBD-NAMING 决定，存在名为 review 的 A。

#### 操作步骤

1. 尝试创建同名 B，并尝试把另一会话重命名为 review。
2. 若规则允许同名，分别用稳定 ID 查询；若拒绝，核对原数据。

#### 预期与不变量

- 重命名允许同名，Tab／列表／搜索的可见名称碰撞时显示可区分的 Session ID 标记；创建部分按另定合同核对。
- 不以名称作为 upsert 键覆盖 A 的 Run、历史或消息关系。
- 不能因重名使操作目标变得不可区分。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-005

**名称边界与特殊字符不会进入后台命令**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0021](../../adr/0021-persistent-session-display-names.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：boundary、negative、security
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-NAMING](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 已固定名称允许字符、空白归一化和长度计量规则 B_NAME。
- 使用测试目录与仅回显输入的夹具，不执行名称中的文本。
- 新增重命名按 ADR-0021 拒绝控制字符／行分隔、无效 UTF-8、空白名及超过 128 code point／512 字节；原始输入最多 4096 字节，不做 NFC/NFD 归一化。

#### 操作步骤

1. 对参数矩阵逐项新建或重命名会话。
2. 比较保存值、显示值、后端目标和夹具收到的字节。

#### 预期与不变量

- 每项明确接受或拒绝，按已定 B_NAME 检查边界；不静默把两个名称合并。
- 名称始终作为数据处理，不因引号、分号或控制字符改变启动/停止目标。
- 被拒绝时原会话及 Run 不变。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 名称数据：`["空串", "纯空白", "首尾空白", "中文/emoji/组合字符", "B_NAME-1、B_NAME、B_NAME+1", "换行、制表、ESC", "单引号、双引号、分号、美元符号"]`


### TC-SESSION-006

**搜索规则与结果寻址分离**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)、[ADR-0016](../../adr/0016-tab-workspaces-and-input-ownership.md)、[ADR-0020](../../adr/0020-local-session-search.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：functional、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 按 ADR-0020 的当前作用域实时 Session 名称／程序路径查询规则建立名称相近的 A、B、C；历史／分组／cwd 不在本轮搜索范围。

#### 操作步骤

1. 按中文片段、大小写差异、组合字符和空查询分别搜索。
2. 选中返回项后进入会话；清空查询再重复。
3. 验证 Unicode 小写、首尾空白裁剪、1–3 rune 精确子串、4–5 rune 一个编辑、更长两个编辑；精确子串优先、同分稳定，不做 NFC/NFD 归一化。
4. 搜索期间模拟发现重排和选中 Run 替换：前者保留原选择，后者清除选择直到明确重选；以慢写入验证进入弹窗前先完成已接受输入。
5. 原生 Tab 中保留 shell 草稿后搜索、取消、零结果 Enter 与选择；核对普通／中文查询、120／63／49 列和界面重开。

#### 预期与不变量

- 结果集合与已定规则一致，选中项使用稳定 ID 寻址。
- 查询文本不进入 agent 输入，清空搜索不启动新 Run。
- 无结果时不能沿用不可见旧选中行执行停止或进入。
- 终端／原生复制模式的斜杠仍交给目标；管理搜索窗消费自己的键鼠／粘贴，q 是查询正文，Esc/Ctrl+C 返回管理焦点。
- Run 替换不将 Enter 自动转到首项或同名新 Run；零结果／不可交互项不进入，不自动启动／停止。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 查询集合：`["空串", "全匹配", "中文子串", "大小写变体", "组合字符变体", "无匹配项"]`


### TC-SESSION-007

**排序过滤变化不把选中目标换成同行号会话**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：state、concurrency
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B、C 运行且输出不同 nonce；当前选中 B。

#### 操作步骤

1. 在操作前触发排序、分组或过滤结果更新，改变 B 的行号。
2. 执行进入或停止动作，并读取目标操作记录。

#### 预期与不变量

- 动作绑定明确选中的 session_id，而不是旧行索引。
- 如果 B 已不在可操作集合中，必须重新确定目标或拒绝动作，不回退为同一行的 C。
- 未选择的会话不收到输入或停止命令。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-008

**移动分组不改变运行与历史归属**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0023](../../adr/0023-persistent-groups-and-membership.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 在项目组甲运行，已有保存输出；另有项目组乙。
- 按 ADR-0023 建立稳定 Group ID、父组和可选默认目录；原 Run 与原 cwd 已独立记录。

#### 操作步骤

1. 用 g 创建根组和子组，Tab 切名称／目录、↑↓ 切层级；同父同清理名同目录返回原组，不同目录拒绝。
2. 选中折叠组标题按 r，预填旧名、光标在末尾，Tab 不进入隐藏目录；分别改根组、子组和默认组名称，取消一次并退出重开。
3. 将 A 移入乙并更改分组显示顺序。
4. 选中展开子树之后的根组，用 Shift/Ctrl+↑↓、K/J、+/- 与同父兄弟交换；再交换两个子组，测试默认组及首尾边界，退出重开并核对备份。
5. 重新打开列表，按 A 的 session_id 读取记录并 attach。
6. 用两个 UI 或连接从同一成员版本并发移动，另执行甲→乙→甲后提交原旧窗口；退出重开并检查名称、组树、成员及原 Run。
7. 两个连接读取同一组版本后改名；执行改名再改回后提交原旧版本，另尝试同级清理后重名及不同父组同名。
8. 两连接从同一完整兄弟快照并发排序；在旧快照之后改邻居名／新增兄弟／顺序改回，再提交旧编辑。UI 保存期间请求退出并分别阻塞排序、已接受输入和偏好保存，另注入偏好／终端失败。

#### 预期与不变量

- 运行、后端目标、历史所属 Session 不变。
- 列表只显示一个 A，不产生复制 Run。
- 分组变化不会暗中修改既有 Run 的 cwd。
- 成员关系归属逻辑 Session，移动使用完整原 Session/Run/Backend 和预期成员版本；ABA 或过期窗口拒绝覆盖。
- 组内新建可预填／覆盖默认目录；Run 与初始归属一并提交，外部终端创建不假定原子。
- 组改名只改显示名和组版本；稳定 Group ID、父组、顺序、目录、子组、成员版本及 Session/Run/Tab 不变，折叠按 ID 保留，子组展示路径随祖先名更新。
- 组改名同值不增加版本，同级名称冲突和过期／ABA 编辑拒绝；不同父组同名独立，默认组可改名但身份及位置保持。较旧回执／发现不回退已观测新版本，弹窗文字不进入终端，失败不自动重试。
- 组排序只在同父兄弟间进行，展开子树不成为相邻排序目标；ordinal 归一化后交换，移动双方和实际改序号的组增加版本，其他组／成员／Run 不改。默认组固定根首位，边界不写版本。
- 完整有序兄弟集合／元数据／版本拒绝并发变化和 ABA，失败事务不留下半次交换；版本耗尽拒绝。UI 不乐观应用未提交顺序或积压旧编辑，回执保留更新组版本及当前稳定选择。
- 正常退出等待已接受排序、终端输入及偏好保存；排序不确定、偏好或终端失败取消该次退出并保留错误。排序失败后先刷新权威元数据，再允许新的显式编辑；不自动重放操作。
- 目录仍保留 designed/not_run；创建／移动／改名／组排序的独立局部证据不宣布整条设计用例或完整会话手动排序通过。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-009

**关闭视图再打开仍指向原会话**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 的受控任务正在递增计数；记录 Session/Run 和初始计数。

#### 操作步骤

1. 关闭 A 的视图，不点击停止。
2. 通过会话列表重新打开并读取计数。

#### 预期与不变量

- 关闭视图不会结束 agent 或后端 session。
- 重新打开保持原 session_id/run_id，计数连续增加。
- 允许创建新 view_id，不能把它替代 Session 身份。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-010

**连续切换仅允许按需创建 attach 客户端**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：functional、state、stability
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B 两个 agent 的启动 nonce 和 PID 已保存；记录后端客户端进程基线。

#### 操作步骤

1. 交替进入 A、detach、进入 B、detach，共执行 20 轮。
2. 每轮核对目标 nonce，结束后核对 Run 与进程树。

#### 预期与不变量

- A、B 各自只有原 Run；agent 未因切换重新启动。
- 允许按需产生 attach 客户端，结束后退出的客户端不误记为 agent 退出。
- 切换本身不提交额外文本给任何 agent。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-011

**已退出会话归档保留可识别的历史**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 已正常退出，保存了带 Run 标识的输出和退出记录。

#### 操作步骤

1. 归档 A，再查看归档记录并读取已保存输出。
2. 重新启动管理器后再次查询 A。

#### 预期与不变量

- 归档收纳的是会话记录，不产生新的 agent Run。
- 历史仍关联原 Session/Run，界面不会提供虚假的活终端状态。
- 已退出事实与原有输出范围不因归档丢失。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-012

**运行中会话归档遵守明确处置策略**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-006](../COVERAGE.md#fl-006)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-05](../COVERAGE.md#fr-05)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-ARCHIVE-DELETE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-ARCHIVE-DELETE 已规定运行中归档允许性及运行处置；A 正在计数。

#### 操作步骤

1. 对 A 发起归档，分别覆盖确认及取消路径。
2. 核对 Run、计数、列表归属与保存记录。

#### 预期与不变量

- 是否继续运行、拒绝归档或要求先停止，严格匹配已记录策略。
- 取消操作不改变运行状态；归档不会隐式等同于未经说明的停止。
- 若采取停止动作，必须保留目标 Run 与退出证据。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-013

**删除范围明确且不波及其他记录**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P2 / M2 / integration
- 类型：functional、negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-ARCHIVE-DELETE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 仅在隔离目录建立 A、B 与共享引用记录；TBD-ARCHIVE-DELETE 已规定删除范围。

#### 操作步骤

1. 发起 A 的删除并先取消一次。
2. 再次按已定规则删除，核对 A 的元数据、日志、运行以及 B 的数据。

#### 预期与不变量

- 取消时元数据、日志和进程均不变。
- 确认后的处置逐项匹配声明的范围，不把删除视图、记录和活 Run 混为一事。
- B 及不属于 A 的文件不被删除；悬空引用按既定规则报告。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-SESSION-014

**并发重命名与过期列表操作不误作用于替代会话**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0001](../../adr/0001-architecture-and-reuse.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)、[ADR-0021](../../adr/0021-persistent-session-display-names.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：concurrency、state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 两个控制入口读取同一 A；第二个入口可重命名或使 A 不再可操作。

#### 操作步骤

1. 入口一保留旧列表，入口二完成重命名或按已定规则移除 A。
2. 入口一提交基于旧列表的重命名、进入或停止操作。
3. 两个编辑入口带同一版本并发提交；另测改名后改回原值，再提交旧版本。

#### 预期与不变量

- 重命名通过 session_id 定位；成功或冲突结果可核对，不能覆盖其他 Session。
- 目标已失效时返回明确错误或刷新目标，不能自动操作相同行号的新会话。
- 并发操作的最终记录内部一致，不出现名称属于 A 而 Run 属于 B。
- 预期名称／版本失配均明确冲突，ABA 不绕过版本；没有自动覆盖或自动重试，失败编辑需返回刷新核对。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-001

**detach 后原任务继续运行**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-005](../COVERAGE.md#fl-005)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-04](../COVERAGE.md#fr-04)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M0 / windows_e2e
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 运行 FX-PROCESS 计数任务，已经 attach；记录 Run 与计数。

#### 操作步骤

1. 通过当前已验证的 detach 操作分离。
2. 等待夹具报告两个后续检查点，再重新 attach。

#### 预期与不变量

- 分离期间任务继续推进，重新连接仍是原 Run/nonce。
- detach 与 agent 终止有不同事件记录。
- 没有重新执行启动命令。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-002

**正常退出管理器保留终端后台**

- 功能：[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B 后台运行，尚未引入自有消息 Host；计数可独立观察。

#### 操作步骤

1. 正常退出 TUI。
2. 从夹具侧确认任务继续，重新启动管理器发现并接回 A、B。

#### 预期与不变量

- 两个 Run 与 nonce 均不变，后台继续读取终端输出。
- 重开界面不重复创建 agent 或后端 session。
- 没有消费者时不额外宣称消息自动处理。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-003

**TUI 被强制结束不连带终止托管任务**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 隔离测试进程树内 A 已运行；可准确区分 TUI、后端和 agent PID。

#### 操作步骤

1. 只强制结束测试 TUI 进程。
2. 通过夹具观察 A 的后续检查点，再重新打开管理器。

#### 预期与不变量

- 后端及 agent 未被 TUI 清理路径连带结束。
- 重开后验证原 Run 身份，不根据残留界面状态重启任务。
- 异常退出被记录但不伪造任务失败或完成。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-004

**attach 客户端异常退出不等于 agent 退出**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 已 attach，夹具控制器知道 attach 客户端与 agent 的不同 PID。

#### 操作步骤

1. 只结束测试 attach 客户端。
2. 返回或重新打开管理器，再连接 A。

#### 预期与不变量

- agent 是否存活依据其真实事件判断，不以 attach 客户端退出码替代。
- 存活的 A 仍能接回原 Run；失败的客户端连接有独立错误。
- 其他会话和后台不受清理影响。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-005

**agent 正常退出后终端不再被视为活跃 agent**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- FX-PROCESS 在检查点后以退出码 0 结束，提供可核对退出证据。

#### 操作步骤

1. 等待该检查点并捕获退出事件。
2. 打开管理器历史与会话状态，尝试进入已退出的 Run。

#### 预期与不变量

- Run 退出与退出码 0 有来源记录，不继续显示 agent 正在运行。
- 查看历史不会启动新进程；重新启动是明确动作。
- 进程正常退出不自动等同于某条跨会话任务已完成。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-006

**非零退出保留失败原因与既有输出**

- 功能：[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- FX-PROCESS 写入可识别输出后以配置的非零退出码结束。

#### 操作步骤

1. 分别执行参数矩阵中的退出方式。
2. 查看会话、Run、错误说明与保存输出。

#### 预期与不变量

- 保留实际退出证据和可读取输出范围，不把非零退出显示为正常任务完成。
- 未获得具体退出码时说明未知原因，不猜测数值。
- 失败不会把该 Session 的旧历史覆盖为另一 Run。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 退出方式：`["显式非零退出码", "测试夹具触发未处理异常", "仅在隔离测试进程中强制结束"]`


### TC-LIFE-007

**agent 退出而 shell 留存时区分两层存活**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M1 / windows_e2e
- 类型：state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 通过显式 shell 启动受控 agent；agent 结束后 shell 保持存在。

#### 操作步骤

1. 让 agent 写出结束检查点并退出。
2. 检查 shell、pane、agent 事件以及管理器状态。

#### 预期与不变量

- 可以显示终端/shell 存活，但不得由此继续声称 agent 活跃。
- 具备明确退出证据时显示退出；无法取得 agent 证据时显示 agent 状态未知。
- shell 提示符不被当成 agent 的 ready 或任务完成回执。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-008

**后端 session 存在但缺少 agent 证据时保持未知**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / contract
- 类型：state、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 后端合同夹具返回有效 session/pane，但不提供可验证 agent 关联。

#### 操作步骤

1. 执行发现、状态读取和列表刷新。
2. 在不追加 agent 证据的情况下重复查询。

#### 预期与不变量

- 后台存在只能证明终端状态，不显示确定的 agent 存活、busy、ready 或 completed。
- 重复查询不会使猜测自动升级为确认状态。
- 恢复或停止入口不使用未证实的 agent 身份。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-009

**明确重启生成新 Run 并保留历史边界**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 的旧 Run R1 已退出并有保存输出；agent 不支持 resume。

#### 操作步骤

1. 明确重新启动 A，取得 R2 的启动 nonce。
2. 查询 A 的当前运行和历史记录。

#### 预期与不变量

- session_id 不变，R2 与 R1 的 run_id 不同。
- 新输出归属于 R2，旧输出仍归属于 R1。
- 界面说明是新进程，不暗示旧模型上下文已经恢复。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-010

**PID 复用不能让旧 Run 关联或停止新进程**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / contract
- 类型：boundary、recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 合同夹具中 R1 已退出；无关测试进程复用相同 PID，但启动标识不同。

#### 操作步骤

1. 保留 R1 的旧记录执行重连、状态刷新和停止请求。
2. 观察对新进程的调用轨迹。

#### 预期与不变量

- PID 相同不足以证明是 R1，必须根据可核对身份拒绝错误关联。
- 不会 attach、发送输入或结束复用该 PID 的无关进程。
- 旧 Run 保持真实退出或无法确认状态。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-011

**后端 pane 标识复用不继承旧 Run**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / contract
- 类型：boundary、state、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- R1 的后端目标已退出；夹具创建同名或同编号的新 pane 归属 R2。

#### 操作步骤

1. 用 R1 的缓存定位信息执行发现和重连。
2. 核对目标 generation/nonce 及最终操作记录。

#### 预期与不变量

- 不能只凭 pane 名称或编号恢复 R1。
- 旧记录不会自动指向 R2；任何重新指定目标都需明确操作。
- 新 pane 不收到旧会话的控制指令。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-012

**同 cwd 同命令的会话重连保持各自身份**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B 使用相同命令、名称规则允许的相近名称和同一 cwd，但 nonce 不同。

#### 操作步骤

1. 关闭管理器，再以不同发现顺序返回 A、B。
2. 按 session_id 分别重连并读取 nonce。

#### 预期与不变量

- 不以 cwd、命令文本或发现顺序合并两个运行。
- 每个 session_id 恢复自己的 run_id 与后端目标。
- 发现结果中的重复行不会造成额外 agent 启动。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-013

**正常停止只结束明确选中的 Run**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：functional、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A、B 均在运行，A 支持已记录的正常退出方式。

#### 操作步骤

1. 明确对 A 的当前 Run 执行停止。
2. 观察退出事件及 B 的后续检查点。

#### 预期与不变量

- 只有 A 的目标 Run 被停止，保留其记录与退出证据。
- B 和管理器所在 shell 不被结束。
- 停止成功必须有对应退出证据，命令提交成功本身不足以证明退出。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-014

**拒绝退出与子进程的停止升级策略**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-LIFECYCLE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-LIFECYCLE 已规定正常停止、等待、强制停止与子进程范围。
- 在隔离测试进程树中建立忽略正常退出信号的 A 及其子进程，另建无关 B。

#### 操作步骤

1. 停止 A，按已定策略等待或明确升级。
2. 记录 A、子进程、后端与 B 的实际状态。

#### 预期与不变量

- 等待和升级严格符合已记录策略；没有退出证据时不显示已停止。
- 清理仅覆盖经验证属于目标 Run 的对象，不按映像名称批量结束进程。
- 尚存子进程、权限失败和未能终止对象有明确结果，不宣称无残留。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-015

**停止旧 Run 与新 Run 启动竞态不误杀新运行**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / contract
- 类型：concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 停止请求明确绑定 A/R1；故障夹具能暂停在目标重新校验之前。

#### 操作步骤

1. 暂停旧停止请求，让 R1 退出并明确启动 R2。
2. 恢复旧停止请求，观察实际目标。

#### 预期与不变量

- 旧请求发现 generation 不匹配时不向 R2 发出终止。
- R2 继续提供后续检查点；R1 的停止结果保留原目标。
- 界面不会把 R2 的存活状态写回 R1。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-016

**重复停止不会操作另一 Run**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A/R1 已完成停止，并保留该操作的目标身份。

#### 操作步骤

1. 重复提交针对 R1 的停止请求。
2. 在其中一次重试前启动 A/R2，再重复旧请求。

#### 预期与不变量

- 返回已退出、无变化或明确的旧目标错误，不结束 R2。
- 不会因重复停止而清除历史、重新启动 agent 或删除后台命名空间。
- 每次结果仍可追溯到 R1。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-017

**并发重启按已决规则保持单一当前 Run**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：concurrency、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-LIFECYCLE](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-LIFECYCLE 已规定重复/并发重启的仲裁与错误结果；A 已退出。

#### 操作步骤

1. 两个管理入口在同一屏障同时请求重启 A。
2. 观察后端创建、Run 写入和最终当前 Run。

#### 预期与不变量

- 按已定规则接受、串行化或拒绝竞争请求，结果对调用者可见。
- 不会留下两个未被记录的活 agent，也不会让当前 Run 指向错误实例。
- 若实际执行多次明确启动，每次 Run 均独立记录，旧运行处置符合单一后端映射约束。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-LIFE-018

**单会话后端 server 故障不错误清理其他会话**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 在独立测试 namespace 中 A、B 使用不同后端 session/server；记录映射。

#### 操作步骤

1. 只结束 A 的测试 server。
2. 查询 A 的 agent/终端状态并继续操作 B。

#### 预期与不变量

- A 的断开、退出或不确定状态以真实证据显示，不继续假称可接回。
- B 的 server、agent 和元数据不被全局清理。
- 不使用开发分支共享 server 假设解释稳定版进程树。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-001

**首次初始化不产生虚假历史或覆盖其他目录**

- 功能：[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：functional、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 为产品显式指定空的隔离数据目录；相邻目录含哨兵文件。

#### 操作步骤

1. 首次打开产品并读取 Session/Run 集合。
2. 正常退出重开，核对生成文件及相邻目录哈希。

#### 预期与不变量

- 初始化结果可读且为空，不生成模拟产品会话。
- 重复启动不重置已有有效数据。
- 写入限制在指定产品目录，相邻哨兵不变。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-002

**元数据 Unicode 与特殊字符持久化往返**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：boundary、security
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)

#### 前置条件

- 准备包含中文、组合字符、引号和换行说明字段的合法会话数据。

#### 操作步骤

1. 经公共接口保存并关闭管理器。
2. 重开后读取 Session、Run 及其关联值。

#### 预期与不变量

- 被接受字段按已声明编码往返，不截断、多次转义或改变身份。
- 字符串不被执行为 SQL、shell 或终端控制操作。
- 非法字段按规则报错，不能部分改写无关记录。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-003

**语法损坏的 TOML 不静默替换为可执行配置**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：negative、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 隔离配置副本中有有效版本备份和已运行 A；修改副本为可重复的语法错误。

#### 操作步骤

1. 分别加载缺失引号、截断数组及重复关键定义的数据。
2. 读取错误并核对文件与现有 A。

#### 预期与不变量

- 解析失败有位置或可定位原因，不把半份数据拼成启动命令。
- 不自动覆盖用户配置为默认内容，也不停止既有 A。
- 错误配置不会触发 agent 启动。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 配置损坏：`["字符串未闭合", "数组截断", "重复关键定义", "非法字节序列"]`


### TC-STORE-004

**缺失或空配置按明确初始化策略处理**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：boundary、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-CONFIG](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-CONFIG 已记录无文件、零字节文件与空文档的行为。

#### 操作步骤

1. 在独立目录逐项提供上述状态并启动管理器。
2. 观察提示、文件修改与 agent 启动情况。

#### 预期与不变量

- 初始化、报错或引导行为符合策略，三种状态不被无说明地等同。
- 不未经明确启动操作执行默认 agent 命令。
- 若创建新配置，明确记录新路径，不覆盖已存在但未指定的用户配置。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 配置状态：`["不存在", "零字节", "仅注释/空文档"]`


### TC-STORE-005

**未知配置键和更高 schema 版本有可预测结果**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：compatibility、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-CONFIG](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-CONFIG 已定义 schema 版本、未知键、类型错误与向前兼容策略。

#### 操作步骤

1. 加载含未知字段、类型错误及未来 schema 的配置副本。
2. 尝试编辑一个已知字段并保存。

#### 预期与不变量

- 加载与保存按已定策略处理，不静默丢失无法理解的关键配置。
- 不把错误类型强制转为可执行参数。
- 不支持的版本有可定位提示，原文件仍可恢复。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-006

**配置保存中断不留下无提示的半份有效配置**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M1 / integration
- 类型：recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-CONFIG](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- TBD-CONFIG 已规定配置写入、替换与恢复方式；准备可核对的旧/新版本。

#### 操作步骤

1. 在临时写入、替换前后等实际实现故障点终止测试写入进程。
2. 重启并读取配置与恢复提示。

#### 预期与不变量

- 恢复后使用完整旧版本或完整新版本，若无法恢复则明确报错。
- 不把截断文件当成成功保存，也不擅自删除唯一有效备份。
- 故障点及恢复选择可追溯，不依赖猜测文件修改时间。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-007

**配置只读或文件被占用时不显示保存成功**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 隔离测试配置可读取但不可写，或由夹具按测试共享模式占用。

#### 操作步骤

1. 修改一个合法字段并保存。
2. 关闭重开后核对实际文件和显示值。

#### 预期与不变量

- 保存失败被明确呈现，不仅更新内存就显示已持久化。
- 文件保留原内容，重开后不会声称失败修改已生效。
- 失败不导致既有会话被停止或重复启动。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-008

**数据库只读时元数据变更不能伪装提交**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M1 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 测试数据库处于只读状态；存在 A 的有效记录。

#### 操作步骤

1. 尝试重命名 A 并写入一个新元数据记录。
2. 重开读取同一数据源。

#### 预期与不变量

- 写入失败对调用者可见，不返回持久化成功。
- 已有数据及关联保持不变，界面不会永久保留伪造提交状态。
- 失败不会被误解释为 A 的 agent 退出。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-009

**数据库容量不足时保留已提交记录**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M1 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 仅对隔离测试存储注入容量不足；已有提交基线可校验。

#### 操作步骤

1. 在新增 Run 或更新元数据事务期间触发写入失败。
2. 解除故障后重开数据库并核对关联。

#### 预期与不变量

- 失败或未知提交状态按证据报告，不假称新记录已完整提交。
- 原有已提交记录不因失败被清空，Session/Run 关联没有半写结果。
- 涉及进程已经创建的场景转入明确的恢复流程，不仅重试创建。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-010

**数据库 busy 与锁等待按已定有界策略处理**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：boundary、concurrency、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)

#### 前置条件

- O-008 已固定事务、busy/超时策略及 B_DB_WAIT；夹具可保持写锁。

#### 操作步骤

1. 持有测试数据库写锁，另一个控制入口修改 A。
2. 分别在 B_DB_WAIT 前解除锁及越过等待上限后解除。

#### 预期与不变量

- 重试或失败符合记录的策略，无永久无提示等待。
- 失败不报告已提交；成功只对应实际提交的数据。
- 锁释放后可继续执行新请求，不需要删除数据库恢复。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 锁保持时长：`["B_DB_WAIT 以下", "B_DB_WAIT 边界", "B_DB_WAIT 以上"]`


### TC-STORE-011

**多个入口并发写不同会话不丢失已提交更新**

- 功能：[FL-004](../COVERAGE.md#fl-004)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-03](../COVERAGE.md#fr-03)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：concurrency、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- A、B 属于同一测试数据库；两个入口可屏障同步提交独立更新。

#### 操作步骤

1. 并发重命名 A、更新 B 的分组，重复交错执行。
2. 从第三个连接读取最终提交结果。

#### 预期与不变量

- 每个成功回执对应的更新均可读，不能用过期全量快照覆盖另一条更新。
- 没有跨会话字段混合或身份改变。
- 若返回冲突或失败，结果明确且不会假称成功。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-012

**损坏数据库不被静默重建成空成功状态**

- 功能：[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 复制含已知 Session/Run 的数据库到隔离目录并注入可重复损坏；保留原副本哈希。

#### 操作步骤

1. 启动产品读取损坏副本。
2. 记录提示及文件变化，按已记录恢复入口再次打开。

#### 预期与不变量

- 错误可定位，不显示空会话列表并伪装成正常无历史。
- 不静默覆盖唯一损坏文件或原备份。
- 任何恢复后可用范围与无法恢复部分明确说明。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-013

**旧产品读取未来数据库版本不会破坏数据**

- 功能：[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0023](../../adr/0023-persistent-groups-and-membership.md)
- 优先级 / 阶段 / 层次：P0 / M4 / integration
- 类型：compatibility、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- O-008 已明确数据库 schema 版本检测及兼容行为；准备未来版本测试副本。

#### 操作步骤

1. 用旧产品打开该副本并尝试写入。
2. 比较数据库与关联日志的前后哈希。

#### 预期与不变量

- 按兼容策略只读、拒绝或安全处理，不对未知 schema 执行破坏性迁移。
- 未支持的写入不返回成功，原记录可供对应新版本继续读取。
- 错误包含可操作的版本信息而不泄露会话正文。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-014

**数据库正常迁移保持身份和引用关系**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M4 / integration
- 类型：compatibility、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- O-008 已定义实际旧/新 schema 与迁移；建立包含多 Run、归档和来源引用的旧数据。

#### 操作步骤

1. 用新版本按记录步骤迁移副本。
2. 逐项比较 Session/Run 身份、状态、来源范围和引用关系。

#### 预期与不变量

- 迁移后的记录可以读取，稳定 ID 不改变或有明确无歧义映射。
- 没有把已退出 Run 变为活跃、把旧记录混入新 Run。
- 重复打开不会重复迁移或创建重复记录。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-015

**数据库迁移中断按已定恢复合同处理**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M4 / integration
- 类型：recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- O-008 已明确迁移原子性、备份及重入策略；仅使用可丢弃测试副本。

#### 操作步骤

1. 在每个实际迁移持久化边界终止测试进程。
2. 重开并按记录流程恢复，再检查表结构与记录关系。

#### 预期与不变量

- 迁移后的记录可以读取，对外稳定的 session_id/run_id 保持不变；内部表结构变更不能改变用户寻址身份。
- 不丢失已承诺保留的数据，也不重复执行非幂等迁移动作。
- 恢复步骤与实际版本可追溯。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-016

**产品回退时兼容限制明确且不逆向破坏数据**

- 功能：[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M4 / integration
- 类型：compatibility、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-RELEASE](../FIXTURES_AND_FAULTS.md#fx-release)

#### 前置条件

- O-008 已确定升级后数据与旧产品的兼容/备份回退规则。

#### 操作步骤

1. 在副本中完成一次升级迁移，再用上一产品版本打开。
2. 按允许的回退路径恢复并读取记录。

#### 预期与不变量

- 不支持直接回退时明确拒绝，不伪装成功读取或擅自反向转换。
- 备份恢复范围与升级后新增数据的处置对用户可见。
- 活 Run 发现与历史恢复不因回退被混为同一能力。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-017

**日志轮转边界按字节预算正确保留范围**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-011](../COVERAGE.md#fl-011)、[FL-015](../COVERAGE.md#fl-015)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：boundary、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- O-007 已定义格式、字节计量和 B_LOG_FILE/B_LOG_TOTAL；夹具输出可编号 UTF-8 记录。

#### 操作步骤

1. 输出至预算前一单位、预算值、预算后一单位，并继续到触发总保留清理。
2. 分页读取当前及已轮转范围，检查文件总量。

#### 预期与不变量

- 轮转与保留符合已定字节规则，不以中文字符数冒充字节数。
- 已移除范围明确报告，不把日志开头误称整个 Run 的起点。
- 可读取记录仍有正确来源和游标，不因切割多字节内容伪造文本。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 边界负载：`["B_LOG_FILE-1、B_LOG_FILE、B_LOG_FILE+1", "B_LOG_TOTAL 边界", "ASCII 与中文多字节混合"]`


### TC-STORE-018

**读取与轮转并发不拼接不同 Run 的文件**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-007](../COVERAGE.md#fl-007)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-06](../COVERAGE.md#fr-06)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：concurrency、boundary
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- O-007 已确定游标及轮转读取合同；A/R1 持续写出编号记录。

#### 操作步骤

1. 暂停一个分页读取，在另一入口触发轮转和保留清理。
2. 恢复读取并用原游标继续读取。

#### 预期与不变量

- 结果为一致有效范围或明确的失效/缺口，不拼接错误文件。
- 返回的数据始终属于 A/R1；不得因文件名复用读入 R2。
- 重复或遗漏边界若存在必须按游标合同可识别，不静默掩盖。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-019

**日志磁盘满或不可写时明确记录缺口**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- O-007 已规定日志失败后的告警和运行处置；只向隔离日志路径注入写入故障。

#### 操作步骤

1. 让 A 持续输出，在记录中途触发写失败。
2. 恢复可写状态后重新读取前后范围。

#### 预期与不变量

- 写入失败和未记录区间明确可见，不显示连续完整历史。
- 运行继续或停止按已定策略处理，不能仅凭日志失败伪造 agent 退出。
- 恢复后来源与游标边界正确，不把新片段无说明接到旧末尾。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-020

**无自有消费者期间只承诺后端保留的历史**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-007](../COVERAGE.md#fl-007)、[FL-011](../COVERAGE.md#fl-011)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-06](../COVERAGE.md#fr-06)、[FR-10](../COVERAGE.md#fr-10)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：state、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-LOAD](../FIXTURES_AND_FAULTS.md#fx-load)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- M2 无常驻 Host；后端历史保留范围可观测，A 输出带序号检查点。

#### 操作步骤

1. 关闭 TUI，让输出超过后端实际保留范围。
2. 重开管理器读取当前屏幕和可用历史。

#### 预期与不变量

- A 的任务继续运行，但应用不会宣称关闭期间所有日志均已持续落盘。
- 只返回实际保留范围，并明确未采集或已淘汰区间。
- 不通过补写模拟记录填补缺口。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-021

**清理日志不跟随外部路径删除无关数据**

- 功能：[FL-007](../COVERAGE.md#fl-007)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-06](../COVERAGE.md#fr-06)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：security、negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-007](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- O-007 已确定日志清理及路径归属检查；隔离目录含 A、B 与外部哨兵。
- 所有路径跳转、链接或 junction 仅在可丢弃测试目录内构造。

#### 操作步骤

1. 对 A 的过期日志执行清理，覆盖普通路径、已替换路径与受测路径跳转。
2. 比较 B 和外部哨兵的内容哈希。

#### 预期与不变量

- 只清理可验证属于 A 且符合保留规则的日志。
- 无法确认路径归属时明确失败，不跨越产品数据边界删除文件。
- B、配置、数据库和外部哨兵不变。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-STORE-022

**配置变更只按已定时机影响后续运行**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：state、compatibility
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-CONFIG](../BUDGETS_AND_DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-CONFIG 已规定加载、热重载和启动模板更新时机；A/R1 使用旧配置运行。

#### 操作步骤

1. 修改配置中的 cwd、参数和环境，再按规定触发加载。
2. 检查 R1，并明确启动新的 R2。

#### 预期与不变量

- R1 已发生的启动描述与身份不被重写成新配置。
- R2 使用符合已定加载时机的配置快照，参数来源可核对。
- 无效新配置不会自动重启 R1 或静默切换工作目录。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-001

**创建开始记录后后端创建失败可以辨认**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 故障夹具在创建后端之前返回可定位失败；没有创建任何 agent。

#### 操作步骤

1. 创建 A，触发失败并刷新会话列表。
2. 明确重试创建，读取旧失败与新尝试的记录。

#### 预期与不变量

- 失败尝试不显示为活跃 Run，错误来源可定位。
- 重试不会把失败记录伪装成仍存活的旧实例。
- 没有未记录的后台 session 或重复 agent。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-002

**后端已创建但元数据提交前崩溃的恢复**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M1 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-LIFECYCLE](../BUDGETS_AND_DECISIONS.md)、[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-LIFECYCLE 与 O-008 已定义创建意图、归属证明和未完成创建的恢复策略。
- 隔离测试后端可记录创建 nonce；暂停点在后端创建成功之后、最终元数据提交之前。

#### 操作步骤

1. 在暂停点结束测试管理进程，再重开。
2. 按已定策略发现、提示恢复或清理未完成对象。

#### 预期与不变量

- 不盲目重发创建命令而产生第二个 agent。
- 恢复仅针对可验证属于该创建尝试的对象，不能按相同名称认领外部会话。
- 无法确认时呈现不确定状态，处置过程和原尝试有可追溯关联。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-003

**创建意图已落盘但进程未启动时崩溃的恢复**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[TBD-LIFECYCLE](../BUDGETS_AND_DECISIONS.md)、[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- TBD-LIFECYCLE 与 O-008 已确定 starting 记录恢复策略；暂停点在后端调用之前。

#### 操作步骤

1. 终止测试管理进程，确认未出现后端创建事件。
2. 重开并执行规定的恢复或重试动作。

#### 预期与不变量

- 不能把残留 starting 当作活进程，也不能无说明永久卡在启动中。
- 恢复行为符合已记录策略，最终每个成功 Run 有对应实际启动证据。
- 不复用旧 PID/后端名猜测启动已经成功。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-004

**外部 psmux 会话与名称碰撞不会被误领养或清理**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：negative、recovery、security
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 测试环境有本产品 A 和独立创建的外部 B，名称相似；B 不含本产品归属证明。

#### 操作步骤

1. 启动发现与重连流程，模拟 A 的元数据缺失。
2. 执行仅针对 A 的停止及产品测试数据清理。

#### 预期与不变量

- B 不因名称、cwd 或 PID 相似被认作 A。
- 本产品操作不向 B 注入内容，不结束 B，也不改其配置。
- 未能证明归属的对象被明确区分，不当作已成功恢复。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-005

**停止命令返回丢失时依据实际状态恢复**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / contract
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 故障夹具能在停止命令处理前或处理后切断回复；A 有独立存活观察。

#### 操作步骤

1. 分别在两个故障点发起停止并丢弃回复。
2. 恢复连接并查询真实 Run 状态。

#### 预期与不变量

- 没有回复时不直接宣称停止成功或自动停止别的 Run。
- 恢复后依据身份匹配的退出/存活证据收敛状态。
- 无法确认时保留不确定结果，重试仍绑定原 Run。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-006

**界面关闭期间 agent 退出后重开不自动续跑**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / integration
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A 在后台执行可控任务；TUI 关闭后 A 将在检查点退出。

#### 操作步骤

1. 关闭 TUI，等待 A 真实退出后重开。
2. 查看原 Run 并读取历史。

#### 预期与不变量

- 不根据上次 alive 状态自动启动一个替代 agent。
- 显示真实退出或有依据的不确定状态，保存的历史归属于旧 Run。
- 需要重新启动时由明确操作产生新 Run。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-007

**重启测试机器后历史记录不伪装成可接回活进程**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P0 / M2 / windows_e2e
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- 仅在可丢弃 Windows 测试虚拟机中建立 A 并保存 Run 身份及日志。

#### 操作步骤

1. 正常重启该虚拟机。
2. 重新打开产品并尝试接回重启前 A。

#### 预期与不变量

- 不声称重启前的活进程仍可接回。
- 历史记录可按实际持久化范围查看；自动恢复对话不作为通用能力。
- 任何新启动都生成新 Run，并清晰区别于读取旧记录。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

说明：只重启专用测试虚拟机，不对用户工作机器或非测试进程执行重启。


### TC-REC-008

**持久化期间虚拟机断电后的恢复边界**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P2 / M4 / windows_e2e
- 类型：recovery、negative
- 设计 / 执行：`designed / not_run`
- 规范状态：待决策（`needs_decision`）
- 待决依据：[O-008](../../DECISIONS.md)
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)、[FX-FAULT](../FIXTURES_AND_FAULTS.md#fx-fault)

#### 前置条件

- O-008 已确定持久化与恢复合同；专用测试虚拟机有可恢复快照。
- 数据库中区分已确认提交、正在提交和未发起三组测试记录。

#### 操作步骤

1. 在实际持久化边界由测试宿主切断虚拟机电源，再恢复启动。
2. 检查存储一致性与各组记录的证据状态。

#### 预期与不变量

- 按声明的持久化合同恢复；若无法证明某次提交，标记不确定而不伪造成功。
- 不把恢复失败静默处理成全新空库。
- 记录实际故障位置、丢失范围和恢复步骤，不能把普通进程 kill 结果替代断电结果。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

说明：限专用可丢弃虚拟机；不假设未选择的 journal/synchronous 配置具有特定断电耐久性。


### TC-REC-009

**睡眠唤醒后的会话状态根据证据重新核对**

- 功能：[FL-006](../COVERAGE.md#fl-006)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-05](../COVERAGE.md#fr-05)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / windows_e2e
- 类型：recovery、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 专用测试 Windows 环境支持受控睡眠；A/R1 正在计数，时钟与后端事件可记录。

#### 操作步骤

1. 使测试环境睡眠后唤醒。
2. 恢复管理器状态刷新并重新连接 A。

#### 预期与不变量

- 若 R1 仍存活，保持原身份；若已退出或无法确认，显示对应事实。
- 不把睡眠期间无输出误判为 agent ready、完成或失效重启。
- 状态刷新恢复后不会重复启动 agent 或复用过期客户端状态。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-010

**历史 cwd 不可用时明确重启失败而不切换到其他目录**

- 功能：[FL-001](../COVERAGE.md#fl-001)、[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-01](../COVERAGE.md#fr-01)、[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M1 / integration
- 类型：negative、recovery
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-ECHO](../FIXTURES_AND_FAULTS.md#fx-echo)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- A/R1 已退出；其测试 cwd 可安全移动或删除，其他目录含可观察哨兵。

#### 操作步骤

1. 使原 cwd 不可用，读取历史后明确请求重新启动。
2. 修正为另一明确选择的有效 cwd 后再次启动。

#### 预期与不变量

- 读取历史不要求原目录仍存在，也不触发 agent。
- 失效 cwd 的启动明确失败，不静默回退到产品目录或另一个会话 cwd。
- 用户明确修正后的成功启动生成新 Run，保留实际启动目录与旧历史边界。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。


### TC-REC-011

**系统墙钟跳变不改变 Run 身份或事件因果**

- 功能：[FL-005](../COVERAGE.md#fl-005)、[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-04](../COVERAGE.md#fr-04)、[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / contract
- 类型：boundary、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-UNIT](../ENVIRONMENTS.md#env-unit)
- 夹具：[FX-CLOCK](../FIXTURES_AND_FAULTS.md#fx-clock)、[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 通过可控时钟而非修改用户系统时钟，建立有序创建、退出和重启事件。

#### 操作步骤

1. 将墙钟前调、后调并注入相同显示时间的事件。
2. 查询 Run 历史和当前活动对象。

#### 预期与不变量

- 事件身份和当前 Run 不仅由墙钟先后决定。
- 相同或回退时间不覆盖记录、不复用 Run ID、不把旧退出事件附到新 Run。
- 时间异常可记录；已知操作因果仍可重建。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。

#### 参数变体

- 墙钟变化：`["向前跳变", "向后跳变", "多事件显示时间相同"]`


### TC-REC-012

**不支持或失败的原生 resume 不伪装成历史恢复成功**

- 功能：[FL-011](../COVERAGE.md#fl-011)、[FL-013](../COVERAGE.md#fl-013)
- 需求：[FR-10](../COVERAGE.md#fr-10)、[FR-12](../COVERAGE.md#fr-12)
- 决策：[ADR-0002](../../adr/0002-terminal-backends-and-attach.md)、[ADR-0003](../../adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](../../adr/0004-cli-agents-and-session-messaging.md)
- 优先级 / 阶段 / 层次：P1 / M2 / integration
- 类型：negative、state
- 设计 / 执行：`designed / not_run`
- 规范状态：依据明确（`ready`）
- 待决依据：—
- 环境：[ENV-WIN-BASE](../ENVIRONMENTS.md#env-win-base)
- 夹具：[FX-PROCESS](../FIXTURES_AND_FAULTS.md#fx-process)、[FX-READY](../FIXTURES_AND_FAULTS.md#fx-ready)、[FX-DATA](../FIXTURES_AND_FAULTS.md#fx-data)

#### 前置条件

- 准备不支持 resume 的通用 CLI，以及有测试 resume 能力但会明确拒绝恢复的合作夹具。

#### 操作步骤

1. 对通用 CLI 查看可用操作；对合作夹具显式请求恢复并触发失败。
2. 检查 Run、原生 conversation ID 与保存输出。

#### 预期与不变量

- 通用 CLI 不提供虚假的恢复原模型上下文成功结果。
- 已声明能力但执行失败时报告失败，不把历史文本显示当作 resume 已成功。
- 若用户另行选择新启动，生成新 Run 并保留与原对话恢复的区别。

#### 需要的证据

- 测试环境、产品提交、依赖版本与隔离数据目录标识。
- 按 session_id/run_id 关联的操作记录、夹具事件及前后状态快照。
