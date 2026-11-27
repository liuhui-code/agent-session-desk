# 五会话短暂 Connecting 与切换后活动标识丢失

2026-11-27；基线 `170d8a9`；[ADR-0067](../../adr/0067-five-session-warm-return-and-partial-activity.md)。

## 复现和修复

用户确认五个会话，Connecting 是短暂的。五会话初次访问完成后循环回切，原实现第一轮就淘汰即将进入的原目标画面。修复为先取目标并移出缓存，再保存原活动屏幕，随后将目标末帧转交当前连接中的只读视图。缓存仍四份／512 KiB；没有增加活动连接、终端输出历史或额外工作区屏幕。取出的缓存仅显示，不产生输入或 live cursor 权限；原连接验证和首帧原子交接不变。

第二条 RED 是一个真实调用链的受控取消：同批一个 peer Preview 已经成功，随后在另一目标等待时触发暂停和当前目标 viewport 变化，原实现整体拒绝该回执。现在只使该批中真正发生原目标 viewport 变化的样本失效，接受未受影响的已完成样本；失败／无效样本仅保留鲜度内原观察，不刷新时间。补充反例确保刚采样的那个目标自身尺寸变化时仍拒绝其旧样本。标记只覆盖原最多八个采样目标，批次结束清除；没有增加轮询或 worker。连续切换可能饿死采样是代码机制，不能推断用户现场所有蓝点变化都只有这一原因。

## 验证

[RED](2026-11-27-five-session-switch-evidence/red.log)同时记录两项失败；[GREEN](2026-11-27-five-session-switch-evidence/green.log)含五会话十五次回切、原身份／缓存上限、布局重排与活动过期回归。额外的[最终定向结果](2026-11-27-five-session-switch-evidence/final-targeted.log)包含原目标自身 viewport 改变仍拒绝旧样本的反例。

本机 `go test ./internal/ui`、`go test -race ./internal/ui`、`go vet ./internal/ui`通过；最后增加反例和一行注释后，仅补跑相关普通／race 用例。按用户 UI 快速迭代约定没有重跑全仓套件。生产文件 500 行、catalog check --generated、diff 检查和 Windows amd64 构建单列；构建不当作原生测试。新增 ADR 导致 COVERAGE.md 的 ADR 数量需重新生成，339 项设计／not_run 状态保持。

Windows 10 19045 使用同一显式开发 bundle，psmux SHA256 `ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b`。测试程序 SHA256 `208e57f0004a338255e66866da76189116db864c47e40b90c8732482f3ff5b70`：11 顶层通过、无 fail/skip、exit=0、stderr 空。两项使用真实后端，其余是 UI 模型；不能算作 11 项桌面流程。

- 原生活动测试验证真实 fixture 输出／静默／重排，并正常 Stop 自建 Run，17.23 秒。
- 五会话原生测试使用独立 namespace 和元数据库，自建五条 fixture，初次访问后回切十五次。每次连接开始前都有该原目标的相同屏幕哈希，无 Connecting 页面／live cursor，缓存总计费与份数不超原上限；验证连接后仍是原目标，17.05 秒。测试结束先关闭 UI 客户端，再对五条原身份分别 Stop；全部无清理错误。
- 这两项通过 UI 模型驱动实际原生连接／渲染，不使用 Windows Terminal 的实体键鼠或真实 Agent；没有桌面 API、剪贴板或日常会话操作。最后新增的反例仅在本机执行，不伪称已纳入前述 Windows 程序。

六个及更多会话、超界／复杂画面和复制状态仍受缓存淘汰规则约束；长时间无有效观察仍显示未知。真实 Agent、整体 Windows 桌面回切节奏、内存 RSS／长 soak 与完整接受仍开放。[原始证据](2026-11-27-five-session-switch-evidence/)不含用户终端正文。

## Windows 日常交付复核（2026-11-28 CST）

正式产品为 `asd-five-5f5e6d3.exe`，源提交 `5f5e6d3`、内嵌 VCS modified=false，SHA256 `6a96f1c34aeb5c1e51bd2a171e46696e7f9e71ef803ecaae3d5ce99c3ca3c060`。原日常 UI 正常退出后，先验证 v11 非覆盖备份，再切换清单并单次调用原快捷方式；未停止 Run、注入终端输入、恢复或迁移数据库。备份保留在 Windows 用户目录 `asd-acceptance-20261122/five-switch-delivery-5f5e6d3/metadata-v11-private.sqlite3`，不进入 Git。

[首次交付回执](2026-11-27-five-session-switch-evidence/delivery-first-result.json)失败在 60 秒内没有确认新 manager，不是交付通过。请求时间 UTC 15:43:13.987；后来观察到新 manager 14940 创建于 15:44:42.489，约晚 89 秒。仅从这两个时间不能断言延迟根因或排除外部再次打开；没有盲目重放启动，也不将后来成功改写成首次等待通过。

[后续只读核验](2026-11-27-five-session-switch-evidence/audit-result.json)确认已有正确镜像的内层 workbench、原 Windows Terminal 路径祖先及唯一 owner binding、同 manager 的一个桌宠；旧 manager 不存在。审计产品启动／输入／剪贴板 API 均为 0，task result=0；没有操作前台、关闭现有窗口或再次启动产品。备份和当前数据库均 v11／integrity=ok。

[严格类型的身份／状态统计](2026-11-27-five-session-switch-evidence/identity-v2.json)从首次只读核验保存的原始私有列表解析：发现六条记录，其中五条 available、一条 exited。备份中六条元数据 Run 身份全部保留；五条 available 身份与备份中“未停止且无 pane exit 记录”的五个候选完全相等。这是元数据候选与当前后端观测的比较，不是更新前 live 集合的独立回执，更不能作为 Agent 仍存活或原 PID／FILETIME 不变的证明。此次实际观测数量也不追溯当作用户已确认的会话数。

首次更新脚本只在内存中留了配置哈希，启动确认超时后未落盘，故无法补称前后哈希完全一致。后续只观察到 ui.toml、launch.toml、last-launch.toml 的写入时间都早于本次更新；另两份被检查的配置当时不存在。原脚本和只读审计均没有执行配置写入。

中间一次 PowerShell 数组投影给出空状态／单项错误统计，已标记为[无效观察器结果](2026-11-27-five-session-switch-evidence/observer-invalid-state-summary.json)，不能据此说会话变成一条或状态未知。改用严格校验 Session／Run／backend／terminal_state 的 Go 只读解析，得到上述 5 available／1 exited；原始用户列表和 SQLite 只留在 Windows 私有目录，公开证据仅有数量、布尔值和镜像哈希。

[最终清理](2026-11-27-five-session-switch-evidence/delivery-cleanup.json)确认两个临时任务不存在，三份未运行的上传／审计镜像按精确哈希移除，日常程序与备份保留，forced=0。以上为版本交付和现有窗口核验，不代替用户整套闪烁现场的复看、完整桌面键鼠接受或长期 RSS 测量。
