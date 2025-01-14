#include <iostream>
#include <fstream>
#include <future>
#include <functional>
#include <chrono>

int sum(int num1,int num2)
{
  return num1 + num2;
}

int main()
{
  int num = 1;
  printf("num is: %d\n",num);
  if(num == 5){
  }else{
    num++;
  }

  return 0;
}