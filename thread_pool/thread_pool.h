#ifndef __THREAD_POOL_H__
#define __THREAD_POOL_H__

#include <thread>
#include <queue>
#include <mutex>
#include <condition_variable>
#include <future>

namespace tools {


class thread_pool {
public:
  thread_pool();
  ~thread_pool();


private:
  using task_queue = std::queue<std::function<void()>>;

  /**
  * @brief 线程运行函数
  */
  void run();

  /**
   * 任务入队函数
   */
  template<typename F,typename... Args>
  auto execute(F&& func,Args&&... args) -> std::future<typename std::result_of<F(Args...)>::type>;

  task_queue                task_queue_;
  std::mutex                mutex_;
  std::condition_variable   condition_;
};

template<typename F,typename... Args>
auto thread_pool::execute(F&& func,Args&&... args) -> std::future<typename std::result_of<F(Args...)>::type>{
  
}


}

#endif // __THREAD_POOL_H__