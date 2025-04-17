c++ notes

- c++ 的纯虚函数是可以被实现的
```cpp
class interface {
public:
  virtual void print() = 0;
};

void interface::print(){
  SPDLOG_INFO("this is a test");
}
```