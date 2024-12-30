#!/bin/bash

g++ -std=c++23 -O3 -g -lspdlog -lgtest -lgmock test.cpp -o test  && ./test
