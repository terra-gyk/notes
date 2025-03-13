#include <spdlog/spdlog.h>
#include <spdlog/logger.h>
#include <spdlog/sinks/daily_file_sink.h>
#include <zookeeper/zookeeper.h>
#include <thread>

// int main() {
//   // auto daily_logger = spdlog::daily_logger_mt("daily_logger", "logs/daily.txt", 2, 30);
//   // spdlog::set_default_logger(daily_logger);
//   // spdlog::flush_every(std::chrono::seconds(10));
//   spdlog::info(123);
//   SPDLOG_INFO(456);
//   return 0; 
// }

#include "spdlog/async.h"
#include "spdlog/sinks/basic_file_sink.h"
#include <iostream>

void async_example()
{
    try {
        // 可以在创建异步日志记录器之前修改默认线程池设置
        // 这里创建一个包含 8192 个项目和 1 个后台线程的队列
        spdlog::init_thread_pool(8192, 1); 

        // 创建一个异步的基本文件日志记录器
        auto async_file = spdlog::basic_logger_mt<spdlog::async_factory>("async_file_logger", "logs/async_log.txt");
        spdlog::daily_logger_format_mt("async_file_logger", "logs/async_log.txt");

        // 另一种创建异步日志记录器的方式
        // auto async_file = spdlog::create_async<spdlog::sinks::basic_file_sink_mt>("async_file_logger", "logs/async_log.txt");

        // 记录一些日志消息
        for (int i = 1; i < 101; ++i) {
            async_file->info("Async message #{}", i);
        }

        // 刷新日志，确保所有消息都被写入文件
        async_file->flush();

    } catch (const spdlog::spdlog_ex& ex) {
        std::cout << "Log init failed: " << ex.what() << std::endl;
    }
}


void watcher(zhandle_t *zzh, int type, int state, const char *path, void *watcherCtx) {
    if (state == ZOO_CONNECTED_STATE) {
        printf("Connected to ZooKeeper server.\n");
    }
}

int main(int argc, char **argv) {
    const char *host = "8.138.98.54:2181";
    int timeout = 30000;
    zhandle_t *zh = zookeeper_init(host, watcher, timeout, 0, NULL, 0);
    if (zh == NULL) {
        fprintf(stderr, "Failed to initialize ZooKeeper handle.\n");
        return EXIT_FAILURE;
    }

    // 保持程序运行一段时间，以便观察连接状态
    std::this_thread::sleep_for(std::chrono::seconds(10));

    zookeeper_close(zh);
    return EXIT_SUCCESS;
}