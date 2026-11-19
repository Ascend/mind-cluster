file(GLOB_RECURSE BMM_SRCS *.cpp *.h)

add_library(fs_bmm_object OBJECT ${BMM_SRCS})
target_link_libraries(fs_bmm_object PUBLIC memfs_core_interface ockio_common_interface)

# bmm_pool.cpp includes <numa.h>. With a toolset compiler the host /usr/include is not
# searched, and `-I /usr/include` is dropped by GCC/CMake for standard system dirs, so
# append it as a system include directory after the standard ones instead.
target_compile_options(fs_bmm_object PRIVATE "-idirafter" "/usr/include")

add_library(fs_bmm_static STATIC $<TARGET_OBJECTS:fs_bmm_object>)
set_target_properties(fs_bmm_static PROPERTIES OUTPUT_NAME "fs_bmm")
set_target_properties(fs_bmm_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)

add_library(fs_bmm_shared SHARED $<TARGET_OBJECTS:fs_bmm_object>)
set_target_properties(fs_bmm_shared PROPERTIES OUTPUT_NAME "fs_bmm")
set_target_properties(fs_bmm_shared PROPERTIES CLEAN_DIRECT_OUTPUT 1)
