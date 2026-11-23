# 会话元数据与历史

日期：2026-11-22。本页是 M1 最小切片，完整阶段尚未完成；合同见 [ADR-0009](adr/0009-session-run-metadata.md)。

启动仍使用自定义 CLI 或 TOML 模板。元数据自动记录 Session/Run 身份、名称、入口和后端关联；运行状态始终向实际后端核对。

```powershell
asd --state-dir C:\asd-data start --name 开发 -- 自定义程序.exe 参数
asd --state-dir C:\asd-data list --json
asd --state-dir C:\asd-data history --limit 20
asd --state-dir C:\asd-data stop 完整session_id
```

history 返回 JSON，即使后端程序缺失也可读取。它显示入库记录而非在线证明，terminal_state 为 unknown；stop_confirmed_at 只说明产品曾确认停止了该后端目标。history 不提供终端正文、agent 退出码或模型上下文；显式屏幕快照另用 [record／output](OUTPUT_RECORDS.md) 查询。list 只显示实际后台记录，历史不会自动变成可输入会话。

`--state-dir` 放在子命令之前，指定 metadata.sqlite3 的父目录。默认使用用户配置目录下 agent-session-desk；如果明确传入 `--runtime-dir` 且未设 state-dir，则在该 runtime 的 namespace 子目录存储，方便隔离。查询时使用相同 runtime-dir/namespace；同一个数据库内不同后台作用域互相隔离。目录搬迁和作用域合并尚无迁移功能。

启动报错若已包含 Session/Run ID 且提示终端已创建、元数据未提交，先用相同参数的 list 检查，避免重复 start。成功发现可以补录后台现有记录。数据库故障会明确报错；不要删除或覆盖故障文件来冒充恢复。重要记录应由用户保存在保留目录，现在可使用下方离线检查与备份／恢复命令；仍需自行保管配置和输出记录。

读取默认最近 100 项，limit 为 1–1000。读取内存和 SQLite 连接有边界，元数据磁盘保留/清理策略和完整资源预算仍待验证。库打开会初始化空库并短暂获取写锁校验 schema；当前不提供只读文件模式。SQLite 文件只用于本机；本次不承诺网络共享目录、完整升级迁移和掉电恢复。

## pane 退出观测（schema v2）

当产品实际观察到原 pane 退出时，list 中可有 pane_exit；history 在记录外层保留首次 pane_exit（pane_id、source、observed_at、exit_code）。退出码目前为 null；psmux 3.4.0 的 pane_dead_status 常量 0 不是真实退出码，observed_at 也不是进程退出时刻。历史 session.terminal_state 仍为 unknown。

这是终端 pane 的观测，不能证明 shell 内子 agent 的状态、某条任务成功或完整进程树已退出。断开、目标消失和显式 stop 不制造 pane_exit；以前已取得的观测继续保留。详见 [ADR-0010](adr/0010-pane-exit-observations.md)。

v1 数据库打开时事务性追加观测表并升级为 v2，原有身份与时间保留；冲突／失败不删库重建。旧 v1 程序会拒绝 v2，没有降级迁移。重要数据升级前保留副本；完整升级和备份／恢复验收仍待做。

## TUI 只读历史

启动管理器后按 h，读取当前作用域最近 100 条记录。↑↓、Home/End、PgUp/PgDown 在已加载列表中选择；Enter 仅查看详情。详情内同样按键滚动，r 按需刷新；Esc/h 从详情返回历史列表，再返回实时列表。q 退出管理器，保留后台。历史中的 Enter 不启动进程、不接回终端、不发送输入。

详情区分首次记录、停止确认和 pane 退出观测；所有时间显示 UTC。历史不显示当前在线证明、agent 成功结果或可靠数值退出码。刷新按 Run ID 保留选择；结果不再包含该 Run 时关闭详情。记录与错误清洗后按窗口宽度换行，完整正常身份可滚动查看。

历史最多保留一个读取请求，关闭时取消，忽略迟到结果；窗口缩放和实时发现轮询不自动重读。TUI 当前仍需初始后端探测成功，随后发现失败时可打开历史；后端程序缺失时可使用独立 CLI history。最近 100 条以外的 TUI 查询、完整数据库分页、日志及原对话恢复尚未实现。见 [ADR-0011](adr/0011-read-only-tui-history.md) 与[局部证据](testing/results/2026-11-22-m1-history-ui.md)。

## 离线检查、备份与恢复

```powershell
asd --state-dir C:\asd-data metadata check
asd metadata check C:\backups\before.sqlite3
asd --state-dir C:\asd-data metadata backup C:\backups\before.sqlite3
asd --state-dir C:\asd-restored metadata restore C:\backups\before.sqlite3
```

三条命令均不依赖后端程序。check 不创建缺失库、不升级 v1，只检查结构、完整性和声明的外键。backup 保存整个数据库的所有 scope，而非只保存当前 namespace；不包含 TOML、输出、活进程或模型上下文。

