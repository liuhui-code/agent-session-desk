# Windows 便携开发预发布（Alpha 2 候选）

完整解压新 ZIP 到一个可写的新目录，双击 **Start.cmd**。包内已附独立 Windows Terminal 与固定终端后台，不需要 WSL、Go、PowerShell 7 或单独安装 psmux；本轮实际环境和限制以发行验证报告为准。不要覆盖正在使用的旧安装目录。

首次打开是英文空工作区，管理模式按 `n` 新建普通系统 PowerShell；之后新建沿用最近成功配置，`N` 显式填写自定义 Agent 的程序、参数、目录和环境。`Alt+J/K` 切换会话，`F6` 切换管理／终端焦点，`Ctrl+S` 显隐侧栏。宿主自带“＋”打开英文标题 `Terminal` 的普通 PowerShell，它不会登记为工作台 Session。

关闭工作台保留后台 Run；再次双击 Start.cmd 接回同一安装目录的会话。电脑重启或 Agent 已退出后，恢复会话配置与恢复 Agent 对话是不同操作；对话恢复需 Agent 声明和明确恢复操作。

## 文件与命令

- `Start.cmd`：双击打开；启动失败时保留错误窗口。
- `asd.cmd`：在 PowerShell 中调用管理 CLI，例如 `.\asd.cmd version`、`.\asd.cmd list --json`。
- `asd-launch.exe`：固定启动入口，按 alpha.json 当前入口选择已校验的程序；更新后不用重新制作快捷方式。
- `alpha.json`：当前程序入口、哈希和便携 namespace；不要手改来绕过失败。
- `terminal-host`：本包独立的 Terminal 和初始英文配置；不会写系统安装的 Terminal 设置。字体使用 Consolas 10，不安装字体。
- `terminal-runtime`：固定修改版 psmux 与 ConPTY 运行库；不要单独替换其中任一文件。
- `data`：首次使用后生成，保存元数据、偏好和运行目录。不得随问题报告公开其中的私人内容。

示例（在解压目录中）：

```powershell
.\asd.cmd version
.\asd.cmd list --json
.\asd.cmd --pet
```

桌宠默认关闭，显式 `--pet` 开启；重复启动只定位原工作台，若要改变启动参数先正常退出 UI 再重开。退出 UI 不等于停止 Agent；停止只对已确认的目标操作。

## 更新与回退

本包支持显式同状态格式程序更新。先 `.\asd.cmd update status`，选择真实已发布且带更新合同的 tag 下载；历史 Alpha 1 缺少合同会被拒绝。下载后正常关闭使用该安装目录的工作台，再 apply。当前数据保持，回退仅换程序；完整操作见仓库 `docs/UPDATES.md`，运行 `asd.exe help` 查看入口。

新 ZIP 包和仅程序更新的范围不同：CLI 更新不会替换固定 launcher／Terminal／外观设置。新机器首次安装用完整 ZIP；从旧 Alpha 1 转换需在保留原目录与数据的前提下按单独迁移说明操作，不把解压新包当作旧会话自动迁移。运行中的安装目录不要移动；移动一个全新未使用目录已纳入验证，存活 backend 的 runtime 身份不能靠搬目录恢复。

## 校验与边界

下载后用 `Get-FileHash -Algorithm SHA256` 比较 SHA256SUMS；manifest.json 覆盖初始包负载。运行后 data／宿主设置等会变化，更新后的当前版本以 `.\asd.cmd version` 和 alpha.json 为准，初始发行清单不跟随用户数据重写。

此包未签名，为实验性预发布，不能据此声称所有 Windows 10／11、字体／DPI／IME、鼠标或真实 Agent 已通过。附第三方许可、官方 Terminal NOTICE 和修改版 psmux 源码附件；整体发行许可审计及稳定版接受继续开放。首次启动失败请保留原目录，提供程序版本、校验值和可脱敏的错误，不要删库反复重试或公开会话内容。
