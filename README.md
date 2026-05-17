# GreyNoise Maltego Go

Go port of the GreyNoise Maltego transforms.

## Build

```powershell
go build -o maltego-server.exe ./cmd/server
```

## Maltego Import

Import the ready-made package:

```text
dist/greynoise-go.mtz
```

It contains:

- GreyNoise custom entities from the upstream package
- a local Maltego server entry named `GreyNoise Go Local`
- all 13 GreyNoise transforms
- local transform settings that run `maltego-server.exe local <TransformName>`

Set your API key before running imported local transforms:

```powershell
$env:GREYNOISE_API_KEY="your_key"
```

If your binary path changes, regenerate the MTZ:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\generate_mtz.ps1 `
  -BinaryPath "C:\absolute\path\to\maltego-server.exe" `
  -Root "C:\absolute\path\to\this\repo"
```

## HTTP Server Mode

```powershell
$env:GREYNOISE_API_KEY="your_key"
$env:PORT="8081"
.\maltego-server.exe
```

Transforms are available at:

```text
POST /run/<TransformName>
```

## Local Mode

Maltego local transforms use this internally:

```powershell
.\maltego-server.exe local GreyNoiseCommunityIPLookup 1.2.3.4
```
