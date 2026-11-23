# 配置页脚四模式的本地还原验证

日期：2026-11-24；基线 `879d94e`。按 [ADR-0034](../../adr/0034-configured-footer-modes.md) 推进 U14/U20：沿来源配置入口接 full／compact／minimal／curated。**没有执行 Windows 原生、桌面、真实 agent 或完整上游视觉验收。**

## 当前体验与配置

默认 full 保留原宽度分级与退出锚点；compact 固定无额外填充的键块＋说明，minimal 固定仅键。curated 用暗色文字、加粗键及两个空格分隔，最多三个与当前选择相关的动作，末尾保留 S 设置／? 帮助。组显示 Tab 展开／折叠，新建、改名；可连接会话显示 Enter 打开、新建、改名，不可连接会话不显示 Enter；空列表仅列可用创建入口。不引入来源品牌重启／fork／队列状态。

小于 50 列时所有模式沿既有窄窗提示。终端焦点继续使用黄色模式／原目标及 F6 提示，管理页脚模式不改变输入归属、框高度或 Tab。curated 隐藏于短提示的其他动作仍可从帮助访问，未改变按键分派。所有新增显示值在 design_tokens.go，styles.go 组合；没有新后台进程、采样、查询、终端输入或模块依赖。

来源本来只通过配置读取 footer，不在 40 项 SettingsPanel 中；产品同样不新增虚构设置行。独立 ui.toml v7 可选 `footer="full"`／`compact`／`minimal`／`curated`，修改后重开 UI。缺省／空／未知值有效显示为 full；已知值忽略大小写和首尾空白。内部零值保持省略，显式非零值保存时规范化。旧 v1–v6 读取不改写，显式保存升级；旧版本夹带 footer、坏类型、未来 v8 拒绝并保留原文。旧 max-v6 reader／writer 拒绝 v7，即使没有 footer 也不覆盖。SQLite 仍 v5，不把两个 schema 混用。

CLI 显式把 footer 传入 Options，再随整份 UISettings 快照保留；主题、密度、导航、比例保存不会主动清除它。不提供配置热重载或偏好多进程 CAS。完整来源设置适用性另见[40 项盘点](../../design/SETTINGS_PARITY_AUDIT.md)；该盘点保留未接入项，不等于完整面板已完成。

## RED → GREEN 与实际检查

- 配置 RED：合法 v7 curated 被旧实现拒绝；GREEN 后保存保留模式和原 Theme／密度／导航／比例。默认零值、四模式、大小写／空白和未知值规范化、v1–v6 只读后升级、版本／类型不匹配及拒绝覆盖通过。原“未来 v7”负例推进到 v8，原 v5 System 拒绝等边界保留。
- UI RED：请求 curated 仍渲染原 verbose 页脚；GREEN 后实际 Model View 列出三个通用动作，设置／帮助在末尾。强制层级另有 RED：宽屏 compact／minimal 仍有 full 填充键块；GREEN 后 compact 保留说明，minimal 仅键，均保留原管理导航提示。
- 两主题的 curated 暗色角色、无填充键块、组展开／折叠、不可连接会话与空列表语义通过。四模式 × 十四宽度 × 两焦点，共 112 个合成完整帧，核对行数、列数、UTF-8 和裁切前的页脚宽度；终端 F6 与 curated 设置／帮助在其适用尺寸内保留。
- 真实临时文件回调贯通 Model 设置：终端焦点 S 字面投递恰好一次；随后管理设置修改主题与密度并保存、返回原 Run，footer 保留 curated。CLI 载入／保存也保留它，取消的写入不改配置或启动模板。它是受控终端服务，不是真实 psmux／shell。
- 全量 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 物理行检查通过；catalog 重新生成并校验，339 项保持 designed/not_run，216 ready／123 needs_decision，不由这些局部结果填成产品通过率。
- config／CLI／UI 测试及产品 Windows amd64 仅交叉构建，没有原生执行。

[证据清单](2026-11-24-footer-modes-evidence/manifest.json)、[最终检查](2026-11-24-footer-modes-evidence/local/checks.json)、[112 帧／模型](2026-11-24-footer-modes-evidence/local/footer-final.log)、[精确角色比较](2026-11-24-footer-modes-evidence/local/role-comparison.json)、[旧版读写](2026-11-24-footer-modes-evidence/local/old-reader.log)和[源指纹](2026-11-24-footer-modes-evidence/local/source-fingerprints.json)单列。归档校验器只做一致性检查，不执行产品或 Windows 测试。

## 固定来源实际对照与保留差异

固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 userconfig.go GetFooter、home_view_frame.go 路由／curatedHint／fitCuratedHints／三项上限选择性适配。保持完整已有 LICENSE、源 SHA 和修改范围；无新增模块。产品中文／通用动作、输入归属和 ui.toml 版本为显式适配。

实际运行来源 helper 与产品 helper，各采集两主题 × 六宽 × 四模式 × 空列表／组／会话，共各 144 个组件参考状态；来源不是完整 Home 运行，产品参考只取真实 View 的页脚行。7 个配置归一化样本完全一致；不宣称本地化文本／不同业务帧逐格相同。来源生产文件未改、前后工作树干净，456 份无关 UI 测试用 Go overlay 排除；不是来源完整测试套件通过。

两主题 × full／compact／curated 的相同英文角色串逐个可见 ASCII 单元格核对文字、前／背景和加粗，标准空 SGR reset 与 reset=0 视为相同状态；不对 RGB 使用容差。6 组中 3 组完全一致，另 3 组明确保留：

| 角色 | 来源实际 RGB | 产品实际 RGB | 当前结论 |
| --- | --- | --- | --- |
| Dark compact／full 键背景（两组） | 121,162,247 | 122,162,247 | 作者值同为 #7aa2f7，精确输出有差异，保持待接受 |
| Light curated 键／说明前景（一组） | 105,109,124 | 106,109,124 | 作者值同为 #6a6d7c，精确输出有差异，保持待接受 |

Dark curated、Light compact／full 三组一致。没有改作者色值去凑运行输出，也没有把三组差异标成精确通过；它们和完整字体／色彩降级／平台视觉属于剩余门槛。

来源执行保持 HOME 不改，四个 XDG 根指向任务目录并预置合成配置，最小环境白名单、GOPROXY off；额外 OS sandbox 禁止真实用户目录文件访问、出站网络及 security／pbpaste／osascript。旧 reader／writer 用基线原 ui.go 的本地 overlay 执行。归档中的工作区路径替换为 `<WORKSPACE>`，原始与归档 SHA 分列；不把归一化配置说成原执行字节。

此前 U19 二进制上传被自动审批以目的地／载荷明确授权不足拒绝，该授权仍未答复。本轮没有远程上传、SSH、CUA、系统剪贴板或系统设置动作，没有换传输绕过。Windows 原生／实体输入、System 通知／清理、真实 agent、目录建议、通用展示开关、错误状态矩阵、完整同尺寸视觉和发布门槛继续开放。

后续：[显示配色修复](2026-11-24-source-colors.md)在新基线上局部关闭本报告记录的三组 TrueColor 样式差异；本报告的原始输出、失败与未执行范围保留。
