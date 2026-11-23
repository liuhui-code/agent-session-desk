# 分组弹窗来源对照与 Windows 桌面验证

日期：2026-11-23；基于 `e84f904`。本轮补齐 ADR-0023 创建／移动切片的实际来源帧与三条普通 shell 桌面证据，并修复中文占位提示生成 NUL 的显示缺陷。实现、报告与证据一同提交，精确源文件指纹见[清单](2026-11-23-groups-desktop-evidence/manifest.json)。**组重命名／删除／重排、完整上游体验和发布验收仍开放**；不把本轮当作 M0/M1/M2 完成。

## 结果与适用范围

- 实际桌面累计 174 条完整回执：60 次只读 state、16 次 capture、10 次 resize、57 次特殊键、18 次 Unicode 文本、8 次粘贴、4 次列表点击及 1 次关闭。回执连续、全部 ok；15 张原始截图经检查归档，另 1 张因显示测试用户路径而排除，未编辑图片。
- 创建“项目 甲”根组及“审核”子组，实际切换根组／子组模式；M 将 A 移入子组、B 移入根组。元数据独立只读观察证实稳定组身份、父组、默认目录、成员版本及原 Run；原两条 Session/Run/Backend/程序未变。
- 在子组内新建 Group C，实际 cwd 等于子组默认目录，PID 与 A/B 不同；只有一个新增 Session/Run，初始归属与 Run 一并可读。既有 A/B 的实际 cwd 保持原值，没有向原 shell 写入 cd。
- 18 次原生完整 capture 都保留编号 1–40 各一次。A/B/C 未提交命令跨移动、窄屏和 UI 退出／重开保留；三个显式 Enter 分别在原 PID 输出 GROUP_A/B/C 标记。两次 UI 是不同进程，Run 与 pane 保持原对象。
- 8 次元数据快照与 8 次偏好快照核对组／成员／Run 和折叠。v4 元数据 backup 含 3 组、3 成员及 3 Run。根组折叠在 UI 重开后保留；切 Tab／搜索临时揭示不保存，后续明确改变密度／模式才携带当前显示状态写入。
- 覆盖 compact/full/minimal/auto、分组／平铺、嵌套元信息行点击和实际 120／63／49 列。Windows Terminal 请求 400 像素时实际最小宽度为 473 像素／49 列，按观测结果记账。
- 最终精确停止 3 个登记的 Session，核对 6 个原 PID/FILETIME/完整映像路径均结束；已知私有映像存活为 0，两个自有任务已删除，专用 settings.json 恢复原哈希。剪贴板文本及格式名恢复，未验证非文本内容等价；Caps Lock 原为 true，结束仍为 true。

代表画面：[创建根组](2026-11-23-groups-desktop-evidence/native/desktop/create-root-empty-120-public.png)、[移动弹窗](2026-11-23-groups-desktop-evidence/native/desktop/move-dialog-120-public.png)、[嵌套成员](2026-11-23-groups-desktop-evidence/native/desktop/tree-full-120-public.png)、[49 列管理](2026-11-23-groups-desktop-evidence/native/desktop/tree-auto-49-public.png)、[重开折叠](2026-11-23-groups-desktop-evidence/native/desktop/tree-folded-reopen-120-public.png)。

## 固定来源组件对照

