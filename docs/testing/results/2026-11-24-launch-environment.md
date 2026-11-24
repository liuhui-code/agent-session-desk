# 新建环境覆盖与原始目标核验

日期：2026-11-24；基线 `82c6545`。关联 [ADR-0039](../../adr/0039-launch-dialog-environment-overrides.md)、U12、FR-01/02/03/11 及 FL-001/002/003/012/013/016/018/019。**本轮执行受控 UI 服务边界、合成表单与模型帧；没有执行真实子进程环境、原生 Windows、终端／agent 或完整视觉接受。**

## 缺口与实现

TC-LCH-001 原有设计要求通过创建表单设置 exe、argv、cwd 和环境；CLI／模板已有 Env，通用新建 UI 只有四字段。现补“环境覆盖”第五字段，放在参数和创建按钮之间。输入可空 JSON 字符串数组，例如 `["ASD_MODE=dev","ASD_EMPTY="]`，每项遵循已有 KEY=value 验证；空白／[] 不提供覆盖，继续继承环境，不清空父环境。字面空值、等号、空格、Unicode、`$HOME` 及重复项顺序传入本次 Launch.Env；最终大小写／重复键与解析限制由既有 ResolveLaunch／后台决定。

按 Enter／Tab 可从环境到创建按钮，Ctrl+S 仍提交；取消／成功后销毁表单，重开不记忆环境。单行粘贴与 64 KiB 字段字节上限沿既有入口；拒绝输入保留已接受草稿。新字段沿原 token／输入／视口布局，不写 TOML、UI 偏好、最近目录或 Session 名称，不更改旧 Run 环境，没有统一聊天输入框。

语法／类型／KEY=value 错误在 Start 前拒绝，错误不回显值。带环境覆盖时不直接展示后端启动错误文本：本项目固定 `executable not found or not executable:`、`working directory`、`environment key ` 前缀仅映射静态字段提示，其余给通用摘要；不从任意 stderr 推断具体原因。代价是部分诊断细节不可见，避免 stderr 展开覆盖值。已产生 Session/Run 的未知结果继续显示原身份、禁止重试。编辑字段本身仍可见，不声称加密输入、内存擦除或整个 CLI 后端的通用脱敏；无覆盖的原错误路径不变。

配置文件继续只读；本轮没有模板选择／保存 UI、持久配置编辑或自动记忆。ui.toml v7、SQLite v5、配置 schema v1、依赖和 design tokens 不变。通用 Env JSON 行是产品适配，源品牌表单没有同一行，不声明五字段整帧与来源相同。

## RED → GREEN 与实际验证

- 入口 RED：管理 n 新建后填第五段，旧 UI 未传任何 Env；GREEN 后只调用一次 Start，名称／程序／目录／argv／Env 保持原边界，值中 Unicode、空字符串、等号、字面变量和重复项均保留。管理器合成环境仍为 outer-sentinel，返回默认视图没有展开覆盖值。该 Start 是受控服务，不是真实子进程。
- 错误 RED：后端错误将合成覆盖值原文带入不确定结果摘要；GREEN 后隐藏原始详情，原 Session/Run 和“不重试”仍在。字段定位 RED：通用摘要不能指出工作目录；GREEN 后三个已知应用前缀给静态程序／目录／环境提示，不包含后端值。
- 10 个非法 JSON／类型／KEY／NUL／语法样本不启动、不回显值，取消重开环境为空；空白、空格和 [] 三种输入可经 Enter 到按钮并提交一次空覆盖。超限／多行粘贴拒绝且原草稿不变，取消不启动。
- 新 Env 长字段加入既有七尺寸、ASCII／Unicode／组合／keycap／中间光标矩阵，resize 后原值／逻辑位置与插入目标保持。创建布局测试按实际五字段加按钮遍历，不沿用旧固定五次循环。
- Env 焦点两主题 × 四颜色能力 × 八尺寸共 **64 模型帧**；核对精确宽高、可用视口光标、值／位置／焦点保留。目录选择及其 128 帧、原新建／长字段回归也通过；没有重新声称旧来源整帧对照适用于这个新行。
- 最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 物理行检查通过。设计用例重新生成／检查，339 项继续 designed/not_run，216 ready／123 needs_decision。
- Windows amd64 产品和 UI／session 两个测试程序仅交叉构建，哈希单列；没有原生执行。

[清单](2026-11-24-launch-environment-evidence/manifest.json)、[入口 RED](2026-11-24-launch-environment-evidence/local/env-red.log)、[隐私 RED](2026-11-24-launch-environment-evidence/local/env-error-red.log)、[字段定位 RED](2026-11-24-launch-environment-evidence/local/env-location-red.log)、[最终定向复验](2026-11-24-launch-environment-evidence/local/env-final.log)、[64 帧](2026-11-24-launch-environment-evidence/local/env-frames.json)、[五项检查](2026-11-24-launch-environment-evidence/local/checks.json)与[源指纹](2026-11-24-launch-environment-evidence/local/source-fingerprints.json)分别保留。证据校验器只检查材料一致性，不执行产品／Windows 测试。

## 对原始目标的核验

用户要求以固定来源还原布局、比例、交互与颜色；Session 以 Tab 管理；agent 和普通 shell／CLI 使用终端原生输入；同步 ADR 与 Feature Ledger。本轮对照当前源码、[逐项差异](../../design/UI_PARITY_AUDIT.md)、[ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)及 [M1 门槛](../M1_ACCEPTANCE.md)，结论如下：

| 要求 | 当前权威证据 | 未证明／未完成 |
| --- | --- | --- |
| 原有布局比例／颜色／交互 | 已有固定来源组件、颜色、Go 模型及早期有限 Windows 画面；当前 tokens、工作区／表单代码和差异单可追溯 | 当前版本完整同尺寸／字体／主题的真实上游对照及用户视觉接受；目录、鼠标、主题等新增切片的完整原生组合 |
| Tab 为 Session 主要入口 | psmux workspace 的创建／选择／切换／关闭重开、身份和队列代码；早期普通 shell 桌面记录与当前回归 | 当前候选整条 Windows A/B/A、重命名／排序、关闭重开／退出重连；Tab 顺序／关闭状态持久化等缺口仍在差异单 |
| agent 与普通终端原生输入 | workspace_input 与 terminal 共同传递原目标输入；通用 exe/args/cwd/env，UI 不增加日常聊天框；本轮补 Env 服务边界 | 当前版本真实子进程／Windows 输入、其他 IME、鼠标、alternate screen 和真实目标 agent；用户目前未安装 agent，fixture 不等价 |
| ADR 与 Feature Ledger | 已接受 ADR-0014 方向，后续实施至 ADR-0039；本轮已同步 FL、用例／报告与提交材料 | 文档不是产品或平台通过证据，相关 FL 仍实现中／未发布 |

这些要求仍按原范围推进；模板 UI、展示项、Tab 状态与剩余输入矩阵继续保留，单个字段接入不关闭整体目标。后端 Env、TOML、数据模型和 v2 组件实现未改；新增解析／摘要为本项目通用适配，没有新的源码复制、模块 API 或依赖。既有固定 Agent Deck 派生文件的版权／MIT 和来源 SHA 保持。

此前 Windows 二进制上传被自动审批以具体目的地／载荷授权不足拒绝，现有明确请求尚未答复；本轮未尝试上传、SSH、CUA、系统剪贴板或真实系统设置，也未换通道绕过。Windows 变量大小写／PATH、实际环境传输、真实 agent、完整上游视觉、资源／长稳和发布仍未验收，整体目标保持未完成。
