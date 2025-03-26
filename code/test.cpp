#include "test.h"

#include <spdlog/spdlog.h>

void node::print(){
  interface::print();
}

void node::test(){
  SPDLOG_INFO("node test func");
}

void interface::print(){
  SPDLOG_INFO("this is a test");
}

int main(int argc, char** argv){
  if(argc < 3) {
    SPDLOG_INFO("argc < 3");
  }
  
  printf("argc: %d\n",argc);

  return 0;
}