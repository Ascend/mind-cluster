
add_library(memfs_core_interface INTERFACE)
target_include_directories(memfs_core_interface INTERFACE
        ${CMAKE_CURRENT_LIST_DIR}
        ${CMAKE_CURRENT_LIST_DIR}/common
        ${CMAKE_CURRENT_LIST_DIR}/bmm
        ${CMAKE_CURRENT_LIST_DIR}/fs
)
target_link_directories(memfs_core_interface INTERFACE memfs_common_interface ockio_common_interface)

add_subdirectory(common)
add_subdirectory(bmm)
add_subdirectory(fs)

file(GLOB_RECURSE MEMFS_CORE_SRC *.cpp *.h)
add_library(ock_memfs_core_object OBJECT ${MEMFS_CORE_SRC})
target_include_directories(ock_memfs_core_object PUBLIC ${CMAKE_CURRENT_LIST_DIR} bmm common fs)
target_link_libraries(ock_memfs_core_object PUBLIC memfs_common_interface ockio_common_interface)

add_library(ock_memfs_core_static STATIC $<TARGET_OBJECTS:ock_memfs_core_object>)
set_target_properties(ock_memfs_core_static PROPERTIES OUTPUT_NAME "ock_memfs_core")
set_target_properties(ock_memfs_core_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)

add_library(ock_memfs_core_shared SHARED $<TARGET_OBJECTS:ock_memfs_core_object>)
set_target_properties(ock_memfs_core_shared PROPERTIES OUTPUT_NAME "ock_memfs_core")
set_target_properties(ock_memfs_core_shared PROPERTIES CLEAN_DIRECT_OUTPUT 1)

if (${CMAKE_INSTALL_PREFIX} MATCHES "/usr/local")
    set(TARGET_INSTALL_INCLUDE ${SRC_DIR}/build/ock_memfs_core/include/ock_memfs_core)
    set(TARGET_INSTALL_LIB ${SRC_DIR}/build/ock_memfs_core/lib)
elseif(NOT CMAKE_INSTALL_PREFIX)
    set(TARGET_INSTALL_INCLUDE ${SRC_DIR}/build/ock_memfs_core/include/ock_memfs_core)
    set(TARGET_INSTALL_LIB ${SRC_DIR}/build/ock_memfs_core/lib)
else ()
    set(TARGET_INSTALL_INCLUDE ${CMAKE_INSTALL_PREFIX}/include/ock_memfs_core)
    set(TARGET_INSTALL_LIB ${CMAKE_INSTALL_PREFIX}/lib)
endif()

file(GLOB INCLUDE_HEADERS memfs_api.h)

install(FILES ${INCLUDE_HEADERS} DESTINATION ${TARGET_INSTALL_INCLUDE} PERMISSIONS OWNER_WRITE OWNER_READ GROUP_READ WORLD_READ)
install(FILES ${C_INCLUDE_HEADERS} DESTINATION ${TARGET_INSTALL_INCLUDE}/capi PERMISSIONS OWNER_WRITE OWNER_READ GROUP_READ WORLD_READ)
install(TARGETS ock_memfs_core_static DESTINATION ${TARGET_INSTALL_LIB}/ PERMISSIONS OWNER_WRITE OWNER_READ GROUP_READ WORLD_READ)
