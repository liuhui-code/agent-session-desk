# 同级组排序、持久顺序与独立退出屏障

日期：2026-11-23；基于 `ffd4d79` 加本轮修改，185 份 Go／模块指纹见[源清单](2026-11-23-group-order-evidence/local/source-fingerprints.json)。管理焦点选中组标题后，Shift/Ctrl+↑↓、K/J、+/- 已接入同级移动和持久保存。**本轮不包含 Session 行手动排序、组删除／收纳或完整上游体验接受。** [ADR-0023](../../adr/0023-persistent-groups-and-membership.md)补充合同，SQLite v4、ui.toml v3 和生产依赖不变。

## 固定来源与产品行为

来源为干净的 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT。home.go 的八种管理别名、稳定选中组与显式保存，以及 session/groups.go 的 moveGroupAmongSiblings 同父过滤／归一化交换为依据；文件 SHA、Git blob 和许可见[来源记录](2026-11-23-group-order-evidence/local/upstream-source.json)，派生文件头和机器复用清单已同步。

隔离辅助测试直接执行未修改的固定 GroupTree，实际保留[四份顺序快照](2026-11-23-group-order-evidence/local/upstream-order-snapshots.json)：初始、根组上移、根组下移、子组上移。第二根组上移会跳过第一根组的完整子树；两个子组交换仍属于原父组。Darwin sandbox 禁止个人目录和出站网络，原测试／TestMain 经只执行本辅助测试的 overlay 隔离；没有运行完整上游应用或个人会话。

产品以稳定 Group ID 和完整有序兄弟快照提交，在 SQLite immediate 事务中核对各组元数据／版本及兄弟集合，拒绝邻居变化、新兄弟、并发或 ABA。有效移动按当前位置归一化 ordinal 再交换；双方及其他序号实际变化的组增加版本，版本耗尽整体拒绝。名称、父组、目录、子树、成员版本和原 Run 不改，失败不留下半次交换。

默认组沿既有产品约束固定根首位，不能自身移动或被跨越；同级首尾无动作／版本变化。并列 ordinal 延续产品稳定 ID 顺序，默认组优先，这与来源路径／名称模型不同。没有迁入来源 remote、conductor／Maestro 或 Session 排序业务。

UI 一次接受一笔排序，期间的额外排序和管理弹窗操作提示等待，不积压旧编辑；原终端事件和 Tab 切换继续处理。只有提交确认才应用顺序，回执按组版本合并，保留当前稳定选择、折叠和原活动 Tab，不连接另一 Run 或注入输入。提交前发现失效，旧结果不能回退新版本或重复应用。

管理退出先等待已接受排序，再沿既有终端输入／偏好屏障处理；排序失败取消该次退出，先读新元数据，再允许新的显式编辑，不自动重放。排序回执到达前的偏好失败或终端失败也取消待退出请求，保留错误。这是来源同步 forceSave 在本产品异步存储中的明确适配。

## RED → GREEN 与本机验证

- 存储／Manager 首条测试因缺少接口失败；接入后，真实数据库排序跳过展开子树，跨重开保留成员／原 Run；Manager 在 nil 终端对象下也可排序。
- 旧的相同 ordinal 可让自定义组排到默认组之前，行为测试失败；保留默认组优先并归一化后通过。存储没有借此迁移或批量重写旧库。
- UI 测试草稿先有未使用导入，原日志保留；去掉导入后，实际 K 行为因未接入而失败。接入后验证稳定目标、完整兄弟快照、折叠、原连接／Tab／草稿和旧发现拒绝。
- 排序提交不确定后可立即再次编辑，测试失败；加刷新屏障后通过，未自动重试失败操作。
- 等待排序退出时，偏好或终端先失败，后到成功回执仍会退出；两项行为测试失败。修正取消待退出请求后，错误保持可见。成功路径另验证同时阻塞排序、已接受输入和偏好保存时依次等待三者。
- 补充真实双连接并发仅一笔提交、ABA、邻居改名／新增兄弟、独立子树变化、子组不跨父组、固定边界、非法方向／缺失快照／作用域、版本耗尽，以及第二笔 SQL 更新失败后完整回滚。UI 覆盖八种管理别名、同名终端输入原用途、忙时不排队、当前选择和较新回执保留；这些增补不全部声称经历独立 RED。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 均通过，含缓存结果；source limits 和 catalog 检查通过。TC-SESSION-008 补充设计，仍为 339 项 designed/not_run、216 ready／123 needs_decision。文档检查不计作产品执行。

## Windows 原生执行

Windows 10 build 19045，实际运行本轮三个测试程序，计数不累加重复构建／执行：

| 套件 | 顶层／含子测试 PASS | FAIL／SKIP |
| --- | --- | --- |
| Store | 58／84 | 0／0 |
| Manager | 15／17 | 0／0 |
| UI | 145／327 | 0／0 |

[原生结果](2026-11-23-group-order-evidence/native/native-tests/outcome.json)保留退出码、完整日志和三个测试映像无残留。Store 使用真实 SQLite／文件；UI 服务 fixture 与受控异步屏障证明模型合同，不冒充桌面存储故障或真实后端断连实验。完整后端七项原生场景本轮没有重跑。

