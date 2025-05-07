#include <memory>
#include <iostream>

class father {};

// 多重继承时，并不共享 public 关键字，每个类前都需要加关键字，否则，默认 private
class son : public father, public std::enable_shared_from_this<son>{};


// c++ 的纯虚函数是可以被实现的

class interface {
public:
  virtual void print() = 0;
};

void interface::print(){
  std::cout << ("this is a test") << std::endl;
}
