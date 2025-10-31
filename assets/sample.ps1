# Sample PowerShell Script for Testing PS-IDE-Go

# Display system information
Write-Host "=== System Information ===" -ForegroundColor Green
Write-Host "Hostname: $env:HOSTNAME"
Write-Host "User: $env:USER"
Write-Host "PowerShell Version: $($PSVersionTable.PSVersion)"
Write-Host ""

# List processes
Write-Host "=== Top 5 Processes by CPU ===" -ForegroundColor Cyan
Get-Process | Sort-Object CPU -Descending | Select-Object -First 5 | Format-Table Name, CPU, PM -AutoSize

# List files in current directory
Write-Host "=== Files in Current Directory ===" -ForegroundColor Yellow
Get-ChildItem | Select-Object Name, Length, LastWriteTime | Format-Table -AutoSize

# Simple calculation
Write-Host "=== Simple Math ===" -ForegroundColor Magenta
$a = 10
$b = 20
Write-Host "$a + $b = $($a + $b)"
Write-Host "$a * $b = $($a * $b)"

Write-Host ""
Write-Host "Script completed successfully!" -ForegroundColor Green
