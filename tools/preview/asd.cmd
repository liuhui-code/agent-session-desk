@echo off
setlocal
if not exist "%~dp0workspace" mkdir "%~dp0workspace"
"%~dp0asd.exe" --terminal-bundle "%~dp0terminal-runtime" --namespace asd-release-alpha1 --runtime-dir "%~dp0data\runtime" --state-dir "%~dp0data\state" %*
exit /b %errorlevel%
