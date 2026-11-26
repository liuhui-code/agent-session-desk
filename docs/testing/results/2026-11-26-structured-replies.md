# 结构化请求／回复的实现与实际验证

日期：2026-11-26。用户同意继续完成请求查询、可信来源关联回复及结果查询，并要求 Agent 接口文档。范围依据 [ADR-0047](../../adr/0047-run-bound-inbox-and-structured-replies.md)，操作合同见 [AGENT_API.md](../../AGENT_API.md)。真实模型 agent 当前未安装，本轮模拟进程不作为真实 agent 兼容性证据。

## 实现

新增 agent-bind、inbox、reply、replies。控制端显式发布随机 Run 凭据文件，数据库只保存摘要；接收方提交凭据，重验原 Session/Run/backend_id/command 与已观察的退出状态。没有任填 --from，也不从终端正文推断回复来源。此为同机同用户内的委派，不声明操作系统级 agent 隔离或防御可改本地数据库的控制者。

SQLite v7 增加凭据和回复表，v1–v6 事务迁移、check 与非覆盖备份／恢复保持。inbox 最多读取 128 条尚无结果的原 Run 请求，不领取任务、不注入输入。每请求一份结果；稳定回复 ID／父请求／来源／正文相同定位原记录及原时间，冲突不覆盖。replies 可离线查询，原 Run 退出后结果保留。回复不改变原消息 queued／unknown／input_written，也不产生 completed 或验收状态。

绑定同库最多 128 份、不自动轮换或删除；回复受每请求一份、原消息最多 128 条与复合外键约束。正文沿用 64 KiB、序列化记录 512 KiB。凭据完整暂存／同步后以同目录硬链接非覆盖发布，再提交数据库；不确定时保留原文件重试，备份不包含私有凭据文件。没有新增依赖、Host、前端或后台自动消费者。

## 源码与构建

实现提交为 `67c505e`。该提交的 344 个 Go／模块文件与正式原生构建指纹逐一相同；本段后续证据补充不更换 Go 源码或二进制。

正式构建从 `5da1413` 的 go.mod/go.sum/internal/cmd clean archive 加本轮修改／新增 Go 文件组成；[344 个 Go／模块指纹](2026-11-26-structured-replies-evidence/source-fingerprints.json)核对最终工作树，刻意排除原有 internal/backend/terminal.go typed-nil 候选及其未提交测试。不能据本轮结果认定此前切换退出根因已解决。产品 SHA-256：`1077b872000fae572e5b3d2912491f83e917cc79fa854ec7f93aa2631e8090cd`；[四个构建物指纹](2026-11-26-structured-replies-evidence/artifact-fingerprints.json)另存。

Go 1.27.4；原生环境 Windows 10 `10.0.19045`／amd64。[运行器](2026-11-26-structured-replies-evidence/native-runner.ps1.txt)先核验当前安装清单与本轮四个构建物散列，再使用现有显式 terminal bundle。测试只用临时 metadata/runtime、独立 namespace 和本轮原 Run；没有迁移日常数据库、更新日常 exe／manifest、桌面／剪贴板 API 或定时任务。

## RED 与本机检查

先增加 Store round-trip 行为测试，缺 AgentCredential、BindAgent、Inbox、RecordReply 时编译 RED；实现后 GREEN。随后 Manager 原 Run 重验测试先因入口缺失 RED，再实现绑定／查询／回复。早期中间编辑出现重复 Store.Replies 声明的编译失败，删掉重复声明后复验；该失误不是产品运行缺陷。原始 RED／首轮 GREEN／中间编译日志在[证据目录](2026-11-26-structured-replies-evidence/)。

