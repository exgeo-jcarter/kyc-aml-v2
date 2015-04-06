project (kyc-aml-doublemetaphone-project C CXX)

cmake_minimum_required(VERSION 2.8)

set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} "${CMAKE_SOURCE_DIR}/cmake/Modules/")

find_package (Boost COMPONENTS system REQUIRED)
find_package (JsonCpp REQUIRED)

list (
    APPEND CMAKE_CXX_FLAGS 
    "-std=c++11 ${CMAKE_CXX_FLAGS}"
    ${JSONCPP_DEFINITIONS}
)

add_executable (
    kyc-aml-doublemetaphone
    double-metaphone/double_metaphone.cc
    KycAmlDoubleMetaphone.cpp
    main.cpp
)

include_directories (
    kyc-aml-doublemetaphone
    ${JSONCPP_INCLUDE_DIRS}
)

target_link_libraries (
    kyc-aml-doublemetaphone
    ${Boost_SYSTEM_LIBRARY}
    ${JSONCPP_LIBRARIES}
)

#set(myExample_test_SOURCES
#        src/file1.cpp
#        test/test_file2.cpp
#        test/test_main.cpp)
#
#add_executable(myTestSuite ${myExample_test_SOURCES})
