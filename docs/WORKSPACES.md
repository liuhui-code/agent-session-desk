# 导航与项目工作区

Windows Ctrl+Alt+Space 恢复原工作台；已有到期事项按原规则核验定位。模态保留；注册冲突只提示。不是后台监测。

Alt+P 回到最近成功连接的另一会话视图，16条内存历史、跳过已关视图，不改行序。

管理 W 启动 launch.toml 的 default workspace；--config可指定。CLI：asd --config FILE workspace NAME。每次显式启动一组新的Run，重复执行会再开；最多8，先验证再串行启动。中途失败停止，已开不撤销，不重试不确定启动。

```toml
schema_version=1
[templates.shell]
executable='pwsh.exe'
cwd='C:\project'
[templates.test]
executable='cmd.exe'
cwd='C:\project'
[workspaces.default]
templates=['shell','test']
```
