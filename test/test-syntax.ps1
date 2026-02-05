# PowerShell Syntax Highlighting Test Script
# This script demonstrates various PowerShell syntax elements

<#
    Multi-line comment block
    Testing block comments
#>

# Variables
$simpleVar = "Hello World"
$number = 42
$decimal = 3.14159
$hexNumber = 0x1A2B
$binaryNumber = 0b101010

# Special variables
$PSVersionTable
$Error
$_
$true
$false
$null

# Strings
$singleQuote = 'This is a single-quoted string'
$doubleQuote = "This is a double-quoted string with $simpleVar"
$hereString = @"
This is a here-string
that spans multiple lines
and can contain $variables
"@

# Arrays and hashtables
$array = @(1, 2, 3, 4, 5)
$hashtable = @{
    Name = "John Doe"
    Age = 30
    City = "New York"
}

# Functions
function Get-CustomData {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$Name,
        
        [int]$Count = 10
    )
    
    begin {
        Write-Host "Starting function..." -ForegroundColor Green
    }
    
    process {
        for ($i = 0; $i -lt $Count; $i++) {
            Write-Output "$Name - Item $i"
        }
    }
    
    end {
        Write-Host "Function completed!" -ForegroundColor Cyan
    }
}

# Control structures
if ($number -gt 40) {
    Write-Host "Number is greater than 40"
} elseif ($number -eq 40) {
    Write-Host "Number equals 40"
} else {
    Write-Host "Number is less than 40"
}

# Loops
foreach ($item in $array) {
    Write-Output "Item: $item"
}

for ($i = 0; $i -lt 5; $i++) {
    Write-Output "Loop iteration: $i"
}

while ($number -gt 0) {
    $number--
    if ($number -eq 20) {
        break
    }
}

# Switch statement
switch ($number) {
    0 { "Zero" }
    1 { "One" }
    default { "Other number: $_" }
}

# Try-Catch-Finally
try {
    $result = 10 / 0
} catch [System.DivideByZeroException] {
    Write-Error "Cannot divide by zero!"
} finally {
    Write-Host "Cleanup completed"
}

# Cmdlets with common verbs
Get-Process | Where-Object { $_.CPU -gt 100 } | Select-Object Name, CPU
Get-ChildItem -Path "C:\Windows" -Recurse -ErrorAction SilentlyContinue
Set-Location -Path $env:USERPROFILE
New-Item -Path "test.txt" -ItemType File -Force
Remove-Item -Path "test.txt" -Force

# Comparison operators
$a = 10
$b = 20
$result = $a -eq $b
$result = $a -ne $b
$result = $a -lt $b
$result = $a -le $b
$result = $a -gt $b
$result = $a -ge $b

# String operators
$text = "PowerShell"
$result = $text -like "Power*"
$result = $text -match "Shell$"
$result = $text -replace "Shell", "Storm"

# Logical operators
$condition = ($a -gt 5) -and ($b -lt 30)
$condition = ($a -eq 0) -or ($b -ne 0)
$condition = -not ($a -eq $b)

# Type casting
[int]$intValue = "123"
[string]$stringValue = 456
[datetime]$date = "2024-01-01"
[System.Collections.ArrayList]$list = @()

# Pipeline and operators
1..10 | ForEach-Object { $_ * 2 } | Where-Object { $_ -gt 10 }
$data = Get-Content "file.txt" | Sort-Object | Get-Unique

# Splatting
$params = @{
    Path = "C:\Temp"
    Filter = "*.txt"
    Recurse = $true
}
Get-ChildItem @params

# Class definition (PowerShell 5.0+)
class Person {
    [string]$FirstName
    [string]$LastName
    [int]$Age
    
    Person([string]$first, [string]$last) {
        $this.FirstName = $first
        $this.LastName = $last
    }
    
    [string] GetFullName() {
        return "$($this.FirstName) $($this.LastName)"
    }
}

# Create instance
$person = [Person]::new("John", "Doe")
$fullName = $person.GetFullName()

# Advanced function with parameter sets
function Invoke-CustomCommand {
    [CmdletBinding(DefaultParameterSetName='Default')]
    param(
        [Parameter(ParameterSetName='File', Mandatory=$true)]
        [string]$FilePath,
        
        [Parameter(ParameterSetName='Content', Mandatory=$true)]
        [string]$Content,
        
        [Parameter()]
        [switch]$Force
    )
    
    switch ($PSCmdlet.ParameterSetName) {
        'File' { 
            Write-Output "Processing file: $FilePath"
        }
        'Content' {
            Write-Output "Processing content: $Content"
        }
    }
}

# Job management
$job = Start-Job -ScriptBlock {
    Get-Process | Where-Object { $_.Name -like "power*" }
}
$result = Receive-Job -Job $job -Wait
Remove-Job -Job $job

Write-Host "Script execution completed!" -ForegroundColor Green
