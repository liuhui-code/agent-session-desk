# 导航与项目工作区

配置总览与独立可复制示例见[配置说明](CONFIGURATION.md)。当前 LoadWorkspace／LoadLaunch 的顶层字段校验不同：含 [workspaces] 的文件用于 workspace 命令，不能同时供快速新建或 start --template 使用。不要把下例直接合并进日常 launch.toml；这是尚待统一的读取器限制。

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
