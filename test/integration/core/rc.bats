#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "create release-candidate machine" {
    run machine -D create -d $DRIVER --release-candidate $NAME
    echo ${output}
    [ $status -eq 0 ]
}

@test "ensure ReleaseCandidate is set" {
    prerelease=$(machine inspect $NAME --format {{.Driver.ReleaseCandidate}})
    [ "$prerelease" = "true" ]
}

@test "ensure install url is test.docker.com" {
    installurl=$(machine inspect $NAME --format {{.HostOptions.EngineOptions.InstallURL}})
    [ "$installurl" = "https://test.docker.com" ]
}
