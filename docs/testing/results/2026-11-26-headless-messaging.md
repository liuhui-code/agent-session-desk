# 后端消息投递与前端边界修正

日期：2026-11-26。基础提交 `e7ad1db`，依据 ADR-0046。用户指出 M 已是 Move，通信属于后端；本轮撤销前端消息入口。

传输接口先 RED：覆盖中文多行 literal payload、精确 pane、无附着／resize／Enter、缺失 ACK 失败；实现后通过。管理边界测试在历史 e80e874 clean archive 上复现小写 m 打开弹窗；当前管理小写 m 不触发通信，大写 M 仍打开 Move。Manager 新入口由后端能力提供 transport，不要求调用方注入 UI Connection 或回调。

局部检查覆盖 CLI 参数／离线 status、核心持久化及不重放、headless 入口、M 边界和原 F2／Alt 模式。Windows 结果后补，不把交叉编译、合成 TCP 或 Model 当实际原生过程。本轮不执行暂缓的稳定性专项，也不纳入已有未提交 typed-nil 后端候选。
