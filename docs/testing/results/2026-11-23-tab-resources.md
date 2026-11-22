# 2026-11-23：Tab 资源测量与脚手架修复（基线未完成）

本轮修复了资源工具对新 Tab 焦点的错误假设和 UTC 时间误判；**新的 0／1／5／10 会话资源基线仍未完成**。第三轮在 1 会话正常输出后无法确认 F6 切回管理焦点，拒绝发送 q 并按失败退出。不能用早期列表 UI 的通过结果替代当前 Tab 路径。

基础提交 `921b67013ac779e5f9d847e796d32acfb22e264a`，产品 Go 代码未改；本轮为 PowerShell 测量脚手架、回归与文档变更。关联 ADR-0003/0016、FL-003/006/015/016/018，保持实现中、未发布。证据见[目录及复核方法](2026-11-23-tab-resources-evidence/README.md)。

## 已落实的测量边界

- 独立 `windows-resource-ui.ps1` 管理启动／退出。非零会话需出现正常终端页脚，再经普通 TUI 输入路径向已知 fixture 提交一次标记，独立 journal 确认恰好一次；连接中／只读错误不能作为开始采样的依据。缺确认不重放输入。这不是通用 agent 就绪协议。
- 退出只读最后一个非空页脚。终端／历史模式先发一次 F6，观察管理页脚后才发 q；已在管理焦点直接 q；未知焦点拒绝发送。纯模拟驱动回归验证路由合同，真实特殊键传输另行验证。
- 共享 JSONL 读取保留 UTC `DateTime` 与小数秒，拒绝不明时区。覆盖校验、25 秒负载、采样时长、身份／归属算法和失败清理规则均未放宽。
- CI 资源步骤加入纯 UI helper 回归。本轮只在指定 Windows 主机实际执行；没有推送或新的 GitHub Actions run，不把 workflow 修改视为已执行。

## 三轮原生执行

| 轮次 | 结果 | 实际范围 |
| --- | --- | --- |
| run1 | 失败，未开始采样 | 外层包装器用数组 splatting 传命名参数，导致 process-metrics helper 参数绑定失败；改为 hashtable。 |
| run2 | 失败，75.85 秒 | 2 个已建场景、4 个登记阶段、53 次快照，其中 48 次属于阶段；在 1 会话输出覆盖检查失败。零会话正常清理一次，后续有 3 次精确持有句柄的失败兜底清理。 |
| run3 | 失败，96.23 秒 | 2 个已建场景、5 个登记阶段、65 次快照，其中 60 次属于阶段；越过输出覆盖检查，完成恢复空闲后，F6 未确认管理焦点，q 未发送。零会话正常清理一次，后续有 3 次失败兜底清理。 |

run2/3 的实际登记阶段分别为：零会话 idle／ui-closed；1 会话 idle／selected-output，run3 再完成 recovery-idle。**没有 5／10 会话或直接管道对照的新结果**。所有快照保留，诊断快照不混入阶段统计；[run2 表](2026-11-23-tab-resources-evidence/run2/evidence/summary.table.md)与[run3 表](2026-11-23-tab-resources-evidence/run3/evidence/summary.table.md)明确标注原实验 fail，仅供部分数据复核，不作为完整资源结论。

## UTC 误判的证据与 TDD

run2 fixture 原始完成时间为 `2026-11-22T19:07:43.6999814Z`，实际输出满 25 秒、2500 行／640000 字节。输出阶段于 `19:07:33.8939113Z` 登记完成，采样耗时 12.169 秒。

