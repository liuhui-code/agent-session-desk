# 历史 UI 证据归档

source-manifest.json 保存父提交、实际执行二进制和 Go 源文件 SHA-256；没有归档二进制或数据库。checks 保存本地检查，tdd 为逐行为 RED/GREEN，windows/run1 是未成功进入测试的参数失败，windows/run2 是实际原生测试；windows/desktop-v2 和 desktop-v3 是桌面执行。

scripts 是该次 SSH 和交互任务所使用的验收驱动，不是产品入口。desktop-v3 恢复脚本绑定该次已结束窗口的精确 PID/FILETIME，不能直接在新环境回放；重跑必须创建新的隔离窗口并取得新身份。v2 的 012 响应被继续脚本重复编号覆盖，不能按整段严格有序回放验收；保留原始 done/commands 与错误，001–011、013–018 的状态及 v3 019–024 分别支持已列明行为。

首次桌面失败另存 windows/desktop-first-failure。所有图像仅用于本轮局部界面核对，不表示用户视觉接受。规范化文本只调整编码、换行及空白；JSON 中的转义值保持原样。SHA256SUMS 校验规范化后的归档字节。
