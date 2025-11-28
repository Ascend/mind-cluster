
aux_source_directory(${CMAKE_CURRENT_LIST_DIR} MEMFS_COMMON_SRCS)
target_sources(bdm_object PRIVATE ${MEMFS_COMMON_SRCS})
target_include_directories(bdm_object PRIVATE ${CMAKE_CURRENT_LIST_DIR})

add_library(ockio_common_interface INTERFACE)
target_include_directories(ockio_common_interface INTERFACE ${CMAKE_CURRENT_LIST_DIR})

set(TIMER_SOURCE ${CMAKE_CURRENT_LIST_DIR}/timer_executor.cpp ${CMAKE_CURRENT_LIST_DIR}/timer_executor.h PARENT_SCOPE)
