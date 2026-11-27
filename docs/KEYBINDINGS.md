# 快捷键配置

main 开发版，重启生效。文件为 `metadata.sqlite3` 同目录的 `keybindings.toml`（一般 `--state-dir`）；不是 launch.toml，也不是 ui.toml。缺文件／缺项沿用默认，空串禁用整个动作及其别名。程序只读，不自动覆盖。

```toml
schema_version = 1
[global]
wake = "ctrl+alt+f9"
[workspace]
next = "alt+x"
previous = "alt+z"
toggle_sidebar = "ctrl+s"
[management]
open_project = "f8"
```

global 是 Windows 注册键；只允许 Ctrl/Alt 加字母、数字、F1–F11或Space。冲突不抢占。workspace 在终端输入时也有效；management 只在管理焦点。两作用域不能抢同一键。无 watcher／新进程；只保留小映射。

名字用 ctrl/alt/shift，顺序可调；字母大小写区分，Shift+w 归一成 W。F12不支持。未知动作／版本、重复修饰、冲突或超64KiB使启动失败；修正或移走文件恢复默认，不会改原文件。IME／宿主截获的键不能靠配置强制送达。

## 动作

| 作用域 | ID |
|---|---|
| global | wake |
| workspace | previous, next, back, attention, next_attention, toggle_sidebar, rename, toggle_focus |
| management | open_project, check_release, attention, copy_visible, delete, move_up, move_down, copy_mode, tab_left, tab_right, quit, up, down, group, first, last, move_group, open, rename, refresh, attach, history, new, new_dialog, help, settings, search, group_mode |

权威默认表为 `internal/keybindings/keys.go`。弹窗字段、Esc/Enter、复制模式内部键、组数字跳转／折叠、拖动不属 v1；不做任意终端重映射。已改旧键不再触发该动作，终端焦点交原终端；切换中仍被输入屏障消费。

帮助有覆盖时显示实际表；缺省仍保留原帮助。删除／改键不会修改会话、后端、配置模板或数据库。
