---
inclusion: always
---
<!------------------------------------------------------------------------------------
   Add rules to this file or a short description and have Kiro refine them for you.
   
   Learn about inclusion modes: https://kiro.dev/docs/steering/#inclusion-modes
-------------------------------------------------------------------------------------> 

请你遵循如下规则:

1. 不要创建任何 exe 或者临时文件或者测试文件. 例如我们不要运行: `go build -o k-admin.exe .`. 但是你可以在没有输出的条件下进行编译, 以确认是否正确.
2. 前端也好, 后端也好, 不要构建任何的单元测试, 这是毫无用处的.