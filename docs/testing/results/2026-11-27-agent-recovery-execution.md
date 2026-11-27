# Agent 恢复声明与显式新 Run：执行记录

日期：2026-11-27；基线提交：2cc50a4；决定：[ADR-0063](../../adr/0063-agent-recovery-declaration-v1.md)、[ADR-0064](../../adr/0064-explicit-agent-recovery-execution.md)。本报告与实现一起提交，最终源码／产品字节由证据清单固定。

## 范围

本次接通 recovery-register／recovery-profile／recover／recovery-result／recovery-status。严格解析 v1 声明与结果，绑定原 Session／源 Run，修订 CAS、相同配置幂等；恢复预分配独立新 Run／attempt／凭据，原逻辑 Session 的名字、组与排序身份不变。成功创建只记 started，匹配结果才记 restored／not_found／failed，不重放旧消息、命令或批准。

SQLite v11 增加每 Run 启动名称／程序、每 Session 最近完整显式启动描述、恢复配置与当前尝试。v1–v10 事务迁移／离线快照沿既有入口；旧程序拒读，无降级迁移。完整 launch 只保存显式参数／环境覆盖，不复制父环境；工作台委派 token 正文和 Agent 历史不进数据库；调用者显式启动参数／环境可能敏感，需保管数据库及备份。配置每份 16 KiB、全库最多 128 份，结果 4 KiB，每 Session 一份当前尝试，完整 launch 每 Session 一份且最多 64 KiB；既有绑定上限 128 不变。

没有新模块依赖、服务、watcher 或每 Session 恢复进程。没有接入真实品牌 Agent、管理 UI 恢复按钮或自动重启恢复；未替换用户日常 exe／数据库。使用方法与接入说明见 [AGENT_RECOVERY.md](../../AGENT_RECOVERY.md)、[AGENT_RECOVERY_INSTRUCTIONS.md](../../AGENT_RECOVERY_INSTRUCTIONS.md)。

## 行为验证与失败保留

- 严格解析：从示例产生声明，保留字面参数；拒绝重复字段、错误大小写、缺字段、null、尾随 JSON、BOM、无效 UTF-8、控制字符和单独 UTF-16 surrogate。版本按精确数值核对，指数文本线性处理，不按大指数分配大整数。
- 登记：真实 SQLite 关开后配置仍在；修订冲突、旧 Run、结束凭据与跨 scope 拒绝；相同内容可重取原修订。
- 恢复：外部调用前先持久化 launch_unknown；重复预约与仍未知尝试拒绝。12 个并发预约仅一个成功、只增加一个 Run。旧 Run 与新 Run 的不同启动程序能分别保存／再次发现。
- 回执：未确认启动不冒充恢复，旧凭据和冲突最终结果拒绝，正确结果幂等；迟到启动收据不回退最终状态。
- Manager／CLI：原终端存活、保留退出屏幕、状态未知、断连、过期修订和缺明确确认时均不调用启动；配置与状态查询在后端不存在时仍走离线仓库。

RED 记录分别包含缺少接口、旧命令未识别、surrogate 接受、未知尝试可被再次确认绕过，以及迟到启动收据被拒绝的问题；逐项修复后复验。最初两个 null／缺 args 测试替换未命中多行示例，纠正测试构造后单独保存真正 RED；初次构建缓存权限／模块目录／网络准备错误属于环境错误，不计产品 RED。证据目录保留这些区分。

首次相关完整检查还发现旧迁移夹具仅降低 user_version、仍留 v11 对象，以及 SQLite writer 的旧 INSERT 形状不再适用；调整夹具使它们实际还原旧结构，再复验。受限环境拒绝回环监听的后端合同测试随后在允许监听的本机环境运行；不把环境拒绝计作通过。

## 原生 Windows

Windows 10 build 19045，使用既有、显式指定的开发 terminal bundle；psmux.exe SHA-256 为 ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b。没有更换官方依赖基线或日常安装。

准备阶段保留日志解码／新目录保护失败，不据此认定任何通过。r2 在全新目录执行：Store 2 pass／3 fail，Manager 0 pass／1 fail，CLI 0 pass／1 fail，均 0 skip。Store／Manager 原因是独立测试 exe 没有相对路径示例文件；Store 首轮错误还暴露未及时关闭测试数据库的清理错误。CLI 已启动并明确停止原合成进程，但模拟环境键用了工作台保留的 ASD_RECOVERY_ 前缀，恢复被拒绝；该轮没有完成恢复。

