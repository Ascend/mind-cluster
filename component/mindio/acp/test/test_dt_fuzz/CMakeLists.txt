include(../conf/comp.cmake)

set(DFS_PROJECT_PATH "${CMAKE_CURRENT_LIST_DIR}/../..")
set(DFS_PROJECT_OUTPUT "${DFS_PROJECT_PATH}/output")
set(DFS_PROJECT_3RD_BINARY "${DFS_PROJECT_OUTPUT}/3rdparty")

INCLUDE_DIRECTORIES(
        ${CMAKE_CURRENT_LIST_DIR}/common
        ${DFS_PROJECT_PATH}/src/include
        ${DFS_PROJECT_PATH}/src/background
        ${DFS_PROJECT_PATH}/src/background/backup
        ${DFS_PROJECT_PATH}/src/common
        ${DFS_PROJECT_PATH}/src/memfs
        ${DFS_PROJECT_PATH}/src/memfs/bmm
        ${DFS_PROJECT_PATH}/src/memfs/common
        ${DFS_PROJECT_PATH}/src/memfs/fs
        ${DFS_PROJECT_PATH}/src/memfs/fuse
        ${DFS_PROJECT_PATH}/src/sdk/memfs/python_sdk
        ${DFS_PROJECT_PATH}/src/sdk/memfs/common
        ${DFS_PROJECT_PATH}/src/sdk/memfs/sdk
        ${DFS_PROJECT_PATH}/src/sdk/memfs/server
        ${DFS_PROJECT_PATH}/src/ufs
        ${DFS_PROJECT_PATH}/output/include
)

INCLUDE_DIRECTORIES(
        ${DFS_PROJECT_3RD_BINARY}/ubs-comm/include/hcom
        ${DFS_PROJECT_3RD_BINARY}/hdagger/include
        ${DFS_PROJECT_3RD_BINARY}/hseceasy/hse/authorization/include/authorization
        ${DFS_PROJECT_3RD_BINARY}/hseceasy/hse/cryption/include/cryption
        ${DFS_PROJECT_3RD_BINARY}/hseceasy/hse/cert/include/authorization
        ${DFS_PROJECT_3RD_BINARY}/hseceasy/hse/utils
        ${DFS_PROJECT_3RD_BINARY}/libboundscheck/include
        ${DFS_PROJECT_3RD_BINARY}/spdlog/include
        ${DFS_PROJECT_3RD_BINARY}/secodefuzz/include
)

LINK_DIRECTORIES(
        ${DFS_PROJECT_OUTPUT}/lib
        ${DFS_PROJECT_3RD_BINARY}/ubs-comm/lib
        ${DFS_PROJECT_3RD_BINARY}/hseceasy/hse/authorization/lib
        ${DFS_PROJECT_3RD_BINARY}/hseceasy/hse/dist_auth/lib
        ${DFS_PROJECT_3RD_BINARY}/cryption/hse/dist_auth/lib
        ${DFS_PROJECT_3RD_BINARY}/hseceasy/hse/cert/lib
        ${DFS_PROJECT_3RD_BINARY}/cryption/hse/tools/lib
        ${DFS_PROJECT_3RD_BINARY}/libboundscheck/lib
        ${DFS_PROJECT_3RD_BINARY}/spdlog/lib64
        ${DFS_PROJECT_3RD_BINARY}/secodefuzz/lib
)

add_compile_options(-fdump-rtl-expand)
add_link_options(-lgcov --coverage)
AUX_SOURCE_DIRECTORY(common TEST_SRC_FILES)
AUX_SOURCE_DIRECTORY(sdk TEST_SRC_FILES)
message("aaa${TEST_SRC_FILES}")
add_executable(dfs_dt_fuzz dt_fuzz_test.cpp ${TEST_SRC_FILES})
target_link_libraries(dfs_dt_fuzz
        fs_sdk_shared fs_server_static fs_common_static
        ock_memfs_core_static bdm_static_tool bdm
)
TARGET_LINK_LIBRARIES(dfs_dt_fuzz pthread dl rt libhcom_static.a libboundscheck.a libSecodefuzz.a hse_cert)
