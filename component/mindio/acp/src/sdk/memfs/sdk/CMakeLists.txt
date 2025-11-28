# fs server library
file(GLOB_RECURSE SDK_SRCS *.cpp *.h)

add_library(fs_sdk_object OBJECT ${SDK_SRCS} ${TIMER_SOURCE})
target_link_libraries(fs_sdk_object memfs_common_interface ockio_common_interface)
TARGET_LINK_LIBRARIES(fs_sdk_object libhcom_static.a libboundscheck.a)

add_library(fs_sdk_static STATIC $<TARGET_OBJECTS:fs_sdk_object>)
set_target_properties(fs_sdk_static PROPERTIES OUTPUT_NAME "fs_sdk")
set_target_properties(fs_sdk_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)
target_link_libraries(fs_sdk_static dl fs_common_static)
TARGET_LINK_LIBRARIES(fs_sdk_static libhcom_static.a libboundscheck.a)

add_library(fs_sdk_shared SHARED $<TARGET_OBJECTS:fs_sdk_object>)
set_target_properties(fs_sdk_shared PROPERTIES OUTPUT_NAME "fs_sdk")
set_target_properties(fs_sdk_shared PROPERTIES CLEAN_DIRECT_OUTPUT 1)
target_link_libraries(fs_sdk_shared dl fs_common_shared)