实际桌面产品 SHA-256 为 `4bc4190270c3e9a34eecb3cbc6e07f9b943d49d03520184295c9b6c1d2645838`；[构建清单](2026-11-23-group-order-evidence/local/binary-hashes.json)与 build-info 保留精确版本。它是基于 ffd4d79 的未提交源码构建，Go 文件与本轮最终指纹一致，不称为提交后重建。使用 ADR-0019 的显式开发 bundle；官方默认依赖、跨权限和完整 Windows／发布接受不升级。

## 两条普通 shell 的实际桌面

普通用户桌面 Session 2，Windows Terminal 1.25.3131.0、PowerShell 7.7.0。160 条连续动作回执全部成功：36 keys、87 state、23 Unicode text、7 capture、3 paste、3 resize、1 finish；其中包含失败等待期间的只读轮询，不把它们算成额外产品能力。

- A 在原终端输入未提交命令，并追加字面量 KJ+-；B 保存另一条未提交命令。创建两个根组和两个子组，将 A 移入“审核”；排序期间 B 留在默认组。
- K、J、Ctrl+↑、Ctrl+↓、Shift+↑、Shift+↓、+、- 各完成一次根组交换。两个根组的版本由 1 依次到 9，子树／成员／Run 不变；折叠乙组按稳定 ID 保留。默认组和首尾边界操作保持整个组快照及版本相同。
- 对两个子组再执行一次 K，子组版本到 2、序号互换，父组／根组顺序保持。元数据快照和 UI 退出／重开保留最终顺序与双组折叠。
- 18 次只读元数据快照分别核对来源身份、各次版本、边界、备份和重开；排序阶段的原成员版本及 Run 记录相同。桌面没有注入并发写故障；该部分由真实双连接及事务失败测试证明。
- 10 次原生 capture 均保留编号 1–40，各一次；8 个未提交阶段的最后非空行与完整草稿逐字相同。最后各明确 Enter 后，仅一次 `ORDER_A:13588`／`ORDER_B:17616`，对应原 shell PID，FILETIME／完整路径前后相同。

第一次 K 后的等待断言误以为选中组仍显示数字快捷键；实际选中标记会替代该 gutter。只读核对任务仍 Running、最后已完成 087 回执以及两个根组 revision=2，确认 K 已提交。保留[原失败](2026-11-23-group-order-evidence/native/flow-failure-first.json)及[只读观测](2026-11-23-group-order-evidence/local/inspect-after-wait-failure.json)，改按标题相对位置核对，从下一条回执继续，没有重放 K、创建、移动或终端输入。[续接记录](2026-11-23-group-order-evidence/native/flow-resume.json)和最终版本证明原目标一直相同。

七张未经编辑的 PNG 已逐张检查：[初始展开子树](2026-11-23-group-order-evidence/native/desktop/order-initial-subtree-120-public.png)、[K 跳过子树](2026-11-23-group-order-evidence/native/desktop/order-root-K-up-120-public.png)、[八种别名之后](2026-11-23-group-order-evidence/native/desktop/order-eight-aliases-120-public.png)、[子组交换](2026-11-23-group-order-evidence/native/desktop/order-child-up-120-public.png)、[63 列](2026-11-23-group-order-evidence/native/desktop/order-grouped-63-public.png)、[49 列](2026-11-23-group-order-evidence/native/desktop/order-grouped-49-public.png)、[折叠后重开](2026-11-23-group-order-evidence/native/desktop/order-folded-reopen-120-public.png)。窄窗使用既有滚动／上下／单列布局，不声称所有组同时可见或完整视觉接受。

## 审计、归档及剩余范围

两条原 Session/Run 经完整身份核对后显式停止；2 shell＋2 UI＋1 Terminal 共五个原 PID／FILETIME／路径确认结束，已知私有映像为 0，两项自有任务不存在。专用 settings.json 恢复原 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。剪贴板文本与格式名恢复，非文本内容等价未验证；Caps Lock 前后均为 true，没有改为关闭。

独立核验脚本首次在本机因无 tomllib 无法启动，原版本和[说明](2026-11-23-group-order-evidence/local/verification-first-failure.json)保留；改用已有 json／re 仅核对本夹具实际五个标量字段，再直接运行归档版通过，不宣称新增通用 TOML 解析器。未重跑产品来掩盖该工具失败。

[407 项证据清单](2026-11-23-group-order-evidence/manifest.json)记录 ZIP SHA、原／归档哈希、BOM／UTF-16 解码和已知用户名／路径脱敏，PNG 保持原字节。运行密钥、数据库／备份、剪贴板原文、设置备份、历史及程序未提交。[独立核验](2026-11-23-group-order-evidence/verification.json)检查精确源指纹、原生计数、连续回执、版本／快照／原草稿、源顺序和清理，不产生额外产品执行或完整验收结论。

继续组删除／收纳、Session 手动排序、上游整体差异清单和界面接受。长目录缩放锚定、其他 IME／焦点／鼠标矩阵、真实 agent、最终资源预算／长稳、WSL 不存在环境和发布门槛仍开放。字面量 K/J/+/- 的 Unicode 注入不代替物理大小写／IME 候选接受；Ctrl／Shift 方向键的本轮实机路由单独由明确按键及实际排序证明。
