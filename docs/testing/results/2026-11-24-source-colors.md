# 固定来源 TrueColor 显示值的精确还原

日期：2026-11-24；基线 `48d44a8`。按 [ADR-0035](../../adr/0035-pinned-source-display-colors.md) 修复前轮记录的三组实际 RGB 差异。**本地 20 配色角色与 6 组合样式精确匹配，不等于完整画面、16／256 色或 Windows 实机验收。**

## 原因与修复

固定来源 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 使用 Lip Gloss v1.1.0、termenv v0.16.0 和 go-colorful v1.4.1。锁定代码中，Hex 通道先乘浮点比例，termenv RGBColor.Sequence 再以 uint8 截断通道乘 255 的结果；v2.1.1 的 Hex 直接解析整数 RGBA。因此同样的作者十六进制值，在所用主题中有两个角色的实际输出不同：

| 角色 | 作者值（保留） | 来源实际显示 | 修复前产品显示 |
| --- | --- | --- | --- |
| Dark accent | #7aa2f7 | 121,162,247 | 122,162,247 |
| Light dim | #6a6d7c | 105,109,124 | 106,109,124 |

修复在 design_tokens.go 记录这两个已测显示值，newStyles 使用局部 token 副本。原 Model 作者 tokens 不改，其余 18 个已用角色不改；不是任意颜色的旧库模拟器，不引入 v1／termenv 或复制其转换实现。此显示映射覆盖已有管理样式、暗文字／表单提示及强调色，不改变布局／宽高、Session/Run、输入链路或后台资源所有权。ui.toml 仍 v7，SQLite 仍 v5，go.mod／go.sum 未变。

原终端的显式 RGB 与索引属性不经过此映射，复制样式与终端默认色查询也不参与。默认前／背景的作者值和来源显示值本来一致，既有 View 的默认色机制保留。主题与 System 观测继续从作者 palette 重建样式，没有颜色逐次漂移或全局主题状态。

## RED → GREEN 与实际证据

- 真实 Model View 的 RED 同时观察到暗色键背景 122,162,247 与亮色 curated 文字 106,109,124，均不等于来源。GREEN 后变为精确的 121／105，作者 accent／dim 仍保持原十六进制值。
- 首次全 UI 回归的三个失败保留：旧断言把作者 RGB 直接当作显示 RGB，影响页脚键、亮色弱文字和分隔符。更新为固定来源的精确预期，未引入容差，其他填充／上下文／拖动／焦点断言保留；强制层级的无填充检查同步用实际来源背景，避免变成弱断言。
- 新增真实 View 回归分别在 Dark／Light 中显示原终端 RGB 122,162,247／106,109,124 与 idx:2／idx:4，原值精确保留。它是合成终端屏幕经过真实 Model 视图，不是真实 shell／psmux 运行。既有主题保存、System 阶段、输入上下文、生命周期和布局回归继续通过。
- 来源与产品 helper 分别保留作者值和显示 SGR。产品取实际 Model styles 的十个角色，再通过锁定 v2 样式渲染；来源实际调用 InitTheme 与 v1 样式。两主题共 20 个作者／显示角色逐一精确匹配；full／compact／curated 的两主题 6 组相同英文组合，逐可见 ASCII 单元格核对文字、前／背景和加粗，6 组全部匹配。仅标准空 SGR reset 与 reset=0 归为同一状态，不忽略任何 RGB 差异。
- 继续各采集来源及产品 144 个页脚组件状态，7 个配置归一化样本一致；本地化文本／通用终端动作和来源业务不同，未宣称这些帧整体相同。
- 全量 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 物理行检查通过；catalog 重新生成并校验，339 项 designed/not_run、216 ready／123 needs_decision 保持，不产生整项产品通过率。
- Windows amd64 UI 测试与产品仅交叉构建；未执行原生测试或桌面操作。

[证据清单](2026-11-24-source-colors-evidence/manifest.json)、[RED](2026-11-24-source-colors-evidence/local/ui-red.log)、[首次回归失败](2026-11-24-source-colors-evidence/local/affected-first.log)、[颜色边界复验](2026-11-24-source-colors-evidence/local/color-regressions.log)、[20＋6 精确比较](2026-11-24-source-colors-evidence/local/role-comparison.json)、[最终检查](2026-11-24-source-colors-evidence/local/checks.json)与[源指纹](2026-11-24-source-colors-evidence/local/source-fingerprints.json)分别保存。前轮 [三组差异](2026-11-24-footer-modes.md)和其原始输出不改写；本轮在新基线上局部关闭。

## 来源、文档与限制

固定源生产文件未改，前后工作树干净；456 份无关 UI 测试用 overlay 排除，仅运行本轮 helper，不是来源完整 UI 或全套测试。保持 HOME 不改，四个 XDG 根为任务目录并预置合成配置，最小环境白名单与 GOPROXY off；额外 OS sandbox 禁止真实用户目录文件访问、出站网络及 security／pbpaste／osascript。归档工作区路径替换为 `<WORKSPACE>`，原始与归档 SHA 分列。

Agent Deck MIT 通知继续保留。转换链的四个锁定模块均为 MIT；版本、源文件／LICENSE 哈希见[诊断来源](2026-11-24-source-colors-evidence/local/conversion-sources.json)，仅作诊断参考，未复制其实现。显示数据及修改范围同步到复用清单。

Context7 匹配 Lip Gloss 主库，返回[官方 v2 升级指南](https://github.com/charmbracelet/lipgloss/blob/main/UPGRADE_GUIDE_V2.md)中 color.Color 的样式接口与主分支颜色解析；可选版本只列 beta，本轮没有依赖它。确切行为依据本地锁定 v2.1.1 及固定来源转换链的实际代码／输出，未以主分支文档取代版本实测。

没有改变框架的终端色彩能力检测；16／256 色转换、真实字体／Windows Terminal／恢复、来源整帧和用户视觉接受仍开放。本轮未进行远程上传、SSH、CUA、剪贴板或系统设置动作。此前上传因目的地／二进制载荷授权不足被自动审批拒绝，授权仍未明确；本地配色与交叉构建不能填补原生、真实 agent／IME／鼠标、资源或完整发布门槛。

后续：[颜色能力验证](2026-11-24-color-profiles.md)补实际 v1/v2 渲染器的有限模式、框架事件与默认色包输出；本报告原 TrueColor／未执行范围保留，不回填真实平台接受。
