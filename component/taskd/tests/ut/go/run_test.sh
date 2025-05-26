#!/bin/bash

# Perform test taskd-go
# Copyright(C) Huawei Technologies Co.,Ltd. 2025. All rights reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================
export GO111MODULE="on"
export GONOSUMDB="*"
export PATH=$GOPATH/bin:$PATH

CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/../../..)
OUTPUT_DIR="${TOP_DIR}"/test/ut/go
GO_PKG=${TOP_DIR}/taskd/go
TEMP_DIR=${TOP_DIR}/taskd/go/test
FILE_INPUT='testTaskdGo.txt'
FILE_DETAIL_OUTPUT='api.html'

function unit_test() {
  if ! (go test -mod=mod -gcflags=all=-l -v -race -coverprofile cov.out ${GO_PKG}/... >./$FILE_INPUT); then
    cat ./$FILE_INPUT
    echo '******go test cases error!******'
    exit 1
  else
    echo ${FILE_DETAIL_OUTPUT}
    gocov convert cov.out | gocov-html >${FILE_DETAIL_OUTPUT}
    gotestsum --junitfile unit-tests.xml -- -race -gcflags=all=-l "${GO_PKG}"/...

    total_coverage=$(go tool cover -func=cov.out | grep "total:" | awk '{print $3}'| sed 's/%//')
    # round up
    coverage=$(echo "$total_coverage" | awk '{if ($1 >= 0) print ($1 == int($1)) ? int($1) : int($1) + 1;\
                                          else print ($1 == int($1)) ? int($1) : int($1)}')
    if [[ $coverage -ge 80 ]]; then
      echo "coverage passed: $coverage%"
    else
      echo "coverage failed: $coverage%, it needs to be greater than 80%."
      # exit 1
    fi
  fi
}

function clean_before() {
    if [ -d "$TEMP_DIR" ]; then
      rm -rf $TEMP_DIR
    fi
}

function clean_end() {
    if [ -d "${OUTPUT_DIR}" ]; then
       rm -rf "${OUTPUT_DIR}"
    fi
    mkdir -p "${OUTPUT_DIR}"
    mv "${TEMP_DIR}"/* "${OUTPUT_DIR}"/
    rm -rf "${TEMP_DIR}"
}

function execute_test() {
    echo "************************************* Start LLT Test *************************************"
    clean_before
    mkdir -p "${TEMP_DIR}"
    cd "${TEMP_DIR}"
    unit_test
    echo "<html<body><h1>==================================================</h1><table border="2">" >>./$FILE_DETAIL_OUTPUT
    echo "<html<body><h1>taskd-go testCase</h1><table border="1">" >>./$FILE_DETAIL_OUTPUT
    echo "<html<body><h1>==================================================</h1><table border="2">" >>./$FILE_DETAIL_OUTPUT
    while read line; do
      echo -e "<tr>
       $(echo $line | awk 'BEGIN{FS="|"}''{i=1;while(i<=NF) {print "<td>"$i"</td>";i++}}')
      </tr>" >>$FILE_DETAIL_OUTPUT
    done <$FILE_INPUT
    echo "</table></body></html>" >>./$FILE_DETAIL_OUTPUT
    echo "************************************* End LLT Test *************************************"
    clean_end
    exit 0
}

execute_test