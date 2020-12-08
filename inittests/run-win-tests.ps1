# dot source env variables from local file, otherwise assume build pipeline provides the files
if (Test-Path -Path "./localvars/win-env.ps1")
{
    . "./localvars/win-env.ps1"
}

if (Test-Path -Path "$HOME\.thy.yml")
{
    Move-Item "$HOME\.thy.yml" "$HOME\.thy2.yml"
}

if ((Test-Path -Path winenv) -eq $false)
{
    virtualenv -p python winenv
}

if (-not (Test-Path env:INIT_CLINAME)) {
    $env:INIT_CLINAME = 'dsv'
}

$env:BINARY_PATH = "$env:INIT_CLINAME"
echo "WIN CLI NAME: $env:INIT_CLINAME"

#/cd..
Set-Location ..

# go build variables
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:GO111MODULE = "on"

if ($env:IS_SYSTEM_TEST -eq "true")
{
    echo "building system test"
    go test -c -covermode=count -coverpkg ./... -o inittests\$env:BINARY_PATH
}
else
{
    go build -o inittests\$env:BINARY_PATH
}

Set-Location ./inittests

winenv/Scripts/activate
pip3 install -r requirements.txt

python tests-win.py

deactivate

Remove-Item $env:BINARY_PATH

if (Test-Path -Path "$HOME\.thy2.yml")
{
    Move-Item "$HOME\.thy2.yml" "$HOME\.thy.yml"
}
