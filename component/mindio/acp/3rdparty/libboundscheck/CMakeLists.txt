set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/libboundscheck)
set(OBJECT_DIR ${CMAKE_CURRENT_SOURCE_DIR}/libboundscheck/obj)
set(BUILD_DIR ${PROJECT_BUILD_PATH}/build/libboundscheck/)
set(INSTALL_DIR ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck)

set(CHECK_FILE_SECUREC ${INSTALL_DIR}/lib/libboundscheck.so)

message(STATUS "securec: check if need to build at ${CHECK_FILE_SECUREC}")

# build the lib if not built yet
if(EXISTS ${CHECK_FILE_SECUREC})
    message(STATUS "securec: ${CHECK_FILE_SECUREC}")
    message(STATUS "securec: has been built, ignored")
else()
    execute_process(COMMAND make -j WORKING_DIRECTORY ${SOURCE_DIR})
    EXEC_PROGRAM(ar ${INSTALL_DIR}/lib ARGS crv ${INSTALL_DIR}/lib/libboundscheck.a ${OBJECT_DIR}/*.o)
    execute_process(COMMAND mkdir -p ${INSTALL_DIR} WORKING_DIRECTORY ${SOURCE_DIR})
    execute_process(COMMAND cp -r ${CMAKE_CURRENT_SOURCE_DIR}/libboundscheck/include ${INSTALL_DIR} WORKING_DIRECTORY ${INSTALL_DIR})
endif()