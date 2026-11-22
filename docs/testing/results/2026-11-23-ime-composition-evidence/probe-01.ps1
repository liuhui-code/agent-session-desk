$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/ime-20261123/desktop'
function Action($id,$value){
 $name=('{0:d3}.json' -f $id);$path=Join-Path $root "commands/$name";$temp=Join-Path $root "$name.tmp"
 [IO.File]::WriteAllText($temp,($value|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false));Move-Item $temp $path
 $response=Join-Path $root "responses/$name";$until=[DateTime]::UtcNow.AddSeconds(15)
 while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $response)){throw 'Driver response timeout'}
 $r=[IO.File]::ReadAllText($response)|ConvertFrom-Json
 if(!$r.ok){throw ($r|ConvertTo-Json -Depth 6)}
 return $r
}

$ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20)
while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if(!(Test-Path $ready)){throw 'Driver not ready'}
$r=Action 1 @{kind='state'}
@{keyboardLayout=$r.state.keyboardLayout;capsLock=$r.state.capsLock}|ConvertTo-Json
$null=Action 2 @{kind='capsOff'}
$null=Action 3 @{kind='paste';text="function prompt { 'shell> ' }; Clear-Host"}
$null=Action 4 @{kind='keys';keys='{ENTER}'}
$null=Action 5 @{kind='paste';text="Set-Content ime-confirm.txt '"}
$r=Action 6 @{kind='scan';text='nihao'}
$null=Action 7 @{kind='capture';label='scan-initial-mode'}
@{keyboardLayout=$r.state.keyboardLayout;capsLock=$r.state.capsLock;text=$r.state.text}|ConvertTo-Json -Depth 3
