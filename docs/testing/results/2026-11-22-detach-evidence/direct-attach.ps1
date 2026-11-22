$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$o=Join-Path $b 'detach-diagnosis'
$s=Get-Content "$o\created.json" -Raw|ConvertFrom-Json
$env:PSMUX_DATA_DIR=Join-Path $b 'detach-runtime/asd-detach-20261122'
$env:PSMUX_CONFIG_FILE=Join-Path $env:PSMUX_DATA_DIR 'psmux-m0.conf'
$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1';$env:PSMUX_INPUT_DEBUG='1'
try {
 [IO.File]::AppendAllText("$o\direct-trials.jsonl",(@{kind='attach_start';utc=[DateTime]::UtcNow.ToString('o');session_id=$s.session_id;run_id=$s.run_id}|ConvertTo-Json -Compress)+"`n")
 & "$b\backend\psmux.exe" -L asd-detach-20261122 -f $env:PSMUX_CONFIG_FILE attach-session -t $s.backend_id
 [IO.File]::AppendAllText("$o\direct-trials.jsonl",(@{kind='attach_return';utc=[DateTime]::UtcNow.ToString('o');exit=$LASTEXITCODE}|ConvertTo-Json -Compress)+"`n")
}finally{foreach($n in @('PSMUX_DATA_DIR','PSMUX_CONFIG_FILE','PSMUX_NO_WARM','PSMUX_ALLOW_NESTING','PSMUX_INPUT_DEBUG')){Remove-Item "Env:$n"}}
