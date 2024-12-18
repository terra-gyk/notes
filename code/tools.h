#ifndef __TOOLS_H__
#define __TOOLS_H__

#include <iostream>

namespace tools{

void print()
{
    std::cout << std::endl;
}

// 这是一个简单的print，用来向print不同类型的数据
template<typename T, typename... Args>
void print(T first, Args... args)
{
    std::cout << first << " ";
    print(args...);
    return ;
}

}


#endif // __TOOLS_H__