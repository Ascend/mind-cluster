cmake_minimum_required(VERSION 3.14.1)

find_package(SWIG REQUIRED)
find_package(Python3 COMPONENTS Interpreter Development REQUIRED)

if (NOT SWIG_FOUND OR NOT Python3_FOUND)
    message(FATAL_ERROR "use swig build, but swig or python3 not found!")
    return()
endif ()

cmake_policy(SET CMP0078 NEW)
cmake_policy(SET CMP0086 NEW)

add_compile_options(-DPy_LIMITED_API=0x03070000)

include(UseSWIG)
include(${SWIG_USE_FILE})
include_directories(${Python3_INCLUDE_DIRS} ${CMAKE_CURRENT_LIST_DIR})
set(CMAKE_SWIG_FLAGS "-c++")
set_source_files_properties(c2python_api.i PROPERTIES CPLUSPLUS ON)
message(STATUS "use swig build module")
swig_add_library(c2python_api TYPE MODULE LANGUAGE python SOURCES c2python_api.i c2python_api.cpp)

target_link_libraries(c2python_api PRIVATE ockio_common_interface memfs_common_interface memfs_core_interface)
target_link_libraries(c2python_api PRIVATE
        -Wl,--start-group
        pthread dl rt hcom_static fs_sdk_static
        -Wl,--end-group)

if(BUILD_FOR_UT)
    add_library(c2python_api_objs OBJECT c2python_api.cpp c2python_api.h)
    target_include_directories(c2python_api_objs PUBLIC ${CMAKE_CURRENT_LIST_DIR})
    target_link_libraries(c2python_api_objs ockio_common_interface memfs_common_interface memfs_core_interface)

    add_library(c2python_api_static STATIC $<TARGET_OBJECTS:c2python_api_objs>)
    target_link_libraries(c2python_api_static PRIVATE pthread hcom_static fs_sdk_static)
    set_target_properties(c2python_api_static PROPERTIES OUTPUT_NAME "acp_c2python_api")
    set_target_properties(c2python_api_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)
endif()