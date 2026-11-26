# 桌宠阻塞提醒

桌宠现在可显示[已有授权／选择提醒](ATTENTION.md)。使用新的 Windows 开发构建，在原工作台参数末尾加 `--pet`，例如：

```text
asd --terminal-bundle DIR --runtime-dir DIR --namespace NAME --state-dir DIR --pet
```

保留原来那套 backend／runtime／namespace／state 参数，才能看到同一作用域的事项；本功能不新增 Agent 接入步骤，也不自动识别所有终端中的授权文字。旧 exe 不支持此开关，独立 `asd-pet-demo.exe` 仍是外观演示。本机日常快捷方式已启用桌宠与系统记事本，数据库备份、升级及安装竞争修复见[交付报告](testing/results/2026-11-26-integrated-desktop.md)；历史公开 ZIP 未更新。

- 空闲猫显示 ASD；待处理显示 WAIT 及数量，待批准为 `!`、待回答为 `?`。
- 点击猫／右键 Open pending item：优先打开未查看事项，重新核验后定位原会话。看过仍待处理，继续在原终端表单中选择。
- 失联或观察失败显示 LINK／未知，保留已有数量；未知不代表已完成。
- 拖动移动；右键隐藏／恢复或关闭提醒。隐藏后用托盘恢复，隐藏期间仍接收状态；关闭宠物不会停止会话或处理请求。
- 工作台退出时宠物一起退出，没有新增常驻后台。

普通状态变化不抢焦点；Windows 若拒绝前台激活或窗口归属无法确认，工作台会提示自行切回原窗口。同宿主多个 Tab 的精确前台选择仍待验；不使用模拟快捷键强行激活。

[ADR-0051](adr/0051-pet-existing-blocking-reminders.md)记录边界，[局部执行报告](testing/results/2026-11-26-pet-blocking.md)记录实际通过范围。原生窗口／后端模拟表单通过不等于实体鼠标、真实 Agent、长期资源或完整 Windows Terminal 桌面验收。

后续 [R5 实机集成](testing/results/2026-11-26-integrated-desktop.md)已在指定宿主中验证从记事本点击猫恢复原 Windows Terminal 前台及原表单；不提高同宿主多个原生 Tab、其他应用或实体鼠标的接受范围。
