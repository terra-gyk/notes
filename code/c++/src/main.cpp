#include <spdlog/spdlog.h>
#include <asio.hpp>

#include <string>
#include <map>

void b_sort(std::vector<int>& list)
{
  size_t size = list.size();
  for(size_t i = 0; i < size; i++)
  {
    for(size_t j = i; j < size; j++)
    {
      if( list[i] > list[j] )
      {
        std::swap(list[i],list[j]);
      }
    }
  }
  return ;
}


int main(int argc, char **argv) 
{
  std::vector<int> list{ 1,3,45,23,4,6,1,23,67};
  b_sort(list);
  for(auto num : list)
  {
    SPDLOG_INFO(num);
  }
  return 0;
}