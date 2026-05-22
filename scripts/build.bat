@echo off
echo [MalteGO] Building all services...
go build -ldflags="-s -w" -o bin\greynoise-api.exe .\cmd\greynoise-api
if %errorlevel% neq 0 ( echo FAILED: greynoise-api & exit /b 1 )
go build -ldflags="-s -w" -o bin\transforms.exe .\cmd\transforms
if %errorlevel% neq 0 ( echo FAILED: transforms & exit /b 1 )
go build -ldflags="-s -w" -o bin\ui.exe .\cmd\ui
if %errorlevel% neq 0 ( echo FAILED: ui & exit /b 1 )
echo [MalteGO] All binaries built in .\bin\
