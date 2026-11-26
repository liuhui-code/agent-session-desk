# ADR-0045：显式消息 outbox 与原视图输入队列投递

- 日期：2026-11-26
- 状态：已接受，用户同意跨会话最小闭环并要求采用成熟方案
- 后续：前端 m／活动视图投递选择已由 [ADR-0046](0046-backend-only-message-delivery.md)替代，当前通信通过后端／CLI；其余来源与持久请求规则保留
- 关联：FR-04/06/07/09/12，FL-005/007/008/013/019
- 扩展 ADR-0003/0004/0013；不关闭 M0–M4 接受或暂缓的稳定性专项

## 依据与选择

采用持久 outbox 的请求记录与幂等键原则、SQLite 事务／条件更新和现有终端粘贴通道。只复用既有依赖与代码，不引入消息中间件、品牌 agent 协议或新常驻进程。参考资料是设计依据，本轮没有复制外部代码：

- [AWS transactional outbox](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html)：先提交持久请求，外部副作用和数据库分开处理；消费者幂等仍需单独保证。
- [SQLite atomic commit](https://www.sqlite.org/atomiccommit.html)：沿用既有 DELETE journal、FULL synchronous 与单写者事务，不把数据库事务解释为终端输入事务。
- [固定 tmux 3.4 手册](https://github.com/tmux/tmux/blob/3.4/tmux.1)：借鉴 literal／buffer paste、显式目标和 paste 与 Enter 分离的语义；Windows 实际仍调用既有 pinned psmux send-paste，不声称两个后端实现等价。

Context7 已核对 tmux 3.4 与 SQLite 官方资料。通用终端不提供幂等消费／agent 接收证据，所以此处采用持久意图、单次显式投递及 unknown 保守恢复，不启用教科书 outbox 常见的自动重试消费。

## 合同

1. `read --lines N SESSION_ID` 显式保存带 Session/Run/record/时间和截断标记的屏幕快照；N 为 1–200，默认 60。返回的行号只相对于该不可变快照，不是完整输出流游标；complete=false。
2. `send` 必须给完整请求 ID、目标 Session/Run；从 UTF-8 文件／stdin 或同作用域保留快照的局部行区间取正文。UTF-8 BOM 文件去 BOM，正文不经 shell 拼接。来源、提交时的正文副本一起入库，来源快照轮转后不影响既有请求。同一 ID／同一参数返回原请求，内容、目标或区间参数不同则拒绝。
3. CLI send 只建单，返回 queued；status 按消息 ID 离线查询。接收／完成状态没有可靠来源时不产生。CLI 不开第二个输入连接，不在用户输入期间旁路注入。
4. psmux 工作区管理焦点选中且打开原目标后按 m，按需加载该 Run 最早 queued 请求。预览完整目标身份、来源、可分页的转义正文，默认 Cancel；Tab 选择 Paste 后 Enter 明确确认。窄窗无法完整呈现目标时禁止确认。弹窗消费鼠标／粘贴／快捷键，不把这些内容发给子程序。
5. 投递经当前视图已有 FIFO 和原 Connection。先重新发现原目标，事务领取 queued 请求并提交 unknown，然后一次调用原生 Paste；成功后才记录 input_written。粘贴不追加 Enter；多行在部分 shell 仍可能执行，界面明确提示，由用户检查实际目标程序。已写入仅指终端传输写出，不是 agent 接收或完成。
6. 条件更新保证同一请求只有一个领取者；attempt token 限定回执所有者。取消／部分写入／写出后回执失败或进程中断保持 unknown，禁止重新领取和自动重放。SQL 事务期间不调用外部终端。来源读取、改名／排序和历史不改变目标身份。
7. 同一 metadata.sqlite3 全部作用域最多保留 128 条请求，正文每条 64 KiB，序列化文档每条 512 KiB，满时明确拒绝，不自动删除幂等回执。这些是首个有界切片的防护上限，不是测得的资源预算；后续显式清理／归档另行设计。状态限定 queued／unknown／input_written。
8. SQLite v5→v6 在原事务内新增 messages 表／索引；v1–v4 先沿原路径升级。只读 check／backup 接受 v1–v6，v6 快照包含消息；旧 v5 程序拒绝新库。ui.toml v7 与依赖版本不变。

## 本轮边界

实现的是单个管理视图输入队列内的显式投递和同一请求的跨调用者 CAS，不声明多个独立 TUI／辅助 attach／外部客户端共同编辑同一终端时的全局输入仲裁，也不声明不同请求的全局 FIFO。共享 Host／多客户端仲裁、队列自动消费、reply_to、agent accepted/completed 及后台提醒仍需后续合同与验证。未知请求需要人工检查，当前接口不能将其恢复为 queued。

验证见[执行报告](../testing/results/2026-11-26-explicit-messaging.md)。用户暂缓的 STAB-001～005 保持，不以本轮功能检查替代长稳、完整输入矩阵、真实 agent 或完整发布接受。
