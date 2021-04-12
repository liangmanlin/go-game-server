@echo off
set codeFile=main.go
@echo on
go build  -o ../bin/CfgExporter.exe %codeFile%
pause