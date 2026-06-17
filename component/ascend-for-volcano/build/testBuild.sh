#!/bin/bash
# Perform  test volcano-huawei-npu-scheduler plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2022. All rights reserved
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

export GO111MODULE=on
export GONOSUMDB="*"
export PATH=$GOPATH/bin:$PATH

VOLCANO_PLUGIN_PKG="${GOPATH}"/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/...

function filter_cov_by_tested_pkgs() {
  local tested_pkgs
  tested_pkgs=$(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' "${VOLCANO_PLUGIN_PKG}")
  awk -v pkgs="$tested_pkgs" '
    NR==1 {print; next}
    {
      file=$1; sub(/:[0-9].*/, "", file)
      n=split(file, p, "/"); pkg=""
      for(i=1;i<n;i++) pkg=pkg p[i]"/"; sub(/\/$/,"", pkg)
      found=0; split(pkgs, arr, "\n"); for(k in arr) {if(arr[k]==pkg){found=1;break}}
      if (found) print
    }
  ' cov.out > cov_filtered.out
}

cd "${GOPATH}"/src/volcano.sh/volcano
go get github.com/agiledragon/gomonkey/v2@v2.8.0
go get github.com/smartystreets/goconvey@v1.6.4
go mod vendor

file_detail_output='api.html'

echo "************************************* Start LLT Test *************************************"
mkdir -p "${GOPATH}"/src/volcano.sh/volcano/_output/test/
cd "${GOPATH}"/src/volcano.sh/volcano/_output/test/
rm -f $file_detail_output

gotestsum --junitfile unit-tests.xml --jsonfile test.jsonl \
  -- -count=1 -v -gcflags=all=-l -coverprofile cov.out "${VOLCANO_PLUGIN_PKG}";

filter_cov_by_tested_pkgs

gocov convert cov_filtered.out | gocov-html >"$file_detail_output"
total_coverage=$(go tool cover -func=cov_filtered.out | grep "total:" | awk '{print $3}'| sed 's/%//')
# round up
coverage=$(echo "$total_coverage" | awk '{if ($1 >= 0) print ($1 == int($1)) ? int($1) : int($1) + 1;\
                                      else print ($1 == int($1)) ? int($1) : int($1)}')

if [[ $coverage -ge 80 ]]; then
  echo "coverage passed: $coverage%"
else
  echo "coverage failed: $coverage%, it needs to be greater than 80%."
fi

echo "************************************* End   LLT Test *************************************"
exit 0
