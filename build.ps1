# Baut HanseKaiserReloaded komplett: beide Clients (dist/kaiser, dist/hanse), die Startseite
# (dist/lobby) und den Server (server/hkd.exe). Aufruf in der Projektwurzel: ./build.ps1
# Go und Node müssen im PATH liegen (oder GO-Pfad über $env:GO angeben).
$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot

foreach ($spiel in 'kaiser', 'hanse') {
    Push-Location (Join-Path $root "clients/$spiel")
    try {
        npm run check
        npm run build
    } finally { Pop-Location }
}

# Startseite ist reines HTML/CSS/JS ohne Build-Schritt
$lobby = Join-Path $root 'dist/lobby'
New-Item -ItemType Directory -Force $lobby | Out-Null
Copy-Item (Join-Path $root 'clients/lobby/*') $lobby -Force

Push-Location (Join-Path $root 'server')
try {
    $go = if ($env:GO) { $env:GO } else { 'go' }
    & $go vet ./...
    & $go build -buildvcs=false -o hkd.exe ./cmd/hkd
} finally { Pop-Location }
Write-Host 'Fertig. Start: server/hkd.exe -addr 127.0.0.1:8080 -data data -static dist -state <Ordner>'
