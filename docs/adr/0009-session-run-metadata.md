# ADR-0009：以 SQLite 保存最小 Session/Run 元数据

- 日期：2026-11-22
- 状态：已接受（Accepted），依据用户批准推进 M1 最小元数据切片的授权
- 关联：FR-04、FR-10、FR-12；FL-005、FL-011、FL-013；细化 ADR-0003/0004/0005
- 实现与验证：[存储与协调核心](../../internal/store/)、[实际执行报告](../testing/results/2026-11-22-m1-metadata.md)；完整恢复、预算与发布验收仍待做

## 驱动与运行设置

固定 `modernc.org/sqlite v1.62.0`，该模块内嵌 SQLite **3.54.1**；使用标准 `database/sql`，CGO_ENABLED=0 的 Windows amd64 构建和原生执行分别验证。按驱动要求精确保留其 `modernc.org/libc v1.80.1` 依赖，不单独漂移 libc。传递模块以 go.mod/go.sum 为准。Context7 的当前页面片段可能滞后，精确语法与内嵌版本同时对照固定模块源码和实际 `sqlite_version()`。

选择理由是原生 Windows 构建不增加 C 编译器或额外运行时安装要求；尚未建立驱动间性能优势结论。对照方案 mattn/go-sqlite3 需要 CGO 构建，ncruces/go-sqlite3 引入 WASM 运行时。当前不加入 ORM、存储服务进程或新的常驻消费者。

本地文件使用 **DELETE journal + synchronous FULL**；每个 Store 最多一个打开/空闲连接，busy_timeout=1000ms、cache_size=-1024（KiB 目标），事务采用 BEGIN IMMEDIATE。连接池等待服从 context。缓存设置不是完整 Go/SQLite 内存硬上限，锁超时也不是任意负载的时延 SLA。现阶段不启用 WAL；不自动转换已有其他 journal 模式。

## schema 与作用域

- 数据库文件名 `metadata.sqlite3`。`--state-dir` 可明确目录；省略时，显式 `--runtime-dir` 的 namespace 目录同时承载测试/隔离元数据，否则使用 UserConfigDir/agent-session-desk。默认元数据不依赖可清理的用户缓存目录。
- scope = 后端种类 + 后端绝对数据目录（含 namespace），Windows 按大小写不敏感归一化。迁移 runtime-dir 会成为另一作用域；不自动合并别名、symlink、junction 或复制目录，不承诺网络共享文件系统。
- application_id=`0x41534431`、user_version=1。空数据库可在单个事务内初始化；未来版本、其他应用标记、外来表或损坏文件返回错误，保留文件，不覆盖重建。已有本版数据库核对必需字段。
- sessions 保存 scope/session_id、名称和 executable 显示信息；runs 保存独立 run_id、所属 Session、backend_id、首次入库时间和显式停止确认时间。复合主键、外键和事务防止半份关联；冲突不静默覆盖。当前名称/入口视为该切片的不可变快照，重命名/重启接口另做。
- 不保存 argv、env 值、cwd、终端正文、会话内容或模型上下文。字段上限沿用后端有界元数据原则（每个显示字段最多 4096 字节）；重复发现先核对已存身份，匹配时只读，不因 TUI 刷新抢写锁。

## 进程操作与提交边界

start 先检查当前可写性，成功创建终端后提交元数据，再报告普通成功。数据库和外部终端创建不是原子事务：若后提交失败，报告准确 Session/Run ID、终端已创建及不要重复 start，不自动杀掉已运行任务，也不伪造落盘成功。进程崩溃仍可能留下“后台存在、入库未完成”的窗口；后续成功的 list 可以按真实后端记录补录。没有持久 starting 状态或自动重放。

list/TUI 始终向后端查询，成功发现才保存，后端错误按原错误返回；未列出某 Run 不写成退出或停止。attach/preview/stop 继续只从实时后台选取、重验目标，历史记录不能成为控制凭据。stop 的后端成功返回后才记录明确 stop 证据；若落盘失败，错误说明后端已停止和元数据未更新。

`asd history [--limit N]` 独立读取数据库，无需找到或启动后端。默认最近 100 项，上限 1000，历史 terminal_state 恒为 unknown，stop_confirmed_at 为独立事件证据，不等于 agent 退出码或任务完成。读取历史不会启动/重启 agent；库打开时会初始化空库并短暂获取写锁验证 schema，当前不提供只读文件模式。历史查询尚无 TUI 页面和完整分页接口。打开库会初始化空库并短暂校验写锁，尚不提供只读文件打开模式。

## 保留的门槛

这是最小元数据能力，不包含消息表、自动恢复、原对话 resume、退出码证据、日志轮转、升级迁移、磁盘配额/保留清理和完整诊断导出。只读/磁盘满/断电/强杀等故障矩阵仍需扩展，尤其不能以 FULL 设置声称已通过断电恢复。TBD-LIFECYCLE、TBD-UPGRADE、TBD-INSTANCE-SCOPE 与 O-008 仅细化本次范围，其余保持未决。

## 许可与依据

- [固定驱动版本](https://pkg.go.dev/modernc.org/sqlite@v1.62.0)、[源码](https://gitlab.com/cznic/sqlite/-/tree/v1.62.0)
- [SQLite 事务](https://sqlite.org/lang_transaction.html)、[PRAGMA](https://sqlite.org/pragma.html)
- 驱动 BSD-3-Clause，Copyright (c) 2017 The Sqlite Authors；SQLite 本体 public domain。实际发行物的直接/间接模块与嵌入代码许可仍须收集，原创整体许可 O-010 未解决。
