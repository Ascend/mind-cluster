#build secure
set(SOURCE_DIR ${PROJECT_3RDPARTY_SRC_DIR}/libboundscheck/libboundscheck)
set(BUILD_DIR ${PROJECT_BUILD_PATH}/3rdparty/libboundscheck/)
set(INSTALL_DIR ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck)

message("-- secure_c: check if need to build")

# build the lib if not built yet
find_file(LIBS_BUILT_SECURE_C libboundscheck.so ${INSTALL_DIR}/lib/)
# check if the file exists or not, if not build it
if(NOT EXISTS ${LIBS_BUILT_SECURE_C})
    set(CMAKE_C_STANDARD 11)
    set(CMAKE_CXX_STANDARD 11)

    EXEC_PROGRAM(cp ${BUILD_DIR} ARGS -rf ${SOURCE_DIR} ${BUILD_DIR})

    # create build dir
    exec_program(mkdir ${BUILD_PATH} ARGS -p ${BUILD_DIR})

    EXEC_PROGRAM(make ${BUILD_DIR}/libboundscheck ARGS -j4)

    # create .a
    EXEC_PROGRAM(ar ${INSTALL_DIR}/lib ARGS crv ${INSTALL_DIR}/lib/libboundscheck.a ${BUILD_DIR}/libboundscheck/obj/*.o)

    # cp include and lib
    EXEC_PROGRAM(cp ${INSTALL_DIR} ARGS -rf ${BUILD_DIR}/libboundscheck/lib/ ${INSTALL_DIR})
    EXEC_PROGRAM(cp ${INSTALL_DIR} ARGS -rf ${BUILD_DIR}/libboundscheck/include ${INSTALL_DIR})

else()
    message("-- secure_c: ${LIBS_BUILT_SECURE_C}")
    message("-- secure_c: has been built, ignored")
endif(NOT EXISTS ${LIBS_BUILT_SECURE_C})