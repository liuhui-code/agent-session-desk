# ADR-0050：原生 Windows 橘白猫桌宠样机

- 日期：2026-11-26
- 状态：已接受，仅限用户授权的可试用样机；不是完整提醒渠道或发布接受
- 关联：FR-14、FL-020；补充 ADR-0006/0049

后续：[ADR-0051](0051-pet-existing-blocking-reminders.md)按用户澄清补齐已有阻塞事项联动，替代本记录的仅演示交付范围；下文保留第一轮样机决定与验证历史。

## 决定与范围

用户同意尝试 Go + Windows 原生透明窗口的橘白小猫，并要求记录 ADR、每个代码文件正文不超过 500 行。本切片先验证桌宠外观、原生窗口、普通点击不抢焦点、拖动、托盘隐藏／恢复和退出。入口为显式 `asd pet-demo`，默认工作台不启动桌宠，不修改日常安装、个人数据库或终端输入。

采用已有 Go 与 golang.org/x/sys/windows，不引入 WebView、WPF、游戏引擎、网络服务或每会话常驻进程。样机在独立命令进程的锁定 OS 线程维护 Win32 消息循环；专用 cmd/asd-pet 入口保证双击只运行样机，不落入工作台。它不是后台 Host，也不监控 Agent。Windows x64 是本轮目标，其他系统明确拒绝原生运行。

窗口使用 WS_EX_LAYERED、WS_EX_NOACTIVATE、WS_EX_TOOLWINDOW 与 WS_EX_TOPMOST；普通点击返回 MA_NOACTIVATE。透明区域由原生逐像素 alpha 命中测试透过；点击猫只轮转演示状态，拖动不发送键盘输入。托盘提供恢复及退出入口，不注册系统通知或自启。使用 UpdateLayeredWindow 和 32-bit top-down DIB，Go premultiplied RGBA 转为 BGRA；GDI 旧对象先恢复，再释放 bitmap/DC，正常退出移除托盘图标。

空闲、演示待批准、演示待回答、连接未知均明确显示 DEMO；未知不制造 ready／completed。空闲低频眨眼，提醒姿态轻微起伏，未知静止；隐藏时清除动画 timer。所有新视觉尺寸、颜色、动画节奏与 DPI 上限由 internal/ui/design_tokens.go 管理，字体点阵属于固定字形数据。DPI 渲染防御上限 400%，实际多屏／缩放接受仍另测。

素材通过内置 imagegen 生成原创橘白猫四姿态透明图，原图保留，运行时按实际尺寸划分 2×2 单元格并缩小缓存，不假设输出为方形／偶数尺寸。没有复用 VPet／Neko 的源码或美术。完整提示与来源见[素材说明](../../internal/pet/assets/PROVENANCE.md)。四张小图缓存和一个窗口 surface 有界；不从素材大小推断总运行内存。

本切片新增 Go、测试与 PowerShell 驱动文件全部不超过 500 物理行，生产文件继续执行仓库既有的 500 物理行门禁（比正文行数更严格）。不因原生窗口或 Demo 另开大文件豁免。

## 未交付的联动

本轮不读取或更改真实 attention，不具备批准、回答、自动回车或实际会话导航能力。点击演示不是用户批准行为。后续若接入，必须复用 ADR-0049 的原事项 ID／版本／Run 核验与输入排空；宠物只展示与请求导航，不能制造来源事件或解决请求。

UI 退出后继续提醒、跨进程导航、Windows Terminal 精确窗口激活、多工作台单实例、Explorer 重启后的托盘恢复、通知降级及真实 Agent hook 仍开放。共享后台与 IPC 在后续接入前单独决策；没有借本样机把 M3 Host 变成产品前置条件。

## 验证依据与剩余风险

TDD、本地透明／alpha／缩放、Windows ABI／GDI、实际桌面试运行与资源结果分别记录在[报告](../testing/results/2026-11-26-pet-demo.md)，不将交叉编译或官方 API 能力当作 Windows 执行证据。用户视觉接受、实体拖动／菜单完整矩阵、多屏热插拔、长稳／GDI／GPU 和发布仍开放。

参考 Microsoft [layered windows](https://learn.microsoft.com/en-us/windows/win32/winmsg/window-features)、[UpdateLayeredWindow](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-updatelayeredwindow)和[BLENDFUNCTION](https://learn.microsoft.com/en-us/windows/win32/api/wingdi/ns-wingdi-blendfunction)；通过 Context7 查询官方 Win32 API 文档并核对本地绑定布局。不复制第三方桌宠实现，不宣称有已实测的低内存数字或 Windows 全面兼容。
