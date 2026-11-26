# 后端消息投递与前端边界修正

日期：2026-11-26。基础提交 `e7ad1db`，依据 ADR-0046。用户指出 M 已是 Move，通信属于后端；本轮撤销前端消息入口。

传输接口先 RED：覆盖中文多行 literal payload、精确 pane、无附着／resize／Enter、缺失 ACK 失败；实现后通过。管理边界测试在历史 e80e874 clean archive 上复现小写 m 打开弹窗；当前管理小写 m 不触发通信，大写 M 仍打开 Move。Manager 新入口由后端能力提供 transport，不要求调用方注入 UI Connection 或回调。

局部检查覆盖 CLI 参数／离线 status、核心持久化及不重放、headless 入口、M 边界和原 F2／Alt 模式。Windows 结果后补，不把交叉编译、合成 TCP 或 Model 当实际原生过程。本轮不执行暂缓的稳定性专项，也不纳入已有未提交 typed-nil 后端候选。

首轮正式 Windows 构建 `fc96f09`：传输 2 项通过，Manager 4 项通过／1 项原生观察失败，随后未执行 CLI／UI。服务端 ACK 和 input_written 已到达，但观察器假定已知单行标记不会被物理折行拆开。单独诊断先等待可控目标 PS 提示符，捕获实际 `PS <controlled-test-directory>> SO\nURCE_<nonce>_中文`，确认默认宽度折行；原始诊断包含临时目录的个人 home 路径，仅保留私有。诊断来自工作树、可能带已有后端候选，不计产品证据。

正式测试改为仅针对该无空白唯一标记去除捕获行间空白，再核对 draft 尾部，不把观察器规则当成通用日志合并算法；显式等待可控 shell 提示符不构成产品 ready 信号。生产投递代码未变。首轮失败与随后正式 clean archive 复验分别记录。
