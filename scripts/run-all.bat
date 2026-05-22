@echo off
echo [MalteGO] Starting all 3 services...
echo   greynoise-api  -> http://localhost:8090
echo   transforms     -> http://localhost:8080
echo   ui             -> http://localhost:3000

start "GreyNoise API Service" cmd /k "go run .\cmd\greynoise-api"
timeout /t 2 /nobreak >nul
start "Transform Service"     cmd /k "go run .\cmd\transforms"
timeout /t 2 /nobreak >nul
start "UI Service"            cmd /k "go run .\cmd\ui"

echo [MalteGO] All services started. Open http://localhost:3000
