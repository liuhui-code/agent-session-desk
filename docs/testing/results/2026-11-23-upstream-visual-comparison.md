# 固定上游实际画面对照与工作区几何修复

日期：2026-11-23（上海时间，回执使用 UTC）。基础提交 `62c57ad317ba7defb04bf0515bb6d5ed18cc53e9`。本报告与代码由包含它的 Git 提交绑定；状态为局部实现与对照证据，完整视觉接受、真实 agent 与发布仍开放。

## 本轮结果

实际构建并运行未修改的 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，版本 1.16.39。根据实际终端单元格修复四项差异：

- 120×32 工作区的终端框上边移至零基第 2 行；底部分隔线、帮助和空白位于第 29／30／31 行。终端内容高度保持 25 行，光标与鼠标几何使用相同坐标。
- 设置框外宽为 66 列，新建框外宽为 86 列。上游 v1 的 Width 64／84 含义与当前稳定 v2 的外框尺寸不同，先前直接照搬数值少了两列。
- 普通新建框上下 padding 恢复两行；首个 Session 为一行。打开时固定首次状态，后台发现不改变正在编辑的框。
- 窄窗按实际单元格宽度选择帮助文本，保留管理／终端焦点标签、`?`、`q`、`F6`，避免完整提示被直接裁掉。

全部尺寸、留白和装饰沿用 `design_tokens.go`，没有修改后端、输入传输、依赖或 Run 所有权。终端输入仍由目标程序处理，没有新增日常聊天框。

## 对照方法与适用范围

[上游指纹](2026-11-23-upstream-visual-comparison-evidence/upstream-source.json)保存 commit、原文件 SHA-256／Git blob、未修改工作树与参考二进制哈希。上游采用自己的 v1 依赖构建，仅作迁移参考；产品仍使用 ADR-0005 的稳定 v2。首次 sparse checkout 缺少 `examples/providers`，构建失败；补齐原仓库目录后成功，未修改源码，失败日志保留。

上游在 Darwin/amd64、tmux 3.3a 的隔离 XDG／tmux 根和 sandbox 中运行。保留 HOME，禁止读取真实 home 文件内容及出站网络；仅使用两个 Review A／B 合成 fixture。参考配置、控制和抓取脚本归档，临时绝对路径需按本机调整。首次外层 tmux 保护拒绝启动，后使用上游支持的 `AGENT_DECK_ALLOW_OUTER_TMUX=1`。个人 tmux 配置不可读导致的错误用空配置排除。

参考 sandbox 限制 `/bin/ps`，上游 process ownership 保持 unknown；没有放宽保护、把未知进程当作已验证或按名称强杀。上游 tmux 警告、嵌入客户端 signal-killed 横幅与 ownership 日志保留。结束时原 CLI 停止登记 Session，独立精确映像扫描无参考程序／fixture，两个私有 socket 均无 server。该实验不能证明上游 Windows 或完整生命周期兼容。

原始 `tmux capture-pane -p -e`、纯文本、列行数及光标可见性／位置保存在 [frames](2026-11-23-upstream-visual-comparison-evidence/frames/manifest.json)。将真实抓取的单元格和 SGR 序列化为 VT，在 Windows Terminal 重放，用同一显示配置比较真实产品前后构建；没有重新绘制参考界面。tmux 3.3a 未提供光标形状，本轮不判定形状一致。

显示配置均为 Consolas 12、bg `#1a1b26`、foreground／cursor `#c0caf5`。主轮产品前后使用同一窗口；参考因编码问题重开单独窗口，但显示配置逐项相同。UIA 实际字体提供者包含 Consolas，也包含 `System.__ComObject`，不据此声称所有字体 fallback 相同。窗口实际尺寸为 120×32、63×25、49×25。

首次参考回放没有设置 UTF-8 code page，出现乱码，八张图仅存于 `invalid-codepage-replay/`，不用于视觉结论。重放显式设置 65001，结束恢复 936；[逐行核对](2026-11-23-upstream-visual-comparison-evidence/utf8-replay/row-comparison-raw.json)只去除 UIA 窗口标题、行尾 ASCII 空格并做 NFC，保留空行／行号：

| 参考帧 | 原始尺寸 | UIA 与来源文本 |
| --- | --- | --- |
| 终端，18%／35% 侧栏 | 120×32，各一帧 | 全部 32 行一致 |
| 管理，35% 侧栏 | 120×32 | 全部 32 行一致 |
| 设置、新建 | 120×32，各一帧 | 全部 32 行一致 |
| 上下管理布局 | 63×25 | 全部 25 行一致；来源含上游错误横幅 |
| 单列管理、裁切设置 | 49×25，各一帧 | 前 24 行一致；最后行右下末格各少一个字符，保留未核验缺口 |

最后两帧不作为逐格完全还原的证据，也未将上游裁切和错误横幅当成需要复制的产品行为。逐行文本一致不能证明所有颜色、字重或交互一致，仍需视觉与操作评审。

## 可查看的真实画面

