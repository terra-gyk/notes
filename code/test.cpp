#include <iostream>
#include <utility>
#include <source_location>
#include <tuple>
#include <vector>
#include <functional>
#include <algorithm>
#include <thread>

template<typename T, T N>
void test()
{
  std::cout << std::source_location::current().function_name() << std::endl;
  std::cout << N << std::endl; 
}

template<typename T, typename F = std::function<bool()>>
void sort(T contain, F cmp){

}

using ve_iter = std::vector<int>::iterator;

bool cmp(int left, int right){
  return left < right;
}

int main()
{
  std::vector<std::string> ddd;
  auto node = ddd.at(1);
  std::this_thread::sleep_for(std::chrono::milliseconds(100));

  return 0;
}
