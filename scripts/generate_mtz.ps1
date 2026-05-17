param(
    [string]$Root = (Resolve-Path "$PSScriptRoot\..").Path,
    [string]$BinaryPath = (Join-Path (Resolve-Path "$PSScriptRoot\..").Path "maltego-server.exe"),
    [string]$Output = (Join-Path (Resolve-Path "$PSScriptRoot\..").Path "dist\greynoise-go.mtz")
)

$ErrorActionPreference = "Stop"

$mtzRoot = Join-Path $Root "dist\greynoise-go-mtz"
$repo = Join-Path $mtzRoot "TransformRepositories\Local"
$servers = Join-Path $mtzRoot "Servers"
New-Item -ItemType Directory -Force $repo, $servers | Out-Null

$transforms = @(
    @{ Name = "GreyNoiseCommunityIPLookup";       Display = "GreyNoise Community IP Lookup";       Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseNoiseIPLookupAllDetails"; Display = "GreyNoise Noise IP Lookup All Details"; Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseNoiseIPLookupGetActor";   Display = "GreyNoise Noise IP Lookup Get Actor";   Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseNoiseIPLookupGetCVEs";    Display = "GreyNoise Noise IP Lookup Get CVEs";    Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseNoiseIPLookupGetOrg";     Display = "GreyNoise Noise IP Lookup Get Org";     Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseNoiseIPLookupGetPorts";   Display = "GreyNoise Noise IP Lookup Get Ports";   Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseNoiseIPLookupGetTags";    Display = "GreyNoise Noise IP Lookup Get Tags";    Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseNoiseIPSims";             Display = "GreyNoise Noise IP Similarity";         Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseRIOTIPLookup";            Display = "GreyNoise RIOT IP Lookup";              Input = "maltego.IPv4Address" },
    @{ Name = "GreyNoiseQueryByASN";              Display = "GreyNoise Query By ASN";                Input = "maltego.AS" },
    @{ Name = "GreyNoiseQueryByActor";            Display = "GreyNoise Query By Actor";              Input = "maltego.Person" },
    @{ Name = "GreyNoiseQueryByCVE";              Display = "GreyNoise Query By CVE";                Input = "maltego.CVE" },
    @{ Name = "GreyNoiseQueryByTag";              Display = "GreyNoise Query By Tag";                Input = "maltego.Phrase" }
)

function Escape-Xml([string]$Value) {
    return [System.Security.SecurityElement]::Escape($Value)
}

foreach ($t in $transforms) {
    $name = "greynoise.$($t.Name)"
    $transformPath = Join-Path $repo "$name.transform"
    $settingsPath = Join-Path $repo "$name.transformsettings"
    $command = Escape-Xml $BinaryPath
    $parameters = Escape-Xml "local $($t.Name)"
    $workingDirectory = Escape-Xml $Root

    @"
<MaltegoTransform name="$name" displayName="$($t.Display)" abstract="false" template="false" visibility="public" description="$($t.Display)" author="GreyNoise" requireDisplayInfo="false">
   <TransformAdapter>com.paterva.maltego.transform.protocol.v2api.LocalTransformAdapterV2</TransformAdapter>
   <Properties>
      <Fields>
         <Property name="transform.local.command" type="string" nullable="false" hidden="false" readonly="false" description="The command to execute for this transform" popup="false" abstract="false" visibility="public" auth="false" displayName="Command line">
            <SampleValue></SampleValue>
         </Property>
         <Property name="transform.local.parameters" type="string" nullable="true" hidden="false" readonly="false" description="The parameters to pass to the transform command" popup="false" abstract="false" visibility="public" auth="false" displayName="Command parameters">
            <SampleValue></SampleValue>
         </Property>
         <Property name="transform.local.working-directory" type="string" nullable="true" hidden="false" readonly="false" description="The working directory used when invoking the executable" popup="false" abstract="false" visibility="public" auth="false" displayName="Working directory">
            <DefaultValue>$workingDirectory</DefaultValue>
            <SampleValue></SampleValue>
         </Property>
         <Property name="transform.local.debug" type="boolean" nullable="true" hidden="false" readonly="false" description="When this is set, the transform's text output will be printed to the output window" popup="false" abstract="false" visibility="public" auth="false" displayName="Show debug info">
            <SampleValue>false</SampleValue>
         </Property>
      </Fields>
   </Properties>
   <InputConstraints>
      <Entity type="$($t.Input)" min="1" max="1"/>
   </InputConstraints>
   <OutputEntities/>
   <StealthLevel>0</StealthLevel>
</MaltegoTransform>
"@ | Set-Content -Encoding UTF8 $transformPath

    @"
<TransformSettings enabled="true" disclaimerAccepted="false" showHelp="true" runWithAll="true" favorite="false">
   <Properties>
      <Property name="transform.local.command" type="string" popup="false">$command</Property>
      <Property name="transform.local.parameters" type="string" popup="false">$parameters</Property>
      <Property name="transform.local.working-directory" type="string" popup="false">$workingDirectory</Property>
      <Property name="transform.local.debug" type="boolean" popup="false"/>
   </Properties>
</TransformSettings>
"@ | Set-Content -Encoding UTF8 $settingsPath
}

$serverTransforms = ($transforms | ForEach-Object { "      <Transform name=`"greynoise.$($_.Name)`"/>" }) -join "`r`n"
@"
<MaltegoServer name="GreyNoise Go Local" enabled="true" description="GreyNoise Go local transforms" url="http://localhost">
   <LastSync>2026-05-17 00:00:00.000 UTC</LastSync>
   <Protocol version="0.0"/>
   <Authentication type="none"/>
   <Transforms>
$serverTransforms
   </Transforms>
   <Seeds/>
</MaltegoServer>
"@ | Set-Content -Encoding UTF8 (Join-Path $servers "GreyNoise Go Local.tas")

if (Test-Path $Output) {
    Remove-Item $Output -Force
}
$zipOutput = "$Output.zip"
if (Test-Path $zipOutput) {
    Remove-Item $zipOutput -Force
}
Compress-Archive -Path (Join-Path $mtzRoot "*") -DestinationPath $zipOutput -Force
Move-Item -Force $zipOutput $Output
Write-Host "Wrote $Output"
