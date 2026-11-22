# 侧栏密度、设置与独立偏好证据

配套[实际报告](../2026-11-23-sidebar-density-preferences.md)。最终原生 `ui-tests.log` 是 78 个顶层／224 个含子测试 PASS，`config-tests.log` 是 12 个顶层 PASS。`receipt.json` 绑定最终构建；`receipt-before-help.json`／`ui-tests-before-help.log` 是补帮助文本前的中间轮次，不叠加结果。`source-before-help.json` 同样只属于中间源码阶段。

`source-before.json` 与 `source-manifest.json` 绑定最终 113 个 Go/module 源指纹，实机后逐项匹配；`build-reproduction.json` 是再次构建三个 EXE 的相同哈希。最终本地检查为 `go-test-final2.log`、`go-race-final2.log`、`go-vet-final2.log`、`mod-verify-final2.log`、`catalog-final.log` 和 `source-limits.log`。

五个 red 文件保存实际失败。早期 green-02 曾因没有等待已排队 resize 而误判；模型合同初版误把合法 resize 也当输入；中间 radio 样式重排遗漏一个 token 字段产生编译错误，均保留并修正。见 fixed 日志及报告。`native-flow.log` 是解析失败轮次的空 stdout；准确的 14 个 MissingLoopStatement 由 `flow-parser-failure.json` 和 `native-inspect-driver.log` 保存，失败脚本为 `flow-parser-failed.ps1`。修正 Do 关键字冲突后未重启桌面任务，继续原窗口，成功轮次为 `native-flow-fixed-parser.log`。

`desktop-events.json.gz` 合并 63 个命令／回执，保存即时设置、只读文件保存失败、Auto 缩放及重开过程。PS5 对 before/after 数组产生 value/Count 包装，保留原样；独立 checks-summary 已核对两个 Session/Run 集合相同。15 张 PNG 是实际最终构建的原始窗口，无改图；只读失败用文本 State 留证，未创建含机器路径的截图。`ui.toml` 是实机保存文件，只有 schema 和密度。

文本按 UTF-8/LF 归档，替换测试用户名和本地仓库路径；原字节 SHA-256 保留在来源清单。图片无归一化。gzip mtime=0。全部内容为合成测试；不含真实 agent 对话。当前目录的用例设计仍 designed/not_run，不把 catalog 工具或本轮局部证据当成全套验收。
