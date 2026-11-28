# Windows 便携候选打包

实现合同见 [ADR-0072](../../docs/adr/0072-reproducible-windows-portable-delivery.md)。此目录供发布者使用，下载 ZIP 的用户不需要 Python 或 Go。工具不下载依赖、不构建程序、不启动 UI，也不创建／公开 Release。

1. 从干净提交分别构建 `./cmd/asd` 与 `./cmd/asd-launch`，采用 `-trimpath -buildvcs=true`、`GOOS=windows GOARCH=amd64 CGO_ENABLED=0`；保存各自 `go version -m` 输出。产品与 launcher 必须相同提交，Windows 原生执行证据另存。
2. 在 Windows 上对这个实际程序运行 `update package --program EXE --tag v0.1.0-alpha.2 --output asd-update.json`。保留原文件，不手填兼容性或为了改显示版本重建已测字节。
3. 准备与合同 runtime 哈希完全相同的三文件目录；当前仍为既有修改 psmux／ConPTY。其来源与原源码 ZIP、许可集可由历史 Alpha 1 的公开资产核对取得，也可以从原构建记录重新收集。不得从个人 Terminal 设置／data 目录制作发行包。
4. Go 许可索引 `go-modules.json` 必须与实际程序内嵌模块集合一致；包含原 Go、Rust crate、复用源和运行库许可。索引匹配不替代完整法律审计。准备[官方 Terminal x64 ZIP](https://github.com/microsoft/terminal/releases/download/v1.25.3131.0/Microsoft.WindowsTerminal_1.25.3131.0_x64.zip)，工具固定核验官方 SHA256，不接受另一个用户目录的宿主文件。
5. 在仓库根目录调用下列命令，用自己的已核验输入路径替换占位值。输出目录必须不存在，失败保留该轮目录以便核对，不覆盖重试。

```text
python3 -B tools/release/package_windows.py --program INPUT/asd-windows-amd64.exe --launcher INPUT/asd-launch.exe --contract INPUT/asd-update.json --runtime INPUT/terminal-runtime --licenses INPUT/licenses --terminal-zip INPUT/terminal.zip --source-zip INPUT/psmux-source.zip --build-info INPUT/build-info.txt --launcher-build-info INPUT/launcher-build-info.txt --output OUTPUT
```

产物为完整 ZIP、同字节独立 `asd-windows-amd64.exe`、`asd-update.json`、初始负载 `manifest.json`、修改 psmux 源码 ZIP 和 SHA256SUMS。Terminal 原始分发文件及 NOTICE 保留，只在全新包目录新增 `.portable` 和 settings；配色引用已记录的完整 iTerm 参考色表，不复制字体／历史／测试启动命令。相同输入的 ZIP 使用固定时间戳生成，不包含构建机器路径。

新目录先核对全部负载摘要，再实际执行 Start.cmd、空首开／默认新建／输入切换／重开及所有原进程清理。CLI 更新需独立验证在线拒绝、备份、入口切换、保留新数据回退与 repair；模拟 HTTP 不能冒充真实 GitHub 下载。准备包／草稿、公开发布与匿名下载核验分别留证，不覆盖旧 Release 或移动已有 tag。

本地快速工具检查：`python3 -B -m unittest discover -s tools/release -p 'test_*.py'`。完整产品与原生检查遵循仓库 AGENTS.md；本工具检查不算 Windows 接受。
