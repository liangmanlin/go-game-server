@echo off

:StartBuild
set /p a=��һ�������ļ��Ͻ�������Enter����

@..\tool\bin\CfgExporter.exe -os  ../config -f %a%

goto StartBuild

