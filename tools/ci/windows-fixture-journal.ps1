# Read completed JSONL records while the controlled fixture still owns a writer.
# File.ReadAllText's default FileShare.Read conflicts with that live writer on
# Windows. Sharing read/write/delete here grants no write access to this reader.
function Get-AsdFixtureJournal {
    param([Parameter(Mandatory = $true)][string]$Path)
    if (-not (Test-Path -LiteralPath $Path)) { return @() }
    $stream = [IO.FileStream]::new($Path, [IO.FileMode]::Open, [IO.FileAccess]::Read,
        [IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete)
    $reader = [IO.StreamReader]::new($stream, [Text.Encoding]::UTF8, $true)
    try { $text = $reader.ReadToEnd() } finally { $reader.Dispose(); $stream.Dispose() }
    $lines = $text -split "`n"
    for ($index = 0; $index -lt $lines.Count - 1; $index++) {
        if ($lines[$index].Trim()) { $lines[$index] | ConvertFrom-Json }
    }
    # A concurrent partial last record is retried by the caller. Malformed
    # completed records must still throw instead of silently disappearing.
}
