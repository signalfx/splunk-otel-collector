BOSH_DIRECTOR_DIR="./bosh-env/virtualbox"
BOSH_DIRECTOR_DEPLOYMENT_DIR="bosh-deployment"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -d $BOSH_DIRECTOR_DIR ] && [ -d $BOSH_DIRECTOR_DIR/$BOSH_DIRECTOR_DEPLOYMENT_DIR ]; then
  cd $BOSH_DIRECTOR_DIR
  ./$BOSH_DIRECTOR_DEPLOYMENT_DIR/virtualbox/delete-env.sh
  cd $SCRIPT_DIR
  rm -rf $BOSH_DIRECTOR_DIR
fi