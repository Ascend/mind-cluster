include(../conf/comp.cmake)

set(DFS_PROJECT_PATH "${PROJECT_SOURCE_DIR}/")
set(DFS_PROJECT_3RD_SOURCE "${PROJECT_SOURCE_DIR}/output/3rdparty")
message(STATUS "DFS_PROJECT_PATH=${DFS_PROJECT_PATH}, DFS_PROJECT_3RD_SOURCE=${DFS_PROJECT_3RD_SOURCE}")
INCLUDE_DIRECTORIES(
    ${DFS_PROJECT_PATH}/src/csrc
    ${DFS_PROJECT_PATH}/src/csrc/common
    ${DFS_PROJECT_PATH}/src/csrc/framework
    ${DFS_PROJECT_PATH}/src/csrc/framework/controller
    ${DFS_PROJECT_PATH}/src/csrc/framework/processor

    ${DFS_PROJECT_3RD_SOURCE}/hdagger/include/
    ${DFS_PROJECT_3RD_SOURCE}/libboundscheck/include/
    ${DFS_PROJECT_3RD_SOURCE}/acc_links/include/
    ${DFS_PROJECT_3RD_SOURCE}/secodefuzz/include/
    )

#link_directories()

link_directories(
        ${DFS_PROJECT_3RD_SOURCE}/libboundscheck/lib/
        ${DFS_PROJECT_3RD_SOURCE}/acc_links/lib
        ${DFS_PROJECT_3RD_SOURCE}/secodefuzz/lib/
        ${DFS_PROJECT_3RD_SOURCE}/../lib/
)

set(FUZZ_DEPEND_LIBS
        -Wl,--start-group
        pthread dl boundscheck acc_tcp_net rt ttp_framework Secodefuzz
        -Wl,--end-group)

file(GLOB_RECURSE FUZZ_TEST_SOURCE_FILES dt_fuzz_test.cpp c2python/*)
add_link_options(-lgcov --coverage)
add_executable(fuzz_ttp ${FUZZ_TEST_SOURCE_FILES})
add_dependencies(fuzz_ttp ttp_framework_static)
target_compile_options(fuzz_ttp PUBLIC -D_GNU_SOURCE -g)
message(STATUS "BUILD_PATH=${CMAKE_BINARY_DIR}")
target_compile_definitions(fuzz_ttp PRIVATE BUILD_PATH="${CMAKE_BINARY_DIR}")

set_target_properties(fuzz_ttp PROPERTIES OUTPUT_NAME "fuzz_ttp")
set_target_properties(fuzz_ttp PROPERTIES CLEAN_DIRECT_OUTPUT 1)

target_link_libraries(fuzz_ttp ${FUZZ_DEPEND_LIBS})
target_include_directories(fuzz_ttp PRIVATE c2python/)