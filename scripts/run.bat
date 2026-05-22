@echo off
cd /d %~dp0..
echo [MalteGO] Starting Transform Service on port 8080...
go run .\cmd\transforms\main.go
