# 显式终端 bundle：产品接入与原生／桌面验证

日期：2026-11-23。结论：**已把历史补丁＋固定 Microsoft 宿主接入真实产品，指定 Windows 10／PowerShell 组合的普通桌面 Tab 流程通过。** 仍是显式开发配置，完整上游视觉、跨权限、真实 agent、资源和发布验收未完成。产品决定见 [ADR-0019](../../adr/0019-explicit-windows-terminal-bundle.md)。

## 改变了什么

`--terminal-bundle ABS_DIR` 选择一组精确固定的 Windows x64 后端与宿主，不与 `--backend` 混用。目录／文件类型和三份 SHA-256 在创建运行目录前检查；继承的 `PSMUX_*` 仍被清除，只注入显式目录。psmux 继续托管终端，没有新增产品伪终端或常驻检查进程。

新建后、发布元数据前，以及连接原 Run 时，直接 Windows API 核对私有 PID registry、原 server PID／FILETIME／EXE 文件身份、实际 DLL 和唯一 OpenConsole 子进程。没有把 psmux 的静默系统回退当作成功。新建失败只清理新生成的精确目标；旧 Run 核验失败时保留原身份，拒绝终端输入。文件身份比较兼容实际扩展 DOS 路径，不放宽进程身份。

固定来源、补丁及 SDK 哈希见 ADR 和[依赖锁](../../../tools/ci/windows-dependencies.json)。官方默认路径未替换，其[输入／历史失败](2026-11-23-vt-host-boundary.md)继续保留。

## TDD 与本机检查

1. CLI 回归先 RED：未知 `--terminal-bundle`；实现显式选择后 GREEN，错误选择不创建运行目录。
2. 原生旧 Run 回归先 RED：相同 psmux EXE 的系统宿主 Run 被错误接入；加入实际宿主核验后 GREEN，拒绝接入但原 Session/Run 仍可发现／停止。最初使用不同 EXE 副本的用例进一步收窄为同一个文件，再次保存 RED，避免只验证文件路径差异。
3. 追加真实 bundle 输入／历史／重连和错误哈希无副作用测试。完整本机 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 通过，包含工具缓存结果；不是重跑全部原生 UI 套件。source limit／catalog 校验另记，不当作产品测试。

使用已有 x/sys v0.49.0 的 Windows API，从间接依赖改为直接依赖，版本未变。最初 EXE 与该 go.mod 声明修正后重建的 EXE 哈希一致。130 项源码／模块指纹前后只变化 go.mod 的直接依赖标记和追加的 Windows 测试；产品 Go 正文未在桌面执行期间改变。最终 Windows 测试 EXE 包含最终测试内容。

## 真实 Windows 合同

环境：Windows 10 build 19045，PowerShell 7.7.0。产品 EXE `bedda5d361fe63adb2231e314ad90234c24c395635eece1666d35cfc9ab02cfc`；最终后端测试 EXE `1502e44f8cd378600f2ffb1afd252b1c8e70112db401f9640d639496df946965`。SDK 文件来自已验证 Microsoft 签名的固定包，本次产品使用精确文件哈希，不在每次接入执行签名助手。

最终执行 10 个后端顶层场景（3 个新 bundle 场景＋7 个已有原生后端场景）、6 个拒绝子测试，以及 1 个 CLI 场景，合计 17 个 PASS、零 FAIL／SKIP。已有后端场景本次使用历史补丁候选；不能把它们描述为重新验证官方发行物。

新 bundle 场景覆盖：

- 在首次连接前输出 40 个短编号，120×32→94×25；通过产品原生连接发送 `ALT输入α` 和独立 Enter，独立 Console.ReadLine 文件逐字相等。
- 进入 `?1049h`，94×25→61×18，第二次同样输入精确；目标独立报告 61×18；返回主屏后编号 1–40 无缺失／重复。
- 关闭连接、重新构造后端、发现原 Session/Run、重新连接并在原 PowerShell 提示符执行命令；核对原根进程 PID／FILETIME／文件身份仍存活，历史仍完整。
- 同一个后端文件但实际系统宿主的旧 Run 被拒绝，且未被替换；非固定哈希文件在创建运行文件前拒绝。

