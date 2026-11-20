function Connect-CollectorQuay {
    if ($env:DRY_RUN) {
        if (-not $env:QUAY_DEV_USERNAME -or -not $env:QUAY_DEV_PASSWORD) {
            throw 'QUAY_DEV_USERNAME and QUAY_DEV_PASSWORD are required for a Docker dry run'
        }
        $env:QUAY_DEV_PASSWORD | docker login --username $env:QUAY_DEV_USERNAME --password-stdin quay.io
    } else {
        $env:CIRCLECI_QUAY_PASSWORD | docker login --username $env:CIRCLECI_QUAY_USERNAME --password-stdin quay.io
    }
    if ($LASTEXITCODE -ne 0) { throw 'Quay login failed' }
}

function Publish-CollectorWindowsManifest($Repository, $Tag, $OsVersion) {
    $image = "${Repository}:${Tag}"
    $pushOutput = @(docker push $image)
    $exitCode = $LASTEXITCODE
    $pushOutput | Write-Host
    if ($exitCode -ne 0 -or $pushOutput[-1] -notmatch 'digest: (sha256:[a-f0-9]{64})') {
        throw "Could not push $image and read its digest"
    }
    $source = "${Repository}@$($Matches[1])"

    # The manifest must use this push's image, even if another job moves its tag.
    docker manifest rm $image | Out-Host
    docker manifest create $image $source | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "Could not create manifest for $image" }
    docker manifest annotate --os windows --arch amd64 --os-version $OsVersion $image $source | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "Could not annotate manifest for $image" }
    $manifestOutput = @(docker manifest push $image --purge)
    $exitCode = $LASTEXITCODE
    $manifestOutput | Write-Host
    $digest = $manifestOutput[-1]
    if ($exitCode -ne 0 -or $digest -notmatch '^sha256:[a-f0-9]{64}$') {
        throw "Could not push the manifest for $image and read its digest"
    }
    return "$image $digest"
}
