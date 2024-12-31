#!/bin/bash

g++ -std=c++23 -O3 -g test.cpp -lspdlog -lgtest -lgmock -o test && ./test
