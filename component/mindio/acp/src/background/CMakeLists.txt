
add_subdirectory(backup)

aux_source_directory(${CMAKE_CURRENT_LIST_DIR} MEMFS_BG_SRCS)
target_sources(bdm_object PRIVATE ${MEMFS_BG_SRCS})
target_include_directories(bdm_object PRIVATE ${CMAKE_CURRENT_LIST_DIR})
