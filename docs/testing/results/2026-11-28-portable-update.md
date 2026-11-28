# 显式便携更新与程序回退验证

2026-11-28；基线 925233d；[ADR-0070](../../adr/0070-explicit-portable-update-and-rollback.md)，[操作及发布方接口](../../UPDATES.md)。源变更见本报告的 Git 引入记录；[源码指纹](2026-11-28-portable-update-evidence/source-hashes.json)覆盖 Go／module 文件，含实验／历史参考文件，不是产品模块数。

## 实现边界

CLI 显式 tag 下载、status／apply／rollback／repair／discard、无私有读取的 probe 和离线发布合同生成 package 已接通。U 保留查询，帮助指向命令说明。一个下载槽，三个程序槽和三个按槽轮转的 metadata 备份；流式缓冲 32 KiB、程序 64 MiB、合同 32 KiB、API 元数据 256 KiB。没有新依赖、常驻更新进程、启动重试或终端输入；已有旧文件不自动删。

下载核对 API SHA256／size／uploaded 及合同 SHA256／tag／revision／状态格式／固定运行库；安装另核对实际探测和本地文件摘要。合同用精确字段 JSON，拒绝重复／大小写别名／typed null。安装固定同格式 v11／UI v7 等，不借此迁移数据库或替换终端库。旧公开包无合同会拒绝；开发构建不按发布时间自动认定较旧。没有自动发布本次功能。

apply／rollback／repair 要求显式关闭确认，实际持有既有启动门禁和 UI 存活 mutex；内核文件锁拒绝并发更新。它只覆盖同一用户／交互登录会话中参与协议的目标 scope，不冒称检测所有自定义 scope 或其他用户／登录会话。旧不参与协议的程序需人工首次升级。

安装前使用原只读 SQLite 校验及新增固定独占 capture 路径 SnapshotFixed，保留 Snapshot 原行为。未知备份槽和程序槽不覆盖；对应版本已登记的非活动备份须通过结构检查才轮换。当前／前版不作准备槽。intent 先于副作用，入口同目录完整写后切换；repair 只辨别旧／新完整入口，不重做安装、备份、启动或恢复数据库。它不是断电持久性或对外部不合作编辑器的原子 CAS 证明。

## RED 与本机验证

证据目录保存 [新 API／备份 RED](2026-11-28-portable-update-evidence/backup-red.log)、[回退 RED](2026-11-28-portable-update-evidence/rollback-red.log)、[损坏下载接口 RED](2026-11-28-portable-update-evidence/download-red.log)、[固定 Snapshot RED](2026-11-28-portable-update-evidence/snapshot-red.log)、[歧义 JSON RED](2026-11-28-portable-update-evidence/json-red.log)、[空安装虚报备份 RED](2026-11-28-portable-update-evidence/empty-red.log)、[UTF-8 BOM 丢失 RED](2026-11-28-portable-update-evidence/bom-red.log)和[未知备份覆盖 RED](2026-11-28-portable-update-evidence/backup-owner-red.log)。

大小写槽的第一版测试用了非法大写 EXE 扩展，因此只是提前拒绝，不能当作该缺陷的 RED；相应两个通过日志保留为无效覆盖。改成合法扩展、登记原槽后，暂去新增拒绝检查的[定向变异失败](2026-11-28-portable-update-evidence/case-corrected-mutant-red.log)证明该保护确有作用；原实现立即恢复。它是变异证据，不是产品发布构建。

[最终局部 race](2026-11-28-portable-update-evidence/final-recheck.log)覆盖损坏／正确下载、HTTP／外域重定向拒绝、备份失败不切换、七次更新仍三个程序槽、保留更新后新数据的程序回退、已提交／未提交 intent 修复、并发拒绝、空安装与未登记备份以及 BOM 中文保留。原始 repair-slot 检查在新增补充之前已通过，不计作 RED。网络响应是可控 transport；不声称已经下载真实公网更新资产。

一次普通全包检查因沙箱禁止既有握手测试绑定 127.0.0.1 失败，见 all-tests-sandbox.log；允许回环后复验。最终本机 [Go 全包](2026-11-28-portable-update-evidence/final-all-tests.log)、[race](2026-11-28-portable-update-evidence/final-race.log)、[vet](2026-11-28-portable-update-evidence/final-vet.log)和[模块校验](2026-11-28-portable-update-evidence/final-mod-verify.log)均通过；未变化包可用缓存，Windows 结果分别追加，测试设计目录 339 条基线不改成已执行；catalog 成功不构成产品执行。

## Windows 原生状态

已准备独立目录的真实双程序字节、实际 CLI package／probe／apply／rollback、SQLite v11 备份、存活 mutex 拒绝和明确终止自有 updater 的恢复用例。此源码提交时尚未执行最终 clean 镜像；不得把交叉编译、上面的模型或已有旧原生结果记为本轮通过。正式执行／交付见后续追加。

长期资源、真实 Agent、跨格式迁移、签名／许可与完整发布接受仍开放。本轮也不关闭完整断电／全部文件锁／多用户／多登录会话矩阵。

