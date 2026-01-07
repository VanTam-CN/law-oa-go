---
active: true
iteration: 1
max_iterations: 50
completion_promise: "COMPLETE"
started_at: "2026-01-07T07:25:37Z"
---

分析并标记系统中的无用文件,生成清理建议清单。

任务:
1. 扫描项目,分类文件:
source,test,config,data,docs,unknown
2. 分析代码依赖关系:
使用 go list 分析包依赖、检查 import 语句、识别孤立文件(未被引用)
3. 生成清理建议报告(cleanup_plan.md):
高风险文件(可能在用)、中风险文件(不确定)、低风险文件(明显无用)、每个文件的删除理由、预计释放空间
4. 对每个建议删除的文件运行模拟验证:
假设删除该文件、检查是否影响编译、检查是否影响测试

输出格式:
清理建议报告
可安全删除 (低风险)
[ ] path/to/file1 - 原因
[ ] path/to/file2 - 原因

完成后输出 <promise>COMPLETE</promise>
