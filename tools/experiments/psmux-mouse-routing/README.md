# 显式目录宿主的鼠标单通道候选

状态：已通过有限原生复验，产品仅采用显式开发 bundle profile。**没有替换远端旧 bundle／旧产品或官方依赖**。最新[实机结果及接入](../../../docs/testing/results/2026-11-25-single-mouse-channel.md)包含授权、首轮失败、Win32／VT／wheel、四项后端测试和清理；以下“本地准备／待授权”保留前轮准备过程，已由本次结果接续。对应真实失败见[当前复验报告](../../../docs/testing/results/2026-11-24-current-ui-followup.md)。

固定来源仍为 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a`，先应用既有 resize-history candidate.patch，再应用本目录补丁。原始 1604 文件清单、缩高补丁与新变更摘要见 [candidate-manifest.json](candidate-manifest.json)；许可沿 [psmux MIT](../../../third_party/psmux/LICENSE)。新代码不升级 Cargo.lock 或依赖。

## 问题与候选

Windows 10 19045 上实际加载现代固定 SDK 后，原生 Win32 应用同时收到正常 SGR→MOUSE_EVENT 和旧 OS 判定触发的额外记录。普通按下／释放重复，修饰和中／右拖动还夹入无修饰提前释放；直接 Windows Terminal 同模式程序没有这些重复。当前候选增加精确值 `PSMUX_CONPTY_PIPE_MOUSE_ONLY=1`，且必须 `portable_pty::win::conpty_source()` 实际为 Directory 才关闭额外记录通道，包括旧有 wheel 例外。仅设置目录或加载失败回落 Kernel32 不生效；未显式启用时默认路由保持。

该参数只是显式宿主传输策略，**不是 SDK 版本／哈希认证**。若未来接产品，必须由产品对固定 bundle 产物、原 PID／FILETIME、实际 DLL／OpenConsole 校验后选择；不能给任意目录自动启用，也不使用 PSMUX_FAKE_WIN_BUILD。VT／Win32／wheel／混合输入读取器与生命周期仍必须实际验证，不以纯函数表和交叉构建通过替代。

## 已完成的本地准备

- 同一生产路由模块 `src/mouse_routing.rs` 的独立 rustc 测试先 RED（旧规则对已确认目录仍开启第二通道），再 GREEN；默认、现代 inbox 及目录加载失败的旧行为保护共 2 个测试通过。
- 完整上游 macOS `cargo test --bin psmux` 在既有平台代码的 3 个编译错误处终止，未运行测试；没有为本实验修无关上游接口。该失败保留。
- Rust/Cargo 1.95.0、cargo-xwin 0.23.0、锁定依赖和既有 MSVC/SDK 缓存，Windows x64 release 交叉构建 exit 0；**不是 Windows 原生测试或可复现构建接受**。
- 本地 `psmux.exe` 9,637,376 bytes，SHA256 `ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b`。本地产物未执行版本命令；源码包版本仍为 3.4.0，不能凭版本号认证来源。

可复验构建：

```sh
git apply --check /path/to/tools/experiments/psmux-mouse-routing/candidate.patch
git apply /path/to/tools/experiments/psmux-mouse-routing/candidate.patch
rustc --edition 2021 --test src/mouse_routing.rs -o /tmp/mouse-routing-tests
/tmp/mouse-routing-tests
cargo xwin build --locked --offline --release --target x86_64-pc-windows-msvc --bin psmux
```

必须在已应用原 resize 补丁的独立固定源码中执行；不要在用户运行目录应用补丁。Windows 新候选文件名必须为 psmux.exe。

## 原准备阶段的具体授权清单（本轮已授权执行）

此前清单只授权四个同名产品／UI／Session／脚本文件，新后端不在其中。待明确授权后：

1. 创建全新 `windows-pc:asd-acceptance-20261122/terminal-bundle-mouse-20261124/`，只上传上述哈希的 `psmux.exe`；目录若已存在先只读核对，不覆盖未知内容。
2. 从既有已核对 bundle 复制相同 `conpty.dll`（`e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492`）和 `OpenConsole.exe`（`b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775`），再次核对。旧 bundle 与旧 asd.exe 全部保留。
3. 继续用已授权同名 `run-current-ui-tests.ps1`，专属 namespace／runtime／state／合成程序，实际核验 server 映像、PID／FILETIME、DLL 与宿主身份；分别验证 VT／Win32 的严格单次按钮、修饰、hover／wheel、原 shell 和正常退出。不执行 Shift 中／右键、剪贴板 API 或未知用户输入。
4. 候选通过后，才在同名产品／测试构建中增加精确 profile 及对应输入路由，重测必要范围。保留所有失败，任何默认／系统宿主路径都不自动切到新策略。
5. 只清理本轮精确任务／原进程／生成物，不发布、部署、安装系统组件或更改系统设置。

以上 not_run 属于上传前准备状态；本轮实际结果另记，FL／339 catalog 与发布状态仍不提升。

## 完整序列离线复核

`evaluate_win32.py` 专用于已固定的 `(5,4) → (9,4)` 三步手势及该脚手架的按钮／修饰编码。它要求记录编号连续、区间无遗漏且恰好覆盖全部记录；仅允许按下前起点上的无按钮定位 motion，随后按下／拖动／释放的坐标、按钮、修饰、顺序及数量必须完全匹配。按下后的无按钮 motion 也不能过滤。它不适用于任意真实鼠标轨迹，不能证明硬件操作或未捕获期间的投递。

```sh
python3 -B -m unittest discover -s tools/experiments/psmux-mouse-routing -p 'test_*.py'
python3 -B tools/experiments/psmux-mouse-routing/evaluate_win32.py packets.ndjson intervals.json
```

退出码 0 表示全部记录区间匹配，1 表示实际序列不符，2 表示证据文件无效。原存在性判定的 RED 把归档错误轨迹误判 9/9；严格 GREEN 拒绝为 0/9，直接 Terminal 对照为 3/3，另覆盖重复、提前释放、乱序、错误目标／修饰／按钮、缺失／重叠区间及 CLI 非零退出。6 个顶层测试是本地工具验证，不是新原生执行。原始历史判定与数据保持不变。

隔离的产品接入补丁见 [product-candidate.patch](product-candidate.patch)，本轮已按该补丁接回主工作区：仅更新显式开发bundle的固定psmux哈希，并在该已核验选择下发送单通道变量；默认后端继续清除继承值。该命令环境边界RED→GREEN和隔离副本Go普通／race／vet／模块输出通过，Windows产品只交叉构建，具体[产品候选哈希](product-candidate-manifest.json)另记。当前远端仍是`baf087dc…cb57`旧profile产品；这是上传前的准备记录；本轮产品使用新目录内的 404d352e…3366，原远端产品保留。早期[本地准备证据](../../../docs/testing/results/2026-11-24-psmux-mouse-candidate-evidence/manifest.json)不提升平台接受。
