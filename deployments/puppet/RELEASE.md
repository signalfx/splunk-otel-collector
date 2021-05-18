# Release Process

To release a new version of the module, run `./release` in this directory.  You
will need access to the SignalFx account on the Puppet Forge website, and the
release script will give you instructions for what to do there.

You should update the version in `metadata.json` to whatever is most appropriate
for semver and have that committed before running `./release`.

The release script will try to make and push an annotated tag of the form
`puppet-vX.Y.Z` where `X.Y.Z` is the version in the `./metadata.json` file.
