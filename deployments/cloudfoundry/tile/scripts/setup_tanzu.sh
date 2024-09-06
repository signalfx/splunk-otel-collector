# USAGE NOTE: The environment variable TAS_JSON must be set for this script to run successfully.
# Set TAS_JSON to the path of the downloaded hammer file from your Tanzu environment

hammer -t $TAS_JSON cf-login
cf target -o system && cf create-space test-space && cf target -o system -s test-space
hammer -t $TAS_JSON cf-login
cf target -o system -s test-space
git clone https://github.com/cloudfoundry-samples/test-app && cd test-app && cf push && cd .. && rm -rf test-app && cf apps
eval "$(hammer -t $TAS_JSON om)"
UAA_CREDS=$(om credentials -p cf -c .uaa.identity_client_credentials -t json | jq '.password' -r)
#UAA_CREDS=$(om credentials -p cf -c .uaa.ssl_credentials -t json | jq '.password' -r)
TAS_SYS_DOMAIN=$(jq '.sys_domain' -r $TAS_JSON)
uaac target https://uaa.$TAS_SYS_DOMAIN --skip-ssl-validation
uaac token client get identity -s $UAA_CREDS
NOZZLE_SECRET="password"
uaac client add my-v2-nozzle --name my-v2-nozzle --secret $NOZZLE_SECRET --authorized_grant_types client_credentials,refresh_token --authorities logs.admin
echo "signalfx-nozzle client secret: $NOZZLE_SECRET"
