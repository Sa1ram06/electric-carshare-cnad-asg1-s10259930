@echo off
echo Starting user service...
start cmd /k "cd user\server-side && go run server.go"

echo Starting vehicle service...
start cmd /k "cd vehicle\server-side && go run server.go"

echo Starting billing service...
start cmd /k "cd billing\server-side && go run server.go"

echo Starting promotion service...
start cmd /k "cd promotion\server-side && go run server.go"

echo All services are running in separate windows.
pause