PowerShell 7.7 默认 JSON 解码已得到 `System.DateTime / Kind=Utc`。旧代码 `DateTime.Parse($done.finished_utc)` 隐式先转字符串，丢失 UTC 与小数秒；在 China Standard Time 主机得到 `2026-11-22T11:07:43Z`，错误地移动了 8 小时。直接保留已解码值则仍为原始 UTC ticks。[官方解码规则](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/convertfrom-json?view=powershell-7.7#notes)支持这一类型判断；未新增 `-DateKind` 的最低版本要求。

RED-03 经真实共享 journal 读回固定 ISO UTC 完成记录，复现时间／精度丢失；GREEN-03 保留 UTC 值后通过。run3 原 fixture 完成 2499 行／639744 字节，完成时间 `19:17:20.512245Z`，输出阶段于 `19:17:10.6918178Z` 登记完成，采样耗时 12.162 秒；真实完整负载和后续 `after-load-1` 标记通过。这只关闭了本次 UTC 缺陷。

此前 RED-01 复现终端焦点盲发 q 无法正常退出，GREEN-01 检查先确认管理焦点；RED-02 为缺少新启动就绪判断的函数错误，GREEN-02 检查连接中拒绝、正常终端与空管理工作区。第二条 RED 不是一次真实 ConPTY 就绪失败，不将其夸大。首次 RED 包装器被外层 PowerShell 5 的错误流提前终止，重取预期失败的日志与首次 bootstrap 日志均保留。

## F6 的最小诊断与未决项

第三轮失败后另建一条受控 fixture 的短诊断，不重复整套资源实验。原正式构建仍可提交普通标记，但命名 F6 未改变页脚。随后独立诊断构建记录应用收到的消息；最终临时 ultraviolet 副本另记录 Windows 原始 `KEY_EVENT_RECORD`，正式产品与依赖未修改。

| 独立探针 | 实际结果 |
| --- | --- |
| 原构建：命名 F6、Win32 F6 按下／释放编码 | 均未确认管理焦点，拒绝 q；不是资源通过。 |
| 应用键消息／所有消息日志 | 普通标记与 Enter 出现；没有观察到 F6 按键或对应未知输入消息。 |
| 将 VT F6 字符封装为 Unicode 输入记录 | 未确认管理焦点。 |
| 十六进制 VT／Win32 F6 | 未确认管理焦点；后续普通文字与 Enter 仍能进入原 fixture。 |
| 小写 f6 与认证后的原生 `send-key f6` | 未确认管理焦点；认证内容未写日志或归档。 |
| 临时 ultraviolet 控制台记录 | `vt=true`，观察到普通字符和 Enter；未观察到 F6 输入记录。 |

这些事实把范围缩到本机测量 driver 的特殊键路径与应用控制台读取边界，**尚未确定根因属于 psmux、系统 ConPTY 还是读取模式**。下一步应在同一个受控 pane 对比独立原始按键读取器与当前 TUI，定位丢失层，再修复并复验原最小场景；随后才能重跑完整资源档位。不能通过重试 F6、盲发 q、杀 UI 后记正常退出，或把诊断构建混进资源结果来绕过门槛。

Microsoft 的 [ConPTY 键盘编码说明](https://github.com/microsoft/terminal/blob/main/doc/specs/%234999%20-%20Improved%20keyboard%20handling%20in%20Conpty.md)仅用于设计编码探针，不被当成该主机已支持的证据。此前 Windows Terminal 桌面上的 F6 局部结果与本次测量用 psmux ConPTY 是不同宿主条件，不能互相覆盖。

第一份 key-only 日志探针的外层回执误读了前一个正式构建摘要；归档内 `focus-key-probe/summary.json` 才是该次诊断构建的实际哈希与结果。另有一次脚本复制因目标已存在而在启动前失败，修正为覆盖自有诊断文件后才运行。诊断材料保留这些限制；最早日志变体未保存当时完整源码指纹，不声称所有诊断构建均可从归档逐字节重建。

## 环境、检查与清理

Windows 10 家庭版 `10.0.19045`，4 逻辑处理器、16,659,688 KiB 可见内存，PowerShell 7.7.0、China Standard Time。正式应用在独立测量用 psmux ConPTY 120×32 中运行，history limit 2000、warm off、warm-pool-size 0；不是 Windows Terminal 桌面实验，也未采整机计数。

Go 1.27.4 在 darwin/amd64 构建 Windows/amd64 EXE，随后在上述 Windows 原生执行。应用 SHA-256 `3ea19887436c4f803e988d64186e54762b86d5879132686023b5d13e7f0d15b4`；fixture `107f94c3e886ef26c32cd4f4e1fa14cc22bbe3163510d74b847e4239dc41bbb1`；psmux 3.4.0／driver `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`。构建和实际执行分别记账。

远端为真实 Git bundle checkout，加本轮明确脚本覆盖；run2 的 113 项、run3 的 115 项 Go／module／工具源码前后指纹一致。run3 新增覆盖 UTC helper 与测试，共五个覆盖文件；本轮未改生产 Go 代码。纯 UI helper 回归脚本提交前去除一处尾部空格，行为不变；归档 `sources/` 保留实测字节，差异哈希见 `post-run-formatting.json`。

最终 `go test ./...`、race、vet、module verify 通过；14 项统计工具回归通过。Windows 实际执行 UI helper、69 项合成归属／路径规则和共享 journal／UTC 回归通过，但它们不等于原生资源实验通过。500 行与 catalog 校验通过，339 项设计保持 designed/not_run。

最终审计核对本轮 36 条精确自有映像路径，未发现剩余活进程；三项测试任务在终态 Ready／LastTaskResult=1 后删除，剩余任务为零。没有按进程名终止其他任务。诊断代码仅在临时副本，正式 `internal/` 无本次 DEBUG 记录器。

资源计量仍分别报告管理器、后端、fixture、辅助进程和 driver／sampler，缺测不补零；CPU 为可见相邻身份计数差的下界，工作集可能重复共享页。run2 两次观察包装器还误序列化了 Get-Content provider 元数据，单份约 13 MB；公开材料仅保留原本需要的身份／阶段字段、原始捕获哈希和大小，其观测干扰不能排除。run3 已使用有界纯文本观察。

这次失败不产生新内存预算、无泄漏或完整 M0/M1 结论，也不与旧 Windows Server 2025 结果进行性能高低比较。真实 agent、完整上游视觉对照、子程序鼠标／选择复制／搜索、长稳及发布门槛继续开放。
