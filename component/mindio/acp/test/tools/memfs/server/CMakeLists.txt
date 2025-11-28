# executable
file(GLOB_RECURSE SHELL_FS *.cpp)
add_executable(shell_fs_server ${SHELL_FS})
target_link_libraries(shell_fs_server
        ock_memfs_core_static fs_common_static fs_core_static fs_bmm_static fs_server_static ock_memfs_core_static
        fs_common_static bdm_static_tool fuse dl rt)
TARGET_LINK_LIBRARIES(shell_fs_server libhcom_static.a libboundscheck.a)
target_include_directories(shell_fs_server PRIVATE ${PROJECT_SOURCE_DIR}/src/common)
target_include_directories(shell_fs_server PRIVATE ${PROJECT_SOURCE_DIR}/src/memfs)
target_include_directories(shell_fs_server PRIVATE ${PROJECT_SOURCE_DIR}/src/memfs/common)
target_include_directories(shell_fs_server PRIVATE ${PROJECT_SOURCE_DIR}/src/sdk/memfs/common)
target_include_directories(shell_fs_server PRIVATE ${PROJECT_SOURCE_DIR}/src/sdk/memfs/server)
target_include_directories(shell_fs_server PRIVATE ${PROJECT_SOURCE_DIR}/3rdparty/libboundscheck/include)

set(TARGET_INSTALL_BIN ${DFS_PROJECT_BINARY}/shellFS)
install(TARGETS shell_fs_server DESTINATION ${TARGET_INSTALL_BIN}/ PERMISSIONS OWNER_WRITE OWNER_READ OWNER_EXECUTE GROUP_READ GROUP_EXECUTE WORLD_READ WORLD_EXECUTE)
