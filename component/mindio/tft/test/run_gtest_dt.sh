#!/bin/bash
# ***********************************************************************
# Copyright: (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
# script for run gtest ut
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

PROJECT_HOME="$( cd "$( dirname "$0" )"/.. && pwd )"
UT_SRC_PATH="${PROJECT_HOME}/test"
BUILD_DIR="${PROJECT_HOME}/test/build"
BIN_DIR="${BUILD_DIR}/bin"
REPORT_DIR="${PROJECT_HOME}/test/build/report"
TEST_EXECUTABLE="test_ttp"

SKIP_FLAG=""
UT_FILTER="*"
ENABLE_COLLECT="1"

print_info() {
    echo "[INFO] $1"
}

print_warn() {
    echo "[WARN] $1"
}

print_error() {
    echo "[ERROR] $1"
}

TARGET_LINE="add_compile_definitions(_GLIBCXX_USE_CXX11_ABI=0)"

# 拉取三方代码
cd ${UT_SRC_PATH}
if [[ ! -d ${UT_SRC_PATH}/3rdparty/googletest ]]; then
    echo "Trying to git clone gtest ..."
    cd ${UT_SRC_PATH}/3rdparty
    git clone https://gitcode.com/GitHub_Trending/go/googletest.git
    cd ${UT_SRC_PATH}/3rdparty/googletest
    git checkout v1.12.0

    # 在 googletest 的 CMakeLists.txt 中添加 ABI 定义
    echo "Adding _GLIBCXX_USE_CXX11_ABI=0 to googletest/CMakeLists.txt"
    sed -i "1i ${TARGET_LINE}" CMakeLists.txt
fi

