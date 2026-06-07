param(
    [string]$ApiBaseUrl = "http://localhost:8080/api/v1",
    [string]$DemoInput = "I am study computer science and I have did a project about robot control."
)

$ErrorActionPreference = "Stop"

function Join-ApiPath {
    param([string]$Path)
    return "$($ApiBaseUrl.TrimEnd('/'))/$($Path.TrimStart('/'))"
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

Write-Host "SpeakMate Mock Demo"
Write-Host "API: $ApiBaseUrl"

$healthBaseUrl = $ApiBaseUrl -replace "/api/v1/?$", ""
Invoke-RestMethod -Method Get -Uri "$($healthBaseUrl.TrimEnd('/'))/health" | Out-Null
Write-Host "[OK] backend health"

$session = Invoke-ApiJson -Method Post -Path "/sessions" -Body @{ scenario_id = 1 }
$sessionId = $session.data.session_id
Write-Host "[OK] created interview session: $sessionId"
Write-Host "Opening: $($session.data.opening_message)"

$message = Invoke-ApiJson -Method Post -Path "/sessions/$sessionId/messages" -Body @{ content = $DemoInput }
$userMessageId = $message.data.user_message.id
Write-Host "[OK] sent demo input"
Write-Host "AI: $($message.data.ai_message.content)"
Write-Host "Score: $($message.data.score_summary.total_score)"

$correction = Invoke-ApiJson -Method Get -Path "/messages/$userMessageId/corrections"
Write-Host "[OK] correction: $($correction.data.corrected_text)"

$score = Invoke-ApiJson -Method Get -Path "/sessions/$sessionId/scores"
Write-Host "[OK] current total score: $($score.data.total_score)"

Invoke-ApiJson -Method Post -Path "/sessions/$sessionId/finish" | Out-Null
Write-Host "[OK] finished session"

$report = Invoke-ApiJson -Method Post -Path "/sessions/$sessionId/report"
Write-Host "[OK] generated report: score=$($report.data.total_score)"
Write-Host "Summary: $($report.data.summary)"

$history = Invoke-ApiJson -Method Get -Path "/sessions?page=1&page_size=5"
Write-Host "[OK] history total: $($history.data.total)"
Write-Host "Demo complete. Open frontend and visit /history or /report/$sessionId."
