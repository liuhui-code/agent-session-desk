# ADR-0071：原生新标签默认打开普通 Terminal

2026-11-28，已接受。关联 FL-016／018，细化 ADR-0053／0055 的便携 Windows Terminal 配置边界。

## 背景与决定

用户保留 Windows Terminal 原生标签栏和窗口按钮，要求“＋”直接打开普通英文终端。当前日常使用的独立便携宿主仍把旧验收脚本 profile 设为 defaultProfile，因此原生新标签会打开测试启动页。

仅调整该便携宿主的 settings/settings.json：增加名为 `Terminal` 的普通 PowerShell profile，用户主目录启动、`-NoLogo -NoProfile`，设置为 defaultProfile；固定英文 tabTitle，界面语言覆盖为 en-US。旧验收 profile 隐藏，保留可回查的原始配置备份。产品启动仍显式选择 local-host.json 中的原工作台 profile，其 GUID 和内容保持。

不启用 focus mode：该模式同时隐藏用户需要的标题栏／窗口按钮。普通终端由 Windows Terminal 管理，原生“＋”不创建产品 Session；产品内部管理 `n`／`N` 的已有行为不变。不修改系统 Terminal 配置、Windows 语言或输入法，不创建常驻辅助进程。

## 代价与验证

本轮是日常便携宿主配置维护，不是产品 exe 新功能或公开 Alpha 包升级。设置修改前备份，按旧哈希复核并原子替换；不为语言切换关闭正在使用的工作台。已打开宿主若未即时应用界面语言，在用户下次正常重开宿主时生效。

原生“＋”的实际打开、标题、子进程和清理结果单独记入[验证报告](../testing/results/2026-11-28-native-terminal-default.md)。配置正确不等于实体输入／完整终端或多原生 Tab 定位接受；尤其桌宠精确选择某个原生 Tab 仍是既有未决项。

## 依据

2026-11-28 已通过 Context7 核查 Microsoft Terminal 的 portable、defaultProfile 和 profile 字段，并核对[官方启动设置](https://learn.microsoft.com/en-us/windows/terminal/customize-settings/startup)与[官方外观设置](https://learn.microsoft.com/en-us/windows/terminal/customize-settings/appearance)的 language 覆盖。实际宿主为便携版 1.25.3131.0；不从最新文档推定未经执行的版本兼容性。
