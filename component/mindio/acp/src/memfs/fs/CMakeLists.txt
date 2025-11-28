file(GLOB_RECURSE FS_SRCS *.cpp *.h)

add_library(fs_core_object OBJECT ${FS_SRCS})
target_link_libraries(fs_core_object PUBLIC memfs_core_interface ockio_common_interface)
target_include_directories(fs_core_object PUBLIC ${CMAKE_CURRENT_LIST_DIR})

add_library(fs_core_static STATIC $<TARGET_OBJECTS:fs_core_object>)
set_target_properties(fs_core_static PROPERTIES OUTPUT_NAME "fs_core")
set_target_properties(fs_core_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)

add_library(fs_core_shared SHARED $<TARGET_OBJECTS:fs_core_object>)
set_target_properties(fs_core_shared PROPERTIES OUTPUT_NAME "fs_core")
set_target_properties(fs_core_shared PROPERTIES CLEAN_DIRECT_OUTPUT 1)
