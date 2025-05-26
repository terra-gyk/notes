#include <fmt/printf.h>

#include <future>
#include <thread>


int print_func(int num)
{
  fmt::print("this is package test, num is {}.\n",num);
  return num;
}

int main()
{
  std::packaged_task<int(int)> package(print_func);

  std::future<int> f1 = package.get_future();

  std::thread t([&package](){ package(123);}); 

  f1.get();
  t.join();

  auto f2 = std::async(print_func,123);
  fmt::print("future 2 num: {}\n",f2.get());

  return 0;
}