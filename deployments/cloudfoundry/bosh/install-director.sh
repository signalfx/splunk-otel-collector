BOSH_DIRECTOR_DIR="./bosh-env/virtualbox"
BOSH_DIRECTOR_DEPLOYMENT_DIR="bosh-deployment"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -d $BOSH_DIRECTOR_DIR/$BOSH_DIRECTOR_DEPLOYMENT_DIR ]; then
  echo "Bosh director already exists"
  echo "If you want to build again, run make reinstall-director"
  exit 0
fi

mkdir -p $BOSH_DIRECTOR_DIR
cd $BOSH_DIRECTOR_DIR

git clone https://github.com/cloudfoundry/bosh-deployment.git

./$BOSH_DIRECTOR_DEPLOYMENT_DIR/virtualbox/create-env.sh

source .envrc
bosh -e vbox env

bosh -e vbox update-cloud-config bosh-deployment/warden/cloud-config.yml
bosh upload-stemcell --sha1 d44dc2d1b3f8415b41160ad4f82bc9d30b8dfdce \
https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-bionic-go_agent?v=1.71

cd $SCRIPT_DIR