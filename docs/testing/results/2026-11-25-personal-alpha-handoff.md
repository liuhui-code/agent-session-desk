# Windows 本机 Alpha 8234c4e 体验交付

日期：2026-11-25（Asia/Shanghai）。用户明确要求把内测版放到 Windows 并查看效果。本次为本机体验交付，不是 GitHub Release、系统安装器或完整发布验收；沿 [ADR-0019](../../adr/0019-explicit-windows-terminal-bundle.md)／[ADR-0040](../../adr/0040-verified-conpty-single-mouse-channel.md) 的显式开发 profile。

## 产物与运行

- Windows 桌面目录 `Agent Session Desk Alpha 8234c4e`，快捷方式 `Agent Session Desk Alpha`；同名 ZIP 和 `.zip.sha256`。
- ZIP 13,189,731 bytes；SHA256 `8af0deb1083692433576b4f68594742ffc0e8dc896968209ce7b64200b03956f`。
- 使用上一轮已测的同一产品字节：`404d352e72def2ed0d36ecb2ca08491271de8652ca589e6a8a13da2dff993366`；psmux `ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b`，两项固定 SDK 文件不变。内置 version 仍为 `0.0.0-m0-dev`，没有为改版本标识而重建未测字节；包清单明确标识 Alpha 和源提交。
- 独立 `data` 保存 runtime／SQLite／偏好／示例 shell 历史，独立 namespace `asd-alpha-8234c4e`，示例 cwd 为包内 `workspace`。不覆盖旧产品、旧 bundle 或既有用户数据。
- `local-host.json` 仅在 Windows 本机生成，引用并校验已验证的 Terminal 1.25.3131.0 和 PowerShell 7；这是依赖本机已有宿主的体验包，不能称为任意电脑可用的独立离线发行包。运行时数据及本机路径清单不进入 Git。

## 实际核验与边界

支持文件先打包，再在 Windows 复制固定二进制，生成清单、ZIP，最后从 ZIP 解压到桌面新目录。解压后 18 个负载文件逐项 SHA256 一致；产品 version exit 0；两份 PowerShell 源文件通过 Windows parser。直接内联执行安装脚本首次因 Windows 命令行过长而未执行，改为上传相同脚本、核对哈希后从文件执行，未绕过执行策略。

在交互用户会话执行最终解压目录的 `start-alpha.ps1`，首次发现空作用域后创建两个真实 PowerShell 示例，之后打开管理 UI。只读 UIAutomation 确认标题和 A／B 可见；CLI 与两个独立 shell PID／creation FILETIME 核验成功。截图已实际查看，具有两个 Tab、会话侧栏和 A 的真实 PowerShell 提示符。本轮没有注入按键或调用剪贴板；不将初始就绪截图当作完整键盘／重开流程或新版 UI 套件执行。鼠标／输入／历史／重连的精确上一轮证据见[候选复验](2026-11-25-single-mouse-channel.md)，不冒充本轮重复执行。

PowerShell 序列化回执把发现数组包装为 `value`／`Count`；观察器外层 Count=1 不是一个会话。独立读取包装后的 value 及实际截图确认两会话，保留此脚手架表示差异。原始截图、实际会话和进程身份只留私有交付目录，不提交用户运行数据。

一次性打开任务在结束后按脚本哈希／动作路径核对删除；manager、Terminal、两个后端和两个示例 shell 为用户主动保留，不能记作泄漏或宣称已全部清理。此后不自动注入输入、停止会话或回收用户窗口。

新增辅助脚本位于 `tools/preview/`，未改产品 Go 源码或依赖；500 行检查与普通文本 whitespace 检查通过。本轮没有重新执行 Go／原生产品测试，沿用上次源与字节对应证据。附带既有第三方源码声明／许可，不宣称完成全部 Go／Rust／SDK 发行物许可审计。实体硬件、真实 agent、资源、完整兼容与公开发布继续开放。
