file(GLOB_RECURSE SRCS *.cpp *.h)
add_library(fs_common_object OBJECT ${SRCS})
target_link_libraries(fs_common_object memfs_core_interface ockio_common_interface)
TARGET_LINK_LIBRARIES(fs_common_object libhcom_static.a)

add_library(fs_common_static STATIC $<TARGET_OBJECTS:fs_common_object>)
set_target_properties(fs_common_static PROPERTIES OUTPUT_NAME "fs_common")
set_target_properties(fs_common_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)

add_library(fs_common_shared SHARED $<TARGET_OBJECTS:fs_common_object>)
set_target_properties(fs_common_shared PROPERTIES OUTPUT_NAME "fs_common")
set_target_properties(fs_common_shared PROPERTIES CLEAN_DIRECT_OUTPUT 1)
