#!/usr/bin/env bash

# Copyright 2022 The Captain Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")

# __PATCH__ if generate more resources, set this field eg."cluster:v1alpha1 cluster:v1alpha2"
GROUPS_WITH_VERSIONS="cluster:v1alpha1"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash ./hack/generate-groups.sh "deepcopy,client,informer,lister" \
  captain/pkg/client captain/apis \
  "${GROUPS_WITH_VERSIONS}" \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/../.." \
  --go-header-file "${SCRIPT_ROOT}"/boilerplate.go.txt
  


# To use your own boilerplate text append:
#   --go-header-file "${SCRIPT_ROOT}"/hack/custom-boilerplate.go.txt