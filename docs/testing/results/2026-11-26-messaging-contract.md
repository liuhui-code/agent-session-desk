# 消息接口模拟、竞争与中断恢复验证

日期：2026-11-26。用户当前没有真实 agent，授权先用模拟把后端接口稳定性验证做好。本轮复用现有 `cmd/asd-fixture`，增加四个测试文件与[验证合同](../MESSAGING_CONTRACT.md)，生产协议、schema v6、前端及 M/Move 均未改变；没有新增模拟服务或产品常驻进程。依据 [ADR-0046](../../adr/0046-backend-only-message-delivery.md)。

## 构建与执行范围

正式 Windows 构建以 `b5bb4169073d5ee72d40b25d7003e59de2e7c5ea` 的 clean git archive 为基础，只覆盖本轮四个新测试文件。逐文件及测试 exe／fixture 散列见[指纹](2026-11-26-messaging-contract-evidence/fingerprints.json)。工作区原有 `internal/backend/terminal.go` typed-nil 候选及其未提交测试未纳入原生构建，也不能据本轮结果认定切换退出根因已解决。

运行器校验原安装清单及本轮四个 exe 的 SHA-256，使用其已有显式 terminal bundle；没有更新日常 exe／manifest 或迁移个人数据库。测试仅使用临时 state/runtime、独立 namespace 与合成数据。无桌面／剪贴板 API、无定时任务。运行器见[首轮](2026-11-26-messaging-contract-evidence/native-message-contract.ps1.txt)和[超时补充](2026-11-26-messaging-contract-evidence/native-message-terminal-final.ps1.txt)。

原生环境为 Windows 10 `10.0.19045`／amd64，Go 1.27.4；补充摘要记录实际 OS 版本。首轮 terminal 测试源码另存为 `.go.txt`，补充版本增加两个超时子项，分别保留源码／exe 指纹，避免将补充测试归入首轮二进制。

| 执行 | 结果 | 能证明的范围 |
| --- | --- | --- |
| 本机五个相关包的消息选择集 | 18 个顶层 PASS、无失败；其中 FaultWorkerProcess 无环境时为空入口，实际行为测试为 17 个 | session／terminal／store／manager／cli；16 个子用例另外包含 ACK 与五个故障点；不是全仓库回归 |
| 本机相同选择集 race | 五包通过 | 实际 SQLite 与模拟传输路径的 Go race；不是 Windows race |
| 首轮 Windows | session 2、terminal 3、manager 8，共 13 顶层 PASS，失败／跳过 0 | 其中只有 SimulatedAgentMessageContract 是真实 psmux／普通受控 CLI 闭环；其余明确为模型／协议模拟／SQLite／工作进程测试 |
| 补充真实期限到期边界 | 本机 terminal race、Windows terminal 3 顶层均通过，13 个传输故障子项 | 加入认证无响应、ACK 无响应，等待真实 200ms context deadline，不用提前 cancel 代替超时 |

首轮 Windows 终端结果由补充版本取代；session 与 manager 未改、不重复计数。最终仍为 13 个不同顶层测试，不能把两轮相加成新覆盖。原生首轮及补充原始日志、摘要均在[证据目录](2026-11-26-messaging-contract-evidence/)。本机日志为工作树局部执行；正式原生源码隔离如上。

## 实际核验

- 身份／内容：完整 Session/Run 和稳定请求 ID；1、64 KiB−1、64 KiB 正文接受，0／超界、无效 UTF-8 与不支持的控制字符拒绝；中文、emoji、组合字符、引号、反斜杠、CR/LF／Tab 按声明合同处理。JSON 保留公开回执字段，不公开 attempt/fingerprint 或伪造 accepted/completed。
- 传输：64 KiB 正文下，分片 ACK 能完成；错误认证、错 nonce、重复／过长／截断 ACK、丢回执、认证／部分命令写入／等待回执阶段取消、实际超时、ACK 后不结束响应均不能报告成功。连接由调用结束关闭；错误不包含认证哨兵值。
- 并发／存储：16 个独立 Store 同时领取同一请求，只有一次输入调用和一次成功回执。真实 SQLite 写锁阻止提交领取时，输入次数为 0，释放后请求仍 queued。
- 中断：五个精确屏障暂停实际独立测试进程，再由父进程 Kill 该归属 worker；不是杀真实 agent 或同名进程。重开 SQLite 后，queued 仍 queued；claimed／partial／written 均 unknown；receipt 为 input_written。独立 fsync 旁路文件分别为无正文、UTF-8 中间截断前缀或完整正文；重用原请求 ID 能找回相同记录，unknown／written 不再调用输入。旁路观察不能升级产品状态。
- Windows 原生闭环：两个实际 asd-fixture 经 psmux 运行，无 TUI。向 A 明确输入合成标记，再经公开 read 建立来源快照，选中一行建单并 deliver 到 B；另一个明确 LF 请求提交受控程序草稿。独立输入 JSONL 核对收到的来源副本。UTF-8 BOM 任务文件的中文／emoji、组合字符、符号和两行正文被实际程序读取；同一 send --deliver 重试返回原回执。两个 CLI 争抢一条 queued 请求仅一个成功，B 最终恰好四条输入行。停止 B 后旧请求投递失败，离线 status 保留 queued 与原 Run；原身份 Stop 清理均返回成功。

fixture Scanner 观察的是终端行规处理后的行文本，不能据此声称原始 CR/LF 字节完全未变化；本机传输与旁路故障测试另核对精确字节。原生 Stop 成功也不等于完整独立 PID／生命周期审计。

本轮原生 CLI 路径在同一测试进程内调用两次公开 cli.Run，各次构建独立后端／Store；不是两个外部 asd.exe 进程竞争。进程中断由另外的真实测试 worker 覆盖，CLI stdout 传输中断与外部命令进程竞争仍可继续扩展。

## 反例与保留边界

在独立源码副本故意将 ClaimMessage 的落盘状态从 unknown 破坏为 queued，claimed 屏障后杀进程的恢复测试稳定失败：`recovered state=queued want=unknown`。见[受控 RED](2026-11-26-messaging-contract-evidence/isolated-claim-mutant-red.log.txt)。正常实现随后在本机及 Windows 通过。此 RED 是证明回归测试能拦截该破坏，不是声称现有产品复现了 queued 重放缺陷；未修改生产文件制造历史 bug。

构建准备曾遇本机 Python tarfile 不支持 filter 参数，改为解包本地 git archive；首个副本测试启动过早、未命中最终变异，已舍弃该通过结果，保留的是准备完成后的明确 RED。文档 catalog 首次误用 `--check` 被参数解析拒绝，改为 `check --generated` 后通过；它只校验 339 项设计结构，不执行产品测试。500 物理行检查通过。

本轮未发现必须修正的生产逻辑失败，因此交付为可重复的接口回归与故障验证。真实模型 agent、FX-READY／accepted／completed、自动消费、共享 Host、多请求全局 FIFO、Windows Terminal 桌面输入、UI 长稳和发布接受继续开放。339 catalog 保持 designed/not_run；本报告仅为 TC-SEND-001/010/013/015/016、TC-CTL-007/010 等相关不变量提供局部证据，不能将共享 Host／完整工具调用／原始字节场景整体判为通过。[STAB-001～005](../DEFERRED_STABILITY.md)的界面／资源专项继续保留。
