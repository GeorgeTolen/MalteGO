@echo off
cd /d %~dp0..
echo [MalteGO] Building all services...
if not exist bin mkdir bin
go build -ldflags="-s -w" -o bin\greynoise-api.exe .\cmd\greynoise-api\main.go
if %errorlevel% neq 0 ( echo FAILED: greynoise-api & exit /b 1 )
go build -ldflags="-s -w" -o bin\transforms.exe .\cmd\transforms\main.go
if %errorlevel% neq 0 ( echo FAILED: transforms & exit /b 1 )
go build -ldflags="-s -w" -o bin\ui.exe .\cmd\ui\main.go
if %errorlevel% neq 0 ( echo FAILED: ui & exit /b 1 )
echo [MalteGO] Done. Binaries in .\bin\
