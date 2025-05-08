#include <spdlog/spdlog.h>
#include <asio.hpp>

#include <string>
#include <map>

int main(int argc, char **argv) 
{
  std::map<int,std::string> param;
  for(int i = 1; i < argc; i++)
  {
    param.emplace(i,argv[i]);
  }

  for( auto node : param )
  {
    SPDLOG_INFO(node.second);
  }
  // SPDLOG_INFO(param[1]);
  return 0;
}