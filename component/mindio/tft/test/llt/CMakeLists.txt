set(DEPEND_LIBS
        -Wl,--start-group
        pthread dl acc_tcp_net rt ttp_framework
        -Wl,--end-group)

file(GLOB_RECURSE UT_SOURCE_FILES *.cpp testcase/*)

add_executable(test_ttp ${UT_SOURCE_FILES})

target_compile_options(test_ttp PUBLIC -D_GNU_SOURCE -g)
target_compile_definitions(test_ttp PRIVATE BUILD_PATH="${OCK_TEST_PROJECT_BINARY}")

set_target_properties(test_ttp PROPERTIES OUTPUT_NAME "test_ttp")
set_target_properties(test_ttp PROPERTIES CLEAN_DIRECT_OUTPUT 1)

target_link_libraries(test_ttp ${DEPEND_LIBS} acc_tcp_net_static)
target_include_directories(test_ttp PRIVATE testcase/controller/
    testcase/
    ${PROJECT_SOURCE_DIR}/src/csrc/acc_links
    ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/common
    ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/security
    ${PROJECT_SOURCE_DIR}/src/csrc/acc_links/under_api/openssl
)