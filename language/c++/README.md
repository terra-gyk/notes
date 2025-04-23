- 使用clangd，需要在cmakelist.txt中指定需要的版本，比如使用c++协程，至少需要 c++20 版本，否则会出现无法识别的关键字

- c++ 中迭代器失效的情况存在于，内存重新分配或者内存失效的情况下，常见的有，新加元素导致扩容，元素被移除

- 三原则
- 如果某个类需要用户定义的析构函数、用户定义的复制构造函数或用户定义的复制赋值运算符，那么它几乎肯定需要全部三个。
- 因为 C++ 在各种场合（按值传递/返回、操纵容器等）对用户定义类型的对象进行复制和复制赋值时，如果这些特殊成员函数可以访问，就会调用它们，并且如果它们没有被用户定义，那么它们就会被编译器隐式定义。
- 如果类管理的资源的句柄是个非类类型对象（原始指针、POSIX 文件描述符等），其析构函数不执行任何操作，且复制构造函数/赋值运算符进行“浅复制”（复制句柄的值，而不复制底层资源），则不应使用隐式定义的特殊成员函数。

```c++
#include <cstddef>
#include <cstring>
#include <iostream>
#include <utility>
 
class rule_of_three
{
    char* cstring; // 以裸指针为动态分配内存的句柄
 
public:
    explicit rule_of_three(const char* s = "") : cstring(nullptr)
    {
        if (s)
        {
            cstring = new char[std::strlen(s) + 1]; // 分配
            std::strcpy(cstring, s); // 填充
        }
    }
 
    ~rule_of_three() // I. 析构函数
    {
        delete[] cstring; // 解分配
    }
 
    rule_of_three(const rule_of_three& other) // II. 复制构造函数
        : rule_of_three(other.cstring) {}
 
    rule_of_three& operator=(const rule_of_three& other) // III. 复制赋值
    {
        // 为简洁采用复制后交换（copy-and-swap）手法实现
        // 注意这样做妨碍潜在的存储重用
        rule_of_three temp(other);
        std::swap(cstring, temp.cstring);
        return *this;
    }
 
    const char* c_str() const // 访问器
    {
        return cstring;
    }
};
 
int main()
{
    rule_of_three o1{"abc"};
    std::cout << o1.c_str() << ' ';
    auto o2{o1}; // II. 使用复制构造函数
    std::cout << o2.c_str() << ' ';
    rule_of_three o3("def");
    std::cout << o3.c_str() << ' ';
    o3 = o2; // III. 使用赋值赋值
    std::cout << o3.c_str() << '\n';
}   // I. 此处调用所有析构函数
```

- 五原则
- 因为用户定义（包括 = default 或 = delete）的析构函数、复制构造函数或复制赋值运算符，会阻止隐式定义移动构造函数和移动赋值运算符，所以任何想要移动语义的类必须声明全部五个特殊成员函数：
```c++
class rule_of_five
{
    char* cstring; // 以裸指针为动态分配内存的句柄
public:
    explicit rule_of_five(const char* s = "") : cstring(nullptr)
    {
        if (s)
        {
            cstring = new char[std::strlen(s) + 1]; // 分配
            std::strcpy(cstring, s); // 填充
        }
    }
 
    ~rule_of_five()
    {
        delete[] cstring; // 解分配
    }
 
    rule_of_five(const rule_of_five& other) // 复制构造函数
        : rule_of_five(other.cstring) {}
 
    rule_of_five(rule_of_five&& other) noexcept // 移动构造函数
        : cstring(std::exchange(other.cstring, nullptr)) {}
 
    rule_of_five& operator=(const rule_of_five& other) // 复制赋值
    {
        // 为简便而实现为从临时副本移动赋值
        // 注意这样做妨碍了潜在的存储重用
        return *this = rule_of_five(other);
    }
 
    rule_of_five& operator=(rule_of_five&& other) noexcept // 移动赋值
    {
        std::swap(cstring, other.cstring);
        return *this;
    }
 
// 此外，还可以将两个赋值运算符都改为复制后交换（copy-and-swap）实现，
// 这样做仍会无法在复制赋值中重用存储。
//  rule_of_five& operator=(rule_of_five other) noexcept
//  {
//      std::swap(cstring, other.cstring);
//      return *this;
//  }
};
```

- 与三原则不同的是，不提供移动构造函数和移动赋值运算符通常不是错误，而只是错过了优化机会。


- 零原则
- 有自定义析构函数、复制/移动构造函数或复制/移动赋值运算符的类应该专门处理所有权（这遵循单一责任原则）。其他类都不应该拥有自定义的析构函数、复制/移动构造函数或复制/移动赋值运算符[1]
```c++
class rule_of_zero
{
    std::string cppstring;
public:
    rule_of_zero(const std::string& arg) : cppstring(arg) {}
};
```
- 当有意将某个基类用于多态用途时，可能需要将它的析构函数声明为 public 和 virtual。由于这会阻止生成隐式移动（并弃用隐式复制），因此必须将各特殊成员函数定义为 = default[2]。
```c++
class base_of_five_defaults
{
public:
    base_of_five_defaults(const base_of_five_defaults&) = default;
    base_of_five_defaults(base_of_five_defaults&&) = default;
    base_of_five_defaults& operator=(const base_of_five_defaults&) = default;
    base_of_five_defaults& operator=(base_of_five_defaults&&) = default;
    virtual ~base_of_five_defaults() = default;
};
```
- 然而这使得类有可能被切片，这是多态类经常把复制定义 = delete 的原因。