
aux_source_directory(${CMAKE_CURRENT_LIST_DIR} MEMFS_BG_BACKUP_SRCS)
if(BUILD_BOTH_MFS_DFS STREQUAL "ON")
    add_compile_definitions(-D__BUILD_BOTH_MFS_DFS__)
else()
    list(REMOVE_ITEM MEMFS_BG_BACKUP_SRCS ${CMAKE_CURRENT_LIST_DIR}/dfs_backup_initiator.cpp)
endif()

target_sources(bdm_object PRIVATE ${MEMFS_BG_BACKUP_SRCS})
target_include_directories(bdm_object PRIVATE ${CMAKE_CURRENT_LIST_DIR})