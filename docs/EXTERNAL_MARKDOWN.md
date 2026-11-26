# 打开外置 Markdown 文档

Windows 开发工作台中，先将焦点放到内容区，再按住 Ctrl 单击 Agent／CLI 输出中的完整文件路径或 Markdown 链接。系统会用 .md／.markdown 的默认编辑器或阅读器打开文件，终端不接收这次点击对应的输入。

支持以下单行输出：

```text
C:\project\docs\plan.md
"C:\work space\设计.md"
`docs/plan.md`
[设计文档](C:/project/docs/设计.md)
[设计文档](<C:/work space/设计.md>)
file:///C:/project/docs/plan.md
```

相对路径按原会话 pane 报告的目录解析；目录未知时提示使用绝对路径。shell 切目录、历史输出或远程目录可能与当前 pane 报告不同，推荐 Agent 输出绝对路径，不按工作台所在目录猜测文件。路径含空格时需要引号或明确链接边界。行／列后缀与锚点只用于识别文件，首版不控制默认程序跳转位置。

文件必须已存在且为普通本地 Markdown 文件。换行、右边缘截断、复杂嵌套链接和有歧义的路径不自动重组；可以加宽窗口或让 CLI 输出短的完整引用。普通单击／拖动继续终端原有选字或程序鼠标行为。

打开失败会在底部显示原因；没有默认应用时，先在 Windows 中为 .md 配置默认编辑器／阅读器。成功提示表示已请求打开，不表示已阅读。切换会话或断连会取消尚未提交的解析结果。

该功能在 main 开发源码与本轮隔离构建中；历史 Alpha ZIP 和日常 exe 尚未更新。默认个人阅读器及实体 Ctrl+单击仍待桌面验证，准确执行范围见[实际报告](testing/results/2026-11-26-external-markdown.md)和 [ADR-0052](adr/0052-open-external-markdown-documents.md)。
