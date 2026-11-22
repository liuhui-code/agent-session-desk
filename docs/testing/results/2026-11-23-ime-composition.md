# 2026-11-23：真实中文 IME 的候选、确认与取消

结论：在限定 Windows 桌面组合上，活动 Tab 终端及新建表单已取得真实拼音预编辑、候选选择与 Esc 取消证据，直接 PowerShell 对照同样成功。**这是基础 IME 局部通过，不是全部输入法、切换键、尺寸或 Windows 兼容验收。** 基础提交 `0c0cd53`，对应 ADR-0016、FL-002/003/012/016/018。

本轮没有修改产品代码，复用上一轮正式 EXE SHA-256 `24f434d6491e9d551030238dd94ef52fd769cd975c43b9a47cfbe7cf607235d2`。当前 Go 正文／测试与 go.mod/go.sum 均逐文件核对为上一轮原生构建的相同指纹。没有重新执行 Go 普通／race 全套或 UI 测试，也不把上一轮 62 个 Windows 顶层测试计为本轮新执行。

## 环境与输入方法

Windows 10 build 19045、Windows Terminal 1.25.3131.0、Consolas 12、120×32、psmux 3.4.0。当前输入语言 `zh-Hans-CN`，TIP `0804:{81D4E9C9-1D3B-41BC-9E6C-4B40BF79E35E}{FA550B04-5AD7-411F-A5AC-CA038EC515D7}`，活动窗口 HKL `0x08040804`；Microsoft IME 的 ChsIME.exe 文件版本 `10.0.19041.4355`。版本与系统进程路径见 `ime-metadata.json`，没有安装或更换语言／输入法。

新建专用窗口，重新记录 PID／FILETIME／路径和前台身份；每次扫描码输入前核对，未复用先前已关闭的 PID。初始 Caps Lock 为开启，暂时关闭后使用真实扫描码事件输入 `nihao`／`zhongwen`；不是把中文通过 Unicode 注入或剪贴板写进去。候选选择用 Space，取消用 Esc。其他命令前缀／引号采用合成 ASCII，命令提交另用显式 Enter。

扫描码标志依据 [Microsoft KEYBDINPUT](https://learn.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-keybdinput)，按键状态检查依据 [SendInput](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendinput)。Context7 已读取官方 API 说明；脚本不改变用户语言配置。Caps Lock 的前后状态单独保存并恢复。

## 实际观察

| 入口／操作 | 证据 |
| --- | --- |
| 活动 Tab 输入拼音、保持候选 | 截图显示 `ni'hao` 与“你好”等候选；独立 CLI preview 的同 Session/Run 屏幕仅有命令前缀，未含预编辑拼音或中文 |
| Space 确认“你好” | 后端屏幕出现一次“你好”，目标文件仍不存在；候选确认没有提交 shell 命令 |
| 再组合 `zhongwen` 并 Esc | 组合前、组合中及取消后的后端已确认屏幕相同，“你好”保留；未把新预编辑正文送进后端 |
| 闭合引号、显式 Enter | 独立文件 `ime-confirm.txt` 内容为“你好”，正文无重复；原 Run 不变 |
| 退出 UI、直接 PowerShell 对照 | 同样显示拼音候选，确认后显式提交，`ime-control.txt` 内容同样为“你好” |
| 返回 UI、新建表单 | 名称预编辑时仍是占位符，确认后变为一次“你好”；再组合并 Esc 后表单仍开着且名称保留，再按 Esc 才关闭表单；CLI 列表仍只有原一条 Run |

[终端真实候选](2026-11-23-ime-composition-evidence/scan-initial-mode.png)、[取消第二段候选](2026-11-23-ime-composition-evidence/candidate-cancelled-public.png)、[直接 PowerShell 对照](2026-11-23-ime-composition-evidence/direct-control-preedit-public.png)、[表单预编辑](2026-11-23-ime-composition-evidence/form-preedit-public.png)及[表单取消后](2026-11-23-ime-composition-evidence/form-cancelled-public.png)均为未经修改的正式二进制画面。候选位置靠近当时真实终端／表单光标；没有把 UIAutomation 不包含预编辑覆盖层误当作没有候选。

## 首轮失败与观察修正

旧报告仅看到 `NIHAO` 且扫描码探针因陈旧窗口身份被拒绝，不能用于判断产品 IME 是否可用。本轮使用新身份、关闭 Caps Lock 并实际发送扫描码后，真实候选可见。因为同时改变了 Caps 状态和注入方法，**不能单独断言旧失败只由 Caps Lock 或只由虚拟键方式导致**；本轮证明的是已记录配置下的真实 IME 行为，没有据此修改产品代码。

最初 `Get-WinUserLanguageList | Select-Object` 没有逐项展开列表，记录为空字段；改用 foreach 后读到上述语言／TIP。两次原件均保留，不把初次空字段解读为没有安装 IME。

表单探针第一次聚合断言失败：它搜索整张 UIAutomation 文本中的“你好”，把后方终端已确认的旧文字也计入。实际名称行仍是占位符，候选在覆盖层，产品没有提前提交。随后把断言限定到名称字段，继续同一窗口中的候选确认／取消，通过并关闭表单。`probe-03.log` 只保存了 stdout，原 SSH stderr 字节未保存；错误已由当时工具回显，原失败谓词、032 原回执和可复验字段检查均保留，不能声称有原始 stderr 日志。

冻结回执的 `verify-snapshots.py` 另核对 42 个操作回执、表单名称字段、后端前后屏幕、两个最终文件与 Caps 恢复；这是证据一致性检查，不是新增产品测试或替代真实键鼠执行。

## 测试设计、清理与限制

TC-TRM-006／TC-FOC-008 的设计入口补充为表单、主 Tab 终端和辅助全屏 attach 分别记录；FR/FL/TC 编号稳定。目录重新生成并 `check --generated` 通过，339 项仍 designed/not_run。本次只执行所列终端／表单和直接 shell 子集；辅助全屏 attach、语言切换键及候选中切换焦点未执行，不将整个用例或切换语言功能标通过。

[证据目录](2026-11-23-ime-composition-evidence/)保存合成输入输出、前后屏幕快照、42 次动作及回执压缩包、原始画面、脚本、源指纹和校验值。文本统一 UTF-8/LF 并替换本地账户／仓库路径，原件哈希另记；没有归档终端密钥、运行数据库或设置备份正文。

最终 Caps Lock 恢复为原开启状态；UI 正常退出，按原 Session/Run 核对后显式停止本轮 Run，私有列表 `[]`、窗口 0，任务删除，原终端设置恢复哈希 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。

尚未验证其他 IME／语言模式、真实键盘硬件、多 DPI／窄窗口候选、候选中 resize／Tab 切换、所有确认键、辅助 attach 与真实 agent。本轮缩小了基础 IME 的未知范围，完整上游视觉与交互还原、scrollback、新路径资源和发布验收继续开放。
