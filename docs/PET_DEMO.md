# 原生 Windows 桌宠样机

此页仅描述独立外观诊断。用户要求的已有阻塞提醒联动已另见[桌宠提醒](DESKTOP_REMINDERS.md)与 [ADR-0051](adr/0051-pet-existing-blocking-reminders.md)；运行工作台 `--pet`，不是本页的 Demo 入口。

[ADR-0050](adr/0050-native-windows-pet-demo.md) 定义第一版可试用样机。它展示橘白猫与合成提醒，固定标有 DEMO；不连接真实 Agent、不批准、不发送选项／回车，不取代现有 Alt+N 提醒。

使用本轮 Windows x64 样机构建，运行：

```text
asd-pet-demo.exe
asd-pet-demo.exe --state approval --count 2
asd-pet-demo.exe --state question --duration 30s
```

Windows GUI 子系统构建不会额外弹出控制台；专用 cmd/asd-pet 入口和桌面快捷方式都只进入 pet-demo，双击不会落入默认工作台或初始化个人状态。主 asd 开发构建仍可显式执行 asd pet-demo。该命令不使用 --config／backend／runtime／namespace／state-dir，非法参数在初始化窗口或后端前拒绝。其他系统返回不支持，不伪造桌面效果。

- 点击猫：依次查看 idle／approval／question／unknown 的演示姿态。
- 左键拖动：换位置，并限制在目标显示器工作区内。
- 右键猫或托盘：选择演示状态、隐藏／显示、退出。
- 隐藏后单击托盘图标恢复；隐藏时停止动画。
- duration 默认为 0（手动退出），可指定不超过 5 分钟的自动退出；report 只用于本轮诊断，新文件非覆盖写出。

没有桌宠自启、养成、聊天、远程服务或实际审批入口。托盘图标暂用 Windows 标准应用图标，猫的视觉与完整菜单／拖动、多屏和缩放仍待用户确认。Explorer 重启、UI 退出后持续提醒及实际 Windows Terminal 导航是后续事项。

代码按状态模型、素材渲染、Win32 绑定、surface、几何、窗口与托盘拆分；所有新增代码与测试／驱动物理行均不超过 500。素材与完整生成提示见[来源说明](../internal/pet/assets/PROVENANCE.md)，实际本地／Windows 结果见[验证报告](testing/results/2026-11-26-pet-demo.md)。
