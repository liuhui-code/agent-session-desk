# ADR-0012：元数据离线检查、完整快照与非覆盖恢复

- 日期：2026-11-22
- 状态：已接受（Accepted），用户授权完成 M1；补齐 ADR-0009/0010 的最小恢复路径
- 关联：FR-01、FR-10、FR-12；FL-001、FL-005、FL-011、FL-013、FL-016
- 实现：[检查](../../internal/store/check.go)、[快照](../../internal/store/recovery.go)、[CLI](../../internal/cli/metadata.go)
- 实际证据：[恢复切片报告](../testing/results/2026-11-22-m1-metadata-recovery.md)；不等于完整 M1 或发布验收

## 决定

提供独立于终端后端的管理命令：`asd metadata check [FILE]`、`asd metadata backup DEST`、`asd metadata restore SOURCE`。未提供 FILE 时检查当前 state-dir 的 metadata.sqlite3。所有路径继续复用现有 metadataPath；不启动 agent、不探测当前 Run，不从恢复的记录创建进程。

检查以 mode=ro 打开已存在的普通文件，在只读事务内核对 application_id、支持的 schema v1/v2、必需字段、DELETE journal、SQLite integrity_check 和 foreign_key_check。错误不显示库内正文，不创建缺失库，不升级 v1。检查结果只表示数据库结构，不能证明 agent 或模型上下文可恢复。

备份对整个数据库产生逻辑一致的快照，包含全部已有 scope 的 Session/Run、停止确认及 pane 退出观测；不包含 TOML、输出记录、消息、凭据或活进程。不使用直接复制正在写入的数据库。恢复将 SOURCE 的同类快照写到当前 metadata.sqlite3，目标必须尚不存在，不能覆盖空文件、损坏文件或已有有效库。重要原库继续保留，用户可以选择新的 state-dir；同一 runtime/namespace 才能读取原作用域，没有目录别名合并。

## 发布完整文件与失败语义

SQLite VACUUM INTO 使用绑定参数，目标是同目录内由 CreateTemp 创建的私有临时文件。先完成快照并校验，再通过同目录硬链接发布到原先不存在的目标，最后删除临时名。硬链接拒绝竞态中新出现的目标；不降级到可能覆盖现有文件的 rename。

成功返回的正式路径已通过检查。临时清理失败时明确报告“已发布”和目标路径，不把重试描述为尚未发生任何副作用。运行中的普通失败／取消删除本次仍拥有的临时文件；原路径不变。强杀可能遗留 .asd-snapshot-*.sqlite3，需要在确认无在途备份后检查／清理；不自动删除来源不明的文件。正式路径只在完整库已就绪后出现，避免强杀产生的空目标被普通 Open 当成新的空库初始化。

硬链接支持是本命令当前的文件系统前提；在支持它的本地文件系统验证，Windows 以实际执行结果为准。不支持的文件系统明确失败，不承诺 FAT／网络共享。FULL 设置会使 SQLite 同步快照文件，但未验证目录项、硬件缓存和断电恢复，不作断电零丢失承诺。

v1 检查／备份／恢复均保留 v1；恢复目标下一次正常 Open 时按 ADR-0010 事务性升级到 v2。备份来源不会被迁移。未来 schema、外来应用、损坏库或外键错误均拒绝，不以删除重建冒充恢复。

## 依据与验证

Context7 查询后对照 [SQLite VACUUM INTO 官方合同](https://www.sqlite.org/lang_vacuum.html)：生成一致快照，但中断生成的输出可能不完整。因此采用临时库校验和非覆盖发布，不把 SQL 调用成功前的临时文件作为正式备份。

TDD 验证跨作用域证据、损坏／未来库保留、离线 CLI 往返和正式路径不暴露空库。另以子进程在真实 SQLite 写事务检查点强杀，再用 Store 重开，核对全部既有身份／时间及未提交记录回滚。Windows 原生执行同一合同；它是进程中断证据，不代表物理断电、磁盘满或任意硬件故障矩阵已通过。
