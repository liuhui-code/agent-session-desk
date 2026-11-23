# 底栏角色与帮助粘贴局部还原

日期：2026-11-23。关联 [ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)、[ADR-0016](../../adr/0016-tab-workspaces-and-input-ownership.md)、[U13/U14 差异清单](../../design/UI_PARITY_AUDIT.md)，以及 FL-002/003/005/012/016/018/019。TC-UI-001 增补设计内容，但全部 339 项仍保留 designed/not_run（216 ready、123 needs_decision），实际执行独立记录。

本轮恢复默认底栏的键／说明层级、模式和实际目标，纠正帮助粘贴未关闭的偏差，并修复 49 列目标标题只剩 ID 尾缀的问题。普通 shell 两轮桌面分别通过；**整体界面、真实 agent、完整 Windows 兼容及发布没有通过验收**。

## 来源与实现边界

固定 Agent Deck 提交 `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT，源码 checkout 干净。机器来源表记录 19 个文件的精确 SHA／blob；实际调用未修改的 `UISettings{}.GetFooter()`，结果为 `full`，没有采用过期注释中的 curated 默认说法。参考 helper 在禁止访问个人 home／出站网络的沙箱中，以 overlay 隔离原测试／TestMain；执行的是来源组件，不是完整上游程序或来源 Windows 兼容测试。

来源 `home_view_frame.go`、`insert_mode.go`、`help.go` 的选择性迁入分别对应 `workspace_footer.go`、`help_overlay.go` 和既有 tokens/styles。来源与产品各运行 49／63／80／100／120／160 列 × 空白／组／会话／终端，共 24 个参考状态；这些帧用于核对角色／宽度适配，并不声明 24 帧逐字符相同。产品 Tab／F6、中文和通用终端为明确适配，只展示实际支持的操作，不仿造 agent 工作状态。

- 管理底栏大于 100 列使用有内边距蓝色键块，70–100 列紧凑键块，50–69 列只列键，低于 50 列有界保留 F6／帮助／退出。说明用独立前景角色，管理上下文为紫色；导航／退出为 dim，退出位于右侧末尾并保留来源留白。
- 组上下文显示实际“展开／折叠”；不可连接 Session 不提示 Enter 打开。错误重查、分组／移动／删除提示按当前对象和能力选择。
- 终端底栏用黄色模式块和实际活动 Tab 名称。49 列空间不足时省略整个可选目标，不显示箭头或 ID 末尾碎片；F6 和 Ctrl+PgUp/PgDn 仍可见。
- 来源 v1 带 Paste 标记的按键使短／全帮助关闭。v2 对应 `PasteMsg` 必须消费整段后关闭并重连原目标，不能将载荷解释为快捷键或 CLI；鼠标仍消费并保持帮助。

全部颜色、粗细、间距、装饰和响应式阈值经 `design_tokens.go`，styles 仅组合角色。没有依赖升级、数据库迁移或终端所有权变化。

## 颜色精确差异

来源和产品作者色值均为 `#7aa2f7`。实际 v1 序列化为 RGB **121,162,247**，稳定 v2 为 **122,162,247**。已对照固定依赖 termenv v0.16.0 的 `RGBColor.Sequence`／go-colorful v1.4.1 转换代码并保存其 SHA；不是凭截图推测，也没有修改产品 token 去补偿旧序列化。

初次严格角色比较因这一差异失败，原 comparator、日志和差异 JSON 均保留。最终比较仅允许这个明确的背景三元组变换；其他文字、背景／前景和粗体要求一致。full 12 格、compact 10 格、badge 17 格中，共 12 格具有这一已知差异（7 格 full 蓝键、5 格 compact 蓝键），黄色 badge 精确相同。不使用一般 RGB 容差，不声称逐 RGB 一致。

Context7 查询了官方 Lip Gloss 通用 hex／profile／v1→v2 指引；目录没有 2.1.1 专属条目。当前 main 文档只提供一般说明，精确数值的依据是上述固定版本代码和实际输出，未将 current main 当作固定版本证据。

## TDD 与本机检查

