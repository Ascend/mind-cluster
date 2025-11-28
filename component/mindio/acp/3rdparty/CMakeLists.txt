set(PROJECT_BUILD_PATH ${CMAKE_CURRENT_BINARY_DIR})

add_subdirectory(ubs-comm)
add_subdirectory(libboundscheck)
add_subdirectory(spdlog)

if(BUILD_FOR_FUZZ)
    add_subdirectory(secodefuzz)
endif()

##### install 3rdparty library
set(TARGET_INSTALL_LIB ${TARGET_INSTALL_DIR}/lib)
 
### huawei secure_c
file(GLOB SUCURE_LIB ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck/lib/libboundscheck.so)
install(FILES ${SUCURE_LIB} DESTINATION ${TARGET_INSTALL_LIB}/ PERMISSIONS OWNER_READ OWNER_EXECUTE)