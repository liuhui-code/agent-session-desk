# ADR-0053：Windows 提醒／文档工作流与显式系统编辑器

- 日期：2026-11-26
- 状态：已接受，用户授权实机集成与日常交付
- 关联：FR-03/14、FL-012/016/018/020；扩展 ADR-0051/0052

## 决定

用户授权用真实 Windows Terminal 跑通“保留 A 草稿 → B 原生表单等待 → 点猫定位 → 打开文档 → 从编辑器点猫回原窗 → 原终端明确选择 → 清除提醒 → UI 关闭重开”，并在通过后接入日常入口。用户明确选择系统记事本先验证打开，不修改默认文件关联。

增加仅 Windows 工作台使用的显式 `--document-app default|system-notepad`。省略／default 仍沿用 Windows 文件关联；system-notepad 复用既有路径规范化、原 Run 核验和取消检查，再用 GetSystemDirectory 定位系统 Notepad，通过独立 argv 传入唯一文件路径。没有任意可执行命令模板、cmd／PowerShell 拼接、注册表写入或新的常驻进程。编辑器启动后只释放父进程 handle，不绑定 UI 请求 context 的寿命；退出工作台不能结束用户正在阅读的编辑器。

开关只用于工作台，非法值、其他平台或带子命令在后端／状态初始化前拒绝。应用提示统一称 external editor，不把“打开请求成功”表达为已读或渲染完成。SQLite 仍 v9、ui.toml 仍 v7，无新增依赖。模拟 CLI 只增加可选受控 Markdown 引用，保留原事件／表单和显式 y/n 输入，不新增真实 Agent 接入。

## Windows 验证与目录选择纠正

[实际报告](../testing/results/2026-11-26-integrated-desktop.md)保留五轮尝试及各自源码／产物范围。注册表快检最初返回“.md 无关联”不能代表有效关联：官方 AssocQueryString API 实际发现有效处理器。没有据此覆盖用户默认应用，也没有把 Shell 的成功返回当作记事本已打开；改用用户明确选择的启动开关。

最终 R5 使用受控窗口、真实 SendInput、UIA／有超时的自有 Edit 控件文本观察及当前固定运行库。已确认原授权表单、查看仍 pending、实际记事本文档标记、点猫恢复同一 Windows Terminal 前台、显式 y/Enter 才 resolved、原草稿和 Run 保持、UI 关闭重开，以及外部编辑器跨 UI 退出存活。输入与鼠标坐标采用测试线程的 Per-Monitor-V2 上下文，结束恢复该线程原值；不修改用户 DPI 设置。仅该宿主组合和受控 CLI 场景，不等于实体鼠标／真实 Agent 或完整发布接受。

## 日常更新边界

复用同一已测产品字节，采用版本化 exe／启动器、哈希校验和可读回的桌面快捷方式，入口显式启用 pet 与 system-notepad。实际交付前先对 Windows 本机私有数据库副本验证 v5→v9，原数据库不先改写；关闭原 UI 后再做最终本机备份、升级和稳定身份／配置／输出文件核对。没有停止原用户 Run，备份、原始会话内容和凭据不得导出或提交。

只有日常窗口与管理器归属唯一、无其他宿主 Tab、满足输入空闲守卫时才关闭旧 UI；不用按键猜测旧模式或强杀进程。新入口验证默认 Windows Terminal 交接、一个工作台／桌宠、原 live 投影和原配置／输出保持；失败时仅在无新 UI／状态没有额外改变的条件下恢复，升级后的文件也保留，未知情形不盲目回滚。具体结果见报告，准备脚本不算已交付。

共享 Host、UI 退出后的持续监控、跨应用多 Tab 精确激活、其他阅读器／动态 cwd、长稳／资源与真实 Agent 继续后置。已有未提交 terminal.go 候选不属于本轮；旧鼠标五项失败仍按 STAB-006/007 保留。

依据：[GetSystemDirectoryW](https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsystemdirectoryw)、[AssocQueryStringW](https://learn.microsoft.com/en-us/windows/win32/api/shlwapi/nf-shlwapi-assocquerystringw)、[ASSOCF](https://learn.microsoft.com/en-us/windows/win32/shell/assocf_str)、[GetWindowTextW](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getwindowtextw)、[SendMessageTimeoutW](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendmessagetimeoutw)。Context7 查询官方文档；有效关联查询不以是否有 exe 命令判断所有现代应用能力。
