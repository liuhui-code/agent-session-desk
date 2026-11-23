# ADR-0015：用固定 psmux 原生持久客户端协议接入 Tab 终端

- 日期：2026-11-22
- 状态：已接受（Accepted），落实 ADR-0014 已授权的终端技术验证与实现
- 实现与验证状态：终端传输适配器已有代码和限定 Windows 原生证据；尚未连接 TUI Tab，不构成 Tab 界面、IME、视觉或完整 M1/M2 验收
- 关联需求：FR-01、FR-02、FR-03、FR-04、FR-05、FR-11；FL-001/002/005/006/012/015/016/018
- 细化：[ADR-0014](0014-upstream-ui-tabs-and-general-terminals.md)；保留 ADR-0002/0003 的终端所有权和资源边界

## 背景与证据

Tab 内交互需要可持续取得终端屏幕、光标、颜色并写入按键和调整尺寸。Context7 当前 psmux 页面主要来自 master，不能作为 3.4.0 的能力证明。本轮核对固定提交 `20aa1d7254ee93f64dafa059ef89341d9767470a`，并在 Windows 10 build 19045 上对既有固定发行 EXE（SHA-256 `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`）实际验证。

稳定源码和发行物具有原生客户端的 `AUTH`、`PERSISTENT`、`client-attach`、`client-size` 与 `dump-state` 协议。屏幕是完整 JSON 快照，包含 session_name、原 pane ID、尺寸、光标和 rows_v2 样式段；段 width 表示整个段的单元格宽度。初轮产品测试因把段宽误当作字形宽度失败，修复和原生复验均保留在[执行报告](../testing/results/2026-11-22-native-terminal-transport.md)。

## 决定

1. Windows 先使用独立 Go 适配器连接产品私有 data-dir/namespace 内的 psmux session endpoint。只连接 loopback；不扫描用户全局会话、不启动另一套 ConPTY、不引入常驻产品 Host。psmux 继续拥有进程、VT 解析、屏幕和滚动历史。
2. 连接前检查后端报告的固定版本、既有 Session/Run 元数据、原 pane 存活和活动 pane；端点文件须为有界普通文件，读取时核对打开后的文件身份。首帧及后续每帧必须匹配原 session_name 和 pane；split、切换 pane 或未知协议立即关闭，不猜测或回退到其他终端。版本字符串仍不能替代发行物哈希认证。
3. 原生协议是固定版本的内部客户端合同，不宣称跨版本稳定。窗口、pane 或 server 被外部替换的完整仲裁仍未实现；不提升原有后端对外部 respawn 等竞争的保证。
4. 接收端持续排空屏幕流，只保留最新一帧；帧、尺寸、输入和端点字段都有防御性硬上限。终端文本、颜色、光标及字符属性属于目标程序的输出数据；应用 chrome 继续 100% 使用 design tokens。后续渲染只能输出受验证的字符属性，不能把收到的任意控制序列直接写到外层终端。
5. 字面输入使用指定 pane 的十六进制字节传输；特殊键使用受限名称，由 psmux 按目标终端模式转换。粘贴使用 base64，并交给指定 pane 的原生粘贴路径，不追加回车。正文与提交分别操作；写入成功只表示通道写入，不证明程序接受或任务完成。失败／断开后不自动重放输入。
6. Resize 只由交互终端客户端发起，被动预览保持只读。原生推送不保证最终一帧送达，因此提供独立无输入／尺寸副作用的 Refresh；UI 需按上游客户端约一秒的有界兜底节奏刷新活动视图，隐藏时停止，不能用持续 resize 代替刷新。Close 只关闭客户端连接，保留 Session/Run；UI 切换和关闭 Tab 时必须沿用该边界。实际 Tab 焦点、输入串行化及关闭／重开交互仍需下一切片实现。
7. 采用直接原生连接，避免额外运行 `psmux -CC` 客户端及其固定源码中的调试输出文件路径。本产品不记录鉴权 key 或终端正文调试日志。仅合成测试内容可进入显式测试证据。
8. Unix 继续保留显式全屏 attach；Tab 内 tmux 协议适配另行验证，不拿 Windows 原生协议冒充通用 tmux 支持。未验证的 UI、鼠标、IME、真实 agent 或资源预算均保留待测。