if [[ ! -d ${UT_SRC_PATH}/3rdparty/mockcpp ]]; then
    echo "Trying to git clone mockcpp ..."
    cd ${UT_SRC_PATH}/3rdparty
    git clone https://gitcode.com/Ascend/mockcpp.git
    cd ${UT_SRC_PATH}/3rdparty/mockcpp
    git checkout v2.7

    # 在 mockcpp 的 CMakeLists.txt 中添加 ABI 定义
    echo "Adding _GLIBCXX_USE_CXX11_ABI=0 to mockcpp/CMakeLists.txt"
    sed -i "1i ${TARGET_LINE}" CMakeLists.txt

    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/include/mockcpp/JmpCode.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/include/mockcpp/mockcpp.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCode.cpp"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCodeArch.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCodeX64.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCodeX86.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpOnlyApiHook.cpp"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/UnixCodeModifier.cpp"
    dos2unix ${UT_SRC_PATH}/3rdparty/patch/*.patch
fi

cd ${PROJECT_HOME}

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

# 清理函数
clean() {
    if [[ "$ENABLE_COLLECT" == "0" ]]; then
        return 0
    fi
    print_info "Cleaning old report directory..."
    rm -rf "${REPORT_DIR}"
}

# 编译函数
build() {
    if [[ "$SKIP_FLAG" == "0" ]]; then
        print_info "Skip the build step."
        return 0
    fi

    print_info "Building main program with build.sh..."
    bash $PROJECT_HOME/build/build.sh -t debug --ut ON
    if [ 0 != $? ];then
        print_error "Failed to build ockio!"
        exit 1
    fi

    print_info "Building with CMake..."
    
    # 创建构建目录
    mkdir -p "${BUILD_DIR}"
    
    # 配置 CMake
    cd "${BUILD_DIR}"
    CMAKE_CMD="cmake ${PROJECT_HOME}/test \
        -DCMAKE_BUILD_TYPE=Debug \
        -DENABLE_UT=ON"
    
    echo "Running CMake command: ${CMAKE_CMD}"
    eval ${CMAKE_CMD}
    
    if [ 0 != $? ]; then
        echo "Failed to configure test project with CMake!"
        exit 1
    fi
    
    # 编译
    N_CPUS=$(nproc)
    JOBS=$((N_CPUS > 2 ? N_CPUS-2 : 1))
    make -j${JOBS}
    
    # 安装/复制可执行文件
    print_info "Installing test executable..."
    mkdir -p "${BIN_DIR}"
    find "${BUILD_DIR}" -name "${TEST_EXECUTABLE}" -type f -executable | while read -r exe; do
        cp "${exe}" "${BIN_DIR}/"
        print_info "Copied ${exe} to ${BIN_DIR}/"
    done
}

# 运行测试
run_tests() {
    print_info "Running tests..."
    
    # 创建报告目录
    mkdir -p "${REPORT_DIR}"
    
    # 检查测试可执行文件
    if [ ! -f "${BIN_DIR}/${TEST_EXECUTABLE}" ]; then
        print_error "Test executable not found: ${BIN_DIR}/${TEST_EXECUTABLE}"
        exit 1
    fi
    
    # 设置库路径
    export LD_LIBRARY_PATH="${PROJECT_HOME}/output/lib:${LD_LIBRARY_PATH}"
    
    # 运行测试
    cd "${BIN_DIR}"
    print_info "Running: ./${TEST_EXECUTABLE} --gtest_output=xml:${REPORT_DIR}/report.xml --gtest_filter=${UT_FILTER}"
    
    if ! ./${TEST_EXECUTABLE} --gtest_output=xml:"${REPORT_DIR}/report.xml" --gtest_filter="${UT_FILTER}"; then
        print_warn "Some tests failed. Check the report for details."
    else
        print_info "All tests passed!"
    fi
}

# 生成覆盖率报告
generate_coverage_report() {
    if [[ "$ENABLE_COLLECT" == "0" ]]; then
        print_info "Skip coverage collection."
        return 0
    fi

    print_info "Generating coverage report..."
    
    # 检查是否安装了 lcov
    if ! command -v lcov &> /dev/null; then
        print_error "lcov not found. Please install it: sudo dnf install lcov"
        exit 1
    fi
    
    cd "${BUILD_DIR}"
    
    # 创建覆盖率报告目录
    mkdir -p gcover_report
    
    # 收集覆盖率数据
    print_info "Collecting coverage data..."
    lcov --directory "${PROJECT_HOME}" \
         --capture \
         --output-file test.info \
         --rc lcov_branch_coverage=1 \
         --rc lcov_excl_br_line="LCOV_EXCL_BR_LINE|TTP_.*|LOG_.*|ASSERT_.*|std::to_string\(.*|std::string\(.*|std::cout.*|std::cerr.*|static.*=.*\{|.*[Ss]slHelper.*|.*SslCtx.*|fclose.*|spdlog::.*|map\..*|errMsg\s=.*|permMsg\s=.*|.*\?.*:.*|.*&&.*\s*\|\|.*|.*&&.*\s*&&.*|if\s*\(.*(Obj|tmp|this).*!=.*\)"
    
    # 过滤不需要的文件
    print_info "Filtering coverage data..."
    lcov --remove test.info \
         "*/build/*" \
         "*7.3.0*" \
         "*/3rdparty/*" \
         "*/test/llt/*" \
         "*/src/csrc/acc_links/common/*" \
         "*/src/csrc/acc_links/security/*" \
         "*/src/csrc/acc_links/under_api/openssl/*" \
         "*generated.h" \
         --output-file coverage.info \
         --rc lcov_branch_coverage=1 \
         --rc lcov_excl_br_line="LCOV_EXCL_BR_LINE|TTP_.*|LOG_.*|ASSERT_.*|std::to_string\(.*|std::string\(.*|std::cout.*|std::cerr.*|static.*=.*\{|.*[Ss]slHelper.*|.*SslCtx.*|fclose.*|spdlog::.*|map\..*|errMsg\s=.*|permMsg\s=.*|.*\?.*:.*|.*&&.*\s*\|\|.*|.*&&.*\s*&&.*|if\s*\(.*(Obj|tmp|this).*!=.*\)"
    
    # 生成 HTML 报告
    print_info "Generating HTML report..."
    genhtml coverage.info \
            --output-directory gcover_report \
            --show-details \
            --legend \
            --rc lcov_branch_coverage=1
    
    print_info "Coverage report generated at: ${BUILD_DIR}/gcover_report/index.html"
}

CURRENT_PATH=$(cd "$(dirname "$0")"; pwd)
cd "${CURRENT_PATH:?}"

# 主函数
main() {
    print_info "Starting GTest UT for test_ttp"
    print_info "Project home: ${PROJECT_HOME}"
    
    # 执行步骤
    clean
    build
    run_tests
    generate_coverage_report
    
    print_info "All tasks completed successfully!"
}

main