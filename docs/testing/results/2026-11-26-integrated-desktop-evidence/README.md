# 本轮证据阅读顺序

此目录包含失败脚本与失败回执，**不是可重复执行的安装工具**。尤其 install-daily-worker 两个版本的超时恢复逻辑已证明不安全，禁止复用。完整结论与限制见[实际报告](../2026-11-26-integrated-desktop.md)。

- 产品桌面最终结果：run5.json；run1～4 是之前失败，不可用最终通过覆盖。
- 原生局部检查：tests.json、各 test.exe.stdout；源码与二进制指纹分 R1／R3，R3～R5 产品相同。
- 安装失败：install-daily-output.json、install-daily-r2-output.json；daily-final-audit-initial.json 只证明当时运行及结构完整性，未证明语义保持。
- 比较器失效：daily-final-comparison.json 的数组计数／快捷方式检查错误；修正后的 daily-final-comparison-r2.json 与 daily-field-diff.json 记录四条分组修订缺失。
- 原件验证：daily-retained-check.json；修复操作：daily-repair-output.json。
- 最终独立复核：daily-final-audit.json、daily-final-comparison-final.json。六条原历史投影一致、两个原示例 shell 身份保持，一个已更新快捷方式，一个工作台／猫／窗口绑定；27 个清单文件哈希一致，临时 fixture／manager 和被检查任务无残留。

数据库、用户历史、私有比较基线、完整桌面文本及本机路径清单均未收录。私有基线仅在 Windows 本机保存。脚本内的目录定位使用环境变量，不应将此目录当作用户运行数据备份。新增 JSON 回执仅统一换行及末尾空行，字段内容保持原样；二进制／源码指纹按各自清单解释。
