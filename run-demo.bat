@echo off
echo Running URL Shortener Demo...
echo.
echo Make sure the server is running in another terminal:
echo   start-server.bat
echo.
timeout /t 3 /nobreak >nul
go run -tags demo demo.go
