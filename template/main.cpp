#include "some.hpp"

#include <iostream>
#include <fstream>
#include <chrono>

//static std::string data = "this is a log test. this is a log test.this is a log test.this is a log test.this is a log test.this is a log test.this is a log test.this is a log test.\n";
char data[156] = "this is a log test. this is a log test.this is a log test.this is a log test.this is a log test.this is a log test.this is a log test.this is a log test.\n";

std::string filename = "fstream.txt";
std::ofstream ostrm(filename, std::ios::out);

void test_fstream(){
  for(int num = 1; num < 10000; num++){
    ostrm.write(data, 154); // binary output
  }
}

std::string c_filename = "c_file.txt";
FILE *file = fopen(c_filename.c_str(), "w");

void test_c_file(){
  for(int num = 1; num < 10000; num++){
    fwrite(data, 1, 154, file);
  }
}


int main(){
  // 获取开始时间
  auto start = std::chrono::high_resolution_clock::now();
  test_fstream();

  // 获取结束时间
  auto end = std::chrono::high_resolution_clock::now();

  // 计算并输出执行时间
  auto duration = std::chrono::duration_cast<std::chrono::microseconds>(end - start).count();
  std::cout << "执行时间: " << duration << " 微秒" << std::endl;

  start = std::chrono::high_resolution_clock::now();
  test_c_file();
  end = std::chrono::high_resolution_clock::now();
  duration = std::chrono::duration_cast<std::chrono::microseconds>(end - start).count();

  std::cout << "执行时间: " << duration << " 微秒" << std::endl;

  ostrm.close();
    // 关闭文件
  fclose(file);
}