最初 PS5 未引用带点的测试参数，被拆成 `-test`，未运行产品场景；引用参数后才得到真实 RED。失败脚本和结果没有删除。

## 桌面权限分界：失败单独保留

第一次从 SSH 创建双 shell，再由普通桌面用户的 TUI 接入，实际宿主核验返回 Access Denied，未进入终端输入。独立同桌面令牌探针证明：两个 server 的 QueryLimitedInformation、QueryLimitedInformation＋Synchronize 和模块快照均返回 Win32 error 5。此失败不是 SDK 未加载或历史回归通过。

原 Run 保留，关闭管理视图后显式停止；没有自动提权或跳过核验。首次使用物理 `q` 未确认退出，后续明确取消组合、核对管理焦点并使用既有 Unicode 文字驱动才退出，未把不确定退出当作完成。该负例未取得原 UI FILETIME，审计只声明两条原 shell／Terminal 身份和私有 UI 镜像消失，不补造六进程身份。

当前不能声称支持从普通桌面操作该 SSH 权限上下文创建的 Run。后续更改该兼容范围需要独立身份与权限方案，不能删除核验来通过测试。

## 普通桌面同权限完整流程

再以同一普通桌面用户／权限（Session 2、非管理员）创建新的隔离双 shell。产品 EXE 与核心代码不变，使用真实 Windows Terminal 1.25.3131.0：

| 行为 | 实际结果 |
| --- | --- |
| Tab、原生搜索／选择／复制，120→63→120 列 | 实际系统剪贴板得到合成 Unicode 标记；13 次主屏 capture 全为 1–40 |
| A 草稿、B 草稿、取消搜索、A→B→A | 无额外提交；仅显式 Enter 得到原 A/B PID |
| 受控全屏子程序缩放 | 独立收到 `ALT输入α`，61×18，返回主屏历史完整 |
| 关闭／重开视图、退出／重开管理器 | 原两条 Session/Run 及 shell PID／FILETIME 不变 |
| 结束与恢复 | 6 个登记原进程结束，私有镜像为零；设置哈希、剪贴板文本／格式名恢复，任务移除 |

共 86 个驱动动作回执、15 个原生观测（2 个受控全屏观测，13 个完整历史 capture）。86 是动作数，不是独立产品测试数。已检查并保留四张未修改 PNG：行选择、窄窗全屏程序、返回原提示符、管理器重开；GetWindowRect 包含 DWM 边缘，不是纯客户区截图。搜索找到合成输出不等于搜索输入的每个字符均经独立收件验证；精确输入依据子程序和原生测试文件。

## 清理、证据与范围

[证据目录](2026-11-23-terminal-bundle-evidence/)分别保存 native、user-desktop 与 refused-desktop，附[哈希清单](2026-11-23-terminal-bundle-evidence/manifest.json)和[独立核验汇总](2026-11-23-terminal-bundle-evidence/verification.json)。文本从原 UTF-8／UTF-16 BOM 正确解码、统一 UTF-8/LF、用户名脱敏；PNG 字节不变。认证 key、运行数据库、原剪贴板和二进制未提交。

最终合并审计确认两个有独立 FILETIME 的 bundle 原根进程结束、所有本轮已知私有测试镜像为零、四项自有任务不存在、Terminal 设置恢复原哈希。没有按全局进程名清理，也不宣称机器没有其他官方 psmux。非文本剪贴板内容仍未验证。

FL-001/002/005/006/012/015/016/018/019 保持实现中／未发布，339 项 catalog 仍 designed/not_run。跨权限、启动后 SDK 加载故障注入、宽行重排、目标鼠标、真实 agent、其他 Windows／IME／shell、这一产品 SDK 配置的资源预算／长稳、安装许可与完整上游视觉接受仍开放。旧测量 driver 的资源结果不能转为本次产品宿主通过。
