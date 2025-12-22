param(
    [Parameter(Mandatory=$true)]
    [string]$AdminApiKey,

    [Parameter(Mandatory=$true)]
    [string]$FirstAdminEmail,

    [string]$BaseUrl = "https://easi.dfds.cloud"
)

$headers = @{
    "Content-Type" = "application/json"
    "X-Platform-Admin-Key" = $AdminApiKey
}

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/api/v1/platform/tenants" -Method Get -Headers $headers
    Write-Host "Get successfull!" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5
}
catch {
    Write-Host "Failed to get tenants:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
    exit 1
}
