cmake_minimum_required(VERSION 3.14.1)
project("ockio" VERSION 23.0.0)

#################################
# global option switches
#################################
option(BUILD_PYTHON_SDK "enable python sdk. " OFF)
option(BUILD_FOR_UT "enable unit test. " OFF)

#################################
# include common cmake modules
#################################
include("cmake_modules/bdm_common.cmake")
include("cmake_modules/bdm_test.cmake")

set(PROJECT_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/output)
set(EXECUTABLE_OUTPUT_PATH "${PROJECT_OUTPUT_PATH}/bin")
set(LIBRARY_OUTPUT_PATH "${PROJECT_OUTPUT_PATH}/lib")
set(PROJECT_3RDPARTY_SRC_DIR ${PROJECT_SOURCE_DIR}/3rdparty)
set(PROJECT_3RDPARTY_BIN_DIR ${PROJECT_OUTPUT_PATH}/3rdparty)

#################################
# build 3rd party dependencies
#################################
add_subdirectory(3rdparty)

include_directories(
        ${PROJECT_3RDPARTY_BIN_DIR}/ubs-comm/include/hcom
        ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck/include
        ${PROJECT_3RDPARTY_BIN_DIR}/spdlog/include
)

link_directories(
        ${PROJECT_3RDPARTY_BIN_DIR}/ubs-comm/lib
        ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck/lib
        ${PROJECT_3RDPARTY_BIN_DIR}/spdlog/lib
        ${PROJECT_3RDPARTY_BIN_DIR}/spdlog/lib64
)

#################################
# build ockio components
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

add_subdirectory(src)
if(BUILD_WITH_HDT_TEST STREQUAL "ON")
    link_directories(${CMAKE_BINARY_DIR}/output)
    add_subdirectory(test)
endif()

if(BUILD_FOR_FUZZ)
    add_subdirectory(test/test_dt_fuzz)
endif()

if(BUILD_WITH_TEST_TOOLS STREQUAL "ON")
    link_directories(${CMAKE_BINARY_DIR}/output)
    add_subdirectory(test/tools)
endif()

