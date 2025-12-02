#!/usr/bin/env bash
# ***********************************************************************
# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
# script for Open Computing Kit DFS to run unit test
# version: 1.0.0
# change log:
# ***********************************************************************
set -e

usage() {
    echo "Usage: $0 [ -h | --help ] [ -s | --skip ] [ -n | --no_collect ] [ -f | --filter ]"
    echo
    echo "Examples:"
    echo " 1. bash run_ut.sh -h"
    echo " 2. bash run_ut.sh -s"
    echo " 3. bash run_ut.sh -f TestBackupFileManager.*"
    echo " 4. bash run_ut.sh -s -f TestBackupFileManager.Initialize"
    echo
    exit 1;
}

PROJECT_HOME="$( cd "$( dirname "$0" )"/.. && pwd  )"
UT_EXE_PATH=${PROJECT_HOME}/output/bin
UT_CONF_PATH=${PROJECT_HOME}/output/conf
GENERATE_DIR=${PROJECT_HOME}/Build/Debug/cov/gen

SKIP_FLAG=""
UT_FILTER="*"
ENABLE_COLLECT="1"

# Parse the argument params
while true; do
    case "$1" in
        -s | --skip )
            SKIP_FLAG="0"
            shift ;;
        -n | --no_collect )
            ENABLE_COLLECT="0"
            shift ;;
        -f | --filter )
            UT_FILTER=$2
            shift 2
            ;;
        -h | --help )
            usage
            exit 0
            ;;
        * )
            break;;
    esac
done

if [[ "$SKIP_FLAG" == "" ]]; then
    bash $PROJECT_HOME/build/build.sh -t debug --ut ON
    if [ 0 != $? ];then
        echo "Failed to build ockio!"
        exit 1
    fi
fi

cd "${PROJECT_HOME}/test"
hdt comp -i
if [ 0 != $? ];then
    echo "Failed to install components!"
    exit 1
fi

hdt build
if [ 0 != $? ];then
    echo "Failed to build HDT!"
    exit 1
fi
cd -

chmod 550 -R ${UT_EXE_PATH}
find ${PROJECT_HOME}/Build/Debug -type f -name "*.gcda" | xargs rm -rf

[ -d "${PROJECT_HOME}/Build/Debug/res_xml" ] && rm -rf "${PROJECT_HOME}/Build/Debug/res_xml"
mkdir -p "${PROJECT_HOME}/Build/Debug/res_xml"

mkdir -p ${UT_CONF_PATH}
\cp -f ../configs/memfs.conf ${UT_CONF_PATH}
sed -i -e 's/data_block_pool_capacity_in_gb.*/data_block_pool_capacity_in_gb = 8/' ${UT_CONF_PATH}/memfs.conf

export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PROJECT_HOME}/output/lib
${PROJECT_HOME}/test/build/dfs_hdt --gtest_filter="${UT_FILTER}" --gtest_output=xml:${PROJECT_HOME}/Build/Debug/res_xml/

[[ $ENABLE_COLLECT == "0" ]] && exit 0

# coverage
rm -rf ${PROJECT_HOME}/Build/Debug/cov/; mkdir -p ${GENERATE_DIR}
lcov --d ${PROJECT_HOME}/Build/Debug/src --d ${PROJECT_HOME}/test/build \
    --c --output-file ${GENERATE_DIR}/coverage.info \
    --rc lcov_branch_coverage=1 \
    --rc lcov_excl_br_line="LCOV_EXCL_BR_LINE|MFS_LOG*|ASSERT_*|BKG_*|gLastErrorMessage"
if [ 0 != $? ];then
    echo "Failed to generate all coverage info"
    exit 1
fi

# filter
lcov -r ${GENERATE_DIR}/coverage.info \
    "/usr/*" "/opt/buildtools/*" "*/test/*" "*gtest*" "*mockcpp*" \
    "*/src/util/*" "*/src/sdk/memfs/python_sdk/*" "*/src/sdk/memfs/sdk/*" \
     "*/src/sdk/memfs/server/*" "*/src/sdk/memfs/common/ipc*" \
    "*/3rdparty/*" "*_generated.h" "*/src/memfs/common/memfs_*" \
    -o ${GENERATE_DIR}/coverage.info --rc lcov_branch_coverage=1

# generate
genhtml -o ${GENERATE_DIR}/result ${GENERATE_DIR}/coverage.info --show-details --legend --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to generate all coverage info with html format"
  exit 1
fi

cd ${PROJECT_HOME}/Build/Debug
echo '<?xml version="1.0" encoding="UTF-8"?>' > test_detail.xml

tests_val=$(cat res_xml/* |grep "<testsuites "|awk -F "tests=" '{print $2}'|awk '{print $1}'|awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
failures_val=$(cat res_xml/* |grep "<testsuites "|awk -F "failures=" '{print $2}'|awk '{print $1}'|awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
disabled_val=$(cat res_xml/* |grep "<testsuites "|awk -F "disabled=" '{print $2}'|awk '{print $1}'|awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
errors_val=$(cat res_xml/* |grep "<testsuites "|awk -F "errors=" '{print $2}'|awk '{print $1}'|awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
time_val=$(cat res_xml/* |grep "<testsuites "|awk -F "time=" '{print $2}'|awk '{print $1}'|awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
timestamp_val=$(cat res_xml/* |grep "<testsuites "| head -n 1|awk -F "timestamp=" '{print $2}'|awk '{print $1}'|awk -F "\"" '{print $2}')

echo "<testsuites tests=\"${tests_val}\" failures=\"${failures_val}\" disabled=\"${disabled_val}\" errors=\"${errors_val}\" time=\"${time_val}\" timestamp=\"${timestamp_val}\" name=\"AllTests\">" >> test_detail.xml

cat res_xml/* | grep -v testsuites |grep -v "xml version" >> test_detail.xml

echo '</testsuites>' >> test_detail.xml
cp -rvf test_detail.xml ${GENERATE_DIR}/result/

rm -rf /usr1/nginx/odfs/result/
cp -r ${GENERATE_DIR}/result/ /usr1/nginx/odfs/

if [ ! -e ${PROJECT_HOME}/build ]; then
    ln -s ${PROJECT_HOME}/Build/Debug ${PROJECT_HOME}/build
fi