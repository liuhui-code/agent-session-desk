# 恢复执行证据

这些是接口／模型与真实 Windows 合成 Agent 的局部证据，不是全产品、真实品牌或桌面验收。

- [源码字节清单](sources.json)：493 个 cmd/internal Go 与模块文件；最终构建后复核无变化。
- [最终产品／测试 SHA-256](artifacts.json)；构建过程见 [build.py](build.py.txt)。
- [前置失败说明](setup-failures.md)；各编号 RED／GREEN 日志保留行为修正及测试构造错误，不把准备失败当产品 RED。
- [r2 失败](native-r2/result.json)、[r3 通过](native-r3/result.json)、[r4 通过](native-r4/result.json)与[最终 r5](native-r5/result.json)分开归档；每轮 artifacts.json 绑定自己的二进制。
- 最终 [native-r5 清理](native-r5/cleanup.json)和[进程审计](native-r5/process-audit.json)记录实际 Windows Store／Manager／CLI、私有测试映像审计和 cleanup；驱动见 [native-r5.ps1](native-r5.ps1.txt)。
- 最终 final-all.log／final-race.log／final-vet.log／final-modules.log 来自 [顺序检查脚本](final-checks.py.txt)，部分包使用 Go 合法缓存；没有重跑的旧 native／桌面不计新通过。
- schema-check.log 为 49 项独立 Schema 合同校验；它不执行 Agent 或恢复命令。

原生测试只使用新 scope、合成标记及独立 CLI。完整命令／验收边界见[报告](../2026-11-27-agent-recovery-execution.md)。Windows 日志转为 UTF-8 并脱敏用户目录；私有 runtime 文件不提交。
