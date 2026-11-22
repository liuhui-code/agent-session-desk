# 默认侧栏卡片还原与实际键鼠验证

日期：2026-11-23（上海时间；原始运行时间为 2026-11-22 UTC）。状态：默认 compact 卡片的局部实现与实测，不是完整上游视觉接受、全部 TC 通过或发布验收。基础提交 `01f491b6797ce7230cda67286ea8c75810a1ce89`，本次代码／文档由包含本报告的 Git 提交记录绑定。

## 结果与范围

当前 Tab 侧栏从扁平单行迁到固定上游默认的两行 compact 卡片：第一行标题，第二行 `╰` 元信息；不足 28 列或只剩一行内容空间时退到标题行。活动终端使用 `┃`，管理选择使用 `▶`；两者均使用 Tokyo Night surface 背景，标题正常 text、元信息 dim。所有值从 design tokens 取得。旧 tmux 辅助列表仍走原单行渲染。

元信息只展示已有终端可连接性与 executable 文件名，不从终端活着推定 agent ready/busy。程序完整路径仍用于启动／身份等既有流程；显示文件名只是标签变更。产品 Tabs、输入 FIFO、Session/Run 身份和终端所有权未改变。

窗口测量与鼠标命中共用每项行高；点击元信息属于同一卡片，提示行和空白不选中目标。全部会话刚好能容纳时不再为不存在的下方指示预留一行。卡片变高后，原滚动测试将总高度从 12 调到 14，以继续同时观察前一项、选择项和上下提示；既有键鼠操作要求保留。

## 来源对照

Agent Deck 固定提交 `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT，Copyright 2025 Ashesh Goplani。已按缓存的 Git tree blob 核对所读文件；没有采用截断的完整 tar 作为来源证据。

| 来源 | SHA-256 | Git blob |
| --- | --- | --- |
| `internal/ui/embedded_sidebar.go` | `f4539e560f61c760b5bc3fc7a30bc7c3ab78ff371262239464a178fa78c835d0` | `ccd6a51df0b3d8caa8abd217bf8721db7fdc8dd7` |
| `internal/session/userconfig.go`（只核对默认配置） | `e0755ee9e49e80ced9727034ff6161de413059f9ea4445e79fd2ff2eccb3b1dc` | `bd662da79c851545b9303a85b2ca8622983a192e` |

对应迁入 `workspace_sidebar.go`、tokens、styles、列表窗口和鼠标命中；逐文件登记在[复用清单](../../REUSE_MANIFEST.md)与 [JSON](../../../third_party/reuse-manifest.json)。上游默认 compact 是两行，完整 full 三行、minimal 一行、auto 自适应及其设置／持久化本轮尚未接入。上游品牌、agent 状态、分组及远程业务没有迁入；通用终端元信息属于 ADR-0014 规定的产品适配。

## TDD 与本地验证

[证据目录](2026-11-23-workspace-sidebar-cards-evidence/)保存原始输入哈希、归一化日志、脚本、事件与源指纹。

- `red-default.log` → `green-default.log`：原单行列表缺少活动卡片／独立元信息，随后恢复默认卡片。
- `red-metadata-click.log` → `green-metadata-click.log`：旧逐行索引无法命中第二张卡片的元信息，改为共享行高。
- `red-executable-label.log` → `green-executable-label.log`：完整路径挤占工具标签；改为文件名，保留原启动数据。
- `regression-markers.log` 记录真实的小窗口容纳／滚动差异。初版 `red-exact-fit.log` 实际通过，原因是断言误匹配了 Tab 文本，不能作为 RED；最终断言改为列表行首匹配。既有工作区测试同步使用源标记，历史／旧列表标记不变。

最终 Go 测试、race、vet、mod verify 均通过。另有 27/28/40/60/79/80/120 列、8/9/10/12/14/24/32 行的选择可见与列边界矩阵；它是模型渲染证据，不能代替 Windows resize。探索时 7 行高没有列表内容空间，该全局极低高度问题仍开放；这里不定义或接受最低支持尺寸。source limits 与 catalog check 通过；339 个目录用例仍 designed/not_run，关联 TC-UI-003/004/010 仅有这些局部证据。

## Windows 原生执行

Windows 10 Home 19045，固定 psmux 3.4.0、PowerShell 7.7.0 与 Windows Terminal 1.25.3131.0；精确二进制哈希和设置恢复见 `intermediate/native-audit.log`。没有使用 resource measurement SDK 替换产品后端。Go 1.27.4 darwin/amd64 构建 windows/amd64 后，测试 EXE 在 Windows 实际执行；不能将交叉构建本身当作 Windows 通过。

最终产品 SHA-256 `874f445b22286e7d24a3217cf13dd74da7b98b8ada8ccbcb33c1f3aa827cdb5a`；测试 EXE `36818d9fc375dd8d60552f3deb11520059f832ef7ec1613c856dd2cfb9377f60`。最终源码再次构建哈希相同，见 `build-reproduction.json`。`final/ui-tests.log` 有 **71 个顶层／217 个含子测试 PASS，0 FAIL、0 SKIP**；其中四个新增顶层测试。首轮中间构建也执行了同数量原生测试，但不与最终结果相加。

真实桌面脚本核对前台 Windows Terminal PID、FILETIME、完整路径后才输入：

1. 两条普通 PowerShell 各保留未提交命令；管理页点击 A 卡片的元信息，A 显示选择、B 仍为活动终端，文件均尚未创建。
2. 管理 Enter 回到 A 原草稿；双击 B 只打开终端，不提交子程序 Enter。
3. 实际 63 列上下布局中双击 A，实际 49 列单列表中选择 B；Enter 后才进入输入。
4. 显式提交后输出为 `A:7580`、`B:17512`、`B2:17512`；对应两个原 shell 的 PID／FILETIME／路径归档，切换没有重启进程或错投。
5. 正常 q 退出 UI、关闭专用窗口；CLI 按本轮登记 Session/Run 停止两条 shell，列表为空；移除专用任务并恢复设置。独立审核确认两轮四个原 PID/FILETIME 均已退出，设置 SHA-256 恢复为 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。

宽屏元信息点击实际观察 120 列、第 6 个零基文本行；63 列双击观察第 5 行，49 列选择观察第 7 行。完整 UI Automation 回执、SendInput 记录和截图见证据；它们只证明这些普通 shell 的流程，不是目标 agent 验证。

原始最终截图：[120 列卡片选择](2026-11-23-workspace-sidebar-cards-evidence/single-selected-a-public.png)、[活动终端与紧凑侧栏](2026-11-23-workspace-sidebar-cards-evidence/double-opened-b-public.png)、[49 列单列表](2026-11-23-workspace-sidebar-cards-evidence/single-list-selected-b-public.png)。图片未改动；首轮含机器路径的截图仅本地核对、未发布。文本日志替换测试用户名并转 UTF-8/LF，原哈希保留。

## 仍待推进

完整同尺寸、字体、主题和合成数据的上游运行画面对照与用户视觉接受仍未完成；源码对照与当前截图不能冒充这项验收。其他密度设置／持久化、搜索、选择复制、目标 TUI 鼠标、Tab 排序和真实 agent 仍开放。本轮没有重测资源，也没有新的 hosted CI；沿用前轮资源缺口与未定预算，不能把旧数值升级为此构建的资源结论。