改为自包含测试声明、ASD_RESUME_FIXTURE_ 模拟环境后，r3 取得 Store 6、Manager 1、CLI 1 个顶层 pass，无 fail／skip。r3 不冒充最终源码：之后清理重复解析／保存块，并增加结束时空作用域断言与合成身份日志；r4 另行通过。随后代码复核把结果收据固定为事务内对象，并核对 recover 返回的当前 attempt，防止后续尝试替换当前行时串回执；最终 r5 再独立复验。不把 r3／r4 结果归到 r5 字节。

最终 r5 在 Windows 10 19045 上执行 Windows amd64 构建：Store 6、Manager 1、完整 CLI 1 个顶层测试通过，失败／跳过均为 0；不累加 r3／r4 的重复通过数。Store 数量包含既有组损坏恢复回归。产品 SHA-256 为 `72dbfb990c96ca92712311d5fc9ab2cb76675e9cd8fb431d9f749cddf800cb41`；[最终清单](2026-11-27-agent-recovery-execution-evidence/artifacts.json)与远端运行前后核对一致，493 个 Go／模块源码指纹复核无变化。

该轮使用现有显式开发 terminal bundle，psmux SHA-256 为 ff9d8893…fb7b16b；不是官方稳定物 bbe825a3…d262ca 的新接受，也不改变依赖默认哈希。两个合成进程的精确 handle 均通过显式 stop 结束。r2／r3／r4／r5 私有测试映像审计均为 0；按清单哈希回收 12＋4 个测试二进制，并清理 r2 失败日志关联的合成临时数据库，强制结束进程数 0。日志／脚本／清单保留，日常安装未替换；不据此宣称完整后端生命周期资源已计量。见各轮 result.json、process-audit.json 与 cleanup.json。

真实原生 CLI 流程由单独的测试 exe 进程保存合成原生历史：

1. 使用产品 start 启动原进程，独立 receipt 记录 PID／原生 ID／唯一标记；打开精确 process handle 并核验映像／创建时间。
2. 使用原 Run 凭据登记声明；原终端仍在时 recover 拒绝。
3. 显式 stop 并等待原 handle 结束；新产品命令重新读取持久配置，以同一 Session 创建不同 Run／进程。
4. 新进程从原文件加载原生 ID／标记，经私有 context 写结果；提交新凭据回执并查询 restored，拒绝旧凭据与旧 previous-run。
5. 核对仅一个实时后端目标、两个历史 Run、元数据快照没有凭据正文；结束新进程并确认原作用域列表为空。

这是实际 Windows／psmux／产品子进程与合成 Agent 的接口验证，不是 Windows Terminal 桌面、真实 Agent、模型上下文完整性或物理电脑重启验收。原生缺历史、不同品牌迁移和交互 UI 仍未执行。原始 fixture 历史、凭据、context 和数据库不入库。

## 最终检查与边界

最终源码在本机 Go 1.27.4 darwin/amd64 依次执行 go test ./...、go test -race ./...、go vet ./...、go mod verify，全部退出 0（模块全部验证；部分无变更包使用 Go 缓存）。另有 49 项离线 Schema 正反例通过，源文件物理行数全都不超过 500，catalog check --generated 通过，本次文档入口／报告的本地链接及 git diff --check 通过。Windows 交叉构建与上述实际原生执行分开，不把本机 Go 测试当 Windows 桌面。没有新增运行时依赖或常驻进程；实际内存增量、长期资源尚未单独测量。

TC-REC-001/002/003 设计补充 v1 恢复屏障与回执；339 项目录继续保留 designed/not_run，实际局部证据不改写整项状态。全量产品、真实 Agent、完整电脑重启、管理 UI 恢复、独立资源预算与长期稳定性／发布接受仍开放。

不确定且已经没有后端的启动尝试，v1 不提供强制清除屏障；保留状态和私有上下文，不能换 ID 盲重试。scope 内外部程序／另一个同用户进程仍可能改动后端，SQL 事务不保证跨后端的原子性。新的 --confirm-ended 是用户对旧 Run 的声明，空发现或 shell 存活都不是 Agent 退出／恢复证据。restored 仍是 Agent 主动声明，需另有原生历史验证。

证据：[目录](2026-11-27-agent-recovery-execution-evidence/README.md)。Go 依赖资料通过 Context7 核对 encoding/json 的重复键／无效 Unicode／未知字段语义，并沿用当前 Go 1.27.4 标准库；未采用索引中的实验性 json/v2 或新增 JSON Schema 运行时库。