逐条 RED→GREEN 保存于 [local 证据](2026-11-23-footer-input-evidence/local/)：帮助原先忽略粘贴、底栏缺蓝色键／黄色模式与目标、退出锚点／留白缺失，分别失败后修正。其后桌面窄窗发现 ID 尾缀，`target-red.log` 保留失败，最终 `target-green.log` 为完整 UI 复验通过。

测试覆盖短／全 × n、q、j、Tab、多行十个 payload，先收集所有旧连接再重连并检查无泄漏写入；另覆盖上下文、不可连接对象、长 Unicode／宽度边界、活动 A/B 与退出锚点。一次测试使用不存在字段而编译失败，改为实际 sidebarCollapsed 字段后继续，保留 `ui-test-field-first-build.log`，不算产品行为失败。

最终源码完成 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、物理行数检查及用例目录 `check --generated`。204 个 Go／module 指纹包含两个历史证据 reference main，不能全部计作产品模块。目录工具只验证资料，不执行产品测试。

## Windows 构建与原生测试

Windows 10 build 19045，真实交互桌面 session 2／非管理员；Windows Terminal 1.25.3131.0、PowerShell 7.7.0、显式 psmux 3.4.0 bundle。Go 1.27.4、稳定 Charm v2 基线不变。实际 bundle／Shell／Terminal 的精确哈希见 preflight、构建与 outcome 文件；交叉构建不计原生执行。

| 构建 | 产品 SHA-256 | UI 测试 SHA-256 | 原生 UI 实际结果 |
| --- | --- | --- | --- |
| 完整输入流程版 | `3c95594246720d21628ef34198acc642f4f7f24acf0cd7b4cbb3f21e25117658` | `bf73af57a4307cd2c74ab090f6c6dfc799567e45b59105ed0ebadfc6af76db9a` | 170 顶层／354 含子测试 PASS，FAIL/SKIP 0，精确测试 image 剩余 0 |
| 最终标签修复版 | `b3962b6bd5eb4f712ae54ace10a02efec9e06883f8d271c1106d2e4ee43b461b` | `60419a76494b2a8e7249069625375981a6a1114e7d7c06e273c871f4eba35e1d` | 171 顶层／355 含子测试 PASS，FAIL/SKIP 0，精确测试 image 剩余 0 |

两套均基于 `d3e1025` 加未提交源码，指纹和 build-info 定义所测状态，不冒称提交后的重建。两套 Go/module 集合相同，最终只改变 `workspace_footer.go` 和 `workspace_footer_test.go`，其余 202 文件哈希一致；不将重复套件计数相加。两套桌面和最终本机检查分开保留。

## 完整输入流程：193 回执

[原始记录](2026-11-23-footer-input-evidence/native-full-flow/)包含连续 001–193 成功回执：112 state、17 capture PNG、19 text、24 keys、7 resize、13 paste、1 finish。两条普通 PowerShell，各自已有 40 条编号输出和独立身份；原草稿为 `Write-Output ('FOOT_A:'+$PID) # nrMdx?+` 与 `Write-Output ('FOOT_B:'+$PID)`。PSReadLine 历史使用测试专属路径，未归档用户历史。

实际步骤覆盖 A→B→A、管理／终端焦点、组折叠／展开、120／160／80／100／101／63／49 列、短／全帮助十种粘贴、关闭原 Tab 后重开和整个管理 UI 退出重开。12 个独立原生 capture 均保留编号 1–40；前十次草稿未提交，最后仅明确 Enter 各执行一次。SID／Run／backend、原 shell PID／FILETIME／path 和草稿均保持。原 A PID 9328、B PID 1640，最后输出分别为 FOOT_A:9328／FOOT_B:1640。

13 份偏好观察、两次明确组折叠操作：帮助粘贴不修改偏好，折叠／展开会显式保存，最终文件字节回到 seed。不能把整个流程称为“偏好零写入”。17 张原始 PNG 均查看；其中早期终端 49 列的 `6]` 尾缀作为缺陷证据保留，不能用最终结果覆盖。

### 观察器失败与安全续接

首段在第 154 回执后等待短暂 `Tab 已关闭` 提示失败；第一个 x 已关闭 A，实际只剩 B。源码 `activateWorkspace` 在切换时立即清除 status，不能把未观察到提示当作动作未执行。原 `flow-resume.json` 中关于 connection completion 的初步猜测原样保留，因果解释以此处实际调用为准。