## 取舍

原生屏幕快照已经包含 psmux 解析过的完整当前屏幕，断开后重连可直接恢复显示，不需要再维护一套 VT 仿真器或流历史。代价是需要严格锁定并维护内部协议适配；UI、键鼠、原生字宽和资源仍须实测。

本轮未新增 Go 依赖，没有复制 psmux 的 Rust 实现；协议依据的固定源码文件和哈希单独记录。Agent Deck 的上游 UI、样式与交互仍是 ADR-0014 的还原来源，不能用传输适配通过代替视觉还原。

## 验证与剩余门槛

局部 Go 测试覆盖鉴权／屏幕、字面 UTF-8、慢视图下合并、resize、粘贴／提交分离、样式段宽度、身份变化和危险帧拒绝。固定 Windows 二进制上实际测试 fixture 的 Unicode 输入、resize、断开／重连，以及普通 PowerShell 连续命令、再次提示符和未提交输入恢复。精确环境、源码指纹、失败与复验见[报告](../testing/results/2026-11-22-native-terminal-transport.md)。

完成标准仍包含 TUI Tab 工作区、上游界面还原、创建表单、焦点／键鼠、IME、A→B→A、关闭／重开、原生资源复测及真实 agent 验证。当前技术闭环只证明传输方案可继续实施。

## 来源

- [固定 psmux 客户端](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/client.rs)、[连接协议](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/server/connection.rs)、[屏幕样式段](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/layout.rs)、[私有端点路径](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/paths.rs)。
- [ADR-0014](0014-upstream-ui-tabs-and-general-terminals.md)、[原生 Windows 测试规范](../testing/NATIVE_WINDOWS.md)。

2026-11-23 补充：ADR-0016 接入首个 Tab UI。原生缩放／长行探针捕获 `cols=54, cursor_col=54, rows=9, cursor_row=8` 的右边界状态；此前将列等于宽度视为非法会误断开。适配器仅接受这一个右边界待换行位置并将显示光标夹至最后单元格；列大于宽度、行等于／大于高度、身份和文本边界仍拒绝。首轮失败、原始帧、RED→GREEN 与复验见[Tab 报告](../testing/results/2026-11-23-tab-workspaces.md)。不因单个已观察协议状态放宽其他帧合同。

2026-11-23 滚动补充：原生连接新增指定 pane／相对坐标的 `pane-scroll`，沿既有串行输入与最新帧通道使用后端历史；读取复制模式／偏移及原生提示，历史屏幕不显示实时光标。键盘仍交给固定后端的模式处理，不伪造上下箭头、不维护另一份历史。限定 Windows shell 的双向滚轮、返回草稿、A→B→A 与 UI 重开证据见[报告](../testing/results/2026-11-23-workspace-scroll.md)；子程序鼠标协议和完整复制 UI 不在本次通过范围。

2026-11-23 历史完整性补充：固定官方 3.4.0 在首次缩高前已有输出时，原生历史会缺行；直接 psmux capture 和 parser 用例均复现，不能用 Go 拼接历史掩盖。独立补丁候选五轮／20 次实时 capture 与七项原生后端场景已有限定通过，但官方依赖门槛仍失败；候选尚未纳入产品或安装基线。方向、终端所有权和原生有界历史不变，若采用带补丁发行物须另记决定及来源／版本政策。详见[定位、RED→GREEN 与失败记录](../testing/results/2026-11-23-resize-history.md)。

2026-11-23 桌面与宿主分界：历史候选配合系统 ConPTY 在 alternate-screen 缩放返回后仍缺 25–28 行，独立收件末字 α 变为 NUL。固定 SDK 与候选组合的直接严格探针通过，官方＋SDK 仍缺历史；产品尚未接入 SDK。原流程失败及单独生命周期边界见[实际报告](../testing/results/2026-11-23-vt-host-boundary.md)，不提升依赖／验收状态。
