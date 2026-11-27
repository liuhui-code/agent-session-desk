# 2026-11-27 清理旧测试示例会话

用户要求移除侧栏中的“开发／守护”等预置。当前日常作用域和默认／最近启动配置并无这些标签；进一步盘点发现两个旧 `asd-output.exe` 工作台仍在使用另一测试作用域。经其原版本 JSON 列表确认是 `A 开发`、`B 守护`、`C 旁观` 三个旧 fixture 会话，并非产品内置角色或当前日常模板。

## 已执行

固定原产品 SHA256 `da88e3ac5622c817cabc44231ab81862a6db7c9af865de7702f96e1fa84acc7d`、原官方 psmux `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca` 和 fixture `f9a30f0dae7bd6a9fffe7deedf8a194405d67d651aad0b99cc20296e8cb5d9d6`，使用旧版本操作原 schema，不迁移旧库。完整 Session／Run／backend ID、原名称和 fixture 路径一致才继续；两个活动 pane 的原 PID 进程 handle 保持到退出观察。

通过原产品 stop 三个精确测试 Session，三次返回 0；原已退出 A 的保留 pane 一并移除，B/C 两个 fixture 的原 handle 均观察到退出。随后该测试作用域 live 列表为 0；日常作用域 Session／Run／backend 身份集合保持，forced=0。没有停止日常会话、删除数据库／历史记录、修改 Terminal 设置或添加名称黑名单。新开旧测试工作台也只发现空列表，不会由产品重新生成示例。

[实际回执与执行脚本](2026-11-27-example-session-cleanup-evidence/)。源码与日常程序构建未变，无需重发 exe；没有将维护清理计为产品／资源／发布验收。

## 保留的前置失败

旧版读取首次遗漏 --backend，PATH 找不到 psmux；第二次 list 未指定 --json，拒绝解析而未输出正文。补齐原参数后才定位三个目标。受限交互任务的 Process.Path 校验失败，stops=[]；同用户 SSH 只读核验两个原 PID 路径一致后，改走纯 CLI 通道。Windows PowerShell 5 数组管道使首轮 scope count 校验失败，stops=[]；改为显式枚举 JSON 对象后完成清理。没有重试任何不确定的 stop，也没有放宽原会话或进程路径身份校验。
