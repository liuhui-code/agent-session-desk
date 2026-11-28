# 指令片段库：Windows 补验、资源观察与本机交付

2026-11-28。产品源码为 clean `c2ac2c0`（完整 revision 以指纹为准），`vcs.modified=false`；Windows amd64 程序 SHA256 为 `3bfe96fafaa82765f474b1e3d32cdd42c514a169b7a3805ed103459e2789dea1`。本轮没有改变产品代码、依赖、schema 或键位；补充 [ADR-0074](../../adr/0074-bounded-instruction-snippets.md)及[首轮报告](2026-11-28-instruction-snippets.md)。以下均为有界局部证据，不提升完整 Windows、真实 Agent 或发布接受。

## 发现／停止的反馈循环

先在 Windows 10 19045／Windows PowerShell 5.1.19041.6456 的独立目录，用同一产品及固定终端库运行三组双夹具循环。实际 57 次 CLI 调用包含六次 start、十五次 list、三十次 preview、六次 stop；每轮以创建回执核对新鲜列表和原 Run，全部成功，六个原夹具正常结束，私有镜像为零。见 [cli-loop.json](2026-11-28-snippets-windows-validation-evidence/cli-loop.json)。

最初观察器用未共享的 ReadAllLines 读取仍在写的夹具日志，失败发生在第一次 Start 已成功之后。改用共享 ReadWrite／Delete 读取并核对原回执后，从原 Run 续跑，没有重新创建它。该仪器失败不计产品失败或额外通过轮。

随后同一原 UI／两个 Run 完成片段操作；UI 关闭后再冷 list 和 stop，仍核对创建时的完整 Session／Run／backend。两条原身份在首开前、续跑、操作后、关闭 UI 后一致；两个 stop 成功，原夹具结束。先前一次“list 可见但 stop 找不到”的真实拒绝本轮没有复现，不能宣称原因已修复，保留 STAB-010。

已确定的仪器错误与该拒绝分开：旧清理脚本假定 backend 名称为 `asd_SESSION`，实际合同和创建回执为 `asd_SESSION_RUN`；它因此在真正 stop 前自行拒绝。原失败 JSON 不改写。用户现场的冷枚举不完整风险仍遵循 ADR-0066／0073 的不确定与目标复核合同，本轮没有放松守卫、增加操作重试或后台清理器。

## 真实 Terminal 中的 F7

使用独立 Windows Terminal 1.25.3131.0 portable 包，复用既有 pinned runtime：psmux `ff9d8893…fb7b16b`、conpty.dll `e971856f…71f492`、OpenConsole.exe `b17e48b0…9a775`。不更换产品依赖决定。原候选 ZIP `c596345b…a6bc8`、官方 Terminal ZIP `273d6f15…d3437` 与实际宿主 exe `337c583d…93fc9` 已再次核对；宿主 FileVersion 为 `1.25.2611.09001`，保留其与发行 tag 的不同格式，不凭资源版本号猜测来源。见[实际文件审计](2026-11-28-snippets-windows-validation-evidence/artifact-audit.json)。完整字节、562 个受跟踪 Go／module 文件和驱动指纹见 [fingerprints.json](2026-11-28-snippets-windows-validation-evidence/fingerprints.json)；其中包含两份 docs 内历史参考 Go 文件，不把它们视作产品模块。

一份 Global 和每个原 pane 目录下的 Project 文件各含 64 项、每个填充正文 3400 字节；均低于每文件 256 KiB 上限。两个同 ID 的 Review 条目仍区分来源。实际通过：

- F7 打开 128 项库，查询项目条目、预览并粘贴；原 `DRAFT_` 与字面 `${HOME}` 保留，确认后夹具仍为零条输入，显式 Enter 后 A 精确收到一条。
- Alt+J 进入原 B；全局条目预览／粘贴后仍为零条输入，显式 Enter 后 B 精确收到一条。查询、确认、取消未串入另一 Run。
- 连续二十次打开、输入无匹配查询、Esc 取消，没有额外终端输入。
- 同一 manager PID 17976／FILETIME 134403289280117080 和两条原 Run 跨所有续跑保持。管理焦点确认后 q，原 manager exit=0、Terminal 正常退出；之后原两个 Run 仍可发现并正常 stop，私有镜像为零。

[desktop-final.json](2026-11-28-snippets-windows-validation-evidence/desktop-final.json)含 76 条阶段／观察回执、11 次 CLI 调用和原身份。整个流程仅启动一次产品 UI；失败续跑不是新产品轮。无剪贴板 API、没有真实 Agent，也没有覆盖全部 IME／实体键盘／图形视觉矩阵。

## 驱动失败与安全续跑

[desktop-attempts.json](2026-11-28-snippets-windows-validation-evidence/desktop-attempts.json)独立保留以下结果：

| 阶段 | 实际失败 | 后续处理 |
| --- | --- | --- |
| 首轮 | PS 5.1 不识别 `[ushort]`，发生在原草稿输入后、第一次 F7 前 | 使用 `[uint16]`，保留草稿，不重放 |
| 续跑 1 | 重复生成容量夹具会超过上限，在写入／按键前拒绝 | 后轮只复用现有配置 |
| 续跑 2／3 | 单条 PSCustomObject 的 Count 为 null，误报提交等待／回执不符 | 独立共享读取确认 A 一条、B 零条；显式数组后断言通过，不重复 Enter |
| 续跑 4 | Ctrl+PgDown 未观察到 B；只读检查仍看到 A、没有 connecting | 未记录原始键，不归因为产品／宿主或改变键位 |
| 续跑 5 | Terminal UIA FindText 抛异常，发生在点击之前 | 不伪造坐标结果；续跑 6 用既有 Alt+J 成功进入 B |

