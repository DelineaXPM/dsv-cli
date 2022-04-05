#!/bin/bash

rm -rf ~/go/src/*
cp -rf $WorkingDirectory/.  ~/go/src/$constants_cliname

rm -rf ~/.thy

cd ~/go/src/$constants_cliname

rm -rf ./coverage
mkdir ./coverage

echo "-------------Running Tests---------------------"
rm -f test.out
go test -v -p 1 -covermode=count -coverprofile=./coverage.out ./...  | tee test.out

rm -f cli-config/.thy.yml

echo "-------------Generating Report----------------"
cat test.out | ~/go/bin/go-junit-report -set-exit-code >  $WorkingDirectory/report.xml


echo "-------------Run Init Tests for CLI--------"
export IS_SYSTEM_TEST=true
cd inittests
rm init_test_output.txt
chmod +x run-tests.sh
./run-tests.sh

cat init_test_output.txt
rm  $WorkingDirectory/TEST-Suite-*

mv test-reports/* $WorkingDirectory

cd ..

echo "merging coverage results"
~/go/bin/gocovmerge -dir coverage -pattern "\.out" > ./coverage_integration.out
~/go/bin/gocovmerge coverage_integration.out coverage.out > ./all.out

cp ./all.out $ArtifactStagingDirectory/

rm -rf $WorkingDirectory/cobertura
mkdir $WorkingDirectory/cobertura


echo "-------------Generating Code Coverage--------"
~/go/bin/gocov convert all.out | ~/go/bin/gocov-xml > $WorkingDirectory/cobertura/codecoverage.xml