Agent Deck 固定提交 `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT，工作树干净。overlay 仅加入调用 GroupDialog 的合成 helper，排除其他测试，以最小 TestMain 运行；没有执行原 TestMain 的个人目录／tmux 清理或启动完整应用。运行 sandbox 禁止个人目录、外网和特定桌面／凭据命令，GOPROXY=off。改用可独立归档的最小 TestMain 后再次生成帧，与首轮 15 帧字节哈希相同；不是整个上游测试或上游 Windows 应用通过。

来源和产品各 15 帧，覆盖 120×32／63×25／49×25、根组／子组、名称／目录焦点、移动列表。外宽及横向位置全匹配；120／63 列全部外框位置／高度匹配。49 列的创建／子组四帧有一行必要差异：固定来源英文 footer 换成两行，产品中文 footer 保持一行，因此居中纵坐标相差一行。移动列表仍匹配。来源／产品 frame、几何表、helper、overlay 与固定源码 SHA/Git blob 均归档。

| 模式 | 120／63 列外宽×高 | 49 列来源／产品外宽×高 |
| --- | --- | --- |
| 根组 | 46×13 | 41×14／41×13 |
| 子组 | 46×15 | 41×16／41×15 |
| 移动（3 组） | 46×12 | 41×12／41×12 |

中文本地化、稳定 ID、v2 真实光标和有界视口是已列明差异。路径建议／最近目录下拉、完整主界面状态与布局、组编辑等仍有范围缺口；本轮只有来源组件真实渲染和产品实机画面，不冒充同机运行完整上游或用户视觉接受。

## 中文占位提示 RED → GREEN

来源对照时发现产品创建、分组、搜索弹窗的原始渲染串含 NUL。固定 `Bubbles v2.2.2` 的 placeholderView 先分配 rune buffer，再以单元格宽度切片，宽字符会把补零 rune 输出。最小原组件公开 API 复现“搜索会话…”、Width=30，值仍为空、真实光标 X=2，但 raw View 含 4 个 NUL；依赖文件 SHA、源码行段和复现 helper 保留。

先增加从实际 Model 打开三个弹窗的测试，三分支均得到真实 RED；然后只在产品输入显示适配层移除空值 View 中组件生成的 NUL。未修改依赖、输入 Value、样式、焦点、Cursor 或终端文本。120／63／49／27 列的宽度／光标及既有输入行为复验通过，最终 15 个产品帧没有 NUL。该适配不把终端输出或用户输入统一“清洗”成另一份数据。

本机最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 四条退出码均为 0；新 UI 定向测试 count=1。Windows 10 build 19045 实际执行 UI 测试程序 `055845ebdb4bfb8e07077edf74d805c44e863d07baf23f93595cbe5b6b39ebff`：131 个顶层、305 项含子测试通过，0 FAIL／0 SKIP，精确测试映像存活为 0。模型结果、真实桌面及 catalog 设计状态分别记录。

## 环境、失败与清理

实际产品 SHA-256：`d8b88e66425fe284386aa5dd3a10d4fd0efadad490e31721d9412f01a1b2a6ed`。Go 1.27.4 windows/amd64；Windows 10 build 19045、普通 InteractiveToken/Session 2、非管理员；Windows Terminal 1.25.3131.0、PowerShell 7.7.0。使用 ADR-0019 的显式开发 bundle：历史候选 psmux `ca9753…`、conpty.dll `e97185…`、OpenConsole.exe `b17e48…`；完整哈希在预检／回执，不改官方默认后端或发布基线。setup、目标 shell 和外层 UI 的历史路径在输入前隔离。

两次脚手架失败保留：

1. PowerShell 5 ParseFile 对无 BOM UTF-8 中文脚本作了错误编码解析，预检停止，尚未创建桌面或会话。改用与实际 ScriptBlock 一致的 UTF-8 ReadAllText/ParseInput 后，六份脚本语法及七个固定文件通过预检。
2. 第 95 条粘贴完成后，观察函数仍解析旧的两会话局部变量，未找到第三条 shell。原任务及窗口身份经只读核对仍存活，响应 095 已完成且无待执行命令；从 096 开始继续。重新读取三条稳定身份、旧观测／偏好、原 UI 记录和 C 的实际草稿后继续，没有重放粘贴、创建、Enter 或重建窗口。产品没有因此改变。

最终 ZIP 原字节哈希 `b0376b122c72f35c0771fc43ec6573368c63d8c788943d0f2a8edf55a969f934`；内部反斜杠路径规范化但保留目录。JSON/日志由原 BOM 编码解码为 UTF-8/LF，再替换测试用户与本机路径；原始与归档哈希均记录。PNG 原字节保持，仅归档检查后的白名单；数据库、备份、auth key/port、设置备份、PSReadLine 历史和剪贴板原文不入提交。

## 剩余门槛

继续组重命名／删除／重排和统一上游差异清单。长目录字段在缩放后的滚动锚定、其他输入／焦点组合和 IME 预编辑仍需专项核对；本轮确认字段全值落盘和实际 cwd，不把截断视口当作完整字段可见。Unicode 文本事件证明本轮实际 g/M/S 等路由，不能替代物理大写键或 IME 候选的广泛验收。真实 agent、其他目标 TUI／鼠标协议、最终终端资源预算／长稳、WSL 不存在环境和发布门槛保持开放。全部 339 项 catalog 仍 designed/not_run；没有提升功能为已验收或已发布。
