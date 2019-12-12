#!/bin/bash

set -eu -o pipefail

PAAS_DIR=${PAAS_DIR:-paas}
WORKDIR=${WORKDIR:-.}

opsfile_args=""
for i in "${PAAS_DIR}"/manifests/cloud-config/operations.d/*.yml; do
  opsfile_args+="-o $i "
done

# shellcheck disable=SC2086
bosh interpolate \
  --var-errs \
  --vars-file="${WORKDIR}/terraform-outputs/vpc.yml" \
  --vars-file="${WORKDIR}/terraform-outputs/bosh.yml" \
  --vars-file="${WORKDIR}/terraform-outputs/cf.yml" \
  --vars-file="${PAAS_DIR}/manifests/variables.yml" \
  ${opsfile_args} \
  "${PAAS_DIR}/manifests/cloud-config/paas-cf-cloud-config.yml"
