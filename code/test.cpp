#include <iostream>
#include <thread>
#include <functional>
#include <memory>

class node : public std::enable_shared_from_this<node>
{
public:
  std::shared_ptr<node> get_shared()
  {
    return shared_from_this();
  }
};

int main()
{
  auto node_p = std::make_shared<node>();

  auto tt = node_p->get_shared();
  return 0;
}