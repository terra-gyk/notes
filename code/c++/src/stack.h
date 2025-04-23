#include <execinfo.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <cstring>  

// 堆栈回溯函数
void print_stacktrace() {
    const int max_frames = 100;
    void* frame_addrs[max_frames];
    
    // 获取当前堆栈帧
    int n_frames = backtrace(frame_addrs, max_frames);
    
    // 解析符号表
    char** symbols = backtrace_symbols(frame_addrs, n_frames);
    if (symbols) {
        printf("Stack trace:\n");
        for (int i = 0; i < n_frames; ++i) {
            printf("%d: %s\n", i, symbols[i]);
        }
        free(symbols);
    }
}

// 信号处理函数
void signal_handler(int signal) {
    printf("Caught signal %d (%s)\n", signal, strsignal(signal));
    print_stacktrace();
    exit(EXIT_FAILURE);
}

int main() {
    // 注册信号处理函数
    signal(SIGSEGV, signal_handler);  // 段错误
    signal(SIGABRT, signal_handler);  // abort()调用
    
    // 触发崩溃（示例）
    int* ptr = nullptr;
    *ptr = 10;  // 空指针解引用，触发SIGSEGV
    
    return 0;
}