@echo off
echo Creating Windows distribution for Board Game Library...

REM Clean previous build
if exist "windows-dist" rmdir /s /q windows-dist

REM Create distribution structure
mkdir windows-dist
mkdir windows-dist\data
mkdir windows-dist\web
mkdir windows-dist\docs

REM Build the executable with Windows optimizations
echo Building executable...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -H=windowsgui" -o windows-dist\board-game-library.exe ./cmd/server

REM Copy required assets
echo Copying assets...
xcopy /E /I web windows-dist\web >nul
xcopy /E /I docs windows-dist\docs >nul

REM Copy configuration files and launchers
copy build\windows\config.windows.env windows-dist\config.env
copy build\windows\run.bat windows-dist\
copy build\windows\run-portable.bat windows-dist\

REM Create README for Windows users
echo Creating README...
(
echo Board Game Library - Windows Distribution
echo =======================================
echo.
echo Quick Start:
echo 1. Choose your preferred mode:
echo    - run.bat: Stores data in user AppData ^(recommended^)
echo    - run-portable.bat: Stores data in local folder ^(portable^)
echo 2. Open your browser to http://localhost:8080
echo 3. No administrator rights required!
echo.
echo Configuration:
echo - Edit config.env to change settings
echo - Database file: data\library.db
echo - Web assets: web\ folder
echo - API documentation: docs\ folder
echo.
echo Manual Start:
echo If you prefer to run manually, use: board-game-library.exe
echo Make sure to set DATABASE_PATH=.\data\library.db environment variable
echo.
echo Support:
echo Check the documentation in the docs folder for API details.
) > windows-dist\README.txt

echo.
echo Windows distribution created successfully in 'windows-dist' folder!
echo.
echo Contents:
echo - board-game-library.exe  ^(main executable^)
echo - run.bat                 ^(standard launcher - uses AppData^)
echo - run-portable.bat        ^(portable launcher - uses local folder^)
echo - config.env              ^(configuration file^)
echo - data\                   ^(database directory for portable mode^)
echo - web\                    ^(web assets^)
echo - docs\                   ^(API documentation^)
echo - README.txt              ^(instructions^)
echo.
echo To distribute: zip the entire 'windows-dist' folder