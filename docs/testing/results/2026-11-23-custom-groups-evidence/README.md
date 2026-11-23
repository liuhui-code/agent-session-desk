# 分组局部验证证据

`manifest.json` 列出全部允许归档文件、原始字节与归档 UTF-8/LF 哈希；不含二进制、凭据、活目录或用户会话。`native/first` 为占位提示前首轮，`native/final-first-failed` 保留共享占用失败，`native/observer-red`／`observer-green` 为确定性探针与各 20 次复验，`native/final` 为最终五套件。`local` 保留 TDD、先后源码指纹、检查和复现脚本。

原生模型不是桌面证据；当前产品仅本地构建并记哈希，未在本轮启动真实终端。源码指纹含 Go 代码与测试、go.mod/go.sum；170→172 是新增平台观察器测试文件，产品运行代码未因此改变。原失败不覆盖，339 项设计状态不提升。
