# Agent 接手与功能适配

更新：2026-11-27。本文是 main 开发版索引，不等于 Windows 日常安装包或公开 Alpha 已包含全部功能。先读仓库 AGENTS.md，再核对 `git log`、`asd version` 和实际执行报告。

首次接入先读[配置说明](CONFIGURATION.md)：固定 exe／bundle／runtime／namespace／state，再配置启动模板。示例位于 docs/examples，不会自动装入日常环境。

## 快速定位

| 功能 | 当前入口 | 源码／合同 | 状态与边界 |
| --- | --- | --- | --- |
| 配置总览 | launch／last-launch／ui／keybindings | [配置说明与示例](CONFIGURATION.md)；`internal/config` | 各文件独立版本；含最近配置优先、便携缺省与工作区读取器限制 |
| 待办中心 | Alt+A；管理 t | `internal/ui/attention_center*.go`；[ADR-0056](adr/0056-bounded-attention-center-and-snooze.md) | All/Due/Snoozed；s/S 暂停10/30分钟，u恢复；Enter只定位原表单 |
| 桌宠提醒 | 启动 `--pet` | [说明](DESKTOP_REMINDERS.md) | 复用待办；不批准；工作台退出即停 |
| 终端活动 | 会话名后转圈／蓝点 | `internal/ui/activity.go`；[ADR-0057](adr/0057-terminal-activity.md) | 比较有界页，不是 Agent忙闲；离屏不监控、断连不报静默 |
| 版本／发布查询 | version；check-update；管理 U | `internal/release`；[ADR-0058](adr/0058-build-and-release-check.md) | 按需联网，只显示公开发布；不下载安装 |
| 全局唤起 | Windows Ctrl+Alt+Space | `internal/hotkey`；`cli/workbench.go` | 恢复原窗／核验待办；冲突提示、不抢占；实机仍待验 |
| 最近会话 | Alt+P | `internal/ui/recent.go` | 最多16个成功连接视图；跳过已关视图；不改变行序 |
| 项目工作区 | 管理 W；workspace NAME | [用法](WORKSPACES.md)；`config/workspace.go` | W为default；最多8模板，串行开新Run；重复执行会再开 |
| 外置 Markdown | Ctrl+单击路径 | [说明](EXTERNAL_MARKDOWN.md) | 外部编辑器；不修改默认关联 |

普通／race／cross-build 与实际 Windows执行必须分开。导航三个功能的当前局限见[报告](testing/results/2026-11-27-navigation.md)；活动标识见[报告](testing/results/2026-11-27-activity.md)。不要把交叉编译记为真机通过。

## 适配步骤

1. 按[配置说明](CONFIGURATION.md)中的 executable/args/cwd/env 配启动；项目组复用 launch.toml 模板，不添加聊天输入框。
2. 调用工具前固定 exe/bundle/runtime/namespace/state；保存完整 Session/Run。命令参数、字段、幂等与恢复规则见 [Agent API](AGENT_API.md)。
3. 授权／问题只上报合作式 attention-event；保留原 ID、版本及结束关联。查看／暂停／静默都不能变成批准、ready或completed。
4. 先跑[模拟合同](testing/MESSAGING_CONTRACT.md)，再记录目标Agent名称、版本、hook和原生表单证据。当前模拟不是品牌兼容声明。
5. 新行为先测试；UI样式只加 design tokens；生产文件不超过500行。同步ADR、Ledger和报告。

项目启动不是事务：已开的Run不因后续失败自动停止。失败／回执不明后检查实际列表，不重放整组。UI会阻止本次不确定启动的再次触发；CLI重复命令仍有真实副作用。

## 快捷键配置

已实现启动读取，见[配置合同](KEYBINDINGS.md)。放在 metadata.sqlite3 同目录；重启生效。

```toml
schema_version = 1
[global]
wake = "ctrl+alt+space"
[workspace]
previous = "alt+p"
next = "alt+j"
toggle_sidebar = "ctrl+s"
[management]
open_project = "W"
```

按 global/workspace/management 分域，缺省继承、空值禁用、冲突报错；对话框／复制模式内部键和数字分组等固定手势不属v1。

适配时先搜索 `workspace_keys.go`、各模态事件及 `hotkey_windows.go`；不要只替换一处字符串。全局事件与终端输入不同，字母大小写可能代表不同动作。配置不能绕过原输入排空、身份核验及不确定结果不重试。

连接行程序标签：CurrentProgram仅内部瞬态；Command仍为启动快照。勿从标签/消失推断就绪或完成，见[ADR-0061](adr/0061-observed-program-label.md)。
