# Windows Alpha 样式反馈：单一侧栏、无框与悬浮提示

日期：2026-11-25。基础提交 `413cd22`。用户在实际 Alpha 上提出五项修改，按 [ADR-0041](../../adr/0041-minimal-workspace-and-iterm-appearance.md) 接续，而非继续接受旧主界面标题、边框和常驻提示。当前已更新 Windows 桌面入口；同款 Meslo 字体仍未验收，系统级字体安装已具体询问用户，尚未得到答复。

## 实现与本地依据

工作区保留一份侧栏会话／分组，移除产品顶部标题／汇总、重复 Tab 条及右侧终端框；逻辑 Tab 身份与 Ctrl+PgUp/PgDn、管理 `[`／`]`、侧栏打开和管理 n/x 保持。删除旧 Tab 条鼠标热区和拖动状态，不能点击不可见控件。终端 geometry 使用全部行数，绘制、尺寸、鼠标和光标共享新原点。

底边提示覆盖一行，普通状态下不预留空间；内部无按钮 hover 显示，移开、键盘、粘贴或尺寸变化后隐藏。关键操作反馈、断连只读说明及原生搜索编辑提示仍可见。覆盖不 resize 或写入终端；覆盖原生末行光标时暂不绘制光标，移开后恢复；提示表面点击／滚轮消费，原程序已开始的拖动／释放继续交回原 pane。

Dark 采用从用户 iTerm Default 读取的 #000000／#bbbbbb 与中性灰 UI，管理文字不普遍加粗；外层 16 色表按同一 iTerm 配置，原程序显式 RGB／索引／粗体保持。Light 和 System 选择的保存／观察生命周期不改；SQLite v5／ui.toml v7 不变，无新依赖。字体由外层 Terminal profile 决定，Go renderer 不假装能设置字体。

[RED](2026-11-25-minimal-workspace-evidence/local/red.log)捕获旧顶部 chrome 及提示不能随 leave 隐藏；[GREEN](2026-11-25-minimal-workspace-evidence/local/green.log)随后通过。[四项提示／边界测试](../../../internal/ui/workspace_minimal_test.go)另覆盖完整末行／光标恢复、关键失败可见、原生程序底边拖动释放及提示表面不泄漏鼠标。既有长字段测试在新背景暴露中文双宽裁切的 prefix 空格不足：overlayLaunchDialog 对被截断前缀补齐精确格宽，保留输入 Value／Position；这是实际生产修复，没有放宽光标断言。

保留六轮全 UI 反馈，包含旧 chrome 合同需要更新、测试脚手架的编译错误，以及实际中文边界修复。旧可视 Tab 拖动测试改为单一侧栏和逻辑视图顺序／原 Run 输入合同，不再声称已移除控件仍可使用。ANSI16 新灰度使用基本 SGR，独立颜色观察器原先只识别扩展 SGR，因此补齐 30–37／40–47／90–107 解析，不把观察器漏读算作产品无色。

普通 Go、race、vet、模块、500 行检查均通过，具体[原始日志与来源哈希](2026-11-25-minimal-workspace-evidence/manifest.json)保留缓存标记。catalog 339 项设计状态不提升；编辑本条设计时先遇到 UTF-8 stdin 拒绝、又误将 ADR 放入仅接受决策项的 decision_refs，均在工具验证阶段停止并修正，最终 render／check --generated 通过。

## 产物与 Windows 原生测试

| 项 | 实际结果 |
| --- | --- |
| 产品 SHA256 | `7cd746fda0b3a78e559232618555ffd70f834f9557c2597caf3d2930033547d4` |
| Windows UI 测试 SHA256 | `44ad985d3e248c62ef51fe8eb3aafd8792dd72820156795eebeea9b63fc3ab59` |
| 原生 UI | 293 顶层 PASS，0 FAIL，0 SKIP，进程 exit 0；[完整 stdout](2026-11-25-minimal-workspace-evidence/native-ui/ui.stdout) |
| 源码集合 SHA256 | `adb29ccfcc97ac14f72aedbd2caccf7b9bea93e4f45342c11e11e56f4a93ae00`；[逐文件清单](2026-11-25-minimal-workspace-evidence/local/style-source-fingerprints.json) |
| 后端／SDK | 沿用上一轮已测的固定 bundle，未替换三文件或新跑后端套件；旧版本证据不冒充本轮执行 |

