$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tab-resources-20261123/run3';$name='AsdTabResources20261123Run3'
$task=Get-ScheduledTask -TaskName $name -ErrorAction Stop;$info=Get-ScheduledTaskInfo -TaskName $name
$state=$null;$alive=$false;$identity=$false
if(Test-Path (Join-Path $root 'runner-state.json')){
 $state=[IO.File]::ReadAllText((Join-Path $root 'runner-state.json'))|ConvertFrom-Json
 $process=Get-Process -Id $state.pid -ErrorAction SilentlyContinue
 if($process){$alive=-not $process.HasExited;$identity=$process.Path -eq $state.path -and $process.StartTime.ToUniversalTime().ToFileTimeUtc().ToString() -eq $state.filetime}
}
$tail=@()
$log=Join-Path $root 'resource-run.log'
if(Test-Path $log){
 try {
  $stream=[IO.FileStream]::new($log,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete)
  $reader=[IO.StreamReader]::new($stream,[Text.Encoding]::UTF8,$true)
  try{$lines=$reader.ReadToEnd().Split([char]10)}finally{$reader.Dispose();$stream.Dispose()}
  for($index=[Math]::Max(0,$lines.Length-4);$index -lt $lines.Length;$index++){$line=$lines[$index].Trim();if($line){$tail+=([string]$line)}}
 } catch {$tail=@('Log observation unavailable')}
}
@{task_state=[string]$task.State;task_result=$info.LastTaskResult;runner_alive=$alive;runner_identity_matches=$identity;runner_state=$state;latest_resource_output=$tail;observed_utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Depth 6
