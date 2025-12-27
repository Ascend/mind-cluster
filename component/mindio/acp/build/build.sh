#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
set -e

usage() {
    echo "Usage: $0 [ -h | --help ] [ -t | --type <build_type> ] [ -b | --builddir <build_path> ] [ -f | --flags <cmake_flags> ]"
    echo "build_type: [debug, release, asan, tsan]"
    echo "docker: enable docker build"
    echo "cmake_flags: customized flags passed to cmake (these arguments must appear after all other arguments)"
    echo
    echo "Examples:"
    echo " 1 ./build.sh -t debug -d -b ./build/debug"
    echo " 2 ./build.sh -t asan"
    echo " 3 ./build.sh -t release -b ./build-ninja --ninja"
    echo " 4 ./build.sh -t release"
    echo
    exit 1;
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
              CMAKE_FLAGS+='-DOCKIO_ASAN_BUILD=ON'
            elif [[ "$type" == 'tsan' ]]; then
              BUILD_TYPE=Debug
              BUILD_FOLDER=TSAN
              CMAKE_FLAGS+='-DOCKIO_TSAN_BUILD=ON'
            fi
            shift 2
            ;;
        --ut )
            USING_UT=$(echo "$2"|tr a-z A-Z|tr -d "'")
            CMAKE_FLAGS+="-DBUILD_FOR_UT=${USING_UT} "
            shift 2
            ;;
        --tools )
            USING_UT=$(echo "$2"|tr a-z A-Z|tr -d "'")
            CMAKE_FLAGS+="-DBUILD_WITH_TEST_TOOLS=${USING_UT} "
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
PROJ_DIR="$(realpath "${PROJ_DIR}/..")"

COMMIT_ID=$(git log -1 --pretty=format:"%H" $PROJ_DIR) || COMMIT_ID="UNKNOWN"

# 拉取三方代码
cd ${PROJ_DIR}
if [[ ! -d ${PROJ_DIR}/3rdparty/ubs-comm/ubs-comm ]]; then
    echo "Trying to git clone ubs-comm ..."
    cd ${PROJ_DIR}/3rdparty/ubs-comm
    git clone https://atomgit.com/openeuler/ubs-comm.git
    cd ${PROJ_DIR}/3rdparty/ubs-comm/ubs-comm
    git checkout master && git submodule update --init
fi

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
    git clone https://github.com/gabime/spdlog.git
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

# 第三方组件asan适配
if [ "X$BUILD_FOLDER" = "XASAN" ]; then
    # ubs-comm
    sed -i '/endif (${CMAKE_BUILD_TYPE} MATCHES "release")/a\add_compile_options(-fsanitize=address -fno-omit-frame-pointer)' $PROJ_DIR/3rdparty/ubs-comm/CMakeLists.txt
    sed -i '/add_compile_options(-fsanitize=address -fno-omit-frame-pointer)/a\add_link_options(-fsanitize=address)' $PROJ_DIR/3rdparty/ubs-comm/CMakeLists.txt
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
CMAKE_CMD="cmake -DCMAKE_BUILD_TYPE=$BUILD_TYPE $CMAKE_FLAGS $PROJ_DIR -DBUILD_PYTHON_SDK=ON"
if command -v ccache >/dev/null 2>&1; then
    CMAKE_CMD="${CMAKE_CMD} -DCMAKE_CXX_COMPILER_LAUNCHER=ccache"
fi

BUILD_CMD="$BUILD_TOOL -j $((N_CPUS-2))"

cd $BUILD_DIR
rm -rf *

echo $CMAKE_CMD
$CMAKE_CMD || {
    echo "Failed to configure ockio build!"
    exit 1
}
echo
echo "Done configuring ockio build"
echo

echo $BUILD_CMD
$BUILD_CMD || {
    echo "Failed to build ockio"
    exit 1
}
echo
echo "Done building ockio"
echo

if pip3 show wheel;then
    echo "wheel has been installed"
else
    echo "Installing wheel..."
    if pip3 install wheel; then
        echo "wheel installed successfully"
    else
        echo "Failed to install wheel"
        exit 1
    fi
fi

if [ -d "$PROJ_DIR/python_whl/mindio_acp/mindio_acp/lib" ]; then
  rm -rf $PROJ_DIR/python_whl/mindio_acp/mindio_acp/lib
fi
if [ -d "$PROJ_DIR/python_whl/mindio_acp/mindio_acp/bin" ]; then
  rm -rf $PROJ_DIR/python_whl/mindio_acp/mindio_acp/bin
fi
mkdir $PROJ_DIR/python_whl/mindio_acp/mindio_acp/lib
mkdir $PROJ_DIR/python_whl/mindio_acp/mindio_acp/bin

\cp -v $PROJ_DIR/output/lib/_c2python_api.so $PROJ_DIR/python_whl/mindio_acp/mindio_acp/
\cp -v $PROJ_DIR/output/lib/libbdm.so $PROJ_DIR/python_whl/mindio_acp/mindio_acp/lib/
\cp -v $PROJ_DIR/output/bin/ockiod $PROJ_DIR/python_whl/mindio_acp/mindio_acp/bin/
\cp -v $BUILD_DIR/src/sdk/memfs/python_sdk/c2python_api.py $PROJ_DIR/python_whl/mindio_acp/mindio_acp/

sed -i "s/{GIT_COMMIT}/${COMMIT_ID}/g" $PROJ_DIR/python_whl/mindio_acp/mindio_acp/VERSION
sed -i "s/{VERSION}/${build_version}/g" $PROJ_DIR/python_whl/mindio_acp/mindio_acp/VERSION

cd $PROJ_DIR/python_whl/mindio_acp
rm -rf build/
rm -rf dist/
python3 setup.py bdist_wheel --py-limited-api=cp37
\mv -v dist/mindio_acp-*.whl $(echo dist/mindio_acp-*.whl | sed -E 's/(mindio_acp-[^ ]+)-[^ ]+-[^ ]+-([^ ]+)\.whl/\1-py3-none-\2.whl/')
cd $BUILD_DIR

# clean env
git checkout -- $PROJ_DIR/python_whl/mindio_acp/mindio_acp/VERSION || echo ""

if [ "$BUILD_TYPE" != 'Debug' ]; then
  chmod 550 $PROJ_DIR/python_whl/mindio_acp/dist/mindio_acp-*.whl
fi

cd ${PROJ_DIR}
if [[ ! -d $PROJ_DIR/output ]]; then
    mkdir ./output
fi
rm -rf ./output/*.whl
cp ${PROJ_DIR}/python_whl/mindio_acp/dist/*.whl ${PROJ_DIR}/output

echo
echo "Done generating tarball!"
echo
echo Success
