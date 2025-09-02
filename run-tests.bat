@echo off
echo Running URL Shortener Tests...
echo.
go test -v
echo.
echo Running tests with race detection...
go test -v -race
echo.
echo Tests completed!
