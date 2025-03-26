#ifndef __TEST_H__
#define __TEST_H__

#include "ii.h"

class node : public interface {
public:
  friend class interface;
  void print() override;

  void test();
};

#endif // __TEST_H__