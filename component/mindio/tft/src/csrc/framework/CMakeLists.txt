include_directories(
        ${PROJECT_SOURCE_DIR}/src/csrc/framework/controller/
        ${PROJECT_SOURCE_DIR}/src/csrc/framework/processor/
        ${PROJECT_SOURCE_DIR}/src/csrc/framework/
        ${PROJECT_SOURCE_DIR}/src/csrc/common/
        ${PROJECT_3RDPARTY_BIN_DIR}/spdlog/include/
)

file(GLOB SRC_FILES
    ${PROJECT_SOURCE_DIR}/src/csrc/framework/processor/*.cpp
    ${PROJECT_SOURCE_DIR}/src/csrc/framework/controller/*.cpp
    ${PROJECT_SOURCE_DIR}/src/csrc/framework/*.cpp
)

message(STATUS "source file:${SRC_FILES}")

add_library(ttp_framework_object OBJECT ${SRC_FILES})
target_link_libraries(ttp_framework_object ttp_common_interface)

add_library(ttp_framework_static STATIC $<TARGET_OBJECTS:ttp_framework_object>)
set_target_properties(ttp_framework_static PROPERTIES OUTPUT_NAME "ttp_framework")
set_target_properties(ttp_framework_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)

add_library(ttp_framework_shared SHARED $<TARGET_OBJECTS:ttp_framework_object>)
add_dependencies(ttp_framework_shared acc_tcp_net_static)
target_link_libraries(ttp_framework_shared
        -Wl,--start-group
        acc_tcp_net.a boundscheck.a libspdlog.a
        -Wl,--end-group
)
set_target_properties(ttp_framework_shared PROPERTIES OUTPUT_NAME "ttp_framework")
set_target_properties(ttp_framework_shared PROPERTIES CLEAN_DIRECT_OUTPUT 1)