# 长列表卡片的固定来源与本地验证

日期：2026-11-24；基线提交 `b43a07d`。落实用户要求的固定上游布局／交互／颜色还原与 U06 长字段核对，沿 [ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)／[0016](../../adr/0016-tab-workspaces-and-input-ownership.md)。这是局部源码合同和模型证据，**不是完整上游体验、Windows 桌面或发布验收**。

## 改变与结果

固定 Agent Deck 的卡片首行通过 fitCellWidth 使用空尾裁切；元信息仅在 wrapSidebarMetadata 的最后一行自行省略。产品此前首行预先加省略号，整行又加省略号，部分 keycap 元信息还会在最后适配时出现额外尾符。现在沿源规则裁切首行及最终适配，元信息独立换行／末行省略，选中背景仍填满整行。

内部标题函数增加尾符参数；Tab／搜索继续使用既有省略策略，卡片使用空尾。可见重名／裁切后重名仍带稳定 Session ID，控制对象继续绑定完整 Session／Run，不照搬源模型丢失产品身份提示。

程序标签先去终端控制序列，再从完整 executable 路径中取文件名，最后沿既有 4096 字节标签防御边界显示。旧代码先截断整条路径，超长目录会令文件名变成目录片段；修正后的回归又确认 OSC URI 内的分隔符不会被误当实际路径。原 Command、启动路径、Session／Run 和输入正文均不修改。模型数据的长路径显示测试不证明该路径可以在真实系统启动。

生产变化只有 `workspace_sidebar.go` 与 `session_title.go`；无存储迁移、依赖、tokens 或终端生命周期合同变化，SQLite v5／ui.toml v4 保持。所有生产文件均不超过 500 物理行。

证据：[白名单清单](2026-11-24-long-cards-evidence/manifest.json)、[最终源指纹](2026-11-24-long-cards-evidence/local/source-fingerprints.json)、[检查](2026-11-24-long-cards-evidence/local/checks.json)、[逐项对照](2026-11-24-long-cards-evidence/local/comparison.json)、[源组件原始帧](2026-11-24-long-cards-evidence/local/upstream-frames.json)。归档校验器仅检查一致性，不执行产品／Windows 测试。

## RED → GREEN 与模型证据

- 首轮两条 RED：28 列首行多出源没有的省略号；两个超过 4096 字节的 Windows／Unix 合成路径显示目录片段。空尾裁切及完整路径文件名投影后 GREEN，原路径保持不变。
- 后续 RED：OSC URI 中的分隔符被错误解释为文件名，显示 `false.exe`。先清理完整控制序列再取 basename 后 GREEN，显示实际合成文件名，无 OSC 重放，原 Command 保留。
- 180 张合成卡片：CJK、组合重音、keycap、ZWJ；27／28／29／42／63 列，1／3／8 层及 1／2／3 行组合。360 行均保持给定单元格宽度，13,608 个选中 surface 单元格无背景缺口。它们是 x/ansi／样式模型测量，不是实体字体／终端宽度验证。
- minimal／compact／full／auto 四密度各验证七组尺寸：120×24、120×14、79×14、63×24、49×14、27×14、返回 120×24。30 个同前缀长名称及长程序名下，选中原目标保持可见，卡片末行单击选择其原身份且不打开终端；显式 Enter 才打开原 Run。原合成终端草稿只写一次、不提交／重放。既有同名 Tab／列表／搜索身份提示回归亦通过。
- 最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 和源码行数检查通过。后续控制序列标签修正前的检查／构建与最终结果分别保留，不累加通过数量。
- 测试目录检查通过，339 项维持 designed/not_run，216 ready／123 needs_decision。TC-UI-003/004/010 仅追加设计预期；目录工具不执行产品测试。

## 固定源合同对照及差异

来源 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT，embedded_sidebar.go 与 cellwidth.go 的精确 SHA 见机器复用清单；来源工作树未修改。实际调用原 fitCellWidth、wrapSidebarMetadata、renderSidebarMetadataLine，并保留原 renderEmbeddedSessionItem 的原始 ANSI 帧。

八组受控场景覆盖长 ASCII、CJK、组合字符、选中／普通 keycap、ZWJ 和深层组。为核对共享文字合同，使用产品前缀，剥离 ANSI，不加入产品尚无证据的源 agent 状态／tool 图标；源裁切后不足一格时为产品完整 surface 补空格。原始组件帧另存，**不能把归一化结果称为源／产品原始全帧、业务字段或 RGB 一致**。

| 文字合同 | 变更前符合数 | 变更后符合数 |
| --- | --- | --- |
| 首行空尾裁切／补齐 | 1／8 | 8／8 |
| 元信息独立换行／末行省略 | 6／8 | 8／8 |

CJK 与普通 keycap 两组源裁切结果比目标宽度少一格，产品显式填齐既有 surface；ZWJ 那组原本已符合，不伪造成修正前失败。源品牌、agent 状态、pane title／目录业务与通用 executable 标签仍是明确适配。

产品前帧通过临时 overlay 复用 `b43a07d` 的两份原生产文件，其余代码不变；后帧使用最终生产代码，共用一个只导出帧的测试 helper。源测试 overlay 排除 456 份无关 UI 测试，加入一个 helper，生产源码未修改。HOME 保持原值，四个 XDG 根指向任务目录并预置合成配置，`AGENT_DECK_TEST_HOME_ISOLATED=1`／`GOPROXY=off`，额外 OS sandbox 禁止真实用户目录读写、出站网络及 security／pbpaste／osascript。只执行这条固定组件合同，不声明来源完整套件通过。

归档文本将产品 overlay 的当前工作区绝对路径归一化为 `<WORKSPACE>`，原文件 SHA 同时保留；归一化 overlay 是审计输入，不可直接当原执行文件。原渲染日志的单元格补齐空格保留，仅证据日志使用 `-text -whitespace`。

## 未执行与后续

本轮没有远程上传、Windows 原生套件、ConPTY／桌面动作，也没有本轮远程运行句柄。前轮 U19 上传仍受自动审批的目的地／载荷授权拒绝约束，已发出的明确授权请求尚待用户回复；未换传输方式绕过。Windows amd64 UI 测试与产品 EXE 仅在本地交叉构建，精确 SHA 已保存，**原生执行为 not_run**。

未调用剪贴板或真实用户设置 API，不回填历史轮次未知恢复结论。U06 的完整桌面长字段／实体鼠标、源字体／色彩接受，U18 鼠标／可见页复制、U20 设置／主题，U19 停止和 U15 比例的原生流程，以及真实 agent、资源／发布门槛均继续开放。目录／状态字段不能用合成文字替代真实产品能力；整体目标按[差异清单](../../design/UI_PARITY_AUDIT.md)持续推进。
