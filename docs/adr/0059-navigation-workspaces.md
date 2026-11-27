# ADR-0059：唤起、最近视图、项目

2026-11-27；用户按序授权。Win32 RegisterHotKey，无轮询/独立进程；Ctrl+Alt+Space冲突不抢占，UI退出取消。Alt+P内存16视图，身份匹配，成功后记录。

项目复用配置v1模板，workspaces最多8；W固定default，CLI命名。一次读入解析后验证，串行启动；失败保留已开，禁止UI自动再开。schema不变，无工作区常驻副本。

接口与[用法](../WORKSPACES.md)。整体资源/实际热键/窗口与多项目仍待验证。
