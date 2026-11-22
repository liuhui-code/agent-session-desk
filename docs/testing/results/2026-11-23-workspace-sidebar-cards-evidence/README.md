# 默认侧栏卡片证据

配套报告：[本轮结果](../2026-11-23-workspace-sidebar-cards.md)。

`final/` 是显示程序文件名后的正式取证轮次，四张 PNG 是该构建的原始生产窗口，没有改图。`intermediate/` 保存首轮日志、TDD 与构建复现／环境核对；首轮含机器路径的截图未发布。JSON 文本只含合成内容，用户名／本地仓库路径替换，UTF-8/LF；原始哈希见来源清单。桌面命令与回执按轮次合并 gzip，mtime=0。

`red-default`、`red-metadata-click`、`red-executable-label` 保存三个实际失败；对应 green 保存复验。`regression-markers` 还记录卡片变高后的小窗口与滚动测试差异。`red-exact-fit` 实际通过，原断言误匹配了 Tab 标题，不能当作 RED 证据；最终改为行首匹配。极低高度 7 行的探索失败只留在工具交互记录，未保留独立原始日志；正式边界矩阵限定有列表内容空间的 8–32 行，不作为最低尺寸接受。339 目录用例仍 designed/not_run。

来源清单绑定最终 Go/module 源码与实测二进制哈希；`build-reproduction.json` 是相同源码的再次构建哈希比对。中间构建只作为该次显示和流程的局部记录，不冒充最终源码构建。未执行本轮资源重测或新 hosted CI。实际测试脚本每个文件均小于 500 行。
