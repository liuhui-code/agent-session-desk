# 2026-11-21 Windows 资源实验原始证据

每轮目录保留 GitHub Actions 上传的**原始 artifact ZIP**，字节数与 SHA-256 均与 GitHub metadata 核对。`index.json` 为 ZIP 每个 entry 记录独立 SHA-256，并核对其中 25 个 Go/module 源文件指纹与该轮 Git commit。重要原始证据已随仓库存档，不依赖 Actions 的 14 天临时保留期。

| Run | 被测提交 | 资源流程 | 已登记阶段 | 说明 |
| --- | --- | --- | ---: | --- |
| [44862887610](https://github.com/liuhui-code/agent-session-desk/actions/runs/44862887610) | `b682d764` | 失败 | 2 | 活 writer 的共享读取冲突；汇总 CLI 的 cp1252 中文输出另失败 |
| [44863261468](https://github.com/liuhui-code/agent-session-desk/actions/runs/44863261468) | `94a3feea` | 通过 | 22 | 完整短流程，常驻辅助进程的路径核验仍有覆盖缺口 |
| [44863968239](https://github.com/liuhui-code/agent-session-desk/actions/runs/44863968239) | `8cbe7234` | 通过 | 22 | 诊断确认 conhost 的扩展 DOS 路径与普通路径写法不同；缺口原样保留 |
| [44865161425](https://github.com/liuhui-code/agent-session-desk/actions/runs/44865161425) | `30eaf058` | 通过 | 22 | 保守识别等价路径后，常驻辅助进程纳入统计；仍保留 14 条进程读取失败，诊断与短命客户端竞争相符 |

“资源流程通过”表示该脚本的身份、负载窗口与正常清理检查完成，不单独表示完整资源总体、低内存预算、长期稳定或 M0 验收通过。前两轮没有完整路径诊断字段，不能根据后续结果反写旧原始数据。

最终原始 ZIP 为 691,820 字节，SHA-256 `296702d605a02d22a8ad94d6f7a0f0490ce1b8afc9f5a2c86f2ba6af66c8258b`，对应 artifact `14691528489`。完整观察覆盖：20 个阶段为 12/12，5 会话恢复空闲为 11/12，10 会话全部输出为 0/12；全生命周期仍未证明完整。

## 每轮内容

- `artifact.zip`：原始上传包，含 `resources/samples.jsonl`、`events.jsonl`、全部 fixture JSONL、实际配置、UI 文本、统计 JSON/Markdown，以及普通 Go / native 测试输出和源码指纹。
- `index.json`：run、commit、artifact 身份，ZIP/entry 哈希，源码比对、合成数据检查与复现入口。
- `resource-summary.json` / `resource-table.md`：从原包逐字节提取的实验结果与可读统计。
- `native-summary.json` / `process-metrics-graph-results.json`：原生检查及采样工具的独立测试结果。
- `ui/`：真实管理器在测试用 ConPTY 中的屏幕文本，不是 Windows Terminal 桌面截图或视觉验收。
- `windows-execution-excerpt.txt`：验证步骤至上传之前的执行摘录；没有复制 checkout 和凭据配置步骤。

例：从仓库根目录解压最终第四轮到一个新目录，再用原始数据重算统计：

```sh
python -m zipfile -e docs/testing/results/2026-11-21-resource-evidence/run-44865161425/artifact.zip build/review-resource-run
python -B tools/ci/summarize_resources.py build/review-resource-run/resources
```

重算会生成派生文件，不改变原始 `summary.json` / `samples.jsonl`。比较时使用相应被测提交的统计工具，并核对 `index.json` 与派生报告中的源文件哈希。第一轮图测试 fixture 在 Windows checkout 中采用 CRLF；该差异已在索引中说明，后续以 `.gitattributes` 固定工具源文件换行，不能把第一轮 JSON fixture 声称为 Git blob 原字节相同。

这些目录只保存测试自己产生的合成内容，没有整个 psmux runtime、认证 `.key` 文件或用户 agent 会话。采样、归属与统计口径见 [WINDOWS_RESOURCES](../../WINDOWS_RESOURCES.md)，结论见 [资源执行报告](../2026-11-21-m0-windows-resources.md)。
