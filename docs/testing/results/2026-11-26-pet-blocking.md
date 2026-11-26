# 2026-11-26：桌宠联动已有阻塞提醒

依据 [ADR-0051](../../adr/0051-pet-existing-blocking-reminders.md)。用户澄清范围为上一轮已开发的阻塞提示；本轮没有新增 Agent 接入。实现入口是 Windows 工作台 `--pet`，消费原 AttentionItem 快照，点击复用原 ID／版本核验、输入排空及会话导航；独立 pet-demo 保留为外观诊断。使用说明见[桌宠提醒](../../DESKTOP_REMINDERS.md)。

## 源码与执行范围

以本地 main 的 e975060 为底建立干净归档，再覆盖本轮相关 Go 文件；共 391 个源码／模块／素材文件逐项散列。未纳入工作区原有的 internal/backend/terminal.go 与未跟踪 terminal_failure_test.go，它们是另一条后端候选，不是本轮提醒修复。最终所有本轮 Go 文件逐字节等于归档对应文件。每轮[源码和二进制指纹](2026-11-26-pet-blocking-evidence/run2/source-fingerprints.json)保留；首轮与最终源／产物分别存储，不混作同一构建。

本地使用既有 Go 1.27.4：pet／cli 包测试通过，UI 只运行 Attention／Companion 相关行为并做 race，相关三包 vet 与模块验证通过。先增加现有 pending 投影／原事项导航测试，入口不存在形成[编译 RED](2026-11-26-pet-blocking-evidence/red.log)，实现后通过；不把编译 RED 当作旧产品运行故障。日志见 [packages](2026-11-26-pet-blocking-evidence/local-packages.log)、[race](2026-11-26-pet-blocking-evidence/local-race.log)、[vet](2026-11-26-pet-blocking-evidence/local-vet.log)、[modules](2026-11-26-pet-blocking-evidence/modules.log)。没有运行全量 UI／长期稳定性测试。

所有生产文件经 check_source_limits.py 保持 ≤500 物理行；本轮测试与 PowerShell 驱动也人工核对 ≤500。test_catalog.py render 与 check --generated 通过，339 项仍 designed/not_run，工具不执行产品测试；ADR 计数刷新为 51，不提高设计用例执行状态。

## Windows 原生局部结果

Windows 10 19045，沿用本机已散列的显式 terminal-runtime 开发候选，不改变默认后端或公开发行物。每轮均为原生 exe 执行；原生宠物测试由临时 Interactive／Limited 任务运行，只有测试自有 HWND 接受受控消息，无剪贴板操作或实体鼠标输入。

| 轮次 | UI | CLI | Pet | 失败／跳过 | 临时任务结果 |
| --- | ---: | ---: | ---: | --- | ---: |
| run1 | 10 | 5 | 9 | 0／0 | 0 |
| run2 最终 | 10 | 5 | 9 | 0／0 | 0 |

计数为顶层测试，每轮合计 24，不把两次重复执行相加为 48 个独立行为。原始摘要：[run1](2026-11-26-pet-blocking-evidence/run1/summary.json)、[最终 run2](2026-11-26-pet-blocking-evidence/run2/summary.json)。最终产品 asd-blocking-pet.exe SHA-256 为 c4fd9ef298a327d302c41b7d34f4ff483cfde731f3c4cf9fda84983f28096590；这是开发候选，不是用户日常或 GitHub 发布入口。

原生联动测试使用两个真实 psmux fixture 子进程、实际 SQLite、独立产品 attention-event CLI、真实 Bubble Tea 更新循环与 Win32 分层窗口：

1. 原会话 A 中保留未提交 DRAFT_A；B 的原 CLI 内打印 [y/N]，由上一轮已有事件接口上报 pending。
2. 猫从 idle／0 变为 approval／1，工作台仍显示 A 草稿。
3. 对精确自有 HWND 发送左键消息作为点击代理；回到 B 原表单，viewed_at 已写入，事项仍 pending／1，B 没有 input 日志。
4. 隐藏猫，显式向原终端输入 y 和 Enter；fixture 仅记录一次 y，并由来源上报 resolved。隐藏期间原生窗口状态仍更新为 idle／0；恢复后显示最新状态，未独立抓取系统托盘 tooltip。
5. 原 A 草稿仍未提交，原 Session／Run／backend pane 身份仍在。正常取消原生窗口、退出 UI，并通过各自 Manager.Stop 结束测试 fixture，无强制结束。

最终[精确资源审计](2026-11-26-pet-blocking-evidence/run2/audit.json)中，本轮五个明确测试／产品文件名与完整路径匹配的进程为 0，两个临时任务均不存在；这不是所有后端进程的全局审计。日常入口、个人状态和剪贴板未改。

另以自有窗口 nonce 属性验证：原绑定有效，移除属性后 Activate 被拒绝且 foreground 不变；没有尝试强行激活其他窗口。纯状态测试覆盖 viewed 不清除、来源 resolve／invalidate、观察未知、调用方切片隔离、旧版本拒绝与导航不发终端输入。原 UI 原生表单测试继续覆盖 Alt+N、pending 阻止自动 deliver、显式选择与草稿保持。

run2 在首轮通过后补充隐藏期间状态更新，并把原生唤醒与窗口销毁解绑串行化，随后重跑上述局部检查。首轮原始通过与最终通过分别保留，没有丢弃失败或复写源指纹。[最终审计](2026-11-26-pet-blocking-evidence/run2/audit.json)对五个本轮精确文件路径和两个临时任务名找到 0 运行镜像、0 任务；未做全局后端进程审计。日常入口、个人状态、终端设置和剪贴板均未更改。

## 尚未证明

此处的原生窗口点击为 HWND 消息，工作台在测试自己的 Tea loop 中运行；没有把它等同实体鼠标、Windows Terminal 前台／多 Tab 精确激活或日常快捷方式完整链路。CaptureOwner 的实际 WT 祖先绑定与前台恢复仍待桌面验证，失败时仅显示自行切回工作台提示，不模拟按键。没有新增真实 Agent 适配，也没有从表单文字推断 pending。

后台 Host／跨进程提醒、系统通知／声音、多工作台单实例、Explorer 重启恢复、多屏／长稳／资源预算保持开放。上一轮独立 Demo 进程的短时内存不能作为集成工作台增量内存。SQLite 保持 v9，无个人数据迁移；完整 FL-020、Windows 兼容或发布接受均未提升。
