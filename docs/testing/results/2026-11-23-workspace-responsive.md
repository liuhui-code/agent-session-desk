# 2026-11-23：Tab 工作区的窄窗口管理布局

结论：管理焦点恢复 50–79 列上下布局，低于 50 列或高度不足时退到单列会话列表；Enter／F6 返回活动终端后展开输入区域，Tab 栏始终保留。基础提交 `48f92c2`，对应 ADR-0016、FL-003/018/019，仍为实现中、未发布。

来源为固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 `home_view_frame.go` 与 `split_config.go`，沿用现有来源指纹与 MIT 归属。所有断点、比例、间距和内容最小高度经 design tokens。上下布局采用管理百分比及至少 5 行列表，原生终端额外保留边框与 3 行内容；单列隐藏终端不销毁 Run，不复制 shell 草稿，也不发起主动刷新。低高度单列回退是本产品的有界适配。

与固定上游存在明确差异：产品 Tab 导航持续可见，终端焦点优先给原生终端全部内容区域；上下布局以终端圆角框替代只读 PREVIEW 标题区。本轮不声称所有上游内嵌状态分支、水平分隔线或窄屏拖动均已还原。

## TDD 与验证

- RED-01：60 列管理模式看不到会话导航；首次 GREEN-01 遗漏 token 字段声明导致编译失败，修正后通过，保留两份日志。
- RED-02：40 列管理模式仍显示终端而非单列列表；GREEN-02 验证单列、隐藏时切 Tab、Enter 返回正确 Run 的终端和光标。
- RED-03：旧终端写入受阻时切 Tab，再输入新目标文字并按 F6，已接收的新文字无法送达；GREEN-03 在切换完成前暂缓焦点变更，验证旧／新文字分别进入原定 Run，没有互串。

`go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 行检查和 catalog `check --generated` 均通过。Windows 原生 UI 测试 57 个顶层、计入子测试 197 条 PASS，0 FAIL／0 SKIP。这些 UI 测试包含模型／假服务，不等同真实后端场景。339 个设计用例继续保留 designed/not_run。

## Windows 实际桌面

Windows 10 build 19045、Windows Terminal 1.25.3131.0、Consolas 12、固定 psmux 3.4.0；正式 EXE SHA-256 `992304979b87de0301061308acfb6f060ab960c7334f3ca1e407771b8bc7e478`，UI 测试 EXE `e8d62ef7700d4d0cc6083db351adf7431ccfb19cc99beb243a093c16cc2f8406`。

一个隔离的普通 PowerShell Run，先改为合成 `shell>` 提示符并清屏，在原终端输入未提交草稿 `Write-Output narrow-owned`：

| 实际观察 | 管理焦点 | 返回终端 |
| --- | --- | --- |
| 63 列，窗口 600×520 像素 | 列表与原草稿上下同时可见 | Enter 后隐藏列表，展开原终端 |
| 49 列，窗口实际 473×520 像素 | 单列列表，终端内容隐藏 | Enter 后草稿仍在，无隐式提交 |
| 恢复宽窗口 | — | 显式提交草稿后，另执行命令生成 `responsive-input.txt`，独立读取为 `same-run-input` |

驱动请求 400 像素宽，但 Windows Terminal 最小宽度夹取为实际 473 像素，不能写成实际 400。63／49 列依据 UIAutomation 的完整标题行显示单元格计数核对，不由像素宽度直接假定。[上下布局](2026-11-23-workspace-responsive-evidence/stacked-public.png)、[单列列表](2026-11-23-workspace-responsive-evidence/single-list-public.png)及两个返回终端画面均为未经修改的正式构建截图。

本轮真实桌面使用单 Run；隐藏时切到另一 Tab 和受阻输入时序由确定模型回归覆盖，不能把它们写成实机双 Run 场景。既有双 Run 证据仍见首个 Tab UI 报告。

## 证据与清理

[证据目录](2026-11-23-workspace-responsive-evidence/)保存脚本、RED/GREEN、实际日志、二进制／源码指纹、规范化事件压缩包和未修改画面。文本统一 UTF-8/LF，并替换本地账户及仓库路径；原始文本哈希与归档哈希分别保存。没有归档终端密钥、运行数据库或设置备份内容。

UI 正常退出后，按原 Session/Run 核对再显式停止本轮 Run；列表最终 `[]`、专用窗口 0、任务删除，原设置恢复哈希 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。

仍未覆盖全部尺寸／DPI、真实 IME、子程序鼠标与 scrollback、列表密度／搜索、Tab 排序、新路径资源预算或真实 agent。上游同尺寸运行画面对照和用户视觉接受仍开放，不能以本次截图关闭完整还原或发布门槛。
