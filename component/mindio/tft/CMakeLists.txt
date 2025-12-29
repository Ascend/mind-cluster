cmake_minimum_required(VERSION 3.12.0)
project(ock-ttp CXX C)
set(CMAKE_CXX_STANDARD 17)

add_compile_definitions(_GLIBCXX_USE_CXX11_ABI=0)

#################################
# include common cmake modules
#################################
include("cmake_modules/ttp_common.cmake")

set(PROJECT_3RDPARTY_SRC_DIR ${PROJECT_SOURCE_DIR}/3rdparty)
set(PROJECT_BUILD_PATH ${PROJECT_SOURCE_DIR}/build)
set(PROJECT_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/output)
set(PROJECT_3RDPARTY_BIN_DIR ${PROJECT_OUTPUT_PATH}/3rdparty)
set(ENV{3RDPARTY_BIN_DIR} ${PROJECT_3RDPARTY_BIN_DIR})
set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/output/bin)
set(LIBRARY_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/output/lib)

#################################
# build 3rd party dependencies
#################################
add_subdirectory(3rdparty)

include_directories(
        ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck/include/
        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/
        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/common/
        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/security/
        ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/under_api/openssl/
        ${PROJECT_3RDPARTY_BIN_DIR}/spdlog/include/
)

link_directories(
        ${PROJECT_SOURCE_DIR}/output/lib/
        ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck/lib/
        ${PROJECT_3RDPARTY_BIN_DIR}/spdlog/lib/
        ${PROJECT_3RDPARTY_BIN_DIR}/spdlog/lib64/
)

#################################
# build ttp components
#################################
IF(OCKIO_ASAN_BUILD OR BUILD_FOR_FUZZ)
    FIND_LIBRARY(LIBASAN NAMES libasan.so.0 libasan.so.1 libasan.so.2 libasan.so.3 libasan.so.4 libasan.so.5 libasan.so.6 NO_CACHE)
    IF(${LIBASAN} STREQUAL "LIBASAN-NOTFOUND")
        MESSAGE(FATAL_ERROR "Have you installed libasan? It cannot be found.")
    ELSE()
        MESSAGE(STATUS "LIBASAN Found: ${LIBASAN}")
    ENDIF()
    SET(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fsanitize=address")
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=address")
    LINK_LIBRARIES(asan)
ENDIF()

IF(OCKIO_TSAN_BUILD)
    FIND_LIBRARY(LIBTSAN NAMES libtsan.so.0 libtsan.so.1 libtsan.so.2 libtsan.so.3 libtsan.so.4 libtsan.so.5 libtsan.so.6 NO_CACHE)
    IF(${LIBTSAN} STREQUAL "LIBTSAN-NOTFOUND")
        MESSAGE(FATAL_ERROR "Have you installed libtsan? It cannot be found.")
    ELSE()
        MESSAGE(STATUS "LIBTSAN Found: ${LIBTSAN}")
    ENDIF()
    SET(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fsanitize=thread")
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=thread")
    SET(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -pie")
    SET(CMAKE_SHARED_LINKER_FLAGS "${CMAKE_SHARED_LINKER_FLAGS} -pie")
    LINK_LIBRARIES(tsan)
ENDIF()

message(STATUS "BUILD_TESTS = ${BUILD_TESTS}")
if(BUILD_TESTS STREQUAL "ON")
    message(STATUS "BUILD_TESTS = ON, add compile gov")
    add_compile_options(-fprofile-arcs -ftest-coverage -DUT_ENABLED)
    add_link_options(-lgcov --coverage)
endif()

add_subdirectory(src)

if(BUILD_FOR_FUZZ)
    add_subdirectory(test/test_dt_fuzz)
endif()

if(BUILD_WITH_HDT_TEST STREQUAL "ON")
    MESSAGE(STATUS "BUILD_WITH_HDT_TEST =${BUILD_WITH_HDT_TEST}")
    link_directories(${CMAKE_BINARY_DIR}/output)
    add_subdirectory(test)
endif()