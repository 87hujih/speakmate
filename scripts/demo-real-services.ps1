param(
    [string]$ApiBaseUrl = "http://localhost:8080/api/v1",
    [string]$AudioFile = "",
    [string]$AudioContentType = "audio/ogg",
    [string]$DemoInput = "I am study computer science and I have did a project about robot control."
)

$ErrorActionPreference = "Stop"

function Join-ApiPath {
    param([string]$Path)
    return "$($ApiBaseUrl.TrimEnd('/'))/$($Path.TrimStart('/'))"
}

function Assert-EnvPresent {
    param([string]$Name)
    if ([string]::IsNullOrWhiteSpace([Environment]::GetEnvironmentVariable($Name))) {
        throw "Missing environment variable: $Name"
    }
    Write-Host "[OK] $Name is set"
}

function Invoke-ApiJson {
    param(
        [string]$Method,
        [string]$Path,
        [object]$Body = $null
    )

    $params = @{
        Method = $Method
        Uri = Join-ApiPath $Path
    }
    if ($null -ne $Body) {
        $params.ContentType = "application/json"
        $params.Body = ($Body | ConvertTo-Json -Depth 8)
    }

    return Invoke-RestMethod @params
}

function Invoke-AudioUpload {
    param(
        [int]$SessionId,
        [string]$Path,
        [string]$ContentType
    )

    $fileName = [System.IO.Path]::GetFileName($Path)
    $fileBytes = [System.IO.File]::ReadAllBytes((Resolve-Path $Path))
    $boundary = [System.Guid]::NewGuid().ToString("N")
    $lineBreak = "`r`n"
    $prefix = "--$boundary$lineBreak" +
        "Content-Disposition: form-data; name=`"audio`"; filename=`"$fileName`"$lineBreak" +
        "Content-Type: $ContentType$lineBreak$lineBreak"
    $suffix = "$lineBreak--$boundary--$lineBreak"
    $body = [System.Collections.Generic.List[byte]]::new()
    $body.AddRange([System.Text.Encoding]::UTF8.GetBytes($prefix))
    $body.AddRange($fileBytes)
    $body.AddRange([System.Text.Encoding]::UTF8.GetBytes($suffix))

    return Invoke-RestMethod -Method Post -Uri (Join-ApiPath "/sessions/$SessionId/audio") -ContentType "multipart/form-data; boundary=$boundary" -Body $body.ToArray()
}

Write-Host "SpeakMate real-service smoke"
Write-Host "API: $ApiBaseUrl"

Assert-EnvPresent "LLM_BASE_URL"
Assert-EnvPresent "LLM_API_KEY"
Assert-EnvPresent "LLM_MODEL"

if (([Environment]::GetEnvironmentVariable("REDIS_ENABLED")) -eq "true") {
    Assert-EnvPresent "REDIS_ADDR"
}
if (([Environment]::GetEnvironmentVariable("STORAGE_MODE")) -eq "mysql") {
    Assert-EnvPresent "MYSQL_DSN"
}
if (([Environment]::GetEnvironmentVariable("ASR_USE_MOCK")) -eq "false") {
    Assert-EnvPresent "ASR_PROVIDER"
    if (([Environment]::GetEnvironmentVariable("ASR_PROVIDER")) -eq "tencent") {
        Assert-EnvPresent "TENCENT_ASR_APP_ID"
        Assert-EnvPresent "TENCENT_ASR_SECRET_ID"
        Assert-EnvPresent "TENCENT_ASR_SECRET_KEY"
    }
}

$healthBaseUrl = $ApiBaseUrl -replace "/api/v1/?$", ""
Invoke-RestMethod -Method Get -Uri "$($healthBaseUrl.TrimEnd('/'))/health" | Out-Null
Write-Host "[OK] backend health"

$session = Invoke-ApiJson -Method Post -Path "/sessions" -Body @{ scenario_id = 1 }
$sessionId = $session.data.session_id
Write-Host "[OK] created session: $sessionId"

$message = Invoke-ApiJson -Method Post -Path "/sessions/$sessionId/messages" -Body @{ content = $DemoInput }
Write-Host "[OK] text conversation turn completed"
Write-Host "AI: $($message.data.ai_message.content)"

if (-not [string]::IsNullOrWhiteSpace($AudioFile)) {
    $audio = Invoke-AudioUpload -SessionId $sessionId -Path $AudioFile -ContentType $AudioContentType
    Write-Host "[OK] audio upload completed"
    Write-Host "Transcript: $($audio.data.transcript)"
} else {
    Write-Host "[SKIP] audio upload not checked. Pass -AudioFile answer.ogg to verify ASR."
}

Invoke-ApiJson -Method Post -Path "/sessions/$sessionId/finish" | Out-Null
$report = Invoke-ApiJson -Method Post -Path "/sessions/$sessionId/report"
Write-Host "[OK] report generated: score=$($report.data.total_score)"

$history = Invoke-ApiJson -Method Get -Path "/sessions?page=1&page_size=5"
Write-Host "[OK] history total: $($history.data.total)"
Write-Host "Real-service smoke complete. Verify SSE in the frontend or with: curl -N $ApiBaseUrl/sessions/$sessionId/stream"
