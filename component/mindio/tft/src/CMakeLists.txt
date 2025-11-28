SET(SECURE_LIB ${PROJECT_SOURCE_DIR}/output/lib/libboundscheck.so)
include_directories(
        ${PROJECT_SOURCE_DIR}/3rdparty/libboundscheck/include)

if(BUILD_TESTS STREQUAL "ON" OR BUILD_FOR_FUZZ)
    SET(CXX_FLAGS
            -g
            -O0
            -fno-omit-frame-pointer
            )
    STRING(REPLACE ";" " " CXX_FLAGS_STR "${CXX_FLAGS}")
    SET(CMAKE_CXX_FLAGS "${CXX_FLAGS_STR} ${CMAKE_CXX_FLAGS}")
    add_compile_options(-fprofile-arcs -ftest-coverage)
    if(BUILD_FOR_FUZZ)
        add_compile_options(-ftest-coverage -fprofile-arcs -fdump-rtl-expand)
        message(STATUS "add fuzz coverage compile options")
    endif()
    add_link_options(-lgcov --coverage)
endif()

add_subdirectory(csrc)