只读检查确认同一 driver PID 10004／FILETIME、原窗口／UI 存活、任务 Running 和最后 154。第一次续接预检又因 PS5 `@((... | ConvertFrom-Json))` 包裹 Object[]，将原本 4 个 capture／12 份偏好误读为外层一个元素而停止；未发布新 UI 动作。`resume-shape.json` 保留实际 observed=4／wrapped=1。

修正直接赋值后恢复原上下文，155 仅新取 state，156 只读截图确认剩余 B，157 才发出本来计划的第二个 x，**重放已接收的变更动作 0 次**。随后完成至 193。原始失败、检查、失败续接、实际续接均保留；`footer-input-flow-fixed.ps1` 是修正后的完整配方，未从头执行，不能将其当成全新通过轮次。

## 最终标签复验：52 回执

窄窗实测后增加 RED，避免以两列预算构造 `sessionTitle` 导致部分 ID；先按有意义预算度量标题，再仅在完整目标适合时显示。导航空间不足时省略可选目标。

[最终原始记录](2026-11-23-footer-input-evidence/native/)包含 52 连续成功回执、9 原生 capture、5 PNG。120 列 A/B 显示原目标，49 列 A/B 与 UI 重开后的 A 不含箭头或 ID 尾缀，保留 F6／Ctrl+PgUp/PgDn。两条新合成 shell 的原身份／Run／40 行输出／草稿保持，最后明确 Enter 各一次，输出 LABEL_A:10416／LABEL_B:19780；偏好字节不变。五张 PNG 均查看。

| 画面 | 实际 PNG |
| --- | --- |
| 120 列 A | [最终 A](2026-11-23-footer-input-evidence/native/desktop/footer-label-A-120-public.png) |
| 49 列 A/B | [A](2026-11-23-footer-input-evidence/native/desktop/footer-label-A-49-public.png)、[B](2026-11-23-footer-input-evidence/native/desktop/footer-label-B-49-public.png) |
| 120 列 B | [最终 B](2026-11-23-footer-input-evidence/native/desktop/footer-label-B-120-public.png) |
| UI 重开后 49 列 A | [原草稿](2026-11-23-footer-input-evidence/native/desktop/footer-label-reopened-A-49-public.png) |

这是最终二文件修复的有界复验，**未在最终二进制重跑前述全部 193 动作**。两套证据及原生计数分别归属各自哈希。

## 清理、归档与仍开放门槛

两轮均按原 SID／Run／backend 和 PID／FILETIME／image path 清理，各结束 2 个原 shell、2 个先后 UI、1 个专属 Terminal；任务移除、已知测试 image 剩余 0、测试 Session 列表为空。专属 Terminal 设置 SHA 恢复为 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。剪贴板文字／格式名与 Caps 恢复；没有验证非文字剪贴板内容逐字节一致，也不作全机进程清理声明。

完整流程导出 ZIP SHA `f1505c57ee5a3bc4eb56a143336f0942d406a3572c2a690ac33baa44cfdc0927`（1,557,206 bytes）；最终 ZIP SHA `e1334847bfdc8bb997317d31c2a2917387d451ef78bf36c79c9ce9c1fc77714f`（421,759 bytes）。不提交二进制／数据库／密钥／端口／用户历史或剪贴板原文。文字归档仅做 UTF-8/LF 和已知测试用户名／本地路径替换，PNG 不改字节；[manifest](2026-11-23-footer-input-evidence/manifest.json)同时保存原始／归档 SHA，[一致性校验](2026-11-23-footer-input-evidence/verification.json)不是新的产品执行。

下一轮先核对 U01–U06 同尺寸主界面、长标题／目录和低高度／缩放锚定，再推进会话手动排序、显式停止 UI、鼠标／可见页复制、比例保存与适用设置。来源其他 footer 配置模式和完整错误／复制提示仍未还原。UnicodeText 驱动不等于物理 Shift 或 IME 测试；真实 agent 未安装、其他 IME／鼠标／alternate-screen／无 WSL、资源／长稳、用户视觉接受和发布继续开放。Session/Run/Tab、关闭视图与停止进程、终端输出与会话上下文边界不变。
