# 缩高后的终端历史缺行：定位与修复候选

日期：2026-11-23（上海时间，原生回执为 UTC）。基础提交 `e0a853d`；本报告随本地提交绑定诊断工具、实验补丁、来源和证据。**官方依赖仍存在缺行，候选仅通过本报告的限定验证，未纳入默认产品或安装依赖。**

## 结论

此前[复制／搜索截图](2026-11-23-native-copy-search.md)中的 `history line 10 → history line 18` 已重复复现。它发生在首次 `120×32 → 94×25` 缩高，**进入复制模式之前**；直接 psmux 的 capture 和原生 rows_v2／content 都有缺口。Go 解码／渲染不能补回后端已经丢失的数据。

固定 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 `tree.rs::resize_window_panes` 调整 ConPTY 及两个 parser；`vt100-psmux/src/grid.rs::set_size` 直接 resize 行数组，没有把缩高挤出的行送入原生历史。公开 parser API 的独立用例也会丢编号。实际 ConPTY 重绘后留下的缺口与单独 parser 裁尾位置不同；本轮没有保存 ConPTY 原始 VT 字节，不声称逐字节证明其重绘时序。

实验补丁只在主屏幕、完整滚动区域、有历史容量且光标会被缩高挤出时，先将上方移出的行送入已有原生有界历史，再调整尺寸与光标。**没有在 Go 中拼接历史、恢复伪造文本或新增历史缓存。** 候选在相同 Windows 原路径及对照中，全部保留 1–40 的精确顺序，无缺失／重复。

## 反馈循环与对照

[参数化脚本](../../../tools/ci/windows-history-resize.ps1)为每轮创建一个新目录／namespace／合成 shell。脚本先登记独立 shell PID／FILETIME／映像，再以认证连接取得指定 Session/pane；只记录合成输出与屏幕投影，不导出认证密钥。默认每个实时 capture 必须严格等于 1–40；启动或清理不确定也非零退出。

诊断先复现两次同样缺口，再按“后端缩高裁切、提示符／ConPTY 重绘、历史索引、产品投影”四个可证伪假设测试。主要记录如下：

| 场景 | 官方 3.4.0 | 同一实验候选 |
| --- | --- | --- |
| 首次连接前已经输出，再缩高 | 第 11–17 行消失；两轮相同缺口，最终严格脚本再次 RED | 两轮完整，含复制提示、第二次缩高、放大和搜索 |
| 先确定尺寸、再输出 | 冻结复制画面看似连续；实时 capture 在后续缩高丢第 18–24 行 | 实时 capture 全部完整 |
| 直接 psmux 启动及 capture；不预先复制／不做第二次缩放 | 同样 10→18，排除产品启动、解码与渲染 | 完整 |
| 输出后只等待，不进入 PowerShell 提示符 | 同样 10→18，排除提示符行编辑作为必要条件 | 完整 |

官方最终严格脚本退出 1：首次 capture **40** 行，首次缩高后 **33** 行，再缩高后 **26** 行，放大与搜索均没有找回。候选五轮流程共 **20 次实时 capture**，每次均精确为 1–40。部分冻结画面连续不能证明背后实时历史完整；早期 `repro1` 的 `no-gap-observed` 仅指当时可见页，本报告按实时 capture 修正了判定范围。

这些是 headless 原生协议／CLI 与合成 PowerShell 流程，**不是 Windows Terminal 桌面或完整 TC-TRM-014 通过**。设计用例仍为 designed/not_run，扩展了首次输出时序、冻结／实时分开核对及编号完整性；339 项目录未改执行状态。

## TDD、来源和构建

[实验目录](../../../tools/experiments/psmux-resize-history/)包含最小公开 API 用例及补丁。使用相同锁定 parser 依赖时，原版用例实际退出 **101**，补丁版退出 **0**；补丁版 vt100-psmux 原有测试 **196 PASS、0 FAIL、0 ignored**（含一项文档测试）。这是 parser seam，不冒充 Windows 执行。

核对固定来源 **1604 个文件**，唯一修改为 `crates/vt100-psmux/src/grid.rs`，原 SHA-256 `d841bf76035f6b35da9e9de5e688177437e9c413dfdb5f35af0615ef0319089b`，候选 `07786772a2964c053ce66fc0969065ab40db134e3a1aee070e987a8fae2536b9`。补丁 SHA-256 `7b75ad87484ccc07ce80c188e27dbe4f1c843b0abcd0386a56a0a6782d277d0d`。MIT 原文、版权和机器来源清单已保留；没有把候选或修改归为官方发布。

Rust 1.95.0／cargo 1.95.0 在 macOS 使用 cargo-xwin、固定 Cargo.lock 和既有 MSVC/SDK 缓存交叉构建。首次 offline 缺 quickcheck 索引，保留失败；允许下载锁定依赖后 release 构建完成，用时 5 分 10 秒。候选 `-V` 报 `psmux 3.4.0 (unknown commit)`；该字符串不能认证其来源，必须结合本报告的补丁、源指纹和二进制哈希。没有验证跨环境可复现构建，也没有新 hosted CI。

