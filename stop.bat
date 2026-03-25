@echo off
:: ─── Social Network — Stop (Windows) ────────────────────────────────────────
echo.
echo  [Social Network] Stopping containers...
echo.

docker compose -f "%~dp0docker-compose.yml" stop 2>nul
if %errorlevel% neq 0 (
    echo  [ERROR] Docker is not running or containers not found.
    pause
    exit /b 1
)

echo.
echo  [OK] Containers stopped.
echo.
pause
