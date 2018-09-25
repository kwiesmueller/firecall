#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cat <<EOF
REGISTRY ${REGISTRY}
PROJECT ${PROJECT}
NAMESPACE ${NAMESPACE}
STABLE_VERSION ${VERSION}
STABLE_BRANCH ${BRANCH}
STABLE_REVISION ${REVISION}
STABLE_REVSHORT ${REVSHORT}
STABLE_USER ${USER}
STABLE_GOVERSION ${GOVERSION}
EOF