#!/bin/bash

set -euo pipefail

PAAS_CF_DIR=${PAAS_CF_DIR:-paas-cf}
AUTOSCALER_BOSHRELEASE_DIR=${PAAS_CF_DIR}/manifests/cf-app-autoscaler/upstream
WORKDIR=${WORKDIR:-.}


opsfile_args=""
for i in "${PAAS_CF_DIR}"/manifests/cf-app-autoscaler/operations.d/*.yml; do
  opsfile_args+="-o $i "
done

#if [ "${SLIM_DEV_DEPLOYMENT-}" = "true" ]; then
#  opsfile_args+="-o ${PAAS_CF_DIR}/manifests/prometheus/operations/scale-down-dev.yml "
#  opsfile_args+="-o ${PAAS_CF_DIR}/manifests/prometheus/operations/speed-up-deployment-dev.yml "
#fi



variables_file="$(mktemp)"
trap 'rm -f "${variables_file}"' EXIT

cat <<EOF > "${variables_file}"
---
bosh_url: $BOSH_URL
system_domain: $SYSTEM_DNS_ZONE_NAME
app_domain: $APPS_DNS_ZONE_NAME
skip_ssl_verify: false
aws_account: $AWS_ACCOUNT
bosh_ca_cert: "$BOSH_CA_CERT"
vcap_password: $VCAP_PASSWORD
external_app_autoscaler_database_password: $DATABASE_PASSWORD
EOF

# shellcheck disable=SC2086
bosh interpolate \
  --vars-file="${variables_file}" \
  --vars-file="${WORKDIR}/terraform-outputs/cf.yml" \
  --vars-file="${PAAS_CF_DIR}/manifests/cf-manifest/env-specific/${ENV_SPECIFIC_BOSH_VARS_FILE}" \
  ${opsfile_args} \
  "${AUTOSCALER_BOSHRELEASE_DIR}/cf-app-autoscaler.yml"
