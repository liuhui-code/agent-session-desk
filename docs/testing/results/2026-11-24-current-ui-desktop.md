# 当前构建的 Windows Terminal 桌面复验

日期：2026-11-24。基础 `88a6e3d` 加目录 continuation 修复，产品 SHA256 `d014526f36cdb02a13fcc57330f10cf4e33a30b4960b886780f09d72d2ce5432`；[原生报告与完整源指纹](2026-11-24-current-ui-native.md)固定同一构建。依据用户“授权此清单及同名修复重测”，始终只更新指定 `run-current-ui-tests.ps1`，各轮使用全新私有目录／namespace／InteractiveToken Limited 任务，不覆盖旧 asd.exe，不操作用户原会话或系统剪贴板。

## 实际结果

Windows 10 `10.0.19045`、固定便携 Windows Terminal `1.25.3131.0`、实际 TextPattern 字体 Cascadia Mono，终端／pwsh／实验 bundle 的完整映像哈希在脚本中逐项核对。两个实际产品后台均独立核对 PID／FILETIME／映像及已加载的固定 conpty.dll，不把指定环境变量当作加载证明。前台输入前核对唯一私有窗口、原进程和 foreground；失配不注入文本。

**run3 通过：60 事件、31 次键／文本动作、15 次截图。** [结果](2026-11-24-current-ui-desktop-evidence/desktop-run3/outcome.json)、[事件](2026-11-24-current-ui-desktop-evidence/desktop-run3/events.json)、[原始截图与状态](2026-11-24-current-ui-desktop-evidence/manifest.json)归档。

- 两条普通 PowerShell 从新建的合成命令进入产品 Tab，用各自原生提示符输入；A→B→A 保留各自未提交草稿。退出 UI 再启动后，Session／Run／backend ID 和两个 shell 的原 PID／FILETIME／映像不变，两个草稿仍存在。
- 重连后分别显式 Enter，独立后端帧确认 `REVIEW_A:<原 PID>`、`REVIEW_B:<原 PID>` 各且仅出现一次；A 连续第二条命令的 `REVIEW_SECOND_中文_α` 单独检查为完整输出行，输入回显不算输出证据。
- 实际 S 面板切 Dark→Light→System→Light→Dark，每次先等待落盘再操作。System 画面显示“跟随 Windows 应用主题 · 当前 Light”，这是实际初始主题读取，不是系统主题切换通知测试。主题操作后原 A 草稿仍在。
- 帮助／新建五字段／目录候选画面已捕获；新建使用明确 Ctrl+C 取消，确认弹窗消失后继续，未创建第三条会话。以 1240×760、820×650、570×550 像素观察宽双栏／上下管理布局及宽度恢复，实际字号／字体由同一终端设置保留。
- 两次 manager 正常 exit 0；2 shell＋2 backend＋2 manager＋2 Terminal 的 **8 个登记原身份全部结束，0 强制结束**。三个本轮任务经动作路径／各轮脚本哈希及非运行状态核对后删除；[最后审计](2026-11-24-current-ui-desktop-evidence/local/final-audit.json)确认当前产品残留 0、旧 asd.exe 哈希仍 `5216bdfa…ac918`、五作用域执行策略仍 Undefined。

私有终端 settings.json 前后 SHA256 均 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`；未写系统设置，CapsLock 前后一致，剪贴板 API 调用 0。焦点曾转入本轮窗口；窗口关闭后的先前 foreground 没有独立恢复观测，不补写“全部焦点状态已恢复”。

## 可审阅画面与边界

人工检查下列原始 PNG：Tab／原生提示符、当前终端边框、表单全部字段、主题选择及窄窗口内容在画面内；截图不是完整来源逐像素一致的结论。

| 画面 | 原始 PNG |
| --- | --- |
| 当前 Tab 与原生输入／光标 | [暗色初始](2026-11-24-current-ui-desktop-evidence/desktop/01-dark-live.png)、[第二条中文命令](2026-11-24-current-ui-desktop-evidence/desktop/12-second-command.png) |
| 暗／亮／System 设置 | [暗](2026-11-24-current-ui-desktop-evidence/desktop/03-settings-dark.png)、[亮](2026-11-24-current-ui-desktop-evidence/desktop/04-settings-light.png)、[System](2026-11-24-current-ui-desktop-evidence/desktop/05-settings-system.png) |
| 五字段与响应式管理 | [表单](2026-11-24-current-ui-desktop-evidence/desktop/08-five-fields.png)、[820 像素](2026-11-24-current-ui-desktop-evidence/desktop/layout-820.png)、[570 像素](2026-11-24-current-ui-desktop-evidence/desktop/layout-570.png) |
| 原 Run 重连与草稿 | [A](2026-11-24-current-ui-desktop-evidence/desktop/10-reopened-original-A.png)、[B](2026-11-24-current-ui-desktop-evidence/desktop/11-reopened-original-B.png) |

15 张截图中的目录候选包含测试用户的绝对目录，只留本地，不提交；其文本状态按路径占位归档。其余 14 张原始 PNG 不改像素。JSON／脚本／结果逐文件记录原字节与归档 SHA256；不包含认证 key、数据库、二进制或 shell history。

本轮没有重新运行完整来源应用或等数据／等列行的来源对照，不能替代已有来源组件结果及用户视觉结论。改名／会话排序本轮原生模型通过，但本轮桌面未操作；实际程序鼠标／选区、IME／UNC、有限颜色真实终端、System 动态通知、目标 agent（目前未安装）和最终资源／发布门槛仍开放。没有读取系统剪贴板，也没有把合成 shell 当作 agent 兼容证据。

## 保留的失败与观察器修正

run1 只到首次 Terminal 启动，前台 Guard 发现原窗口句柄／身份就绪不足，立即拒绝后续输入；原身份最终均结束。run2 增加独立外层启动收据及一秒窗口就绪观察后走到主题／帮助／表单／缩放，随后失败：目录编辑中 Esc 只退出候选浏览，脚本把弹窗背后的“管理模式”页脚误作关闭确认。当前 run3 改成明确 Ctrl+C 取消并等待“新建终端”消失，再观察管理／终端焦点；没有修改产品规则或在失败现场重放按键。前两轮结果仍为 failed，即使 Scheduled Task 的 LastTaskResult 为 0（worker 在 catch 中落盘后正常结束）也不能升级为通过。

首次只读任务状态查询未返回，另用精确文件取回真实结果后，仅终止已识别的本地 SSH 观察器，未重复桌面任务。首次清理观察器的路径正则转义错误在任何删除前失败；改用 GetFileName 比较，核验三个历史脚本哈希后完成精确删除。两者均是脚手架失败，不隐去或视作产品通过依据。

后续纠正：此前只结束本地SSH的观察器仍在远端运行，后续占用约42GiB并阻止新PowerShell初始化；已按原句柄／FILETIME／完整命令哈希精确回收，资源恢复。实际边界与未完全定位的内部增长原因见[后续报告](2026-11-24-current-ui-followup.md)，不能把原清理段落解释为当时观察器也已结束。