backup 的目标、restore 的 metadata.sqlite3 必须尚不存在，包括空文件也拒绝覆盖。恢复可选择新的 state-dir，同时继续使用原 runtime-dir/namespace 读取原 scope。不要删除原库再重试；损坏／未来版本会报错并保留原文件。

快照先在目标同目录生成并检查，再以硬链接发布完整文件。当前要求支持硬链接的本地文件系统；不支持则明确失败。中途强杀可留下 .asd-snapshot-*.sqlite3 临时文件；确认没有在途备份后再检查和处理，不把临时文件当成备份成功。v1 快照保持 v1，正常打开恢复目标时再事务性升级。详见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md)。

## 离线检查与备份恢复

```powershell
asd --state-dir C:\asd-data metadata check
asd --state-dir C:\asd-data metadata backup C:\backups\metadata-copy.sqlite3
asd metadata check C:\backups\metadata-copy.sqlite3
asd --state-dir C:\asd-restored metadata restore C:\backups\metadata-copy.sqlite3
```

这些命令无需后端程序。check 只检查已有文件的完整性、外键、产品标记和支持的 schema，不初始化、不迁移、不探测存活。backup 保存整个数据库的所有作用域，不是仅当前 namespace，也不包含输出文件、TOML、agent 会话内容或活进程。

backup／restore 的目标数据库路径必须不存在，空文件同样拒绝；恢复不会覆盖故障库。保留原库，用新的 state-dir 恢复，再用原来的 runtime-dir／namespace 查询相同作用域。恢复不会自动启动任何 Run。v1 快照仍为 v1，正常打开恢复库时才执行已有 v1→v2 迁移；原备份不变。

操作先在目标目录创建临时 SQLite 快照并验证，再通过同目录硬链接发布，不提供覆盖式替换。Windows NTFS 已有本轮原生证据；不支持硬链接的文件系统明确失败，不降级成覆盖重命名。正常失败清理自有临时文件；进程强杀可能留下 .asd-snapshot-*.sqlite3 临时文件，不把它视为正式备份，不自动删除其他文件。若错误明确表示快照已发布但临时清理失败，先 check 目标，不重复覆盖。

FULL 同步设置和进程强杀回归不等于物理断电验收。完整升级／回退、磁盘满、网络文件系统、目录迁移和输出记录恢复仍有各自边界。见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 与[执行记录](testing/results/2026-11-22-m1-metadata-recovery.md)。

屏幕记录与 SQLite 元数据分开保存；数据库备份不包括 output-records。停止后能读取已保存的有限画面，但没有连续录制和全流历史承诺。用法见[输出记录](OUTPUT_RECORDS.md)。

## 持久分组（schema v4）

按 [ADR-0023](adr/0023-persistent-groups-and-membership.md)，当前正常打开会在一个事务中将 v1/v2/v3 升级至 v4；原名称、Run、停止／观测时间不变。旧程序拒绝新格式，没有降级迁移。离线 check／backup／restore 可只读处理 v1–v4，旧备份只在正常打开恢复目标时升级；输出文件仍不包含在数据库快照中。分组树的循环、缺失父组和防御边界无效时拒绝检查／备份，保留源库。

当前 psmux Tab 管理焦点下 g 创建根组或子组；Tab／Shift+Tab 切换名称和默认目录，↑↓ 切换根组／子组，Enter 确认、Esc 取消。M 将所选 Session 移到稳定目标组，原 Run、后端目标及 cwd 不变。空组保留，重开仍有原组／成员关系；历史展示当前逻辑组，并不改写原启动证据。组内 n 创建的新终端预填默认目录，可覆盖或清空。提交错误要求返回、刷新核对，不自动重试可能已完成的操作。

成员版本防止旧窗口／多 UI 的 ABA 覆盖；名称版本独立。每作用域最多 256 组（含默认组）、最多 16 层，原始名称 4096 字节、清理后 50 code point／200 字节、目录 4096 字节是防御边界，未作为资源预算验收。组重命名、删除／重排及全量迁移异常矩阵仍待后续实现／验证；当前范围见[分组报告](testing/results/2026-11-23-custom-groups.md)。

## 会话行手动排序（schema v5）

[ADR-0025](adr/0025-persistent-session-order.md) 为逻辑 Session 保存同组位置／版本，分组归属与原 Run 不改。管理 Shift/Ctrl+↑↓、K/J 或 +/- 与同组相邻行交换，边界不跨组；Tab 排序仍独立。新建／移入排在已有手动位置之后，移组清除旧位置，删除组收纳由外键清除原组位置；历史页仍按 Run 时间查询。平铺只交换同组槽位。

v1–v4 通过同一事务升级；v5 的 check／backup／restore 包含位置，输出仍排除。旧二进制不支持 v5，不直接降级打开；保留升级前快照。未确认提交显示错误并重查，不自动重试；管理退出等待排序／原输入／偏好完成。实际验证与限制见[报告](testing/results/2026-11-24-session-order.md)。
