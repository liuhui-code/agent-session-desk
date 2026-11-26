# ADR-0054：桌宠空闲唤起与眨眼帧对齐

- 日期：2026-11-26
- 状态：已接受，用户报告后的局部修复
- 关联：FR-14、FL-016/020；扩展 ADR-0051 的空闲点击，保留 ADR-0050 的眨眼

用户报告点猫不能拉起终端，并澄清左右动发生在眨眼时。旧版本只在存在可达 pending 时发送导航意图，空闲点击直接返回 false；闭眼素材相对睁眼素材横向偏移。当前 Windows 只读观察为一个 idle 猫、一个宿主绑定，窗口坐标不动，不能归因为重复桌宠或窗口自行移动。

点击存在已知、可达 pending 时仍走原 ID／版本核验及输入排空。否则发送独立 WorkbenchActivate，只请求恢复已经绑定的原宿主窗口；不切 Run、不关闭模态、不标记 viewed、不注入输入，也不把未知状态当成新的授权。现有进程 handle／窗口 nonce 检查、前台策略与失败提示保留，不新增标题匹配或强制按键激活。宠物的零值内部 Activation 只表示窗口意图，由 CLI 转为独立 UI 消息。

保留原眨眼节奏与四帧缓存。生成有界缩放缓存时，按实心 alpha 横向外缘中心将闭眼帧平移到睁眼帧的锚点；不移动窗口、不改缩放或纵向位置，不在 timer 中重新解码／对齐。阈值来自 design_tokens.go 的 SpriteAnchorAlpha，图片和素材归属不变；整数像素舍入允许外缘中心最多半像素差，不要求两张不同表情逐像素相等。

[实际报告](../testing/results/2026-11-26-pet-interaction.md)记录 RED→GREEN、四个 DPI、23 项 Windows 定向检查及受控 Windows Terminal／记事本／最小化流程。日常更新沿用同一已测字节，只切换版本化 exe／清单并保留原件，SQLite v9／ui.toml v7 无迁移；不复用 ADR-0053 已暴露的超时数据库回退脚本。未知启动结果只停止观察，不移动／恢复数据库或自动重复启动。

本切片不提高多原生 Tab 精确定位、多屏／硬件输入、真实 Agent、资源或完整发布接受。API 边界已通过 Context7 查阅 Microsoft [SetForegroundWindow](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setforegroundwindow)与 [ShowWindow](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-showwindow)：用户点击后仍可能被系统拒绝，必须核对实际前台，不制造键盘输入。

## 名字上方残留碎片修正

用户确认前一修复有效，随后报告猫身与名字之间仍有跳点。原始生成图的上方格子包含与主体分离的邻帧／透明背景碎片，横向对齐没有将它们去除。只在上方两个蜷卧姿态的缓存生成阶段保留主体最长连续有像素行段，清空段外像素，再做原横向对齐；画布、缩放及主体像素位置不变，素材原文件不改，不在 timer 中新增处理。眨眼、提醒姿态、名字与点击流程沿用原路径。

四档 DPI、主体像素保持、透明／预乘与原生眨眼／GDI 定向证据见[补充报告](../testing/results/2026-11-26-pet-fragment.md)。这次是已接受渲染方案的局部缺陷修正，没有新数据库／协议／依赖或后台进程。
