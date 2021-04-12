@echo off

:StartBuild
set /p a=把一个配置文件拖进来，按Enter键：

@..\tool\bin\CfgExporter.exe -os  ../config -f %a%

goto StartBuild

