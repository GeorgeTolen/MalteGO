@echo off
cd /d %~dp0..
echo [MalteGO] Building Linux binaries (amd64)...
if not exist bin mkdir bin
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o bin\greynoise-api .\cmd\greynoise-api\main.go
if %errorlevel% neq 0 ( echo FAILED: greynoise-api & goto fail )
go build -ldflags="-s -w" -o bin\transforms .\cmd\transforms\main.go
if %errorlevel% neq 0 ( echo FAILED: transforms & goto fail )
go build -ldflags="-s -w" -o bin\ui .\cmd\ui\main.go
if %errorlevel% neq 0 ( echo FAILED: ui & goto fail )
set GOOS=
set GOARCH=
echo [MalteGO] Done. Linux binaries in .\bin\
exit /b 0
:fail
set GOOS=
set GOARCH=
exit /b 1
