#!/bin/bash
# Copyright: (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
set -e

usage() {
    echo "Usage: $0 [ -h | --help ] [ -t | --type <build_type> ] [ -b | --builddir <build_path> ] [ -f | --flags <cmake_flags> ]"
    echo "build_type: [debug, release, asan, tsan]"
    echo "cmake_flags: customized flags passed to cmake (these arguments must appear after all other arguments)"
    echo
    echo "Examples:"
    echo " 1 ./build.sh -t debug -b ./build/debug"
    echo " 2 ./build.sh -t asan"
    echo " 3 ./build.sh -t release -b ./build -f <cmake_flags>"
    echo " 4 ./build.sh -t release"
}

BUILD_DIR=""
BUILD_FOLDER=Release
BUILD_TYPE=Release
BUILD_TOOL=make
CMAKE_FLAGS=""

# Parse the argument params
while true; do
    case "$1" in
        -b | --builddir )
            if [[ ! -d "$2" ]]; then
                echo $2 does not exist!
                exit 1
            fi
            BUILD_DIR=$(realpath $2)
            shift 2
            ;;
        -t | --type )
            type=$2
            type=${type,,}
            [[ "$type" != "debug" && $type != "release" && $type != "asan" && $type != "tsan" ]] && echo "Invalid build type $2" && usage
            if [[ "$type" == 'debug' ]]; then
              BUILD_TYPE=Debug
              BUILD_FOLDER=Debug
            elif [[ "$type" == 'release' ]]; then
              BUILD_TYPE=Release
              BUILD_FOLDER=Release
            elif [[ "$type" == 'asan' ]]; then
              BUILD_TYPE=Debug
              BUILD_FOLDER=ASAN
              CMAKE_FLAGS+="-DTTP_ASAN_BUILD=ON "
            elif [[ "$type" == 'tsan' ]]; then
              BUILD_TYPE=Debug
              BUILD_FOLDER=TSAN
              CMAKE_FLAGS+="-DTTP_TSAN_BUILD=ON "
            fi
            shift 2
            ;;
        --ut )
            USING_UT=$(echo "$2"|tr a-z A-Z|tr -d "'")
            CMAKE_FLAGS+="-DBUILD_TESTS=${USING_UT} "
            shift 2
            ;;
        --dtfuzz )
            CMAKE_FLAGS+="-DBUILD_FOR_FUZZ=ON "
            shift ;;
        -f | --flags )
            while [[ "$2" ]]; do
                CMAKE_FLAGS+="$2 "
                shift
            done
            ;;
        -h | --help )
            usage
            exit 0
            ;;
        * )
            break;;
    esac
done

# Retrieve project top directory
PROJ_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROJ_DIR="$(realpath "${PROJ_DIR}"/..)"

# 拉取三方代码
cd ${PROJ_DIR}

if [[ ! -d ${PROJ_DIR}/3rdparty/libboundscheck/libboundscheck ]]; then
    echo "Trying to git clone libboundscheck ..."
    cd ${PROJ_DIR}/3rdparty/libboundscheck
    git clone https://gitee.com/openeuler/libboundscheck.git
    cd ${PROJ_DIR}/3rdparty/libboundscheck/libboundscheck
    git checkout v1.1.16
fi

if [[ ! -d ${PROJ_DIR}/3rdparty/spdlog/spdlog ]]; then
    echo "Trying to git clone spdlog ..."
    cd ${PROJ_DIR}/3rdparty/spdlog
    git clone https://gitcode.com/GitHub_Trending/sp/spdlog.git
    cd ${PROJ_DIR}/3rdparty/spdlog/spdlog
    git checkout v1.12.0
fi

cd ${PROJ_DIR}

VER_FILE="${PROJ_DIR}"/service_config.ini
build_version="7.3.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  temp=${line#*=}
  build_version="${temp//.SPC/+SPC}"
  if [[ $build_version == *.T* ]]; then
    build_version="${build_version//.T/+t}"
  fi
  echo "build version in service_config.ini:  ${build_version}"
fi
export BUILD_VERSION=${build_version}
echo "build version is ${BUILD_VERSION}"

if [ -z "$BUILD_DIR" ]; then
  BUILD_DIR=$PROJ_DIR/Build/$BUILD_FOLDER
fi

# Setup the build directory
if [[ ! -d "$BUILD_DIR" ]]
then
  mkdir -p $BUILD_DIR
fi

if echo "$CMAKE_FLAGS" | grep -q "BUILD_FOR_FUZZ"; then
  git clone --recurse-submodules --branch v2.4.8 secodefuzz.git 3rdparty/fuzz/secodefuzz
