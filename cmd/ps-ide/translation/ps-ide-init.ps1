# PS-IDE Initialization Script
# This script is loaded at startup to enable ANSI colors

# Enable ANSI rendering
$PSStyle.OutputRendering = [System.Management.Automation.OutputRendering]::Ansi

# Set UTF-8 encoding
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# Override Write-Host to use ANSI colors
function global:Write-Host {
    param(
        [Parameter(ValueFromPipeline=$true)]
        $Object,
        $ForegroundColor,
        $BackgroundColor,
        [switch]$NoNewline,
        $Separator = ' '
    )
    
    process {
        $output = if ($Object -ne $null) { $Object.ToString() } else { '' }
        
        if ($ForegroundColor) {
            $ansiCode = switch ($ForegroundColor) {
                'Black'       { "`e[30m" }
                'DarkRed'     { "`e[31m" }
                'DarkGreen'   { "`e[32m" }
                'DarkYellow'  { "`e[33m" }
                'DarkBlue'    { "`e[34m" }
                'DarkMagenta' { "`e[35m" }
                'DarkCyan'    { "`e[36m" }
                'Gray'        { "`e[37m" }
                'DarkGray'    { "`e[90m" }
                'Red'         { "`e[91m" }
                'Green'       { "`e[92m" }
                'Yellow'      { "`e[93m" }
                'Blue'        { "`e[94m" }
                'Magenta'     { "`e[95m" }
                'Cyan'        { "`e[96m" }
                'White'       { "`e[97m" }
                default       { '' }
            }
            $output = "$ansiCode$output`e[0m"
        }
        
        if ($NoNewline) {
            [Console]::Write($output)
        } else {
            [Console]::WriteLine($output)
        }
    }
}
