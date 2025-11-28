#build spdlog
set(SOURCE_DIR ${PROJECT_3RDPARTY_SRC_DIR}/spdlog/spdlog)
set(BUILD_DIR ${PROJECT_BUILD_PATH}/spdlog)
set(INSTALL_DIR ${PROJECT_3RDPARTY_BIN_DIR}/spdlog)

message(STATUS "spdlog: check if need to build")

# build the lib if not built yet
set(CHECK_FILE_SPDLOG ${INSTALL_DIR}/lib/libspdlog.a)
if(NOT EXISTS ${CHECK_FILE_SPDLOG})

    # create build dir
    execute_process(COMMAND mkdir -p ${BUILD_DIR} WORKING_DIRECTORY ${SOURCE_DIR})

    # configure
    execute_process(
        COMMAND cmake -DCMAKE_BUILD_TYPE=release -DBUILD_SHARED_LIBS=OFF -DCMAKE_INSTALL_PREFIX=${INSTALL_DIR} ${SOURCE_DIR}
        WORKING_DIRECTORY ${BUILD_DIR})

    # execute make && make install
    execute_process(COMMAND make clean WORKING_DIRECTORY ${BUILD_DIR})
    execute_process(COMMAND make -j8 WORKING_DIRECTORY ${BUILD_DIR})
    execute_process(COMMAND make install WORKING_DIRECTORY ${BUILD_DIR})
else()
    message(STATUS "spdlog: ${CHECK_FILE_SPDLOG}")
    message(STATUS "spdlog: has been built, ignored")
endif(NOT EXISTS ${CHECK_FILE_SPDLOG})
