cmake_minimum_required(VERSION 3.13)
project(TTP_TEST VERSION 1.0 LANGUAGES CXX)
set(CMAKE_CXX_STANDARD 17)

include(conf/comp.cmake)

message(STATUS "build from test path")
message(STATUS "BUILD_WITH_HDT_TEST = ${BUILD_WITH_HDT_TEST}")
if(BUILD_WITH_HDT_TEST STREQUAL "ON")
    set(OCK_PROJECT_SOURCE "${CMAKE_SOURCE_DIR}")
    set(OCK_PROJECT_BINARY "${CMAKE_BINARY_DIR}")
    set(OCK_TEST_PROJECT_BINARY "${CMAKE_BINARY_DIR}")
else ()
    set(OCK_PROJECT_SOURCE "${CMAKE_SOURCE_DIR}/..")
    set(OCK_PROJECT_BINARY "${CMAKE_BINARY_DIR}/build_target")
    set(OCK_TEST_PROJECT_BINARY "${CMAKE_BINARY_DIR}")
endif ()
message(STATUS "OCK_PROJECT_SOURCE = ${OCK_PROJECT_SOURCE}")
message(STATUS "OCK_PROJECT_BINARY = ${OCK_PROJECT_BINARY}")
message(STATUS "OCK_TEST_PROJECT_BINARY = ${OCK_TEST_PROJECT_BINARY}")
set(OCK_PROJECT_3RD_SRC ${OCK_PROJECT_SOURCE}/3rdparty)
set(OCK_PROJECT_3RD_BINARY "${OCK_PROJECT_SOURCE}/output/3rdparty")
set(ENV{3RDPARTY_BIN_DIR} ${OCK_PROJECT_3RD_BINARY})

add_definitions(-DUT_ENABLED)
add_compile_options(-fprofile-arcs -ftest-coverage)
# add crc compile flags
if(CMAKE_HOST_SYSTEM_PROCESSOR MATCHES "aarch64")
    add_compile_options(-march=armv8-a+crc)
endif()

add_link_options(-lgcov --coverage)

include_directories(
        ${OCK_PROJECT_SOURCE}/src
        ${OCK_PROJECT_SOURCE}/src/csrc
        ${OCK_PROJECT_SOURCE}/src/csrc/common
        ${OCK_PROJECT_SOURCE}/src/csrc/framework
        ${OCK_PROJECT_SOURCE}/src/csrc/framework/controller
        ${OCK_PROJECT_SOURCE}/src/csrc/framework/processor

        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/
        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/common/
        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/security/
        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/under_api/openssl/
)

link_directories(
        ${OCK_PROJECT_SOURCE}/output/lib

        ${PROJECT_BINARY_DIR}/build_target/src/csrc/controller
        ${PROJECT_BINARY_DIR}/build_target/src/csrc/processor
)
message(STATUS "path=${PROJECT_BINARY_DIR}/build_target/src/csrc/common")

add_subdirectory(llt)