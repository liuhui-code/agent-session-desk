# ConPTY 分界与 Tab 短资源证据

对应[执行报告](../2026-11-23-tab-resource-host.md)。`system-probe/` 失败、`candidate-observer-failure/` 失败、`candidate-probe/` 成功、`resource-sdk1/` 为完整受控短基线通过；各目录独立保留，未改写旧结果。

- `resource-sdk1/evidence/summary.json`、压缩 JSONL、统计表与压缩分析：7 场景／22 阶段／280 快照，含缺测和正常清理。SDK 仅给测量 driver 使用。
- 三个 probe 目录：独立原始控制台读取、正式产品焦点与宿主加载记录；不混入资源阶段。
- `drivers/`：读取器源码（`.go.txt`，复制回 `.go` 后按仓库模块构建）、实际包装器、探针和清理脚本；单文件不超过 500 行。
- `sources/`：资源执行时最终工具／锁定文件；完整 117 项运行前后指纹在 resource-sdk1 中。
- `provenance/`：官方 NuGet catalog／登记／nuspec、包 SHA-512 一致性、PE 导出、构建指纹、8 个宿主与严格归属快照的交叉核对及精简统计。
- 根目录与 `local/`：实际 RED／GREEN、Go／目录检查、已核验同一运行任务的有界观察和最终导出。
- `cleanup.json`：157 个已出现身份的终态审计；私有实验映像／候选宿主剩余精确匹配为零，任务删除。
- `manifest.json`：原始字节／规范化正文／公开文件各自哈希。`checksums.json` 包含除自身外全部公开证据。

文本只进行 UTF-8／BOM 标记 UTF-16 → UTF-8/LF 与账户／仓库路径替换。JSONL 与分析使用 mtime=0 的 gzip。原 ZIP 因含账户路径仅本地保留，哈希入清单；不发布 EXE、DLL、NuGet 包、控制密钥、运行数据库或整个 runtime。

分析内的 source_sha256 对应规范化前字节。复算时把 `summary.json` 与解压的 `samples.jsonl` 放到新临时目录，运行 `python3 -B tools/ci/summarize_resources.py 临时目录`；归档校验另验证统计值一致。来源哈希因脱敏变化是预期，统计不能变化。完整复现参数与候选包位置见 [WINDOWS_RESOURCES](../../WINDOWS_RESOURCES.md)。
