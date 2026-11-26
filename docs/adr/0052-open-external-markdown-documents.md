# ADR-0052：从终端输出打开外置 Markdown 文档

- 日期：2026-11-26
- 状态：已接受，开发切片；默认阅读器／实体鼠标桌面接受仍开放
- 关联：FR-03、FL-012/016/018；沿用 ADR-0016 的原生终端输入和原 Run 身份

## 决定

用户需要 Agent 输出文件路径后，能打开已有 Markdown 文档。采用 Windows 成熟文件关联：终端焦点下 Ctrl+左键单击完整可见的本地 .md／.markdown 引用，交由系统默认编辑器／阅读器打开。不增加 Agent 协议、内嵌 Markdown 阅读器、WebView、文件索引、后台扫描或新依赖。

识别普通路径、单／双引号与反引号包围的路径、基本 [标题](路径)、[标题](<含空格路径>)、本地 file:/// URI。普通路径只接受明确单行边界，带空格的路径需有引号或上述链接边界；不从含糊文本中猜一个同名文件。位置后缀 :行[:列] 与 #锚点只用于提取文件，首版不要求默认程序跳到指定位置。不是完整 CommonMark 解析器，复杂嵌套／带标题链接仍不支持。

按实际显示文本拼接原生样式 runs，以已有 x/ansi 的单元格宽度映射坐标，覆盖中文／组合字符／emoji；隐藏文字不形成链接。只对当前聚焦、已连接、无错误且非转换中的终端生效，位于导航／留白／底部覆盖区的点击不触发。完整行达到屏幕右边界、上一行疑似续行、截断或歧义时不猜测重组。当前没有添加独立链接着色，帮助和 README 提供入口说明，其他 UI 样式仍由 design tokens 管理。

## 路径、身份和副作用

绝对路径原样解析；相对路径通过原 Session／Run／pane 的 display-message 查询 pane_current_path，查询前后核对原记录、pane 和可用状态。未知／非绝对目录拒绝，不使用工作台 cwd 或搜索同名文件。它是 pane 报告的目录，不证明历史输出产生时的目录；shell provider cd、OSC 目录集成、远程／WSL 目录和并发 cd 仍有边界，推荐 Agent 输出完整绝对路径。没有使用 master 文档的新特性替换固定后端；本机固定候选实际查询结果另记。

UI 每次至多一个解析／打开意图。异步解析结果必须仍属于原视图 generation 与 Session／Run／backend/command；切会话或断连取消，旧结果不能转投新会话。Ctrl 点击与配套释放／移动被消费，不开始原程序鼠标手势或选区；无 Ctrl 和其他普通输入保持原路由，不发送路径、选择或回车。

文件必须存在且为普通本地文件，先规范化并解析符号链接，再检查最终 Markdown 扩展名。拒绝网络／设备路径、非 file URI、驱动器相对路径、ADS、控制字符和可执行扩展；不展开环境值或命令表达式。解析与系统关联打开之间不是文件系统原子快照，文档内容不冻结。使用 ShellExecuteExW 的 open 动作，独立 lpFile、NULL lpParameters，无 cmd／PowerShell／拼接命令；锁定线程初始化 STA COM，SEE_MASK_NOASYNC 保证无消息循环线程的委派完成，错误返回现有状态栏。成功只声称系统打开请求已返回，不能声称用户已读或渲染成功。

## 验证和开放项

[实际报告](../testing/results/2026-11-26-external-markdown.md)分别记录 TDD、路径／坐标模型、旧会话结果拒绝、Windows 原生 COM／ABI 与关联参数观察、真实 psmux 原 pane 目录。默认个人阅读器未被自动启动，真实 Windows Terminal Ctrl+单击和文档渲染尚待桌面验证。历史鼠标用例失败在无变更 main 上同样复现，保留并后置，不因本功能声称全部鼠标测试通过。

SQLite v9、ui.toml v7、依赖版本和日常 Windows 入口不变；所有新增生产／测试／驱动文件 ≤500 物理行。默认应用选择由用户系统文件关联负责，缺失时显示错误并提示设置 .md 默认应用。

依据：Microsoft [ShellExecuteExW](https://learn.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shellexecuteexw)、[SHELLEXECUTEINFO](https://learn.microsoft.com/en-us/windows/win32/api/shellapi/ns-shellapi-shellexecuteinfow)、[ShellExecuteW 参数与 COM 说明](https://learn.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shellexecutew)。Context7 查询官方 Win32 文档，绑定沿用既有 x/sys/windows；未复制第三方实现。
