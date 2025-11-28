SET(GLOBAL_DEFS_H_IN ${CMAKE_CURRENT_SOURCE_DIR}/global_defs.h.in)
SET(GLOBAL_DEFS_H ${CMAKE_CURRENT_BINARY_DIR}/global_defs.h)

CONFIGURE_FILE(${GLOBAL_DEFS_H_IN} ${GLOBAL_DEFS_H} @ONLY)
 
SET(SECURE_LIB ${PROJECT_SOURCE_DIR}/output/lib/libboundscheck.so)
include_directories(
        ${PROJECT_SOURCE_DIR}/3rdparty/libboundscheck/include)

if(BUILD_FOR_UT OR BUILD_FOR_FUZZ)
    SET(CXX_FLAGS
            -g
            -O0
            -fno-omit-frame-pointer
            )
    STRING(REPLACE ";" " " CXX_FLAGS_STR "${CXX_FLAGS}")
    SET(CMAKE_CXX_FLAGS "${CXX_FLAGS_STR} ${CMAKE_CXX_FLAGS}")
    add_compile_options(-fprofile-arcs -ftest-coverage -fdump-rtl-expand)
    add_link_options(-lgcov --coverage)
endif()

find_program(CCACHE_FOUND ccache)
if(CCACHE_FOUND)
    set(CMAKE_C_COMPILER_LAUNCHER ccache)
    set(CMAKE_CXX_COMPILER_LAUNCHER ccache)
endif(CCACHE_FOUND)

add_library(bdm_object OBJECT)
add_subdirectory(common)
add_subdirectory(ufs)
add_subdirectory(background)
target_include_directories(bdm_object PUBLIC
        memfs
        memfs/bmm
        memfs/common
        memfs/fs
        sdk/memfs/client
        sdk/memfs/common
        sdk/memfs/sdk
        sdk/memfs/server
        ${PROJECT_SOURCE_DIR}/output/include
        )
target_sources(bdm_object PRIVATE  ${GLOBAL_DEFS_H})

add_subdirectory(memfs)
add_subdirectory(sdk)
add_subdirectory(util)

set (bdm_build_options
        -rdynamic
        -Wno-format
        -Wformat=0
        -Werror=reorder
        -Werror=parentheses
        -Werror=return-type
        -Werror=sign-compare
        -Werror=unused-variable
        -DRTE_FORCE_INTRINSICS
        -DRTE_ARCH_64)
target_compile_options(bdm_object PRIVATE ${bdm_build_options})

#1. libbdm.so
add_library(bdm_shared SHARED $<TARGET_OBJECTS:bdm_object>)
set_target_properties(bdm_shared PROPERTIES OUTPUT_NAME "bdm")
target_include_directories(bdm_shared INTERFACE ufs)
TARGET_LINK_LIBRARIES(bdm_shared libhcom_static.a libboundscheck.a libspdlog.a pthread)

#2. libbdm_static_tool.a, bdm tool use
message("-- BDM Build tool lib.")
add_library(bdm_static_tool STATIC $<TARGET_OBJECTS:bdm_object>)
target_include_directories(bdm_static_tool INTERFACE ufs)
TARGET_LINK_LIBRARIES(bdm_static_tool libhcom_static.a libboundscheck.a libspdlog.a pthread)

add_library(hlog_static STATIC common/hlog.cpp common/hlog.h)
target_include_directories(hlog_static PUBLIC
        ${PROJECT_SOURCE_DIR}/output/include)
TARGET_LINK_LIBRARIES(hlog_static libspdlog.a pthread)

# manually install leftover 3rdparty binaries and libs
