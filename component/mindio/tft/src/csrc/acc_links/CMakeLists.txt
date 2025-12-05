# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
include_directories(
        ${PROJECT_SOURCE_DIR}/src/csrc/common/
)

file(GLOB_RECURSE ACC_TCP_NET_SRCS *.cpp)
add_library(acc_tcp_net_object OBJECT ${ACC_TCP_NET_SRCS})
add_library(acc_tcp_net_static STATIC $<TARGET_OBJECTS:acc_tcp_net_object>)
set_target_properties(acc_tcp_net_static PROPERTIES OUTPUT_NAME "acc_tcp_net")
set_target_properties(acc_tcp_net_static PROPERTIES CLEAN_DIRECT_OUTPUT 1)
target_link_libraries(acc_tcp_net_static
        -Wl,--start-group
        pthread dl rt
        -Wl,--end-group
        )