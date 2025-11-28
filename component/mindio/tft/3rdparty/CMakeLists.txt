INCLUDE(ExternalProject)

add_subdirectory(libboundscheck)
add_subdirectory(spdlog)

if(BUILD_FOR_FUZZ)
    add_subdirectory(fuzz)
endif()

##### install 3rdparty library
set(TARGET_INSTALL_LIB ${TARGET_INSTALL_DIR}/lib)

### huawei secure_c
file(GLOB SUCURE_LIB ${PROJECT_3RDPARTY_BIN_DIR}/libboundscheck/lib/libboundscheck.so)
