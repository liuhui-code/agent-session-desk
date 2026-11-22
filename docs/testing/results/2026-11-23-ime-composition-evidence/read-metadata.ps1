$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'ime-20261123'
$languages=Get-WinUserLanguageList
$rows=@();foreach($l in $languages){$rows+=,@{tag=$l.LanguageTag;tips=@($l.InputMethodTips)}}
$engines=@();foreach($name in @('ChsIME','TextInputHost','InputApp')){foreach($p in @(Get-Process -Name $name -ErrorAction SilentlyContinue)){if($p.Path){$f=Get-Item $p.Path;$engines+=,@{name=$p.ProcessName;pid=$p.Id;session=$p.SessionId;path=$p.Path;file_version=$f.VersionInfo.FileVersion;description=$f.VersionInfo.FileDescription}}}}
@{language_list_type=$languages.GetType().FullName;languages=$rows;engines=$engines}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'ime-metadata.json') -Encoding utf8
Get-Content (Join-Path $root 'ime-metadata.json') -Raw -Encoding utf8
