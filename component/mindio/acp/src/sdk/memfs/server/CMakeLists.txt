# fs server library
file(GLOB_RECURSE SERVER_SRCS fs*.cpp fs*.h)

add_library(fs_server_object OBJECT ${SERVER_SRCS})
target_link_libraries(fs_server_object ockio_common_interface memfs_common_interface memfs_core_interface)
TARGET_LINK_LIBRARIES(fs_server_object libhcom_static.a libboundscheck.a)

add_library(fs_server_static STATIC $<TARGET_OBJECTS:fs_server_object>)
TARGET_LINK_LIBRARIES(fs_server_static libhcom_static.a libboundscheck.a)
target_include_directories(fs_server_static PUBLIC ${CMAKE_CURRENT_LIST_DIR})
target_include_directories(fs_server_static PUBLIC ${PROJECT_SOURCE_DIR}/src/common)
target_include_directories(fs_server_static PUBLIC ${PROJECT_SOURCE_DIR}/src/sdk/memfs/common)

set_target_properties(fs_server_static PROPERTIES OUTPUT_NAME "fs_server")
set_target_properties(fs_server_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)

add_library(fs_server_shared SHARED $<TARGET_OBJECTS:fs_server_object>)
TARGET_LINK_LIBRARIES(fs_server_shared libhcom_static.a libboundscheck.a)
target_include_directories(fs_server_shared PUBLIC ${CMAKE_CURRENT_LIST_DIR})
target_include_directories(fs_server_shared PUBLIC ${PROJECT_SOURCE_DIR}/src/common)
target_include_directories(fs_server_shared PUBLIC ${PROJECT_SOURCE_DIR}/src/sdk/memfs/common)
set_target_properties(fs_server_shared PROPERTIES OUTPUT_NAME "fs_server")
set_target_properties(fs_server_shared PROPERTIES CLEAN_DIRECT_OUTPUT 1)
