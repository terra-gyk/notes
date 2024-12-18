#include "tools.h"

#include <iostream>
#include <vector>
#include <algorithm>
#include <thread>

using ve_iter = std::vector<int>::iterator;

bool cmp(int left, int right){
  return left < right;
}

template<typename T>
T sum(T num1,T num2){
  return num1 + num2;
}

int main()
{
  std::vector<int> arr{ 1,3,5,44,2,56,6,22,7};
  std::sort(arr.begin(),arr.end(),cmp);
  for( auto num : arr ){
    std::cout << num << " " ;
  }
  std::cout << std::endl;

  tools::print(sum(1,2));
  
  return 0;
}
