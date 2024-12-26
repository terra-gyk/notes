#!/bin/bash

g++ test.cpp -o test -std=c++23 -O3 -g -lspdlog -lgtest -lgmock  && ./test