- 完整 internal/store、internal/manager、internal/cli 三个相关包本机通过；没有运行全仓库或全 UI 套件。
- 消息／回复选择集的三个相关包 race 通过；其后新增单线程的 128 绑定容量边界单独通过，不将之前 race 冒充新容量测试已执行。
- 检查错误凭据、合法但错误来源、未知 reply_to、同 ID 改正文、第二个结果、跨 scope、绑定重试／隐式轮换拒绝、已停止和已观测 pane 退出；保存原记录不变。
- 16 个独立 Store 并发：同 reply_id／同正文全部定位同一结果；不同 reply_id 竞争同一请求只一个成功，不覆盖。
- v6 只读检查不改源、v6→v7 保留请求、快照恢复保留回复时间与幂等；备份包含摘要而不含随机凭据明文；冲突表导致整个迁移回滚且源文件散列不变。既有 v1–v5 迁移预期同步到最新 v7，历史报告不改写。
- 新 CLI 无效参数／正文不触碰 runtime/state；私有凭据文件发布可重试、不替换无关文件、正常无残留 stage；公开 JSON 不含 token。replies 不需要后端在线，错误 namespace 不返回原结果。
- 相关包 go vet、go mod verify、500 物理行检查通过。339 catalog 仍为 designed/not_run；只补 ADR-0047 引用／范围说明并重新生成，工具未执行产品测试。

## Windows 原生结果

[摘要](2026-11-26-structured-replies-evidence/native-summary.json)：Store 17、CLI 3、Manager 2，共 22 个顶层 PASS，失败／跳过均 0。Store 包含本轮凭据／回复／竞争与历史迁移检查；不把它们都称为 psmux 端到端测试。只有 TestWindowsNativeReplyTwoAgentProcesses 是新的原生工具闭环。

该闭环实际完成：

1. 真实 psmux 托管 A、B 两个独立测试 CLI 进程，保留不同 Session／Run；控制端通过新产品 asd.exe 对 B 原 Run 绑定凭据。
2. A 进程启动独立 asd.exe send，保存实际 queued 回执；正文是公开合成的中文、组合字符和 emoji。
3. B 进程启动独立 asd.exe inbox，读取实际 JSON 请求与 message_id，从读到的正文计算 SHA-256 并形成带中文的结果；不是预填回复 UI。
4. B 启动独立 asd.exe reply 提交，再启动另一 asd.exe 以完全相同参数重试；reply_id 和 recorded_at 保持一致。
5. A 独立启动 asd.exe replies，按原 message_id 得到同一 reply_id、B 原 Session/Run 与由真实请求算出的正文。控制端再次查库，结果恰好一份，原消息仍 queued。
6. 两份受控报告和后端原身份核对完成后，按原 Session／Run Stop 清理返回成功。私有配置、凭据和带本机路径的 worker 报告仅存临时目录，未纳入公开证据。

这次没有在同一测试进程内用两次 cli.Run 代替独立工具进程。与此同时，queued 未变仅证明该元数据合同；本轮不是完整终端原始输入审计，也不外推独立 PID／完整生命周期或资源接受。父测试通过与归属 Stop 成功不等于已采集每个子程序的原生退出码。

## 文档、环境问题与剩余范围

接口文档补齐四命令的参数、JSON、凭据来源边界、单结果幂等、离线结果查询、与终端投递不同的回复重试规则，以及 A/B 调用步骤。SQLite schema v7 与公共 API 版本分别说明，README、Feature Ledger、路线图和 ADR 同步。FL-010／014 为局部实现中，不宣布 M3 或发布完成。

构建准备期间本机磁盘不足，首次运行器生成 here-document 未执行；随后只删除本任务此前的三个可重建 archive 副本，保留证据／当前源码／测试文件，并重新生成运行器和指纹后开始正式原生执行。该准备失败不是产品结果，没有拿未生成的运行器算通过。

仍开放：真实 agent、任务执行 exactly-once、任务领取／自动消费、进度／多结果、凭据轮换／撤销、退出与 SQL 的跨资源原子性、完成／验收证据、共享 Host、桌面交互、UI 长稳／资源与完整发布接受。拥有同用户文件或数据库控制权者不在凭据隔离声明内。unknown 请求不能因出现在 inbox 而自动再执行；这些边界均写入接口参考。
