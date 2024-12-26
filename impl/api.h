#ifndef __API_H__
#define __API_H__

#include <iostream>
#include <memory>

class node_api {
public:
  void print(){
    std::cout << msg;
  }
private:
  std::string msg = "this is a node api msg\n";
};

class impl_api {
public:
  impl_api():impl_(new node_api){}

  void print(){
    impl_->print();
  }

private:
  std::unique_ptr<node_api> impl_;
};

#endif // __API_H__