fi

# Verify the build directory is in place and enter it
cd $BUILD_DIR || {
  echo "Fatal! Cannot enter $BUILD_DIR."
  exit 1
}

# Check number of physical processors for parallel make
N_CPUS=$(grep processor /proc/cpuinfo | wc -l)
echo "$N_CPUS processors detected."

# Now do the build job
CMAKE_CMD="cmake -DCMAKE_BUILD_TYPE=$BUILD_TYPE -DCI_BUILD=$CI_BUILD $CMAKE_FLAGS $PROJ_DIR"
if command -v ccache >/dev/null 2>&1; then
    CMAKE_CMD="${CMAKE_CMD} -DCMAKE_CXX_COMPILER_LAUNCHER=ccache"
fi

BUILD_CMD="$BUILD_TOOL -j $((N_CPUS-2))"

echo "PROJ_DIR=${PROJ_DIR}"

cd $BUILD_DIR
rm -rf *

echo $CMAKE_CMD
$CMAKE_CMD || {
    echo "Failed to configure mindio_ttp build!"
    exit 1
}
echo
echo "Done configuring mindio_ttp build"
echo
echo $BUILD_CMD
$BUILD_CMD || {
    echo "Failed to build mindio_ttp"
    exit 1
}
echo
echo "Done building mindio_ttp"
echo

if pip3 show wheel; then
    echo "wheel has been installed"
else
    pip3 install wheel
fi

GIT_COMMIT=`git log -1 --pretty=format:"%H" $PROJ_DIR` || true
{
  echo "mindio_ttp version info:"
  echo "mindio_ttp version: ${BUILD_VERSION}"
  echo "git: ${GIT_COMMIT}"
} > VERSION

echo "|=========================Begin to build mindio_ttp whl=========================|"
mkdir -p $PROJ_DIR/python_whl/mindio_ttp/
mkdir -p $PROJ_DIR/python_whl/mindio_ttp/framework_ttp
mkdir -p $PROJ_DIR/python_whl/mindio_ttp/controller_ttp
mkdir -p $PROJ_DIR/python_whl/mindio_ttp/utils
mkdir -p $PROJ_DIR/python_whl/mindio_ttp/mindspore_api
\cp -v $PROJ_DIR/output/lib/libttp_c_api.so $PROJ_DIR/python_whl/mindio_ttp/mindspore_api/
\mv -v $PROJ_DIR/output/lib/_ttp_c2python_api.so $PROJ_DIR/python_whl/mindio_ttp/controller_ttp/
\cp -v $PROJ_DIR/output/lib/libttp_framework.so $PROJ_DIR/python_whl/mindio_ttp/framework_ttp/
\cp -v $PROJ_DIR/src/python/setup.py $PROJ_DIR/python_whl/
\cp -v $PROJ_DIR/src/python/__init__.py $PROJ_DIR/python_whl/mindio_ttp/
\cp -v $PROJ_DIR/src/python/framework_ttp/* $PROJ_DIR/python_whl/mindio_ttp/framework_ttp/
\cp -v $PROJ_DIR/src/python/controller_ttp/ttp_c2python_api.py $PROJ_DIR/python_whl/mindio_ttp/controller_ttp/ttp_c2python_api.py
\cp -v $PROJ_DIR/src/python/controller_ttp/* $PROJ_DIR/python_whl/mindio_ttp/controller_ttp/
\cp -v $PROJ_DIR/src/python/utils/* $PROJ_DIR/python_whl/mindio_ttp/utils/
\cp -v $BUILD_DIR/VERSION $PROJ_DIR/python_whl/mindio_ttp
cd $PROJ_DIR/python_whl/
python3 setup.py bdist_wheel
\mv -v dist/mindio_ttp-*.whl $(echo dist/mindio_ttp-*.whl | sed -E 's/(mindio_ttp-[^ ]+)-[^ ]+-[^ ]+-([^ ]+)\.whl/\1-py3-none-\2.whl/')
\mv -v mindio_ttp/controller_ttp/_ttp_c2python_api.so $PROJ_DIR
\mv -v dist/mindio_ttp-*.whl $PROJ_DIR/output/
echo "|==========================End building mindio_ttp whl==========================|"
echo

cd $PROJ_DIR
\mv -v _ttp_c2python_api.so output/lib/
echo
echo "mindio_ttp*.whl build success"
echo

if [ $BUILD_TYPE != 'Debug' ]; then
  chmod -R 550 $PROJ_DIR/python_whl/dist/
fi