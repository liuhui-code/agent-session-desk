# Tab 资源执行与诊断证据

对应[报告](../2026-11-23-tab-resources.md)。run1/2/3 均失败；本目录不提供新的完整资源通过结论。

- `run1/`：包装器参数绑定失败；尚未采样。
- `run2/evidence/`、`run3/evidence/`：原始摘要、gzip JSONL、合成 fixture、配置／屏幕与失败状态；统计表和压缩分析仅汇总已登记阶段。
- `focus-*/`：七次独立短探针，各自有真实构建哈希与失败摘要；不计入资源阶段。最早 key-only 探针外层打印了旧摘要，真实摘要以对应子目录为准。
- 根目录 RED／GREEN 日志及 `local/`：失败／复验、Go 检查、UTC 类型观察、任务状态与有界观察。最早两份大观察输出只存过滤字段和原始捕获哈希。
- `sources/`：本轮最终五个资源脚本；`drivers/` 与根目录 probe 脚本：实际包装器、诊断与清理流程。
- `diagnostics/`：临时诊断改动与二进制哈希；最早 logger 源码指纹不完整，限制见报告。诊断构建不作为正式产品或资源结果。
- `cleanup.json`：36 条精确自有映像路径剩余为零，三项任务删除。
- `checksums.json`：除自身外全部公开证据的 SHA-256，包含派生文件。
- `manifest.json`：原始文件、规范化正文与公开文件各自 SHA-256；不归档 EXE、认证文件、SQLite 或整个 runtime 目录。

Windows ZIP 条目分隔符统一成目录；BOM 标记的 UTF-16／UTF-8 文本转为 UTF-8/LF，并替换本地账户／仓库路径。JSONL gzip 的 mtime=0。原 ZIP 含账户路径，仅本地保留，其哈希入清单。派生分析内部 source_sha256 对应规范化前原始字节；复算公开数据得到不同来源哈希是预期变化，统计值应一致。

复算时将对应 `summary.json` 与解压后的 `samples.jsonl` 放到新的临时目录，运行 `python3 -B tools/ci/summarize_resources.py 临时目录`。输出的原实验状态必须仍为 fail；不要把“汇总成功”解释成产品测试通过。当前复现入口及负载／口径见 [WINDOWS_RESOURCES](../../WINDOWS_RESOURCES.md)。
