
# executable
add_executable(shell_fs_client_test client_main.cpp)
target_include_directories(shell_fs_client_test PRIVATE ${PROJECT_SOURCE_DIR}/src/common)
target_include_directories(shell_fs_client_test PRIVATE ${PROJECT_SOURCE_DIR}/src/memfs/common)
target_include_directories(shell_fs_client_test PRIVATE ${PROJECT_SOURCE_DIR}/src/sdk/memfs/sdk)
target_include_directories(shell_fs_client_test PRIVATE ${PROJECT_SOURCE_DIR}/3rdparty/libboundscheck/include)
target_link_libraries(shell_fs_client_test -Wl,--start-group
        dl pthread fs_sdk_static hcom_static libboundscheck.a rt
        -Wl,--end-group)

add_executable(mem_fs_cli mem_fs_cli_main.cpp ${PROJECT_SOURCE_DIR}/src/common/hlog.cpp)
target_include_directories(mem_fs_cli PRIVATE ${PROJECT_SOURCE_DIR}/src/common)
target_include_directories(mem_fs_cli PRIVATE ${PROJECT_SOURCE_DIR}/src/memfs/common)
target_include_directories(mem_fs_cli PRIVATE ${PROJECT_SOURCE_DIR}/src/sdk/memfs/sdk)
target_link_libraries(mem_fs_cli -Wl,--start-group
        dl pthread fs_sdk_static hcom_static libboundscheck.a rt
        -Wl,--end-group)

set(TARGET_INSTALL_BIN ${DFS_PROJECT_BINARY}/shellFS)
install(TARGETS shell_fs_client_test DESTINATION ${TARGET_INSTALL_BIN}/ PERMISSIONS OWNER_WRITE OWNER_READ OWNER_EXECUTE GROUP_READ GROUP_EXECUTE WORLD_READ WORLD_EXECUTE)
install(TARGETS mem_fs_cli DESTINATION ${TARGET_INSTALL_BIN}/ PERMISSIONS OWNER_WRITE OWNER_READ OWNER_EXECUTE GROUP_READ GROUP_EXECUTE WORLD_READ WORLD_EXECUTE)
