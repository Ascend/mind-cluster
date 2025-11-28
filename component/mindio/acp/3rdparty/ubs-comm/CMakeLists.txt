set(SOURCE_DIR ${PROJECT_3RDPARTY_SRC_DIR}/ubs-comm/ubs-comm/)
set(BUILD_DIR ${PROJECT_BUILD_PATH}/ubs-comm/)
set(INSTALL_DIR ${PROJECT_3RDPARTY_BIN_DIR}/ubs-comm)

set(CHECK_FILE_HCOM ${INSTALL_DIR}/include/hcom/hcom.h)

message(STATUS "hcom: check if need to build at ${CHECK_FILE_HCOM}")

# build the lib if not built yet
if(EXISTS ${CHECK_FILE_HCOM})
    message(STATUS "hcom: ${CHECK_FILE_HCOM}")
    message(STATUS "hcom: has been built, ignored")
else()
    # create build dir
    execute_process(COMMAND mkdir -p ${BUILD_DIR} WORKING_DIRECTORY ${SOURCE_DIR})
    execute_process(COMMAND mkdir -p ${INSTALL_DIR} WORKING_DIRECTORY ${BUILD_DIR})

    # configure
    execute_process(
        COMMAND cmake -DBUILD_WITH_RDMA=off -DCMAKE_DEPENDS_USE_COMPILER=false -DBUILD_TESTS=off -DBUILD_EXAMPLE=off -DBUILD_PERF=off -DCMAKE_INSTALL_PREFIX=${INSTALL_DIR} ${SOURCE_DIR}
        WORKING_DIRECTORY ${BUILD_DIR}
    )

    # execute make && make install
    execute_process(COMMAND make clean WORKING_DIRECTORY ${BUILD_DIR})
    execute_process(COMMAND make -j8 WORKING_DIRECTORY ${BUILD_DIR})
    execute_process(COMMAND make install WORKING_DIRECTORY ${BUILD_DIR})
endif()