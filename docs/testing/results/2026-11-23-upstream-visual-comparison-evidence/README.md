# 证据索引

正文与结论见[对照报告](../2026-11-23-upstream-visual-comparison.md)。本目录不包含二进制、数据库、用户会话或整个上游仓库。

- `execution/`：实际运行的脚本、RED/GREEN、本地 Go 和 Windows 原生日志、源码指纹、失败／续接与清理。`*-final2.log`、`receipt.json`、`ui-tests.log`、`final-audit.json` 为最终产品；before-padding 等是中间版本，不能叠加统计。
- `frames/`：未修改固定上游运行得到的原始 ANSI、纯文本和序列化 VT。manifest 保存来源哈希与尺寸。VT 仅加入清屏、逐行 CRLF 和原光标定位，不是重新绘图；cursor shape 不可得。
- `product-desktop/`：实际前后产品截图与逐条操作／UIA 回执。命令行只运行合成 Review A/B。非 public 的启动画面未归档。
- `invalid-codepage-replay/`：主轮八张参考图编码错误，仅用于失败诊断，不参与视觉结论。
- `utf8-replay/`：设置 code page 65001 后重新捕获的参考画面、显示配置、回执与清理。`row-comparison-raw.json` 保存六帧逐行一致、两帧右下末格缺口。
- `upstream-source.json`：完整固定提交、来源文件 SHA-256／Git blob、参考构建哈希与干净状态。v1 只作参考，不是产品依赖。
- `archive-manifest.json`：从临时工作目录归档的原字节／归档 SHA-256。文字统一 UTF-8/LF 并替换测试用户名、实际 home／仓库路径；PNG 和有效来源帧未修改。清单本身及后续核验摘要不自包含哈希。
- `checks-summary.json`：独立核对测试数量、显示配置、Session/Run、源码指纹和最终清理。

执行 `python3 -B <本目录>/compare_rows.py` 可重新计算八帧逐行差异。只移除 UIA 标题一行、行尾 ASCII 空格并 NFC；不丢弃空白行，不补写最后一格。它核对文本，不测量字重、颜色或完整交互。

Windows 脚本使用 `$env:USERPROFILE/asd-acceptance-20261122` 的专用安装和隔离状态目录；原始 Darwin 脚本中的 `<LOCAL_TEST_HOME>`、`<REPO>` 需按本机替换，临时根必须独立创建。先查看脚本和已有进程，不直接重放已执行的输入或复用旧运行目录。保留精确 PID／FILETIME／路径核对、前台窗口保护、尺寸保护和恢复原设置；不按程序名杀进程。

未运行的完整目录用例仍为 designed/not_run；fixture 不等于目标 agent，文本回放不等于上游在 Windows 运行，也不等于用户视觉接受。

Git 字节保护：本目录 frames 在 `.gitattributes` 中标记 `-text`，保留 VT 的 CRLF 和原始 ANSI；提交索引／Git blob 逐项与 archive manifest 核对，避免检出时的自动换行转换破坏哈希。首次提交出现 CRLF 规范化提示，后续证据修正提交恢复原字节；实际运行所用临时文件未改变。
