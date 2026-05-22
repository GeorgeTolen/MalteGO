@echo off
echo [MalteGO] Running tests...
go test ./internal/... -count=1
if %errorlevel% neq 0 ( echo TESTS FAILED & exit /b 1 )
echo [MalteGO] All tests passed.
