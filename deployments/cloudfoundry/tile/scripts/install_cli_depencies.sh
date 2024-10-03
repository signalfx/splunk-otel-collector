# Note: This script installs all CLI tools that are required for local Tanzu Tile related deevelopment
# This script was originally written for MacOS and has not been tested on other platforms.

brew install jq
brew install wget
brew install cloudfoundry/tap/bosh-cli

# Navigate to: https://github.com/cf-platform-eng/tile-generator
# Download the tile and pcf release asssets, these will need to be an old version on MacOS, darwin support was dropped
chmod +x ~/Downloads/tile_darwin-64bit
chmod +x ~/Downloads/pcf_darwin-64bit
mv ~/Downloads/pcf_darwin-64bit /usr/local/bin/pcf
mv ~/Downloads/tile_darwin-64bit /usr/local/bin/tile

brew tap pivotal/hammer https://github.com/pivotal/hammer
brew install hammer
brew install cloudfoundry/tap/cf-cli@8
brew tap pivotal-cf/om https://github.com/pivotal-cf/om
brew install om

brew install rbenv ruby-build
echo 'if which rbenv > /dev/null; then eval "$(rbenv init -)"; fi' >> ~/.bash_profile
source ~/.bash_profile
rbenv install 3.1.3
rbenv global 3.1.3
ruby -v
gem install cf-uaac

gem install license_finder
