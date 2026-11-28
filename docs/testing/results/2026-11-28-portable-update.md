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
