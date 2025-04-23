cmake_minimum_required(VERSION 3.11)

set(PROJECT test)

project(${PROJECT})

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Werror ")

find_library(zookeeper_lib zookeeper_mt)

if(NOT zookeeper_lib)
    message(FATAL_ERROR "ZooKeeper library not found.")
endif()

# 查找 librdkafka 库
set(rdkafka_lib rdkafka++)

find_package(spdlog REQUIRED)
find_package(leveldb REQUIRED) 

# ---------------------------------------------------------------------------------------
# Example of using pre-compiled library
# ---------------------------------------------------------------------------------------
file(GLOB_RECURSE TERRA_SOURCE "src/*.cpp")
add_executable(${PROJECT} ${TERRA_SOURCE})
# add_executable(test test.cpp)
target_link_libraries(${PROJECT} PRIVATE ${zookeeper_lib} ${rdkafka_lib} spdlog::spdlog leveldb::leveldb $<$<BOOL:${MINGW}>:ws2_32>)

# ---------------------------------------------------------------------------------------
# Example of using header-only library
# ---------------------------------------------------------------------------------------
# if(SPDLOG_BUILD_EXAMPLE_HO)
#     add_executable(example_header_only example.cpp)
#     target_link_libraries(example_header_only PRIVATE spdlog::spdlog_header_only)
# endif()