[observer-contract.json](2026-11-28-snippets-windows-validation-evidence/observer-contract.json)为实际 PS 5.1 RED／GREEN：旧单项 Count=null、旧断言 false，显式数组 Count=1 且精确正文断言 true。该结果不代表产品提交曾失败。

Context7 先查询官方 Terminal 索引，但没有给出所需固定版本结论；再核对[固定 v1.25.3131.0 defaults.json](https://github.com/microsoft/terminal/blob/v1.25.3131.0/src/cascadia/TerminalSettingsModel/defaults.json)：下一原生 Tab 默认是 Ctrl+Tab，没有 Ctrl+PgDown 默认绑定。此源码不足以证明本次实际原始键或个性化覆盖；Ctrl+PgDown 的前端观察仍计入 STAB-002，不外推为宿主拦截。Alt+J 成功只证明本次采用的路径。

## 内存口径和结果

[memory-summary.json](2026-11-28-snippets-windows-validation-evidence/memory-summary.json)保存原始字节口径；MiB 为 1,048,576 字节。C# 直接读取同一 Go UI 进程的 PID／完整 FILETIME／镜像、PrivateMemorySize64 与 WorkingSet64，并拒绝空值。24 个点样本均有效，含二十次取消后的样本；样本最大值不是生命周期峰值。

| Go UI 样本 | 私有提交 MiB | 工作集 MiB |
| --- | ---: | ---: |
| 初始基线 | 59.25 | 26.93 |
| A 流程续跑前基线 | 60.67 | 29.26 |
| 本轮所有点样本的最大值 | 62.77 | 31.85 |
| 第二十次取消后 | 61.08 | 20.71 |

最后十次取消后，私有提交在 60.57–62.38 MiB 间变化，未观察到逐次持续上涨。末值比初始基线高 1.83 MiB、比续跑前基线高 0.41 MiB。不能据此声称零分配、没有任何泄漏、长稳接受或已经确定全产品预算。

另有四次私有路径镜像快照，每次七个镜像／零读取缺口：Go UI、Terminal、一个 OpenConsole、两个 psmux、两个合成夹具。最后一次私有提交和为 243.60 MiB，其中两个夹具占约 93.09 MiB，Terminal 约 80.46 MiB；产品镜像子集合排除夹具后约 150.51 MiB。这是按私有路径识别的镜像集合，不包含路径外系统辅助进程、控制器或整个生命周期，不能称完整进程树。工作集合计含共享页重算，不是独占物理内存。

## 本机更新与回收

用户同意补验后更新 Windows 测试版。先只返回 manifest 的实际入口、程序摘要、固定配置摘要与进程身份，确认原入口为 `asd-updater-db63c5a.exe`／`069e4d2e…970e9`，原 UI PID 24184。上轮将这一摘要与旧的裸 `asd.exe`／`404d352e…3366` 比较，比较对象不同；因此上轮“日常文件变化”警报不能作为原入口程序被替换的证据。保留旧审计原件，当前核对见 [daily-preflight.json](2026-11-28-snippets-windows-validation-evidence/daily-preflight.json)。

旧 UI 唯一原生 Tab 经身份核对后收到正常 WM_CLOSE，exit=0，不停止后台 Run。元数据在远端私有目录做一致备份并校验；保留旧程序及 manifest 备份，按完整临时文件切换入口为 `asd-snippets-c2ac2c0.exe`。原快捷方式仅启动一次，新 UI PID 9144／FILETIME 134403303028856241；原 live 身份投影和配置摘要前后一致，没有发送终端正文或复制用户配置内容到公开资料。见 [delivery.json](2026-11-28-snippets-windows-validation-evidence/delivery.json)。

日常新 manager 的一次点样本为私有 64.41 MiB、工作集 38.32 MiB，仅 Go UI，不计宿主／后端或完整资源预算。没有往用户库写测试片段，没有替换 GitHub Alpha 草稿附件或公开发行。

一次试图回传日常安装脚本及配置全文的只读检查被自动审批拒绝，理由是可能包含敏感本地信息且不是隔离测试所必需。没有执行或绕过该读取；更安全的入口／哈希／进程最小检查获准，更新所需私有备份和配置保留均在远端完成，没有待用户批准的阻塞。

最终按实际 action 核对并移除九个已结束任务；保留任务零，本轮私有镜像零，日常新 UI 按意图继续运行，见 [cleanup-audit.json](2026-11-28-snippets-windows-validation-evidence/cleanup-audit.json)。本轮无强制产品／宿主／夹具结束，无剪贴板 API。旧 stop 拒绝 STAB-010、STAB-002 的 Ctrl+PgDown 观察、完整真实 Agent／资源／发布门槛继续开放。

## 仓库检查

本轮仅补文档与证据，产品 Go 与模块文件保持 c2ac2c0。该源码的全量、race、vet、模块验证日志在[首轮 F7 证据](2026-11-28-instruction-snippets-evidence/)；本轮未把它们重新记为执行。补充文档后执行生产行数、catalog `check --generated` 和 Git whitespace 检查。339 项设计库仍保留 designed／not_run，不将本轮局部通过改为全部 catalog 已执行。Git 以独立验证提交归档，连同 c2ac2c0 推送 main。
