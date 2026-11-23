# 同组 Session 行持久排序

日期：2026-11-24（原始 Windows UTC 为 2026-11-23）。基线提交 `c9bf8e4`。关联 FR-03/04/10/11/12、FL-003/004/005/011/012/013/016/018/019、[ADR-0025](../../adr/0025-persistent-session-order.md)及差异清单 U16。当前为局部实现与复验，不是整体视觉或发布接受。

## 行为与范围

管理焦点的 Shift/Ctrl+↑↓、K/J、+/- 现在按所选行分派：会话与同组相邻会话交换，跳过其他组，边界不跨组；组标题保留同父组排序。稳定选择跟随完整原 Session/Run，活动 Tab、View 顺序、连接和终端草稿不因排序改变，终端焦点的同名键仍交给原程序。

SQLite v5 增加独立位置表，保存逻辑 Session 所属组、非负位置与正版本。首次排序归一化当前同组行再交换；新建／移入成员跟在已定位成员之后。平铺沿名称／ID生成跨组槽位，再应用同组顺序，不将 Tab 顺序当作会话列表顺序。移组同事务清除旧位置，删除组的旧位置通过受约束外键清理；历史查询仍按 Run 时间，不把历史行加入当前排序，也不复制终端／对话历史。

Workspace 流式计算每组全部逻辑成员及 Run／停止事实、名称／成员／位置版本的摘要，不在 UI 保留第二份历史。Manager 重新发现真实终端；Store 在一个写事务内重验该摘要与完整有序同组行，拒绝新增／缺失／替换、改名、移组、ABA 和过期版本。摘要补上“发现返回后、提交前新增元数据成员”的缺口。SQL 事务里没有终端调用，元数据与实时后端仍不具有跨资源原子性。

UI 一次接受一笔排序，与组排序互斥；终端事件与 Tab 导航继续处理。回执只向仍存在且完整原身份／成员／名称版本匹配的行合并更高位置版本，不新增或复活会话。退出等待已接受排序、输入和偏好；任一失败取消该次退出，不自动重试不确定提交。成员版本推进时，UI 同步清除旧组位置并重排；同一成员版本的迟到移组回执保留后来已保存的位置。

v1–v4 在同一初始化事务升级到 v5；冲突回滚。只读 check 与非覆盖 backup／restore 包含位置元数据，输出文件仍排除，ui.toml 仍 v3。旧二进制拒绝 v5，不能直接降级打开。

## 来源与必要适配

固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT。home.go 的管理键分派、groups.go 的 MoveSessionUp/Down 同组交换／归一化选择性改编到 `internal/ui/session_order.go` 和 `internal/store/session_order.go`；逐文件哈希见证据及复用清单。本产品没有来源的父子 Session／归档分区，没有为迁入排序虚构这些业务。SQLite CAS、摘要、平铺槽位与异步退出是明确适配。

来源首次选择现有五个纯排序测试时，未缓存的 testify 使 GOPROXY=off 执行在 setup 阶段失败，不能写成五项通过。随后仅覆盖无关测试文件，用一个公开合成 helper 调用**未修改的生产排序方法**，核对 initial、up-boundary、swap-up、rebuild、swap-down、down-boundary 六个状态及外组不变；该 helper 一个顶层测试通过，不等于六个来源测试或全套上游接受。保留四个临时 XDG 根、上游真实 home 保护和禁止个人文件数据读写／出站网络的 OS sandbox；HOME 保持原值。来源 checkout 与两个生产文件哈希前后未变。

Context7 核对 SQLite 官方 foreign_key_check／外键说明，未用通用文档替代锁定 SQLite 的执行证据；实际 FK 清理、迁移和回滚由真实 SQLite 测试另证。没有生产依赖或 design token 变更。

## TDD 与回归

保存关键 RED→GREEN：原 Store 没有排序操作；原管理键只分派组导致会话排序未发生；位置组与当前成员不一致仍被 Workspace 接受；不核对摘要的版本接受发现后的元数据插入；UI 移组回执保留旧位置。各失败日志保留，最终相关回归通过。

本轮测试覆盖同组交换与重开、八种源键、边界／非法方向无写入、外组槽位不动、新成员／移组／收纳、完整快照缺失／新增／替换／改名／成员／停止／外作用域拒绝、两连接并发只提交一次、ABA、触发器中途失败整体回滚、待写入版本耗尽、v4 只读检查及事务迁移、冲突源文件字节不变、v5 备份恢复和损坏位置拒绝。UI 另以明确屏障验证元数据／已接受输入／偏好退出顺序，三个失败取消退出、迟到回执不覆盖新状态、终端 K/J 输入与鼠标新建隔离。

旧的三条“升级后 schema=4”测试在首轮广测中因实际 5 失败，按新 ADR 更新为 5；独立冻结 v4 建表夹具仍明确检查升级前 4 及只读字节不变，不将旧夹具整体改写成 v5。

完整 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、生产源码 ≤500 行和 catalog/generated 检查通过；成员回执补充后全部复跑。339 项目录保持 designed/not_run（216 ready／123 needs_decision），工具检查不是产品执行。

## 构建与 Windows 原生执行

222 份最终 Go／模块指纹包含两份历史 reference main 和四份嵌套历史模块文件，不全是产品模块。成员回执补充前后只改变 `internal/ui/group_dialog.go` 和 `internal/ui/session_order_lifecycle_test.go`；其余 220 份一致，Session／Store／Manager 三套测试二进制哈希也一致。