Windows 10 19045／固定 Terminal 1.25.3131.0，产品使用版本化 `asd-minimal-7cd746fd.exe`；桌面同一 start-alpha.cmd 通过新 manifest 入口启动，原 asd.exe／数据／namespace／终端 runtime 保留。没有创建新的日常聊天输入框。

## 实机切换、字体失败与保护

初轮控制器未能确认旧界面的管理模式，拒绝提交 q，保存失败；没有可靠回执证明此前是否发送过 F6，不写成全轮零按键。随后只读审计已查不到旧 manager／专属 Terminal，无法推断离开原因、作者或旧 manager exit code；后续没有重放退出动作，直接在同一原数据上打开新版本。

字体的当前用户注册首次因 Get-ItemPropertyValue 对缺失值抛错而停止；换为安全读取属性表后完成已核验同字节字体注册，GDI 确认 Meslo family。Windows Terminal 实际截图仍出现“找不到字体”警告，因此 **GDI 枚举、配置存在以及 UIAutomation 返回请求的 Meslo 名称，都不算字体已真实使用**。又通过系统字体文件的“当前用户安装”动词和原生 Terminal 设置页刷新字体列表，均未解决实际警告。没有运行 all-users 安装、重启 FontCache 服务、提权或改系统默认字体。

当前无弹窗可用版本采用专属 profile 的 **Consolas 10／正常字重／黑底灰字**；原默认 profile 未切换。iTerm 默认实际为 Meslo LG S DZ for Powerline 12、非 ASCII 覆盖关闭；Windows 的 10pt 是更紧凑的试值，不是跨平台像素／中文回退完全一致。Meslo 安装到系统字体目录需要管理员权限，具体授权问题保持待答复，不默认批准。

字体用户目录问题的参考是 Microsoft Terminal 官方 [14941](https://github.com/microsoft/terminal/issues/14941)／[15344](https://github.com/microsoft/terminal/issues/15344)；与本机警告相似，不能仅凭 issue 认定同一根因已完全证明。外观参数依据[官方 profile 文档](https://learn.microsoft.com/en-us/windows/terminal/customize-settings/profile-appearance)及实际固定宿主执行。Mac 字体只作用户个人资产复制，没有进入 Git 或 ZIP；当前用户目录／注册中的本轮字体条目保留，未称作已清理或系统发行依赖。

## 实际可用画面与悬浮验证

最初截图检查发现字体警告遮挡，不能将其视作通过。回退后，鼠标观察器又把“最后有文字的行”误当作 viewport 底边，实际没有触及提示热区；这些失败保持 failed。最终 verify-run5 使用真实控件 bounds 与行高：hover 出现单行弱提示，移到上一行隐藏；三张截图已实际检查，主产品标题、重复 Tab 条与圆角框均不存在。

最终核对更新前观察到的 **3 个 Session／Run／backend 元组均保留**；两个原示例 PowerShell 的 PID／creation FILETIME 独立一致。没有宣称第三个用户程序的子 agent 状态已独立验证。最终 verifier 无键盘／剪贴板 API 调用，只作原生鼠标移动和只读观察；这不追认更早控制器动作。UIAutomation 字号属性返回混合 COM 对象，10pt 来自实际 profile 字段和截图变化，不能冒充精确字号／DPI 测量。

截图包含脚本以外的真实终端内容，全部留私有目录；不推断其作者、无干扰或内容完整性，不将任何实际 Session/Run ID、原进程身份和终端正文提交 Git。公开记录仅有[无运行身份的摘要](2026-11-25-minimal-workspace-evidence/desktop-summary.json)。

桌面新增 `Agent Session Desk Alpha Style 1.zip`：22,550,864 bytes，SHA256 `1b79df36d13ba0b53308117b77af815f0c342d910902a9549e64176aadefdc88`。按明确的负载文件清单打包，不含 data／workspace 用户内容；旧 ZIP 保留。本包仍依赖本机已有专属 Terminal/profile，未进行公开发行。

本轮 15 个一次性任务按动作文件及脚本哈希逐个核对删除，保留的新版 manager 数量为 1，窗口及原会话按用户要求保留；[最终任务审计](2026-11-25-minimal-workspace-evidence/final-audit.json)保留每个辅助任务的失败／成功原结果；不把辅助任务 exit 0 当作产品接受，也不把字体错误或表面观察失败改写为通过。

同款字体、窗口外无事件的 leave、完整输入／鼠标／IME／DPI 矩阵、真实 agent、资源和公开发布继续开放。本轮不扩大到修改原终端显式配色或恢复用户已退出程序的对话。
