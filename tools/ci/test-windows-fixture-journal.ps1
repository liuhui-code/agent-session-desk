[CmdletBinding()]
param()
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'windows-fixture-journal.ps1')
$path = Join-Path ([IO.Path]::GetTempPath()) ('asd-journal-' + [Guid]::NewGuid().ToString('N') + '.jsonl')
$writer = $null
try {
    if (@(Get-AsdFixtureJournal $path).Count -ne 0) { throw 'Missing journal was not empty.' }
    $writer = [IO.FileStream]::new($path, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::ReadWrite)
    $first = [Text.Encoding]::UTF8.GetBytes("{`"event`":`"started`",`"pid`":42}`n{`"event`":")
    $writer.Write($first, 0, $first.Length)
    $writer.Flush()
    $records = @(Get-AsdFixtureJournal $path)
    if ($records.Count -ne 1 -or $records[0].event -ne 'started' -or $records[0].pid -ne 42) {
        throw 'A live writer or partial final record corrupted the completed record.'
    }
    $rest = [Text.Encoding]::UTF8.GetBytes("`"input`",`"text`":`"中文`"}`n")
    $writer.Write($rest, 0, $rest.Length)
    $writer.Flush()
    $records = @(Get-AsdFixtureJournal $path)
    if ($records.Count -ne 2 -or $records[1].text -cne '中文') { throw 'Appended UTF-8 record was not read exactly.' }
    $invalid = [Text.Encoding]::UTF8.GetBytes("invalid-json`n")
    $writer.Write($invalid, 0, $invalid.Length)
    $writer.Flush()
    $rejected = $false
    try { $null = Get-AsdFixtureJournal $path } catch { $rejected = $true }
    if (-not $rejected) { throw 'Malformed completed record was silently accepted.' }
    'Journal sharing/partial record/UTF-8/corruption checks passed; native Windows sharing is exercised only on Windows.'
} finally {
    if ($null -ne $writer) { $writer.Dispose() }
    if (Test-Path -LiteralPath $path) { Remove-Item -LiteralPath $path }
}
