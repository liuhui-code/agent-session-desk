# Windows Alpha 1 使用说明

以下是历史 Alpha 1 包说明；后来的 main 开发版已有 [update CLI](UPDATES.md)，须先人工升级到支持新合同的受管理安装。不要拿缺少更新合同的历史包覆盖当前开发版。

这是供体验的 Windows x64 预发布，沿用 `7691a2b` 样式版本的已测程序字节。没有系统安装器或自动升级；需要已安装的 Windows Terminal，核心不要求 WSL。包内附带固定的实验 psmux／ConPTY 运行库，不需要另装 psmux。不要用官方 psmux 文件替换其中任意文件。

## 解压与启动

1. 下载 `asd-windows-x64-alpha.1.zip` 和 `SHA256SUMS`。在 PowerShell 中运行 `Get-FileHash .\asd-windows-x64-alpha.1.zip -Algorithm SHA256`，与校验文件比较。
2. 将 ZIP 完整解压到一个可写的新目录；不要覆盖当前正在运行的内测目录。保留 `asd.exe`、`asd.cmd` 与 `terminal-runtime` 的相对位置。
3. 在 Windows Terminal 的 PowerShell 页进入解压后的目录。首次可以先创建一个普通 shell，再打开管理界面：

```powershell
.\asd.cmd start --name PowerShell --cwd "$PWD\workspace" -- powershell.exe -NoLogo -NoProfile -NoExit
.\asd.cmd
```

以后在同一个目录运行 `.\asd.cmd` 即可发现仍在运行的会话。也可以先直接打开空工作区，按 `n` 填写程序、参数和工作目录。包内 `workspace` 是默认试用目录，`data` 保存本版本的运行身份、SQLite 元数据与偏好；上传问题报告前不要发送其中的私人内容。移动目录、删除 data 或换用别的版本不会自动迁移后台进程。

## 界面与操作

- 只有侧栏会话／分组导航，没有重复顶部标题或右侧边框。底部普通提示仅在鼠标移到底边时显示，移开／输入后隐藏；错误和搜索输入保留可见。
- `F6` 切换管理／终端输入，`Ctrl+PgUp/PgDn` 切换会话视图。管理 `Enter` 打开，`n` 新建，`x` 只关闭视图，`?` 帮助，`q` 退出管理界面；退出不会停止后台会话。
- 停止时在管理焦点选择会话，按 `d` 并核对确认。命令行可先 `.\asd.cmd list --json`，再 `.\asd.cmd stop 完整SESSION_ID`；不要停止不认识的会话。
- 管理 UI 使用 iTerm 黑底灰字。字体属于 Windows Terminal：在当前 profile 的“外观”中选择 **Consolas、10pt、正常字重**。包内 `iterm-colors.json` 是参考色表，可在 Terminal 设置 JSON 的 `schemes` 数组中加入该对象，并在所用 profile 设置 `"colorScheme": "ASD iTerm Classic"`。只修改自己明确选择的 profile，先保存原设置。

本包不自动写入 Terminal 设置、不安装字体、不修改默认终端。Meslo 当前用户安装在已测机器仍出现“找不到字体”，因此本次不附带个人字体，也不声称 Windows 中文回退／DPI 与 macOS iTerm 完全一致。

## 验证与限制

产品与上一轮 Windows 10 19045／Terminal 1.25.3131.0 实机样式复验的 SHA256 相同；原生 UI 293 顶层测试通过，实机悬浮出现／内部移开隐藏以及原有三会话保留已验证。发布包的解压、相对路径、CLI 启动／预览／停止另记发布报告；这些不构成任意 Windows Terminal 版本的完整桌面接受。

此程序未签名，内置 `version` 仍显示开发版本；以 `manifest.json` 的 Alpha 标识、提交和文件哈希识别，不为改版本字符串替换未测字节。实验后端包含 resize-history 与单通道鼠标补丁；原官方依赖与旧失败记录保持历史。真实 agent、其他 IME／DPI／鼠标／宿主矩阵、完整资源预算、长稳和正式发布验收仍未完成。SQLite schema v5／ui.toml v7 的数据不要直接交给旧程序读取。

第三方许可及来源附在 `licenses`、`THIRD_PARTY_NOTICES.md` 与来源清单中；原创代码许可证决定和完整发行审计仍开放。本包用于用户已授权的预发布体验，不构成稳定版接受。
