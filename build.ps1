<#
.SYNOPSIS
    Compila o goBaDig e monta o pacote de release em dist/.

.DESCRIPTION
    Baixa as dependencias do modulo Go, roda "go vet"/"go test",
    compila o binario e monta dist/goBaDig-vX.Y.Z/ + o zip
    dist/goBaDig-vX.Y.Z.zip, onde X.Y.Z vem direto de
    internal/version/version.go (fonte da verdade da versao atual).

    Isto so faz o build/empacotamento local - NAO cria tag Git nem
    release no GitHub. Esses passos continuam manuais/deliberados,
    como ja vinha sendo feito (ver CHANGELOG.md e docs/releases/).

.PARAMETER SkipDownload
    Pula "go mod download" (util para recompilar offline quando o
    cache de modulos ja esta populado).

.PARAMETER SkipTests
    Pula "go vet"/"go test" antes de compilar.

.EXAMPLE
    ./build.ps1
.EXAMPLE
    ./build.ps1 -SkipTests -SkipDownload
#>
[CmdletBinding()]
param(
    [switch]$SkipDownload,
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"

function Invoke-Checked {
    param(
        [Parameter(Mandatory)][string]$Exe,
        [Parameter(Mandatory)][string[]]$ExeArgs,
        [Parameter(Mandatory)][string]$Description
    )
    Write-Host "==> $Description" -ForegroundColor Cyan
    & $Exe @ExeArgs
    if ($LASTEXITCODE -ne 0) {
        throw "$Description falhou (exit code $LASTEXITCODE)"
    }
}

$repoRoot = $PSScriptRoot
Push-Location $repoRoot
try {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "go nao encontrado no PATH. Instale Go (https://go.dev/dl/) antes de rodar este script."
    }
    if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) {
        throw "gcc nao encontrado no PATH. O pacote GUI (Fyne) precisa de CGO " +
              "habilitado com um compilador C (ex. MinGW-w64 no Windows). " +
              "Ver docs/MANUAL.md, secao 4."
    }

    # 1. Versao atual: ler direto do codigo-fonte, nao do git, para o
    #    pacote sempre bater com o binario que ele mesmo gerou.
    $versionFile = Join-Path $repoRoot "internal\version\version.go"
    $versionSrc = Get-Content -Raw $versionFile

    function Get-ConstValue([string]$Name) {
        $m = [regex]::Match($versionSrc, "$Name\s*=\s*""([^""]+)""")
        if (-not $m.Success) { throw "Nao encontrei a constante $Name em $versionFile" }
        return $m.Groups[1].Value
    }

    $version = Get-ConstValue "Version"
    $block = Get-ConstValue "Block"
    $feature = Get-ConstValue "Feature"
    $tag = "v$version"

    Write-Host "goBaDig $tag (bloco $block, feature $feature)" -ForegroundColor Green

    $existingTag = & git tag --list $tag
    if (-not $existingTag) {
        Write-Warning ("A tag Git '$tag' ainda nao existe neste repositorio. " +
            "O build continua normalmente; crie a tag/release separadamente " +
            "quando for publicar.")
    }

    # 2. Dependencias
    if (-not $SkipDownload) {
        Invoke-Checked -Exe "go" -ExeArgs @("mod", "download") `
            -Description "Baixando dependencias (go mod download)"
    }

    # 3. Testes
    if (-not $SkipTests) {
        Invoke-Checked -Exe "go" -ExeArgs @("vet", "./...") -Description "go vet"
        Invoke-Checked -Exe "go" -ExeArgs @("test", "./...") -Description "go test"
    }

    # 4. Build
    $onWindows = $true
    if (Get-Variable -Name IsWindows -ErrorAction SilentlyContinue) { $onWindows = $IsWindows }
    $binaryName = if ($onWindows) { "goBaDig.exe" } else { "goBaDig" }

    $distDir = Join-Path $repoRoot "dist"
    New-Item -ItemType Directory -Force -Path $distDir | Out-Null
    $binaryPath = Join-Path $distDir $binaryName

    Invoke-Checked -Exe "go" -ExeArgs @("build", "-o", $binaryPath, "./cmd/goBaDig") `
        -Description "Compilando $binaryName"

    # 5. Montar dist/goBaDig-vX.Y.Z/
    $pkgName = "goBaDig-$tag"
    $pkgDir = Join-Path $distDir $pkgName
    $zipPath = Join-Path $distDir "$pkgName.zip"

    if (Test-Path $pkgDir) { Remove-Item -Recurse -Force $pkgDir }
    if (Test-Path $zipPath) { Remove-Item -Force $zipPath }

    New-Item -ItemType Directory -Force -Path (Join-Path $pkgDir "docs\releases") | Out-Null

    Copy-Item $binaryPath $pkgDir
    foreach ($f in @("README.md", "CHANGELOG.md", "LICENSE", "OUTLINE.md")) {
        Copy-Item (Join-Path $repoRoot $f) $pkgDir
    }
    Copy-Item (Join-Path $repoRoot "docs\MANUAL.md") (Join-Path $pkgDir "docs\MANUAL.md")
    Copy-Item (Join-Path $repoRoot "docs\RESOURCES.md") (Join-Path $pkgDir "docs\RESOURCES.md")

    $releaseNotes = Join-Path $repoRoot "docs\releases\$tag.md"
    if (Test-Path $releaseNotes) {
        Copy-Item $releaseNotes (Join-Path $pkgDir "docs\releases\$tag.md")
    }
    else {
        Write-Warning "Nao encontrei docs/releases/$tag.md - pacote gerado sem as notas desta versao."
    }

    Copy-Item (Join-Path $repoRoot "examples") (Join-Path $pkgDir "examples") -Recurse

    # 6. Zip
    Write-Host "==> Gerando $zipPath" -ForegroundColor Cyan
    Compress-Archive -Path $pkgDir -DestinationPath $zipPath -Force

    Write-Host ""
    Write-Host "Conteudo do pacote:" -ForegroundColor Cyan
    Get-ChildItem -Recurse $pkgDir | ForEach-Object {
        Write-Host "  $($_.FullName.Substring($pkgDir.Length + 1))"
    }

    Write-Host ""
    Write-Host "Pronto." -ForegroundColor Green
    Write-Host "Pasta: $pkgDir"
    Write-Host "Zip  : $zipPath"
}
finally {
    Pop-Location
}
