cmake_minimum_required(VERSION 3.18)
find_package(Python3 COMPONENTS Interpreter Development REQUIRED)
find_package(PythonInterp REQUIRED)
find_package(pybind11 CONFIG)
if (pybind11_FOUND)
    message(STATUS "use python11 build module")
    pybind11_add_module(memfs_pybind memfs_pybind.cpp)

    target_link_libraries(memfs_pybind PRIVATE
            pthread dl rt hcom_static hlog_static fs_sdk_static libboundscheck.a)
    set_target_properties(memfs_pybind PROPERTIES OUTPUT_NAME "memfs")
    target_compile_options(memfs_pybind PUBLIC -g)
    target_include_directories(memfs_pybind PRIVATE ${PROJECT_SOURCE_DIR}/src/memfs/common)
    target_include_directories(memfs_pybind PRIVATE ${PROJECT_SOURCE_DIR}/src/common)
    target_include_directories(memfs_pybind PRIVATE ${PROJECT_SOURCE_DIR}/src/sdk/memfs/sdk)

    set(TARGET_INSTALL_BIN_PS ${PROJECT_SOURCE_DIR}/build/python_sdk/)
    install(TARGETS memfs_pybind DESTINATION ${TARGET_INSTALL_BIN_PS} PERMISSIONS OWNER_WRITE OWNER_READ OWNER_EXECUTE GROUP_READ GROUP_EXECUTE WORLD_READ WORLD_EXECUTE)
endif ()

Python3_add_library(memfs_client MODULE memfs_python_api.cpp)
target_link_libraries(memfs_client PRIVATE
        pthread dl rt hcom_static hlog_static fs_sdk_static libboundscheck.a)
set_target_properties(memfs_client PROPERTIES OUTPUT_NAME "memfs")
target_compile_options(memfs_client PUBLIC -g)
target_include_directories(memfs_client PRIVATE ${PROJECT_SOURCE_DIR}/src/memfs/common)
target_include_directories(memfs_client PRIVATE ${PROJECT_SOURCE_DIR}/src/common)
target_include_directories(memfs_client PRIVATE ${PROJECT_SOURCE_DIR}/src/sdk/memfs/sdk)

set(TARGET_INSTALL_BIN_PS ${PROJECT_SOURCE_DIR}/build/python_sdk/)
install(TARGETS memfs_client DESTINATION ${TARGET_INSTALL_BIN_PS} PERMISSIONS OWNER_WRITE OWNER_READ OWNER_EXECUTE GROUP_READ GROUP_EXECUTE WORLD_READ WORLD_EXECUTE)
