SET(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -pie")

add_executable(ockiod ockio_daemon.cpp ${depend_files})

set (bdm_util_build_options
        -D_FILE_OFFSET_BITS=64
        -fPIC
        -fPIE
        -Werror
        ${OCKIO_CPLUSPLUS_STANDARD}
		-Wno-format
        -Werror=reorder
        -Werror=parentheses
        -Werror=return-type
        -Werror=sign-compare
        -Werror=unused-variable
        -fstack-protector-all
        -DRTE_FORCE_INTRINSICS
        -DRTE_ARCH_64)

target_compile_options(ockiod PRIVATE -DRTE_ARCH_64 -D_FILE_OFFSET_BITS=64 -O3 ${OCKIO_CPLUSPLUS_STANDARD})

set(mylibs
	pthread
	z
	dl
	rt)

set(ockiod_libs
    pthread
    rt
    dl)

target_link_libraries(ockiod ${ockiod_libs} fs_server_static fs_common_static ock_memfs_core_static)
target_link_libraries(ockiod memfs_common_interface memfs_core_interface bdm_shared)

# The toolset linker does not search the host library directories, so `-lz` fails even when
# zlib-devel is installed (same root cause as the numa.h issue). Locate libz.so explicitly and
# link it by absolute path: only zlib is affected and the linker search order for every other
# library stays untouched. Falls back to the original `-lz` when libz.so cannot be found.
find_library(OCKIO_ZLIB_LIBRARY NAMES z
        PATHS /usr/lib64 /usr/lib /lib64 /lib
              /usr/lib/x86_64-linux-gnu /usr/lib/aarch64-linux-gnu)
if(OCKIO_ZLIB_LIBRARY)
    message(STATUS "ockiod: link zlib from ${OCKIO_ZLIB_LIBRARY}")
    target_link_libraries(ockiod "${OCKIO_ZLIB_LIBRARY}")
else()
    message(WARNING "ockiod: libz.so not found, falling back to -lz")
    target_link_libraries(ockiod z)
endif()
target_include_directories(ockiod PRIVATE ${PROJECT_SOURCE_DIR}/src/ufs ${PROJECT_SOURCE_DIR}/src/background)
TARGET_LINK_LIBRARIES(ockiod libhcom_static.a libspdlog.a pthread)
