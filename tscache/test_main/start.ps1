go build -o server.exe
Start-Process -FilePath ".\server.exe" -ArgumentList "-port=8001" -NoNewWindow
Start-Process -FilePath ".\server.exe" -ArgumentList "-port=8002" -NoNewWindow
Start-Process -FilePath ".\server.exe" -ArgumentList "-port=8003", "-api=1" -NoNewWindow

Start-Sleep -Seconds 2

Write-Output ">>> start test"

$job1 = Start-Job -ScriptBlock { Invoke-WebRequest -Uri "http://localhost:9999/api?key=key1" }
$job2 = Start-Job -ScriptBlock { Invoke-WebRequest -Uri "http://localhost:9999/api?key=key1" }
$job3 = Start-Job -ScriptBlock { Invoke-WebRequest -Uri "http://localhost:9999/api?key=key1" }

$job1 | Wait-Job
$job2 | Wait-Job
$job3 | Wait-Job

$result1 = Receive-Job $job1
$result2 = Receive-Job $job2
$result3 = Receive-Job $job3

$result1.Content
$result2.Content
$result3.Content

Remove-Job $job1
Remove-Job $job2
Remove-Job $job3

Wait-Process -Name "server"