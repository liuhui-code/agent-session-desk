# M1 首个配置切片与 UI design tokens

日期：2026-11-22。状态：局部实现和验证，未发布，M0/M1 整体未完成。

## 实现

- FL-001 / FR-01：明确 `--config PATH start --template NAME`，只读 schema v1，参数数组保真，相对模板 cwd 锚定文件目录；CLI name/cwd/env 可覆盖。严格未知字段／类型／版本检查；解析错误不回显配置内容，防御性读取上限 1 MiB。合同、依赖及后续边界见 [ADR-0007](../../adr/0007-explicit-launch-configuration.md)。
- FL-003 / FR-03、FR-11：当前管理器样式全部从 [design_tokens.go](../../../internal/ui/design_tokens.go) 取值，在 styles.go 组合角色样式；保留默认配色和布局。颜色、粗体、间距、尺寸、断点及装饰标记集中，预览内容的 ANSI 不重染。见 [ADR-0008](../../adr/0008-ui-design-tokens.md)。
- 生产 Go 后端命令 helper 和资源脚本进程 helper 原样拆文件。检查生产 Go、Python、PowerShell 文件总物理行数不超过 500（Go 测试文件除外），接入 Verify。新行为采用 TDD，规则写入 AGENTS.md。

## 验证依据

被测工作树基于 `1564919`，源文件指纹及候选二进制哈希见 [evidence/source-manifest.json](2026-11-22-config-tokens-evidence/source-manifest.json)。这不是把未提交候选二进制称为最终 Git 提交构建。证据中的合成 argv/env 用于核对，不包含真实 agent 私有内容。归档文本统一 LF，build-info 行尾空白规范化，二进制 SHA 不变。

### RED → GREEN 与回归

逐行为推进，原始输出保存在 [TDD 日志](2026-11-22-config-tokens-evidence/tdd/)：

| 行为 | 发现／处理 |
| --- | --- |
| 命名模板、参数／cwd／env | config 公共入口与 CLI 路径逐步接通 |
| schema、未知字段、错误类型、重复键 | 拒绝不受支持的配置，不能静默忽略拼写错误 |
| 解析错误脱敏 | 默认解析错误会包含用户键名；包装后只保留行／列 |
| 大小写 | 解析库默认大小写宽容；补充精确 schema 键检查 |
| 缺失模板／无效启动值、读取边界 | 启动前拒绝；1 MiB 边界及超限测试 |
| CLI 覆盖 | 首轮 cwd/env 仍为模板值而失败；合并显式覆盖后通过 |
| UI 样式 | 提取沿用已有视口回归；补充 tokens 标记／菜单间距的 RED→GREEN、尺寸变体和 Model 隔离回归 |
| 500 行检查 | 500 通过、501 拒绝；脚本与 Go 共用边界 |

CLI subprocess observer 只记录后端 argv 后故意拒绝创建，不是 ConPTY 或原生 Windows 替代品。首次 observer 错报版本行、未考虑 Unix `/usr/bin/env --` 包装，测试本身也修正；对应日志保留。部分首次测试因缺失入口而编译失败，只证明尚未接通接口。没有把所有初次失败都称为产品行为缺陷。

最终本地 Go 普通测试、race、vet、模块验证、catalog 派生校验、Python 工具测试与源文件行数检查的实际结果保存在 [checks/](2026-11-22-config-tokens-evidence/checks/)。Go 工具链 1.27.4，宿主 macOS amd64。编译 Windows 候选仅是构建证据，下面的原生执行单独记账。

### 原生 Windows 配置启动

通过已授权的 `ssh windows-pc`，Windows 10 build 19045、便携 PowerShell 7、官方 psmux 3.4.0（exe SHA-256 `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`）。执行可复用 [windows-config.ps1](../../../tools/ci/windows-config.ps1)，并接入未来 Windows CI；工作流接入本身不代表该 CI 已运行。

真实候选程序创建两个后台会话，由独立 fixture 的共享读取 JSONL 核对：空参数、空格、中文、内嵌引号与尾反斜杠保持；模板相对中文 cwd 生效；CLI cwd/env/name 覆盖生效；Session/Run 身份不同。未来 schema、未知字段和模板混用原始 executable 三类请求被拒绝，未增加会话，合成敏感值未回显。

首次启动器使用了错误的 Windows 用户目录路径，未进入测试；修正后 run1 又因 PowerShell 单行对象缺少 Count 失败，已按记录的 Session ID 清理。脚本把共享读取结果显式转数组后 run2 通过。最终代码重新构建并复验，最终结果与清理记录归档在 [windows/](2026-11-22-config-tokens-evidence/windows/)，失败轮次与较早候选的通过结果单独保留。每个已观测 fixture 记录 PID、创建 FILETIME 和映像路径核对所有权；停止后核对该身份已消失，不按进程名批量结束。

资源脚本提取的进程 helper 在该 Windows 主机另执行实际子进程 smoke；没有重跑完整资源实验，也不把该 smoke 称为资源预算通过。

| 轮次 | 结果与范围 |
| --- | --- |
| run1 | 脚本把单条 JSONL 的标量当作数组读取 Count，失败；已记录的会话执行 finally stop。不算模板验收通过 |
| run2 | 修正数组包装后，模板和覆盖两次启动、三项拒绝通过；较早候选，只作为中间证据 |
| run3 | 最终 Go 源码候选重新构建并真实执行：两次启动、三项拒绝通过，两个 Session/Run 身份均不同，fixture 进程路径／PID／启动 FILETIME 独立核验，stop 后原身份消失，随后 list 为零 |

最终候选 asd.exe SHA-256：`521261996358c342644a5032e93ebd027f5d805d9863b193bfb664154ac7d937`。fixture.exe：`69a8c17f829eb18eb84bda9c45094fc705617804ec9013c570182eb22b177736`。

[run3 原始结果](2026-11-22-config-tokens-evidence/windows/m1-config-run3/result.json)、[CLI 调用](2026-11-22-config-tokens-evidence/windows/m1-config-run3/commands.json)、[清理复核](2026-11-22-config-tokens-evidence/windows/m1-config-run3/cleanup.json)与此前失败轮次均归档。资源 helper 的拆分仅另做实际 Windows 子进程／事件写入 smoke，未重跑整个短资源实验，不能继承为新资源结论。新增 CI 步骤尚未宣称在 GitHub 运行。


## 范围限制与下一步

目录关联 TC-LCH-001/002/003、TC-STORE-004/005/022、TC-UI-003/004/010 的部分行为；339 项用例仍为 designed/not_run，本报告不是完整用例验收。TBD-CONFIG 仅细化显式只读启动部分，保存中断、旧文件恢复、迁移、主题配置等继续 open。

没有配置编辑表单、SQLite 元数据／消息、持久 Run 历史、运行时主题切换、可靠 readiness 或真实目标 agent 验收。本次未重新执行 Windows 桌面视觉、IME、attach/detach；之前的桌面证据不能自动升级为本次 UI 改动验收。系统 DPI、其他 IME、完整资源预算与长稳门槛继续保留。

下一切片优先通过该模板接入用户实际使用的 agent，明确其原生 CLI／脚本入口和退出行为；再进入最小元数据存储，先确定驱动和事务合同。
