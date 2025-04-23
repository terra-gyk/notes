#include <fmt/printf.h>

int main()
{
  int num = std::atoi("  123");
  fmt::print("num: {} \n",num);
  return 0;
}