| 执行 | 顶层通过 | 含子测试通过 | 范围 |
| --- | ---: | ---: | --- |
| Session | 9 | 9 | Windows 原生 |
| Store | 73 | 115 | Windows 真实 SQLite、并发／迁移／恢复及既有 killed-writer 等 |
| Manager | 17 | 19 | Windows 模拟终端服务＋真实元数据 |
| 补充前 UI | 181 | 427 | Windows 模型／组件套件；单列历史构建 |
| 最终 UI | 182 | 428 | 最终构建重跑；包含成员回执补充 |

各套件均 exit 0、fail 0、skip 0、精确测试镜像剩余 0。环境 Windows 10 build 19045；不是交叉编译替代执行，不相加重复套件。

最终产品 SHA-256：`32e0d77538709ef269d8c7f743851df02c6bf7947b7346da960b0104f2038289`；最终 UI 测试：`0d3f1397081cc45b93d0ab28d0a2d5a9f302efd6b86487ee204a6e3461de8021`。补充前分别为 `062dca624139610b2134ae3885db4c8ab8d0f8f92e179a3b7d211ec7dd2b87ec`、`32bd9e393dfbca0ee2a19e6558afa0940b8158fa461378b2849121ef8d97d05e`；其他三套精确哈希见 manifest。

## 正式产品的隔离 ConPTY 流程

沿 ADR-0019 的显式开发 bundle，固定补丁 psmux `ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937`、ConPTY DLL `e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492`、OpenConsole 文件及 PowerShell 7 哈希。driver 观测实际 DLL和原 PID/FILETIME；不扩展官方默认系统宿主兼容结论。

最终第五轮使用最终正式产品，固定 120×32，保留 12 个连续编号捕获和 32 条已执行事件（text 11、key 19、manager-exited 2）：

1. 用前一正式构建 `2f022aad…` 的 CLI 启动 A/B 两条普通 PowerShell，真实 metadata check 为 v4，记录原 Session/Run/backend、PID/FILETIME 与历史。
2. 新 UI 打开并事务升级 v5；A、B 各保留一份未提交原生草稿。管理 K/J/K 排序，持久顺序依次 B/A、A/B、B/A；Tab 顺序不改。
3. 创建 Other 根组，用 M 将 B 移入，再移回默认组。两次观察 B 的旧位置版本为 0，成员版本推进到 2；A 的手动位置保留，B 作为未定位成员排后。再 K，B/A 得以保存（B 位置版本 1，A 位置版本 4）。
4. 管理 q 正常退出，原 shell 仍存活；旧构建 check 拒绝 v5，数据库 SHA 前后相同；新构建 check 为 5。backup／restore 后同作用域的顺序、位置与版本一致。
5. 重开自动进入排序首项 B，A/B 原草稿仍在；分别回到原终端显式 Enter 一次。两个原 PID 的输出与独立 PSReadLine history 各只有对应受控命令一次，原 Session/Run/backend 未变。再次管理 q 正常退出，两次退出码均为 0，再显式 Stop 两条所属 Session。

最终六个已登记原身份（两 shell、两 manager、两 driver）全部结束，强制结束产品 0；精确产品镜像剩余 0，无剪贴板 API，也没有访问桌面。这是实际产品的隔离 ConPTY 证据，**不是 Windows Terminal 字体、物理 Shift/Ctrl、IME 或用户视觉接受**。真实 agent 当前未安装，不能用 shell 替代其兼容验收。

## 早期流程与驱动失败

| 轮次 | 构建 | 结果／证据边界 |
| --- | --- | --- |
| 1 | 补充前 | 5 捕获，排序、迁移、旧版拒读及备份恢复已执行；第二 UI 启动时将旧注册文件误算为唯一当前 driver 而失败。第二 driver 未完成初始身份登记，不能补造；所选第二 manager 私有清理终止，forced product=1。 |
| 2 | 补充前 | 8 捕获；等待 Tab 切换完成前就发 F6，产品正确拒绝并显示“正在切换 Tab”，不重发该动作。第二 manager 清理终止，forced product=1。 |
| 3 | 补充前 | 9 捕获、20 事件完整通过，两条原 Run 草稿及正常退出成立；不归给最终构建，也未包含新增移组流程。 |
| 4 | 最终 | 0 捕获／0 UI 输入，fixture B 身份文件仍被写入时非共享读取失败；该轮已按两条所属 Session 清理，只有 A 完成初始进程登记，其他身份不由后轮回填。forced product=0。 |
| 5 | 最终 | 使用同目录完整发布、共享读取及只重试观察的 fixture；12 捕获、32 事件完整通过，如上。 |

每次新轮前均已得到前轮终结及精确产品剩余 0；未因观测超时重启仍活跃的同一流程，也没有重放不确定的输入。跨五轮的已登记身份、尚存的专属 driver 注册线索和精确镜像审计另存；缺失登记不是零成本／完整生命周期证明。路径审计沿用既有窄扩展 DOS 等价规则，不删除 PID／FILETIME／路径检查。

## 归档与剩余门槛

[证据 manifest](2026-11-24-session-order-evidence/manifest.json)包含原始／归档字节 SHA。UTF-16／BOM 日志转 UTF-8，用户目录和本机路径替换为占位符，辅助 Go 改为 `.go.txt`；只收合成资料，不收 runtime、数据库、注册密钥或二进制。只读一致性命令：

```sh
python3 docs/testing/results/2026-11-24-session-order-evidence/local/verify-session-order.py
```

该命令核对归档、源码／构建、测试计数、状态／身份及清理，不执行产品或 Windows 场景。完整资源预算和摘要流式扫描成本尚未测量；不把绿测或快照当作完整资源接受。

继续 U19 显式停止管理 UI、U06 长卡片、U15 比例保存、U18 鼠标／可见页复制、U20 适用设置和剩余主界面／输入矩阵。U16 的完整桌面／上游视觉、真实 agent、最终兼容与发布仍开放，原目标保持未完成。
