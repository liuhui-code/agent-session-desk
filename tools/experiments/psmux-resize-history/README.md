# psmux 缩高历史保留实验

这是隔离验证材料。产品、安装脚本和 `windows-dependencies.json` **没有应用此补丁**，仍固定官方 psmux 3.4.0。完整结果、失败轮与适用范围见[实际报告](../../../docs/testing/results/2026-11-23-resize-history.md)。

固定来源为 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a`；`candidate.patch` 只修改 `crates/vt100-psmux/src/grid.rs` 的 `set_size`：主屏幕、完整滚动区域、有历史容量且光标会被缩高挤出时，先把上方移出的行送入原生有界历史，再调整尺寸。没有新增客户端历史副本或放宽历史上限。许可证见 [psmux MIT](../../../third_party/psmux/LICENSE) 和 [vt100 MIT](../../../third_party/vt100-psmux/LICENSE)。

## 构建与核对

在固定源码的独立副本中操作，先核对原 `grid.rs` SHA-256 为 `d841bf76035f6b35da9e9de5e688177437e9c413dfdb5f35af0615ef0319089b`，不要修改用户正在使用的后端目录：

```powershell
git -C $source checkout 20aa1d7254ee93f64dafa059ef89341d9767470a
git -C $source apply --check "$repo/tools/experiments/psmux-resize-history/candidate.patch"
git -C $source apply "$repo/tools/experiments/psmux-resize-history/candidate.patch"
Set-Location $source
cargo test --locked -p vt100-psmux
cargo build --locked --release --bin psmux
```

本次实际使用 Rust 1.95.0、cargo-xwin 和已有 MSVC/SDK 缓存，在 macOS 交叉构建后于 Windows 10 执行；命令是 `cargo xwin build --locked --release --target x86_64-pc-windows-msvc --bin psmux`。这不是 Windows 原生构建、官方发行物或可复现构建接受结论。构建日志、1604 个原始来源指纹及唯一修改项保存在报告证据中。

`probe.rs` 可放入独立的 Cargo bin 项目，用 `vt100-psmux = { path = "固定源码绝对路径/crates/vt100-psmux" }` 和 `arrayvec = "=0.7.8"` 作为依赖，锁定其余依赖至证据中的版本后运行。它通过公开 parser API 输出 40 行，再把 32 行缩为 25 行，比较全部保留编号。原版实际退出 101，补丁版退出 0；这是解析器 seam，不能代替 ConPTY 测试。

## 原生回归

将候选放在**新独立目录，文件名必须为 `psmux.exe`**。固定源码的 `session.rs::PSMUX_SERVER_IMAGE_NAMES` 只识别 psmux／tmux／pmux；本轮误命名启动失败与认证清理已保留。不要覆盖官方发行物。

运行[诊断脚本](../../ci/windows-history-resize.ps1)需要 PowerShell 7；本次驱动在 5.1 中通过 `ReadAllText` + `ScriptBlock.Create` 明确按 UTF-8 解析。`ShellBinary` 本轮使用 PowerShell 7.7.0：

```powershell
pwsh -NoProfile -File "$repo/tools/ci/windows-history-resize.ps1" `
  -ProductBinary $product -ProductSHA256 $productHash `
  -PsmuxBinary $candidate -PsmuxSHA256 $candidateHash `
  -ShellBinary $shell -ArtifactRoot $freshRoot
```

目录必须不存在，路径必须绝对，哈希必须对应所选产物。每轮仅创建一条合成会话，正常停止明确的原 Session，核对原 shell PID／FILETIME／路径；认证密钥和运行数据库留在本地，不归档。默认每个实时 capture 必须严格等于编号 1–40，缺失、重复、启动或清理不确定均非零退出。`ObserveOnly` 仅允许记录预期历史失败，不放宽启动和清理检查。

`OutputTiming AfterAttach` 验证确定尺寸后再输出；`DirectBackend` 绕过产品启动／capture；`IdleWriter` 使输出后不进入提示符。`SkipEarlyCopy` 和 `SkipSecondaryResize` 只隔离复制及第二次缩放，**首次 120×32→94×25 仍执行**。

候选的五轮流程／20 次实时 capture 和七项原生后端场景已有局部通过结果。宽行重排、部分滚动区域的完整交互、alternate screen 的原生应用、真实 agent、桌面视觉／输入和资源尚未完成；不作为默认依赖或发布接受。
