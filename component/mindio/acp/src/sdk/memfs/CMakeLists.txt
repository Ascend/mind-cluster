include_directories(
        client
        common
        python_sdk
        sdk
        server
        )

link_directories(
        ${THIRD_PARTY_OUT_LIB}/
        ${PROJECT_3RDPARTY_BIN_DIR}/ubs-comm/lib/
)

add_subdirectory(common)
add_subdirectory(sdk)
add_subdirectory(server)

if(BUILD_PYTHON_SDK)
    add_subdirectory(python_sdk)
endif()
