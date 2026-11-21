# 原生 Windows 执行证据

对应[执行报告](../2026-11-21-m0-native-windows.md)。三个目录分别保留失败、修复后部分成功和最终成功的实际记录；不覆盖或重写失败轮次。

| Run | 被测提交 | Windows native 结果 | 索引 |
| --- | --- | --- | --- |
| [44859255945](https://github.com/liuhui-code/agent-session-desk/actions/runs/44859255945) | `9d5061b19b50cc4cbbc90bd40926af347eb09239` | 5 fail，均停在版本探测 | [evidence-index.json](run-44859255945/evidence-index.json) |
| [44859681759](https://github.com/liuhui-code/agent-session-desk/actions/runs/44859681759) | `9bbacc42bf75b1abac42c763493d10df0892d4fc` | 4 pass、1 fail；删除前提未成立 | [evidence-index.json](run-44859681759/evidence-index.json) |
| [44860102659](https://github.com/liuhui-code/agent-session-desk/actions/runs/44860102659) | `b6f2cd2adf24a77d52244aeed1afd1c2ab5dad87` | 5 顶层与 6 子测试 pass，0 fail、0 skip | [evidence-index.json](run-44860102659/evidence-index.json) |

每轮 `summary.json` 是 Windows 脚本实际写出的摘要，记录系统、ImageVersion、工具链、psmux 身份、被测提交与各命令退出码。`source-before.json` / `source-after.json` 原样保留；两份相同，且已逐个核对对应 Git 提交的 25 个受测文件。

`native-events.jsonl` 保留完整 native 测试 stdout；`unit-test-results.jsonl` 只抽取普通测试的 pass/fail/skip 事件，未改变事件字段。各原始 artifact 文件的大小、SHA-256、保存位置和转换方式在索引中。零字节输出只保存大小与哈希。`native/` 只包含测试生成的 fixture 报告，没有真实会话内容或凭据。

`linux-checks.log` 从同轮 Linux job 的验证步骤提取，包含模块、普通测试、race、vet 与目录检查；不包含 checkout 认证和缓存步骤。提取规则、下载日志及片段哈希写入索引。

下载 ZIP 已与 GitHub artifact API 返回的 SHA-256 核对。CI artifact 保留 14 天，到期时点在各索引中；关键文本证据已入库，不依赖 ZIP 永久可下载。仓库中没有提交产品二进制、psmux 二进制或用户运行数据。

结果范围仅限报告列出的 Windows Server 2025 原生进程／后端场景。交互式 Windows Terminal、IME、焦点、无 WSL 环境、全进程树与资源预算均不由这些文件证明。
