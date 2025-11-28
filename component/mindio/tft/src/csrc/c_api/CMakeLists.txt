include_directories(
        ${PROJECT_SOURCE_DIR}/src/csrc/framework/controller/
        ${PROJECT_SOURCE_DIR}/src/csrc/framework/processor/
        ${PROJECT_SOURCE_DIR}/src/csrc/common/
)

file(GLOB C_API_FILES
        ${PROJECT_SOURCE_DIR}/src/csrc/c_api/*.cpp
)
add_library(ttp_c_api OBJECT ${C_API_FILES})
add_library(ttp_c_api_shared SHARED $<TARGET_OBJECTS:ttp_c_api>)

target_link_libraries(ttp_c_api_shared PRIVATE pthread dl rt ttp_framework_shared)
set_target_properties(ttp_c_api_shared PROPERTIES OUTPUT_NAME "ttp_c_api")
target_compile_options(ttp_c_api_shared PUBLIC -g)
target_include_directories(ttp_c_api_shared PRIVATE ${PROJECT_SOURCE_DIR}/src/csrc/common)
target_include_directories(ttp_c_api_shared PRIVATE ${PROJECT_SOURCE_DIR}/src/csrc/framework/controller)
target_include_directories(ttp_c_api_shared PRIVATE ${PROJECT_SOURCE_DIR}/src/csrc/framework/processor)