## 正式 Windows 执行与首版交付

最终源码 **db63c5a0a99ef77f075798ebead9c31814341a89**，两个 Windows amd64 产品均由干净工作树 Go 1.27.4 构建，VCS modified=false。普通产品与去符号候选为同源码不同字节，分别见[产物哈希](2026-11-28-portable-update-evidence/native-hashes.json)及两个 build.txt；候选不是手工伪造 build/probe。日常产品 SHA256：`069e4d2ed87ed08696ea12ee0cfddb0ab0a0d42efd44534d5324a84a207970e9`。

[执行器](2026-11-28-portable-update-evidence/native.ps1)、[回执](2026-11-28-portable-update-evidence/native-result.json)记录 Windows **10.0.19045**：更新包 **17 顶层**、实例门禁 **1 顶层**、CLI **3 顶层**全部通过，零失败／跳过，三个 stderr 为空。17 项使用真实 Windows 文件／锁／替换及受控回调；CLI 包含两个无后台／确认边界和一个实际终端流程，不能全部算作实体桌面场景。详见 updater.log、instance.log、cli.log。

实际流程证明：

- 同一 clean 源码的不同 PE 字节，经实际 CLI package／probe 生成合同与校验；HTTP 提供方为受控 transport，程序内容从真实文件流式读取，不是假称已从公网 GitHub 下载。
- 独立 namespace／安装根目录，固定开发 psmux 运行库（ff9d8893…，不是官方 3.4.0 原包）；实际 fixture JSONL 独立给出 PID，并核对完整镜像／FILETIME，更新全程保留原 Run／backend 身份。
- 实际协议监听器持有 UI 存活 mutex 时 apply 拒绝；这是原生门禁合同，不是物理工作台键盘测试。退出监听后，真实 SQLite v11 快照与校验、程序文件／入口切换通过。
- 更新后写入新的逻辑显示名称，回退程序后该数据仍在；未恢复旧数据库或重启原终端。
- 另起一个明确自有 updater helper，在真实固定 SQLite 备份完成、发布入口前写出独立边界回执；刻意结束这一个 helper，等待结束后 apply 拒绝重放，显式 repair 保留旧入口并清理准备，discard 结束候选。**主动故障终止 updater=1**，不是零故障终止。未终止原 fixture；最后通过正常产品 Stop 清理该自有 Run 并等待原进程结束。

这是同状态格式、同源不同字节的更新事务／失败合同；不代替未来跨源码发布的兼容性回归、签名验证、实体 Windows Terminal／IME 或完整资源预算。不同状态格式明确拒绝。

### 日常安装

首次支持协议的版本按原人工交付机制进入 `asd-updater-db63c5a.exe`；没有用不支持 update 的 44aa154 自行升级。原 manager 21436 的精确路径／创建时间、宿主 hash、唯一 owner binding 和恰好一个原生 Tab 守卫成立后，正常关闭旧视图；验证 v11 非覆盖备份，再切换清单并仅调用原快捷方式一次。新 manager 19012 已运行，见[交付回执](2026-11-28-portable-update-evidence/delivery-result.json)。没有停止日常 Run、发送终端输入或调用剪贴板 API。

关闭前已落盘保存原始身份投影与五份配置哈希；[独立审计](2026-11-28-portable-update-evidence/audit-result.json)逐字段验证本轮前后 **7 条**完整 Session／Run／backend 身份相等、配置哈希相等、新 manager 对应唯一宿主和一个桌宠、旧 manager 不存在。七条是本轮观测身份数，不是七个活 Agent；不倒推上轮六条的状态或新增来源。原始投影／配置哈希／备份只留 Windows `asd-acceptance-20261122/update-delivery-db63c5a/`，不提交 Git。

[实际安装冒烟](2026-11-28-portable-update-evidence/installed-smoke.json)再次证明 update status 可读当前入口、pending=false／repair_required=false，私有备份 v11／integrity=ok；[已装程序 probe](2026-11-28-portable-update-evidence/installed-probe.json)具有 protocol=1、准确 clean revision 与全部格式／运行库声明。status 首次建立专用有界锁目录，不改元数据、偏好或运行程序。

UTC 02:55:06 的 manager 单次工作集 18,919,424 bytes（约 18.0 MiB）、私有 63,537,152 bytes（约 60.6 MiB）；不含 Terminal／psmux／子程序，不是更新前后配对测量、长期预算或优化成果。新增更新逻辑只按命令运行，无常驻更新进程。

[精确清理](2026-11-28-portable-update-evidence/cleanup.json)核对零上传镜像进程、两个临时任务不存在，按哈希移除六份测试镜像和四份控制器，cleanup_forced=0，日常镜像及备份保留。它不抹去上面的一个主动故障 helper 终止。

公开发布流水线未改，也没有发布新 GitHub 资产。旧公开包缺合同的拒绝行为保持；真正公网资产／CDN 下载接受须随新合同包发布后单独验证。
