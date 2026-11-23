# ADR-0032：固定暗／亮配色与可保存的主题设置

- 日期：2026-11-24
- 状态：已接受方向；实现与平台验收另记
- 关联：FR-03/11/12；FL-003/012/013/016/018/019
- 扩展 ADR-0008/0014/0017/0022/0027，不改变终端输入或进程托管

## 决定

1. 沿固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 styles.go 与 settings_panel.go，先接 Tokyo Night Dark／Light 两套已用角色、主题首行／密度次行和行焦点标记。System 跟随、其他适用设置、完整面板和平台视觉接受继续开放，不显示尚未接入的选项。
2. 颜色仍全部定义在 design_tokens.go，styles.go 组合；每个 Model 持有独立 tokens／styles，没有可变全局主题、后台监视器或新依赖。保留布局／几何及原生屏幕的显式字符属性；应用默认前／背景沿既有 View 机制变化，不声称给目标程序注入主题配置或改写其已输出颜色。
3. 独立 ui.toml 升到 v5，新增可选 theme：缺省为暗色，显式值仅 dark／light；空字段、未知值／类型、旧版本夹带该字段及未来版本拒绝。内部零值保持缺省暗色，保存时省略该字段，兼容原调用者的零值设置。v1–v4 加载不写文件；显式修改保存后写 v5。旧版拒读 v5 的迁移边界单列验证。
4. 主题和密度共用既有整份串行偏好写入器。上下／j/k／Tab 选行，左右／h/l 调整，空格轮换；Esc/S 返回。变化在成功保存后应用，选中请求与已保存主题分别显示；错误保留上次成功主题、可见错误并取消等待退出。待写入时的最后选择有界合并，旧回执不回退新意图；关闭等待已接受写入。
5. SaveSidebarDensity 旧回调只能保存密度；没有整份 SaveUISettings 能力时拒绝主题写入并显示错误，不假称持久保存。创建模板／SQLite／输出记录不参与主题保存。主题设置不向终端输入、切 Run、恢复会话上下文或发起剪贴板请求。

## 验证

配置读写／旧版、固定源码调色板对照、真实 v2 模型／保存顺序、窄短布局、原生屏幕和 Session／Run 连续性分别记账。保持 Windows 上传授权缺口，不以本地配色或交叉构建宣告 Windows、系统跟随或完整上游体验接受。

固定 [配色](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/styles.go)／[设置](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/settings_panel.go)；实际范围见[本地报告](../testing/results/2026-11-24-themes.md)。

后续：[ADR-0033](0033-observed-system-ui-theme.md) 扩展 System 与 ui.toml v6；本 ADR 的 Dark/Light、成功保存后应用、模型局部样式及原终端边界继续有效。原 v5／120 帧证据保持其当时范围。

后续：[ADR-0035](0035-pinned-source-display-colors.md) 保留作者色值，并在 token 显示副本中校正固定来源两个 TrueColor 角色；前轮三组输出差异已有独立本地复验，原报告／平台范围不改写。
