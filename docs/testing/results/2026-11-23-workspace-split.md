# 2026-11-23：上游比例调整与分隔符拖动

结论：恢复固定 Agent Deck 的 5% 步进、10%–90% 夹取、鼠标位置向下取整、手动解除紧凑布局和短暂比例提示；增加明确 `=` 恢复默认。终端焦点下 `<`／`>` 原样传给 CLI，管理焦点才操作布局。基础提交 `5f18d1b`，对应 ADR-0016、FL-003/018/019。

来源为 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 `split_config.go`，MIT／版权、固定哈希和新文件已同步复用清单。2 列抓取容差、1500ms 提示和全部数值由 design tokens 管理；提示最多一个待触发计时器，不为每次移动累积计时器。普通点击分隔符不改变比例，拖动和尺寸变化均走既有串行终端通道，不伪造按键。

**与原体验的差异：**比例只保留在当前 UI 内存，不隐式写 config.toml，不把当前只读配置边界变成配置编辑器。退出后恢复默认管理 35%／终端紧凑 18%。拖动当前只覆盖宽屏左右布局；窄屏辅助布局仍待做。

## TDD 与检查

RED-01 复现管理 `<` 不调整，GREEN-01 验证 35%→40% 及原生 resize 请求；同时检查终端 `<` 只发送一次，管理操作不写进子程序。RED-02 复现拖动无变化，GREEN-02 验证拖到 50% 且没有子程序文字／按键。GREEN-03 是边界、复位、普通点击不移动的合同回归，没有虚构单独 RED。

全量普通／race／vet／模块／500 行与 catalog 检查通过，日志见[证据目录](2026-11-23-workspace-split-evidence/)。339 项设计记录仍 designed/not_run。Windows 原生 UI 测试 54 个顶层、计入子测试 194 条 PASS，0 FAIL／0 SKIP；模型／假服务测试不等同真实后端场景。

## Windows 实际执行

Windows 10 build 19045、Windows Terminal 1.25.3131.0、Consolas 12、120×32。正式应用 SHA-256 `612fd63983e776f69b582d9eddb337d40dffcbfc2c085225a59b23eb297405f2`，测试 EXE `6df748e6a71ea1eabb136f121b842124300e3539f71c433bc295d08292513b80`。

一条私有普通 PowerShell Run 的独立输出文件核对：

| 操作 | 可见布局 | 子终端实际宽度 |
| --- | --- | --- |
| 管理 `<` 一次 | 35%→40% 会话侧栏 | 67 列 |
| 真实鼠标拖动到第 60 列 | 50% 侧栏 | 55 列 |
| 管理 `=` 后回终端焦点 | 恢复自动紧凑 18% | 94 列 |

正式构建另外再次拖到 60% 侧栏，实际分隔符位置为第 72 列。独立 shell 通过 `$Host.UI.RawUI.WindowSize.Width` 记录尺寸，不只根据外框截图推算。终端焦点输入 `<>` 原样出现在 shell 草稿，没有触发管理比例操作。

[恢复紧凑布局](2026-11-23-workspace-split-evidence/production-reset-public.png)与[手动 60%](2026-11-23-workspace-split-evidence/production-sixty-public.png)是未经编辑的正式二进制截图；不是上游同尺寸视觉对照或用户接受。

## 首轮失败与诊断

原桌面驱动通过 SetCursorPos 移动，应用只收到 MouseClick／MouseRelease，没有 MouseMotion，首轮拖动断言失败。临时诊断构建仅记录专用窗口的鼠标事件类型／按钮，不记录文字或鉴权；探针代码已从产品源码移除。

第一次 SendInput 对照脚本的 `Move` 函数与 PowerShell 内置别名冲突，在输入前报路径不存在，不能算已测试移动。改为独有函数名后，SendInput 的 MOVE／ABSOLUTE／VIRTUALDESK 对照实际产生连续 6 个 MouseMotion，分隔符跟随到 50%。随后退出诊断构建、恢复正式二进制，以同样移动方式完整重跑通过。不能将诊断构建当作正式二进制，也不能把驱动没发移动消息写成产品拖动失败。

移动标志依据 [Microsoft MOUSEINPUT](https://learn.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-mouseinput)。Context7 返回的 Bubble Tea 鼠标页仍为 beta／v1 文档，实际消息字段与 Windows reader 另核对锁定的 Bubble Tea 2.1.2／ultraviolet 源码及原生事件轨迹，未据旧文档升级依赖或修改框架。

## 清理与限制

UI 正常退出后本轮 Run 按原 Session/Run 核对再显式停止；最终私有列表 `[]`、窗口 0，本轮驱动／诊断任务删除，便携终端设置恢复原哈希 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。未触碰其他用户会话。脚本不可直接复用旧窗口身份或回执。

只证明指定终端、字体、宽窗口与普通 shell；未验证全部 DPI、窄屏拖动、真实 IME、子程序鼠标／scrollback、Tab 排序或新资源预算。比例提示显示的是意图百分比；极端尺寸还会按最小 pane 宽度夹取。完整产品与发布验收继续开放。
