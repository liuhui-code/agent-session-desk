# ADR-0019：显式选择并核验 Windows 终端开发 bundle

- 日期：2026-11-23
- 状态：已接受，落实 ADR-0014 的可交互 Tab 与通用终端要求；仅限显式开发候选，不代表发布依赖已验收
- 关联：FR-01/02/04/05/11；FL-001/002/005/006/012/015/016/018/019
- 部分细化 ADR-0005/0015/0018：允许下述可追溯补丁构建和固定 Microsoft SDK 的显式组合；官方基线和未通过项继续保留

## 背景

[缩高诊断](../testing/results/2026-11-23-resize-history.md)定位官方 psmux 3.4.0 主屏历史丢行；[桌面与宿主对照](../testing/results/2026-11-23-vt-host-boundary.md)进一步证明补丁单独不足：系统 ConPTY 下 alternate-screen 缩放仍缺行，独立 Console.ReadLine 将末字 α 读为 NUL。官方后端＋SDK 仍缺历史，补丁＋固定 SDK 的严格直接探针才同时通过。

上游将 `PSMUX_CONPTY_DIR` 标为诊断逃生接口，且加载失败会静默回落系统宿主。产品当前清除全部继承的 `PSMUX_*`，仅在外部设置变量不能解决产品输入要求，也不能把目录存在当作实际加载证据。

## 决定

1. 增加显式 `--terminal-bundle ABS_DIR`，仅支持 Windows x64。目录同时包含本次固定的 `psmux.exe`、`conpty.dll` 和 `OpenConsole.exe`。不能与 `--backend` 混用；相对路径、不可用目录、链接／非普通文件、错误哈希均在创建后端运行文件前拒绝。没有自动下载或搜索任意 DLL。
2. bundle 的后端来自固定提交 `20aa1d7254ee93f64dafa059ef89341d9767470a` 加仓库 `tools/experiments/psmux-resize-history/candidate.patch`。本构建 EXE 哈希 `ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937`；版本输出仍为 `3.4.0 (unknown commit)`，不能称作官方稳定发行或仅凭版本认证来源。完整补丁／工具链／原源码指纹沿用诊断报告。升级或重建须更新固定构建及对应原生证据。
3. SDK 固定为 Microsoft.Windows.Console.ConPTY `1.25.261109001`，DLL 哈希 `e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492`，OpenConsole 哈希 `b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775`。实际包的 Microsoft 签名与 MIT 声明已核对；运行时通过精确文件哈希选择同一内容，不执行额外签名助手。包哈希见依赖锁；发布包许可收集仍须完成。
4. 产品继续清除继承的后端路由／宿主变量，只向自己的 psmux 命令注入显式验证后的目录。psmux 仍拥有 ConPTY 和子进程；不增加产品伪终端、后台守护或每会话检查助手。
5. 新建成功后、发布 Session 元数据之前，核对原私有 registry 的 PID＋creation FILETIME、活进程 EXE 文件身份、实际加载的指定 DLL，以及属于该 server 的唯一指定 OpenConsole 子进程。检查使用直接 Windows API、临时句柄和有界快照。失败只清理刚生成的精确新 Session，不留下可交互但未核验的视图。
6. 接入已有 Run 时重复检查，不靠当前环境推断旧 server 的宿主。选择 bundle 不能把旧系统宿主 Run 偷换、重启或升级；核验失败不发送终端输入，原 Run 仍可查看和显式停止。重连保留原 Session/Run，不自动重放不确定输入。文件身份比较处理实际普通／扩展 DOS 路径，不取消 PID／FILETIME／存活约束。
7. 无 bundle 参数时保留既有官方／系统宿主路径作为对照和开发兼容路径，其已知缺行／输入失败仍为失败。bundle 是本轮可验证开发配置，不自动成为安装默认、所有 Windows 的支持承诺或上游支持配置。

## 验证与限制

必须通过：错误选择在运行副作用前拒绝；同一个 psmux EXE 下的旧系统宿主 Run 被拒绝且仍存在；指定 bundle 新建后实际加载已核对；普通／alternate-screen Unicode 独立收件精确；连接前输出、缩高／返回后的 40 行历史完整；关闭／重开原 Run 后还能在同一 PowerShell 提示符执行命令；桌面 Tab、复制／搜索、草稿、视图与管理器重开，以及精确清理。

Go 单元／race／vet／模块校验、实际 Windows 执行和桌面验证分别记账，见[产品接入报告](../testing/results/2026-11-23-terminal-bundle.md)。该决定不承诺宽行重排、其他 Windows／CPU／shell／IME、真实 agent、目标鼠标协议、资源预算、长稳或完整视觉还原。尚无真实 agent 样本，不以 fixture 替代。

检查只覆盖启动／接入时点，不提供抵抗同用户持续修改后端文件、外部 respawn 或替换运行进程的原子仲裁。发布仍需安装布局、完整第三方许可及资源／兼容验收。

## 来源

- 固定 psmux [ConPTY 实现](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/crates/portable-pty-psmux/src/win/psuedocon.rs)；当前上游诊断文档仅作接口解释，不证明旧版本能力，旧版能力已核对源码和真实 EXE。
- Microsoft [固定 NuGet 包](https://www.nuget.org/packages/Microsoft.Windows.Console.ConPTY/1.25.261109001)，[依赖锁](../../tools/ci/windows-dependencies.json)，[第三方归属](../../THIRD_PARTY_NOTICES.md)。

## 2026-11-25 显式开发 profile 接续

[ADR-0040](0040-verified-conpty-single-mouse-channel.md)在既有 resize-history 补丁之上采用单通道鼠标候选，固定 psmux SHA256 更新为 ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b，两项 SDK 哈希不变；仅已核验目录选择注入精确变量。此前报告和旧目录保持历史，不自动接受旧哈希或迁移原 Run。默认系统宿主与官方安装依赖不变，有限原生输入／历史／重连／拒绝及鼠标结果见[新报告](../testing/results/2026-11-25-single-mouse-channel.md)。