| Windows 实际执行产物 | SHA-256 |
| --- | --- |
| 复用的产品（Go 代码未改） | `a2518bc0c28c4ff23ac544b440fe1fcc704c1d0e682bbfaf3cb7da91d65934b0` |
| 官方 psmux 3.4.0 | `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca` |
| 实验候选 psmux | `ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937` |
| 本轮后端测试 EXE | `9e0ff3d77fa12c50e18e87886015ab95eaf3695d7381dd895f4d16e0a6301663` |
| 本轮 fixture EXE | `470ff13355fe4360623a70cf75d28435b8c34d3b1d638c02591523b08efc2442` |

Windows 10 Home 19045、PowerShell 7.7.0；本轮产品／测试构建环境为 Go 1.27.4 darwin/amd64，测试 EXE 为 windows/amd64。候选执行全部 `TestWindowsPsmuxNative`：**7 个顶层／13 个含子测试 PASS，0 FAIL、0 SKIP**，覆盖真实 fixture 启动／退出、隔离与拒绝、快速退出、原 pane 拒绝回退、CLI 结束后重连、Tab 传输／resize／重连和普通 shell 草稿。没有重新声称 UI、IME 或资源通过。

## 中间失败与最终清理

- 首个直接对照把 `-L namespace` 错写为 registry base，未找到预期端点；该轮为 harness-failed，独立核对原 fixture 已结束后重试。
- 审计脚本曾把 PowerShell `-Name` 附带的对象元数据展开为大量输出，随后改为仅记录计数；归档保留修正后的有限审计，不保存无关文件系统对象。
- 候选最初命名为 `psmux-history-candidate.exe`。固定源码的 `PSMUX_SERVER_IMAGE_NAMES` 不识别该 stem，导致 server／shell 已启动但产品后续发现失败。该轮不能计入历史通过。通过原认证端点核对 Session 后清理；第一次只读 dump 连接已经关闭，后续 kill 未被消费，失败审计保留；改用持久认证连接后正常退出，**未按名称或 PID 强杀**。同一哈希放入新独立目录、保留 `psmux.exe` 文件名后才进入成功矩阵。
- 原生 Go 套件本身退出 0，但与矩阵并行的外层“同映像总数”审计当时看到 1 个进程而退出 1。这不是清理通过，也不足以判为泄漏。两者结束后另行统一只读审计，原生套件回执不改写。

[最终审计](2026-11-23-resize-history-evidence/native/final-audit.json)核对 **14 条原 fixture 身份全部结束**、七个有可读 namespace 的产品列表为空，误命名／标准命名候选、后端测试、fixture 和产品的私有精确映像匹配均为 **0**。误命名 server 的原 PID／FILETIME／路径及端点移除还由专用清理回执独立核对。其余旧诊断的 exact target 正常停止和直接端点关闭分别保存在各轮结果中；没有 desktop 任务、设置或剪贴板变更。

Go 普通测试、race、vet、mod verify 通过；124 个 Go/module 文件前后指纹相同。source limits、补丁适用性、归档哈希、独立 capture 顺序和 catalog 检查单独记录。目录校验不执行产品测试。

## 证据与剩余门槛

[证据目录](2026-11-23-resize-history-evidence/)包含官方失败、候选五轮、原生套件／fixture、来源指纹、parser RED→GREEN、构建和清理。文本规范化为 UTF-8/LF，日志／驱动脚本的结束空行规范为一个换行，并替换测试用户名／home／仓库路径。`stages.json` 是合成屏幕投影，归档省略重复逐单元格 content，保留 content_text、rows_v2、原文件及原 content 的哈希；不是完整协议原始帧。原导出包 SHA-256 为 `c2e0eaee06d75e803f8baf2f952b48688452941c17f8f08362565a98d524f611`，各原件／归档哈希见 manifest。运行密钥、数据库、二进制和用户内容未提交。

官方依赖的历史保留门槛仍失败。是否接纳带补丁的发行基线，需单独记录决定与发行物来源／版本／许可证／构建政策，不能静默覆盖官方哈希或以候选通过提升 FL 状态。还需补丁下的真实桌面复制／选择／缩放／输入组合、宽行重排、部分滚动区域／alternate screen、真实 agent 与资源回归。主界面上游视觉评审及其余会话组织能力继续推进，完整目标仍未完成。

后续桌面复验发现候选单独不足：alternate-screen 缩放返回仍缺四行，且独立输入有 NUL；固定 SDK＋补丁组合的直接探针通过但产品未接入。请同时阅读[宿主分界报告](2026-11-23-vt-host-boundary.md)，不要将本报告的主屏限定结果外推为完整桌面通过。
