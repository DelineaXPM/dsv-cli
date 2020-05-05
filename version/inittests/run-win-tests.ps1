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

#/cd..
Set-Location ..

# go build variables
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:GO111MODULE = "on"

if ($env:IS_SYSTEM_TEST -eq "true")
{
    echo "building system test"
    go test -c -covermode=count -coverpkg ./... -o inittests/thy-win-x64.exe
}
else
{
    go build -o inittests/thy-win-x64.exe
}

Set-Location ./inittests

winenv/Scripts/activate
pip3 install -r requirements.txt

python tests-win.py

deactivate

Remove-Item thy-win-x64.exe

if (Test-Path -Path "$HOME\.thy2.yml")
{
    Move-Item "$HOME\.thy2.yml" "$HOME\.thy.yml"
}
