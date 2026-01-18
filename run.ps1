# Beautiful Mess - Development Runner

Write-Host "--- Initializing Environment ---" -ForegroundColor Cyan
go mod tidy

Write-Host "`n--- Running Unit Tests ---" -ForegroundColor Cyan
go test ./...

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n[!] Tests failed. Aborting launch." -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "`n--- Launching Beautiful Mess ---" -ForegroundColor Green
go run main.go