| 状态 | 固定上游的实际帧回放 | 修改前产品 | 修改后产品 |
| --- | --- | --- | --- |
| 120 列终端 | [参考](2026-11-23-upstream-visual-comparison-evidence/utf8-replay/desktop/reference-replay-upstream-focused-120-public.png) | [之前](2026-11-23-upstream-visual-comparison-evidence/product-desktop/before-focused-120-public.png) | [之后](2026-11-23-upstream-visual-comparison-evidence/product-desktop/after-focused-120-public.png) |
| 设置 | [参考](2026-11-23-upstream-visual-comparison-evidence/utf8-replay/desktop/reference-replay-upstream-settings-120-public.png) | [之前](2026-11-23-upstream-visual-comparison-evidence/product-desktop/before-settings-120-public.png) | [之后](2026-11-23-upstream-visual-comparison-evidence/product-desktop/after-settings-120-public.png) |
| 新建 | [参考](2026-11-23-upstream-visual-comparison-evidence/utf8-replay/desktop/reference-replay-upstream-new-dialog-120-public.png) | [之前](2026-11-23-upstream-visual-comparison-evidence/product-desktop/before-new-dialog-120-public.png) | [之后](2026-11-23-upstream-visual-comparison-evidence/product-desktop/after-new-dialog-120-public.png) |
| 63 列管理 | [参考，含上游错误横幅](2026-11-23-upstream-visual-comparison-evidence/utf8-replay/desktop/reference-replay-upstream-manager-63-public.png) | [之前](2026-11-23-upstream-visual-comparison-evidence/product-desktop/before-manager-63-public.png) | [之后](2026-11-23-upstream-visual-comparison-evidence/product-desktop/after-manager-63-public.png) |

PNG 为未修改窗口截图；仅文字证据做 UTF-8／LF 和测试用户／本地 home／仓库路径占位替换，原字节与归档哈希另记在 [archive manifest](2026-11-23-upstream-visual-comparison-evidence/archive-manifest.json)。有效原始帧字节保持原样。初次上传被自动审批拒绝，理由是远端可信和终端内容外传风险；核验八份合成帧哈希、无私人路径／URL／凭据并缩小上传集合后获准继续，没有上传真实用户会话。

## TDD 与实际验证

四组 RED→GREEN 分别覆盖工作区行带、弹窗外宽、首次／普通 padding、27／49／63 列帮助提示；实际行为失败、最初测试 import 编译错误和修复日志均保留。最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 通过，日志名为 `*-final2.log`；较早日志是中间构建，不能叠加统计。source limits 与 catalog 单独核验；目录仍为 339 项 designed/not_run。

Go 1.27.4 darwin/amd64 构建 windows/amd64，Windows 10 Home 19045 实际执行最终 UI 套件：**82 个顶层／235 个含子测试 PASS，0 FAIL、0 SKIP**。配置／后端未改，本轮没有重新执行它们的原生套件。最终 114 个 Go/module 指纹前后一致，再构建产品与 UI 测试得到相同哈希。

| 实际产物 | SHA-256 |
| --- | --- |
| 修改前产品（62c57ad） | `986a6d7edc805edd97c1d2b0cc9c0c80c082dc18e74a916f67dd31201bc3415f` |
| 修改后产品 | `74b3064eedee746b595f24bfd394293bd2d217bd7fca74828a6fb6d45e8cf5aa` |
| 最终 UI 测试 | `60b727e8c478307b759dcfb244514e40863578212b969ce2f577b46ade4f5a55` |
| 合成视觉 fixture | `b76b041cf484c9ccc71a576bfd9d1273c4d0ac7c10b81cb93afd435a2bccc0e7` |

后端 psmux 3.4.0、Windows Terminal 1.25.3131.0、OpenConsole、PowerShell 7.7.0 的实际哈希见[独立最终审计](2026-11-23-upstream-visual-comparison-evidence/execution/final-audit.json)。fixture 只实现自己的输出与行输入合同，不是目标 agent。

Windows 主流程 153 个命令回执均成功；过程有两次停下检查并续接同一窗口，失败另存：首次像素高度只得到 63×24，尺寸保护拒绝回放，改为实际 63×25；49→120 缩放后 Enter 遇到连接中状态被产品拒绝，未向子程序提交，检查后用选择／F6 续接，没有重试不确定输入。最终通过 A→B→A 保留 alpha／beta 未提交草稿、显式 Enter 后各自 echo；元信息点击只改选择，活动终端仍保持原目标。前后两条 Session/Run 和两个 fixture PID／FILETIME／完整路径相同。

主轮与 UTF-8 参考轮均正常结束。原两个 fixture 及参考重放进程的登记身份已结束，精确测试映像无存活、两个任务已删除、CLI 列表为空，原终端设置恢复 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。没有名称强杀。

## 仍存在的差异与下一步

| 项目 | 来源／当前差异 | 后续处理 |
| --- | --- | --- |
| 主导航 | 上游过滤条；产品为 Session Tab 与新建／关闭 | 用户明确要求的适配，继续验证排序、长标题与重新打开 |
| 顶部／侧栏内容 | 上游系统指标、品牌和 agent 状态；产品为通用终端观察 | 不推断 agent ready/busy，视觉层级与标记继续评审 |
| 侧栏比例 | 上游实际内嵌默认 18%；产品管理 35%、输入 18%；上游可拖到 35% | 当前未认定为视觉接受；比例持久化仍开放 |
| 管理切换 | 上游 Ctrl+Q 关闭嵌入客户端回预览；产品 F6 保留活动 Tab | 产品输入归属适配，保留既有冲突与焦点验证 |
| 终端内部状态行 | 上游嵌套 tmux 有自己的状态行；产品内容不增加该行 | 不伪造上游后端状态；比较时单独标记 |
| 新建表单 | 外宽与 padding 对齐；品牌／worktree 等字段替换为 executable／args／cwd | 字段内部比例和可用性仍需评审，未宣称完全相同 |
| 49 列设置 | 上游最小宽度导致裁切；产品适应可视宽度 | 保留明确适配及本轮最后单元格核验缺口 |

先完成上述视觉差异的用户评审与必要修正，再推进完整选择／复制／搜索、Tab 排序、普通 shell 与真实 agent 的组合流程。真实 agent 当前缺失，不能以 fixture 替代。当前短资源基线、完整预算／长稳、其他 shell／IME／后端与发布门槛继续按既有报告保留；本轮没有新资源测量或 hosted CI。FL-003/012/018/019 继续“实现中／未发布”。
