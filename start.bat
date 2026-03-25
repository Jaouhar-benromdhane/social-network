@echo off
:: ─── Social Network — Start (Windows) ───────────────────────────────────────
echo.
echo  [Social Network] Starting containers...
echo.

docker compose -f "%~dp0docker-compose.yml" up -d 2>nul
if %errorlevel% neq 0 (
    echo  [ERROR] Docker is not running. Please start Docker Desktop first.
    pause
    exit /b 1
)

echo.
echo  [OK] Containers started!
echo  Frontend : http://localhost:3000
echo  Backend  : http://localhost:8080/api/health
echo.

docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | findstr "social-network"
echo.
pause
