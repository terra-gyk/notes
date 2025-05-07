# C++ Coroutines: Understanding Symmetric Transfer

May 11, 2020

The Coroutines TS provided a wonderful way to write asynchronous code as if you were writing synchronous code. You just need to sprinkle `co_await` at appropriate points and the compiler takes care of suspending the coroutine, preserving state across suspend-points and resuming execution of the coroutine later when the operation completes.

However, the Coroutines TS, as it was originally specified, had a nasty limitation that could easily lead to stack-overflow if you weren’t careful. And if you wanted to avoid this stack-overflow then you had to introduce extra synchronisation overhead to safely guard against this in your `task<T>` type.

Thankfully, a tweak was made to the design of coroutines in 2018 to add a capability called “symmetric transfer” which allows you to suspend one coroutine and resume another coroutine without consuming any additional stack-space. The addition of this capability lifted a key limitation of the Coroutines TS and allows for much simpler and more efficient implementation of async coroutine types without sacrificing any of the safety aspects needed to guard against stack-overflow.

In this post I will attempt to explain the stack-overflow problem and how the addition of this key “symmetric transfer” capability lets us solve this problem.

---
协程技术规范（Coroutines TS）提供了一种奇妙的方法，让你可以像编写同步代码一样编写异步代码。你只需要在适当的位置加上`co_await`，编译器就会负责挂起协程、在挂起点之间保存状态，并在操作完成时恢复协程的执行。

然而，按照最初的规范，协程技术规范有一个讨厌的限制，如果不小心的话很容易导致栈溢出。如果你想避免这种栈溢出，则必须引入额外的同步开销，以在你的`task<T>`类型中安全地防止这种情况发生。

幸运的是，2018年对协程的设计进行了调整，增加了一项称为“对称转移”的功能，允许你在不消耗额外栈空间的情况下挂起一个协程并恢复另一个协程。这一功能的加入消除了协程技术规范的一个关键限制，使得异步协程类型的实现更加简单和高效，同时没有牺牲防止栈溢出所需的安全性。

在这篇文章中，我将尝试解释栈溢出问题以及这个关键的“对称转移”能力如何帮助我们解决这个问题。

---

## First some background on how a task coroutine works

Consider the following coroutines:

```c++
task foo() {
  co_return;
}

task bar() {
  co_await foo();
}
```

Assume we have a simple `task` type that lazily executes the body when another coroutine awaits it. This particular `task` type does not support returning a value.

Let’s unpack what’s happening here when `bar()` evaluates `co_await foo()`.

- The `bar()` coroutine calls the `foo()` function. Note that from the caller’s perspective a coroutine is just an ordinary function.
- The invocation of `foo()` performs a few steps:
  - Allocates storage for a coroutine frame (typically on the heap)
  - Copies parameters into the coroutine frame (in this case there are no parameters so this is a no-op).
  - Constructs the promise object in the coroutine frame
  - Calls `promise.get_return_object()` to get the return-value for `foo()`. This produces the `task` object that will be returned, initialising it with a `std::coroutine_handle` that refers to the coroutine frame that was just created.
  - Suspends execution of the coroutine at the initial-suspend point (ie. the open curly brace)
  - Returns the `task` object back to `bar()`.
- Next the `bar()` coroutine evaluates the `co_await` expression on the `task` returned from `foo()`.
  - The `bar()` coroutine suspends and then calls the `await_suspend()` method on the returned task, passing it the `std::coroutine_handle` that refers to `bar()`’s coroutine frame.
  - The `await_suspend()` method then stores `bar()`’s `std::coroutine_handle` in `foo()`’s promise object and then resumes the `foo()` coroutine by calling `.resume()` on `foo()`’s `std::coroutine_handle`.
- The `foo()` coroutine executes and runs to completion synchronously.
- The `foo()` coroutine suspends at the final-suspend point (ie. the closing curly brace) and then resumes the coroutine identified by the `std::coroutine_handle` that was stored in its promise object before it was started. ie. `bar()`’s coroutine.
- The `bar()` coroutine resumes and continues executing and eventually reaches the end of the statement containing the `co_await` expression at which point it calls the destructor of the temporary `task` object returned from `foo()`.
- The `task` destructor then calls the `.destroy()` method on `foo()`’s coroutine handle which then destroys the coroutine frame along with the promise object and copies of any arguments.

Ok, so that seems like a lot of steps for a simple call.

To help understand this in a bit more depth, let’s look at how a naive implementation of this `task` class would look when implemented using the the Coroutines TS design (which didn’t support symmetric transfer).

---
假设我们有一个简单的`task`类型，当另一个协程等待它时懒惰地执行其主体。这个特定的`task`类型不支持返回值。

让我们解开当`bar()`评估`co_await foo()`时发生了什么：

- `bar()`协程调用了`foo()`函数。请注意，从调用者的角度来看，协程只是一个普通函数。
- 调用`foo()`执行几个步骤：
  - 分配协程帧的存储（通常在堆上）
  - 将参数复制到协程帧中（在这种情况下没有参数，因此这是空操作）。
  - 在协程帧中构造承诺对象
  - 调用`promise.get_return_object()`以获取`foo()`的返回值。这将生成一个`task`对象，并通过一个指向刚刚创建的协程帧的`std::coroutine_handle`初始化它。
  - 在初始挂起点（即打开的大括号处）挂起协程的执行。
  - 返回`task`对象给`bar()`。
- 接下来，`bar()`协程对从`foo()`返回的`task`上的`co_await`表达式进行求值：
  - `bar()`协程挂起，然后调用返回的任务上的`await_suspend()`方法，传递一个指向`bar()`协程帧的`std::coroutine_handle`。
  - `await_suspend()`方法然后将`bar()`的`std::coroutine_handle`存储在`foo()`的承诺对象中，然后通过调用`foo()`的`std::coroutine_handle`上的`.resume()`恢复`foo()`协程。
- `foo()`协程执行并同步完成。
- `foo()`协程在最终挂起点（即关闭的大括号处）挂起，然后恢复由保存在其承诺对象中的`std::coroutine_handle`标识的协程，即`bar()`的协程。
- `bar()`协程恢复并继续执行，最终到达包含`co_await`表达式的语句末尾，在这一点上调用从`foo()`返回的临时`task`对象的析构函数。
- `task`析构函数随后调用`foo()`的协程句柄上的`.destroy()`方法，销毁协程帧及其承诺对象和任何参数的副本。

好的，对于一个简单的调用来说，这看起来有很多步骤。

为了更深入地理解这一点，让我们看看使用协程技术规范设计（不支持对称转移）实现的简单`task`类的朴素实现会是什么样子。

---

## Outline of a `task` implementation

The outline of the class looks something like this:

```c++
class task {
public:
  class promise_type { /* see below */ };

  task(task&& t) noexcept
  : coro_(std::exchange(t.coro_, {}))
  {}

  ~task() {
    if (coro_)
      coro_.destroy();
  }

  class awaiter { /* see below */ };

  awaiter operator co_await() && noexcept;

private:
  explicit task(std::coroutine_handle<promise_type> h) noexcept
  : coro_(h)
  {}

  std::coroutine_handle<promise_type> coro_;
};
```

A `task` has exclusive ownership of the `std::coroutine_handle` that corresponds to the coroutine frame created during the invocation of the coroutine. The `task` object is an RAII object that ensures that `.destroy()` is called on the `std::coroutine_handle` when the `task` object goes out of scope.

So now let’s expand on the `promise_type`.

---
`task`对在协程调用期间创建的协程帧对应的`std::coroutine_handle`拥有独占所有权。`task`对象是一个RAII对象，确保当`task`对象超出作用域时在`std::coroutine_handle`上调用`.destroy()`。

现在让我们详细说明`promise_type`。

---

## Implementing `task::promise_type`

From the [previous post](https://lewissbaker.github.io/2018/09/05/understanding-the-promise-type) we know that the `promise_type` member defines the type of the **Promise** object that is created within the coroutine frame and that controls the behaviour of the coroutine.

First, we need to implement the `get_return_object()` to construct the `task` object to return when the coroutine is invoked. This method just needs to initialise the task with the `std::coroutine_handle` of the newly create coroutine frame.

We can use the `std::coroutine_handle::from_promise()` method to manufacture one of these handles from the promise object.

---
从[前一篇文章](https://lewissbaker.github.io/2018/09/05/understanding-the-promise-type)我们知道，`promise_type`成员定义了在协程框架内创建的**Promise**对象的类型，并控制协程的行为。

首先，我们需要实现`get_return_object()`来构造当协程被调用时返回的`task`对象。此方法只需要用新创建的协程框架的`std::coroutine_handle`初始化任务。

我们可以使用`std::coroutine_handle::from_promise()`方法从promise对象制造这些句柄之一。

---

```c++
class task::promise_type {
public:
  task get_return_object() noexcept {
    return task{std::coroutine_handle<promise_type>::from_promise(*this)};
  }
```

Next, we want the coroutine to initially suspend at the open curly brace so that we can later resume the coroutine from this point when the returned `task` is awaited.

There are several benefits of starting the coroutine lazily:

1. It means that we can attach the continuation’s `std::coroutine_handle` before starting execution of the coroutine. This means we don’t need to use thread-synchronisation to arbitrate the race between attaching the continuation later and the coroutine running to completion.
2. It means that the `task` destructor can unconditionally destroy the coroutine frame - we don’t need to worry about whether the coroutine is potentially executing on another thread since the coroutine will not start executing until we await it, and while it is executing the calling coroutine is suspended and so won’t attempt to call the task destructor until the coroutine finishes executing. This gives the compiler a much better chance at inlining the allocation of the coroutine frame into the frame of the caller. See [P0981R0](https://wg21.link/P0981R0) to read more about the Heap Allocation eLision Optimisation (HALO).
3. It also improves the exception-safety of your coroutine code. If you don’t immediately `co_await` the returned `task` and do something else that can throw an exception that causes the stack to unwind and the `task` destructor to run then we can safely destroy the coroutine since we know it hasn’t started yet. We aren’t left with the difficult choice between detaching, potentially leaving dangling references, blocking in the destructor, terminating or undefined-behaviour. This is something that I cover in a bit more detail in my [CppCon 2019 talk on Structured Concurrency](https://www.youtube.com/watch?v=1Wy5sq3s2rg).

To have the coroutine initially suspend at the open curly brace we define an `initial_suspend()` method that returns the builtin `suspend_always` type.

---
接下来，我们希望协程在打开大括号处初始挂起，以便当返回的`task`被等待时可以从这一点恢复协程。

懒启动协程有几个好处：

1. 这意味着我们可以在开始执行协程之前附加延续的`std::coroutine_handle`。这意味着我们不需要使用线程同步来仲裁稍后附加延续与协程运行至完成之间的竞争。
2. 这意味着`task`析构函数可以无条件地销毁协程框架——我们不需要担心协程是否可能在另一个线程上执行，因为协程直到我们等待它时才会开始执行，在它执行期间调用协程会被挂起，因此不会尝试调用`task`析构函数，直到协程执行完毕。这给了编译器更好的机会将协程框架的分配内联到调用者的框架中。更多关于堆分配省略优化（HALO）的信息，请参阅[P0981R0](https://wg21.link/P0981R0)。
3. 它也提高了协程代码的异常安全性。如果你不立即对返回的`task`执行`co_await`并做其他可能抛出异常的操作，导致栈展开和`task`析构函数运行，那么我们可以安全地销毁协程，因为我们知道它还未开始执行。我们不必在分离、可能留下悬空引用、在析构函数中阻塞、终止或未定义行为之间做出艰难的选择。我在我的[CppCon 2019关于结构化并发的演讲](https://www.youtube.com/watch?v=1Wy5sq3s2rg)中对此有更详细的介绍。

要使协程在打开大括号处初始挂起，我们定义一个返回内置`suspend_always`类型的`initial_suspend()`方法。

---
```c++
  std::suspend_always initial_suspend() noexcept {
    return {};
  }
```

Next, we need to define the `return_void()` method, called when you execute `co_return;` or when execution runs off the end of the coroutine. This method doesn’t actually need to do anything, it just needs to exist so that the compiler knows that `co_return;` is valid within this coroutine type.

---
接下来，我们需要定义`return_void()`方法，当执行`co_return;`或执行运行到协程末尾时调用。此方法实际上不需要做任何事情，它只需要存在，以便编译器知道在这个协程类型中`co_return;`是有效的。

---

```c++
  void return_void() noexcept {}
```

We also need to add an `unhandled_exception()` method that is called if an exception escapes the body of the coroutine. For our purposes we can just treat the task coroutine bodies as `noexcept` and call `std::terminate()` if this happens.

---
我们还需要添加一个`unhandled_exception()`方法，如果协程主体中有异常逃逸时会调用此方法。对于我们的目的，我们可以将任务协程主体视为`noexcept`，如果发生这种情况，调用`std::terminate()`。

---
```c++
  void unhandled_exception() noexcept {
    std::terminate();
  }
```

Finally, when the coroutine execution reaches the closing curly brace, we want the coroutine to suspend at the final-suspend point and then resume its continuation. ie. the coroutine that is awaiting the completion of this coroutine.

To support this, we need a data-member in the promise to hold the `std::coroutine_handle` of the continuation. We also need to define the `final_suspend()` method that returns an awaitable object that will resume the continuation after the current coroutine has suspended at the final-suspend point.

It’s important to delay resuming the continuation until after the current coroutine has suspended because the continuation may go on to immediately call the `task` destructor which will call `.destroy()` on the coroutine frame. The `.destroy()` method is only valid to call on a suspended coroutine and so it would be undefined-behaviour to resume the continuation before the current coroutine has suspended.

The compiler inserts code to evaluate the statement `co_await promise.final_suspend();` at the closing curly brace.

It’s important to note that the coroutine is not yet in a suspended state when the `final_suspend()` method is invoked. We need to wait until the `await_suspend()` method on the returned awaitable is called before the coroutine is suspended.

---
最后，当协程执行到达关闭大括号时，我们希望协程在最终挂起点挂起，然后恢复其延续，即恢复等待此协程完成的协程。

为此，我们需要在promise中添加一个数据成员来保存延续的`std::coroutine_handle`。我们还需要定义`final_suspend()`方法，该方法返回一个可等待对象，在当前协程在最终挂起点挂起后将恢复延续。

重要的是要延迟恢复延续直到当前协程已经挂起，因为延续可能会立即调用`task`析构函数，这将对协程框架调用`.destroy()`。只有在协程已挂起的情况下调用`.destroy()`才是有效的，因此在当前协程挂起之前恢复延续会导致未定义行为。

编译器会在关闭大括号处插入代码来评估语句`co_await promise.final_suspend();`。

需要注意的是，当调用`final_suspend()`方法时，协程尚未处于挂起状态。我们需要等到返回的可等待对象上的`await_suspend()`方法被调用后，协程才会挂起。

---

```c++
  struct final_awaiter {
    bool await_ready() noexcept {
      return false;
    }

    void await_suspend(std::coroutine_handle<promise_type> h) noexcept {
      // The coroutine is now suspended at the final-suspend point.
      // Lookup its continuation in the promise and resume it.
      h.promise().continuation.resume();
    }

    void await_resume() noexcept {}
  };

  final_awaiter final_suspend() noexcept {
    return {};
  }

  std::coroutine_handle<> continuation;
};
```

Ok, so that’s the complete `promise_type`. The final piece we need to implement is the `task::operator co_await()`.

---
Ok，这就是完整的`promise_type`。我们最后需要实现的是`task::operator co_await()`。

---

## Implementing `task::operator co_await()`

You may remember from the [Understanding operator co_await() post](https://lewissbaker.github.io/2017/11/17/understanding-operator-co-await) that when evaluating a `co_await` expression, the compiler will generate a call to `operator co_await()`, if one is defined, and then the object returned must have the `await_ready()`, `await_suspend()` and `await_resume()` methods defined.

When a coroutine awaits a `task` we want the awaiting coroutine to always suspend and then, once it has suspended, store the awaiting coroutine’s handle in the promise of the coroutine we are about to resume and then call `.resume()` on the `task`’s `std::coroutine_handle` to start executing the task.

Thus the relatively straight forward code:

---
您可能还记得在[理解operator co_await()帖子](https://lewissbaker.github.io/2017/11/17/understanding-operator-co-await)中提到的，当评估一个`co_await`表达式时，如果定义了`operator co_await()`，编译器会生成对该函数的调用，并且该函数返回的对象必须定义了`await_ready()`、`await_suspend()`和`await_resume()`方法。

当协程等待一个`task`时，我们希望等待的协程总是暂停，然后，在它暂停之后，将等待的协程的句柄存储在我们将要继续执行的协程的承诺中，然后对`task`的`std::coroutine_handle`调用`.resume()`以开始执行任务。

因此相对直接的代码：

---

```c++
class task::awaiter {
public:
  bool await_ready() noexcept {
    return false;
  }

  void await_suspend(std::coroutine_handle<> continuation) noexcept {
    // Store the continuation in the task's promise so that the final_suspend()
    // knows to resume this coroutine when the task completes.
    coro_.promise().continuation = continuation;

    // Then we resume the task's coroutine, which is currently suspended
    // at the initial-suspend-point (ie. at the open curly brace).
    coro_.resume();
  }

  void await_resume() noexcept {}

private:
  explicit awaiter(std::coroutine_handle<task::promise_type> h) noexcept
  : coro_(h)
  {}

  std::coroutine_handle<task::promise_type> coro_;
};

task::awaiter task::operator co_await() && noexcept {
  return awaiter{coro_};
}
```

And thus completes the code necessary for a functional `task` type.

You can see the complete set of code in Compiler Explorer here: https://godbolt.org/z/-Kw6Nf

---
这就是实现功能性`task`类型所需的全部代码。

您可以在Compiler Explorer这里看到完整的代码：https://godbolt.org/z/-Kw6Nf

---

## The stack-overflow problem

The limitation of this implementation arises, however, when you start writing loops within your coroutines and you `co_await` tasks that can potentially complete synchronously within the body of that loop.

For example:

---
然而，当您开始在协程中编写循环，并且`co_await`可能在循环体内同步完成的任务时，这种实现的局限性就出现了。

例如：

---

```c++
task completes_synchronously() {
  co_return;
}

task loop_synchronously(int count) {
  for (int i = 0; i < count; ++i) {
    co_await completes_synchronously();
  }
}
```

With the naive `task` implementation described above, the `loop_synchronously()` function will (probably) work fine when `count` is 10, 1000, or even 100’000. But there will be a value that you can pass that will eventually cause this coroutine to start crashing.

For example, see: https://godbolt.org/z/gy5Q8q which crashes when `count` is 1’000’000.

The reason that this is crashing is because of stack-overflow.

To understand why this code is causing a stack-overflow we need to take a look at what is happening when this code is executing. In particular, what is happening to the stack-frames.

When the `loop_synchronously()` coroutine first starts executing it will be because some other coroutine `co_await`ed the `task` returned. This will in turn suspend the awaiting coroutine and call `task::awaiter::await_suspend()` which will call `resume()` on the task’s `std::coroutine_handle`.

Thus the stack will look something like this when `loop_synchronously()` starts:

---
使用上述简单的`task`实现，当`count`为10、1000甚至100,000时，`loop_synchronously()`函数可能会正常工作。但是存在一个值，当您传入该值时，最终会导致此协程开始崩溃。

例如，参见：https://godbolt.org/z/gy5Q8q 当`count`为1,000,000时会发生崩溃。

崩溃的原因是栈溢出。

要理解为什么这段代码会导致栈溢出，我们需要看一下这段代码在执行时发生了什么，特别是栈帧发生了什么。

当`loop_synchronously()`协程刚开始执行时，是因为其他某个协程对返回的`task`进行了`co_await`操作。这将依次挂起等待的协程并调用`task::awaiter::await_suspend()`，后者会在任务的`std::coroutine_handle`上调用`resume()`。

因此，当`loop_synchronously()`开始执行时，栈看起来像这样：

---

```
           Stack                                                   Heap
+------------------------------+  <-- top of stack   +--------------------------+
| loop_synchronously$resume    | active coroutine -> | loop_synchronously frame |
+------------------------------+                     | +----------------------+ |
| coroutine_handle::resume     |                     | | task::promise        | |
+------------------------------+                     | | - continuation --.   | |
| task::awaiter::await_suspend |                     | +------------------|---+ |
+------------------------------+                     | ...                |     |
| awaiting_coroutine$resume    |                     +--------------------|-----+
+------------------------------+                                          V
|  ....                        |                     +--------------------------+
+------------------------------+                     | awaiting_coroutine frame |
                                                     |                          |
                                                     +--------------------------+
```

> Note: When a coroutine function is compiled the compiler typically splits it into two parts:
>
> 1. the “ramp function” which deals with the construction of the coroutine frame, parameter copying, promise construction and producing the return-value, and
> 2. the “coroutine body” which contains the user-authored logic from the body of the coroutine.
>
> I use the `$resume` suffix to refer to the “coroutine body” part of the coroutine.
>
> A later blog post will go into more detail about this split.

Then when `loop_synchronously()` awaits the `task` returned from `completes_synchronously()` the current coroutine is suspended and calls `task::awaiter::await_suspend()`. The `await_suspend()` method then calls `.resume()` on the coroutine handle corresponding to the `completes_synchronously()` coroutine.

This resumes the `completes_synchronously()` coroutine which then runs to completion synchronously and suspends at the final-suspend point. It then calls `task::promise::final_awaiter::await_suspend()` which calls `.resume()` on the coroutine handle corresponding to `loop_synchronously()`.

The net result of all of this is that if we look at the state of the program just after the `loop_synchronously()` coroutine is resumed and just before the temporary `task` returned by `completes_synchronously()` is destroyed at the semicolon then the stack/heap should look something like this:

---
注意：当编译协程函数时，编译器通常将其分为两个部分：

1. “启动函数”，处理协程帧的构造、参数复制、承诺构造和返回值的生成；
2. “协程体”，包含用户在协程主体中编写的逻辑。

我使用`$resume`后缀来指代协程的“协程体”部分。

后续的博客文章将详细介绍这种拆分。

然后，当`loop_synchronously()`等待从`completes_synchronously()`返回的`task`时，当前协程被挂起并调用`task::awaiter::await_suspend()`。`await_suspend()`方法接着在对应于`completes_synchronously()`协程的协程句柄上调用`.resume()`。

这恢复了`completes_synchronously()`协程，然后该协程同步运行至完成，并在最终挂起点处挂起。之后，它调用`task::promise::final_awaiter::await_suspend()`，后者在对应于`loop_synchronously()`的协程句柄上调用`.resume()`。

所有这些操作的净结果是，如果我们在`loop_synchronously()`协程恢复后立即查看程序状态，并在`completes_synchronously()`返回的临时`task`在分号处销毁之前，堆栈/堆应如下所示：

---

```
           Stack                                                   Heap
+-------------------------------+ <-- top of stack
| loop_synchronously$resume     | active coroutine -.
+-------------------------------+                   |
| coroutine_handle::resume      |            .------'
+-------------------------------+            |
| final_awaiter::await_suspend  |            |
+-------------------------------+            |  +--------------------------+ <-.
| completes_synchronously$resume|            |  | completes_synchronously  |   |
+-------------------------------+            |  | frame                    |   |
| coroutine_handle::resume      |            |  +--------------------------+   |
+-------------------------------+            '---.                             |
| task::awaiter::await_suspend  |                V                             |
+-------------------------------+ <-- prev top  +--------------------------+   |
| loop_synchronously$resume     |     of stack  | loop_synchronously frame |   |
+-------------------------------+               | +----------------------+ |   |
| coroutine_handle::resume      |               | | task::promise        | |   |
+-------------------------------+               | | - continuation --.   | |   |
| task::awaiter::await_suspend  |               | +------------------|---+ |   |
+-------------------------------+               | - task temporary --|---------'
| awaiting_coroutine$resume     |               +--------------------|-----+
+-------------------------------+                                    V
|  ....                         |               +--------------------------+
+-------------------------------+               | awaiting_coroutine frame |
                                                |                          |
                                                +--------------------------+
```

Then the next thing this will do is call the `task` destructor which will destroy the `completes_synchronously()` frame. It will then increment the `count` variable and go around the loop again, creating a new `completes_synchronously()` frame and resuming it.

In effect, what is happening here is that `loop_synchronously()` and `completes_synchronously()` end up recursively calling each other. Each time this happens we end up consuming a bit more stack-space, until eventually, after enough iterations, we overflow the stack and end up in undefined-behaviour land, typically resulting in your program promptly crashing.

Writing loops in coroutines built this way makes it very easy to write functions that perform unbounded recursion without looking like they are doing any recursion.

So, what would the solution look like under the original Coroutines TS design?

---
接下来的操作是调用`task`的析构函数，这将销毁`completes_synchronously()`帧。然后它将增加`count`变量并再次循环，创建一个新的`completes_synchronously()`帧并恢复它。

实际上，这里发生的是`loop_synchronously()`和`completes_synchronously()`最终相互递归调用。每次发生这种情况时，我们都会消耗更多的栈空间，直到经过足够多的迭代后栈溢出，导致未定义行为，通常会使程序崩溃。

以这种方式构建协程中的循环非常容易编写执行无界递归的函数，而看起来并没有进行任何递归。

那么，在原始的协程技术规范（Coroutines TS）设计下，解决方案会是什么样的呢？

---

## The Coroutines TS solution

Ok, so what can we do about this to avoid this kind of unbounded recursion?

With the above implementation we are using the variant of `await_suspend()` that returns `void`. In the Coroutines TS there is also a version of `await_suspend()` that returns `bool` - if it returns `true` then the coroutine is suspended and execution returns to the caller of `resume()`, otherwise if it returns `false` then the coroutine is immediately resumed, but this time without consuming any additional stack-space.

So, to avoid the unbounded mutual recursion what we want to do is make use of the `bool`-returning version of `await_suspend()` to resume the current coroutine by returning `false` from the `task::awaiter::await_suspend()` method if the task completes synchronously instead of resuming the coroutine recursively using `std::coroutine_handle::resume()`.

To implement a general solution for this there are two parts.

1. Inside the `task::awaiter::await_suspend()` method you can start executing the coroutine by calling `.resume()`. Then when the call to `.resume()` returns, check whether the coroutine has run to completion or not. If it has run to completion then we can return `false`, which indicates the awaiting coroutine should immediately resume, or we can return `true`, indicating that execution should return to the caller of `std::coroutine_handle::resume()`.
2. Inside `task::promise_type::final_awaiter::await_suspend()`, which is run when the coroutine runs to completion, we need to check whether the awaiting coroutine has (or will) return `true` from `task::awaiter::await_suspend()` and if so then resume it by calling `.resume()`. Otherwise, we need to avoid resuming the coroutine and notify `task::awaiter::await_suspend()` that it needs to return `false`.

There is an added complication, however, in that it’s possible for a coroutine to start executing on the current thread then suspend and later resume and run to completion on a different thread before the call to `.resume()` returns. Thus, we need to be able to resolve the potential race between part 1 and part 2 above happening concurrently.

We will need to use a `std::atomic` value to decide the winner of the race here.

Now for the code. We can make the following modifications:

---
为了解决这种无界递归问题，我们可以利用返回`bool`的`await_suspend()`版本。如果返回`true`，则协程挂起，并执行返回到调用`resume()`的地方；如果返回`false`，则协程立即恢复，但不会消耗额外的栈空间。

为了避免无界的相互递归，我们需要做的是在任务同步完成时，通过从`task::awaiter::await_suspend()`方法返回`false`来恢复当前协程，而不是使用`std::coroutine_handle::resume()`递归地恢复协程。

为了实现这一解决方案，有两个部分：

1. 在`task::awaiter::await_suspend()`方法内部，可以通过调用`.resume()`开始执行协程。然后，在`.resume()`调用返回时，检查协程是否已经运行完成。如果它已经完成，则返回`false`，表示等待的协程应立即恢复；否则返回`true`，表示执行应回到`std::coroutine_handle::resume()`的调用者。
2. 在`task::promise_type::final_awaiter::await_suspend()`中（当协程运行至完成时执行），需要检查等待的协程是否（或将要）从`task::awaiter::await_suspend()`返回`true`，如果是这样，则通过调用`.resume()`来恢复它。否则，需要避免恢复协程，并通知`task::awaiter::await_suspend()`它需要返回`false`。

然而，有一个额外的复杂性：协程可能在当前线程上开始执行，然后挂起，之后在对`.resume()`的调用返回之前，在另一个线程上继续并运行至完成。因此，我们需要能够解决上述第1部分和第2部分并发发生时可能出现的竞争情况。

我们将需要使用一个`std::atomic`值来决定这里竞争的胜者。

现在来看代码，我们可以进行以下修改：

---

```c++
class task::promise_type {
  ...

  std::coroutine_handle<> continuation;
  std::atomic<bool> ready = false;
};

bool task::awaiter::await_suspend(
    std::coroutine_handle<> continuation) noexcept {
  promise_type& promise = coro_.promise();
  promise.continuation = continuation;
  coro_.resume();
  return !promise.ready.exchange(true, std::memory_order_acq_rel);
}

void task::promise_type::final_awaiter::await_suspend(
    std::coroutine_handle<promise_type> h) noexcept {
  promise_type& promise = h.promise();
  if (promise.ready.exchange(true, std::memory_order_acq_rel)) {
    // The coroutine did not complete synchronously, resume it here.
    promise.continuation.resume();
  }
}
```

See the updated example on Compiler Explorer: https://godbolt.org/z/7fm8Za Note how it no longer crashes when executing the `count == 1'000'000` case.

This turns out to be the approach that the `cppcoro::task<T>` [implementation](https://github.com/lewissbaker/cppcoro/blob/master/include/cppcoro/task.hpp) took to avoid the unbounded recursion problem (and still does for some platforms) and it has worked reasonably well.

Woohoo! Problem solved, right? Ship it! Right…?

---
查看Compiler Explorer上的更新示例：https://godbolt.org/z/7fm8Za 注意，在执行`count == 1'000'000`的情况下，它不再崩溃。

这实际上是`cppcoro::task<T>`[实现](https://github.com/lewissbaker/cppcoro/blob/master/include/cppcoro/task.hpp)为避免无界递归问题所采取的方法（并且在某些平台上仍然如此），并且这种方法工作得相当不错。

太好了！问题解决了吧？可以发布了，对吧？……？

---

## The problems

While the above solution does solve the recursion problem it has a couple of drawbacks.

Firstly, it introduces the need for `std::atomic` operations which can be quite costly. There is an atomic exchange on the caller when suspending the awaiting coroutine, and another atomic exchange on the callee when it runs to completion. If your application only ever executes on a single thread then you are paying the cost of the atomic operations for synchronising threads even though it’s never needed.

Secondly, it introduces additional branches. One in the caller, which needs to decide whether to suspend or immediately resume the coroutine, and one in the callee, which needs to decide whether to resume the continuation or suspend.

Note that the cost of this extra branch, and possibly even the atomic operations, would often be dwarfed by the cost of the business logic present in the coroutine. However, coroutines have been advertised as a zero cost abstraction and there have even been people using coroutines to suspend execution of a function to avoid waiting for an L1-cache-miss (see Gor’s great [CppCon talk on nanocoroutines](https://www.youtube.com/watch?v=j9tlJAqMV7U) for more details on this).

Thirdly, and probably most importantly, it introduces some non-determinism in the execution-context that the awaiting coroutine resumes on.

Let’s say I have the following code:

---
虽然上述解决方案确实解决了递归问题，但它也有一些缺点。

首先，它引入了对 `std::atomic` 操作的需求，这可能会相当昂贵。在挂起等待的协程时，调用方需要进行原子交换；当被调用方运行完成时，也需要进行原子交换。如果您的应用程序仅在一个线程上执行，则您正在为同步线程支付原子操作的成本，即使这从来都不是必需的。

其次，它引入了额外的分支。一个是在调用方中，需要决定是挂起还是立即恢复协程；另一个是在被调用方中，需要决定是恢复延续还是挂起。

请注意，这种额外分支的成本，甚至可能是原子操作的成本，通常会被协程中存在的业务逻辑成本所掩盖。然而，协程已被宣传为一种零开销抽象，甚至有人使用协程来暂停函数的执行以避免等待L1缓存未命中（详见Gor关于nanocoroutines的[CppCon演讲](https://www.youtube.com/watch?v=j9tlJAqMV7U)以获取更多细节）。

第三，可能也是最重要的是，它在执行上下文中引入了一些非确定性。

假设我有以下代码： 

---

```c++
cppcoro::static_thread_pool tp;

task foo()
{
  std::cout << "foo1 " << std::this_thread::get_id() << "\n";
  // Suspend coroutine and reschedule onto thread-pool thread.
  co_await tp.schedule();
  std::cout << "foo2 " << std::this_thread::get_id() << "\n";
}

task bar()
{
  std::cout << "bar1 " << std::this_thread::get_id() << "\n";
  co_await foo();
  std::cout << "bar2" << std::this_thread::get_id() << "\n";
}
```

With the original implementation we were guaranteed that the code that runs after `co_await foo()` would run inline on the same thread that `foo()` completed on.

For example, one possible output would have been:

---
使用原始实现，我们确保在 `co_await foo()` 之后运行的代码将在执行完成 `foo()` 的同一个线程上内联运行。

例如，一个可能的输出结果是：

---

```
bar1 1234
foo1 1234
foo2 3456
bar2 3456
```

However, with the changes to use atomics, it’s possible the completion of `foo()` may race with the suspension of `bar()` and this can, in some cases, mean that the code after `co_await foo()` might run on the original thread that `bar()` started executing on.

For example, the following output might now also be possible:

---
然而，随着使用原子操作的更改，`foo()` 的完成可能会与 `bar()` 的挂起发生竞争。这在某些情况下意味着，`co_await foo()` 之后的代码可能在 `bar()` 开始执行的原始线程上运行。

例如，现在以下输出也可能是可能的：

---

```
bar1 1234
foo1 1234
foo2 3456
bar2 1234
```

For many use-cases this behaviour may not make a difference. However, for algorithms whose purpose is to transition execution context this can be problematic.

For example, the `via()` algorithm awaits some Awaitable and then produces it on the specified scheduler’s execution context. A simplified version of this algorithm is shown below.

---
对于许多用例而言，这种行为可能不会有影响。然而，对于那些旨在转换执行上下文的算法来说，这可能会成为一个问题。

例如，`via()`算法等待某个可等待对象，然后在指定调度器的执行上下文中产生它。下面是该算法的一个简化版本。

---

```c++
template<typename Awaitable, typename Scheduler>
task<await_result_t<Awaitable>> via(Awaitable a, Scheduler s)
{
  auto result = co_await std::move(a);
  co_await s.schedule();
  co_return result;
}

task<T> get_value();
void consume(const T&);

task<void> consumer(static_thread_pool::scheduler s)
{
  T result = co_await via(get_value(), s);
  consume(result);
}
```

With the original version the call to `consume()` is always guaranteed to be executed on the thread-pool, `s`. However, with the revised version that uses atomics it’s possible that `consume()` might either be executed on a thread associated with the scheduler, `s`, or on whatever thread the `consumer()` coroutine started execution on.

So how do we solve the stack-overflow problem without the overhead of the atomic operations, extra branches and the non-deterministic resumption context?

---
在原始版本中，对`consume()`的调用始终保证在与线程池`s`相关的线程上执行。然而，在使用原子操作的修订版本中，有可能`consume()`会在与调度器`s`相关的线程上执行，或者在启动`consumer()`协程的任何线程上执行。

那么，我们如何在不增加原子操作、额外分支和非确定性恢复上下文的开销的情况下解决栈溢出问题呢？

---

## Enter “symmetric transfer”!

The paper [P0913R0](https://wg21.link/P0913R0) “Add symmetric coroutine control transfer” by Gor Nishanov (2018) proposed a solution to this problem by providing a facility which allows one coroutine to suspend and then resume another coroutine symmetrically without consuming any additional stack-space.

This paper proposed two key changes:

- Allow returning a `std::coroutine_handle<T>` from `await_suspend()` as a way of indicating that execution should be symmetrically transferred to the coroutine identified by the returned handle.
- Add a `std::experimental::noop_coroutine()` function that returns a special `std::coroutine_handle` that can be returned from `await_suspend()` to suspend the current coroutine and return from the call to `.resume()` instead of transferring execution to another coroutine.

So what do we mean by “symmetric transfer”?

When you resume a coroutine by calling `.resume()` on it’s `std::coroutine_handle` the caller of `.resume()` remains active on the stack while the resumed coroutine executes. When this coroutine next suspends and the call to `await_suspend()` for that suspend-point returns either `void` (indicating unconditional suspend) or `true` (indicating conditional suspend) then call to `.resume()` will return.

This can be thought of as an “asymmetric transfer” of execution to the coroutine and behaves just like an ordinary function call. The caller of `.resume()` can be any function (which may or may not be a coroutine). When that coroutine suspends and returns either `true` or `void` from `await_suspend()` then execution will return from the call to `.resume()` and

Every time we resume a coroutine by calling `.resume()` we create a new stack-frame for the execution of that coroutine.

However, with “symmetric transfer” we are simply suspending one coroutine and resuming another coroutine. There is no implicit caller/callee relationship between the two coroutines - when a coroutine suspends it can transfer execution to any suspended coroutine (including itself) and does not necessarily have to transfer execution back to the previous coroutine when it next suspends or completes.

Let’s look at what the compiler lowers a `co_await` expression to when the awaiter makes use of symmetric-transfer:

---
论文[P0913R0](https://wg21.link/P0913R0)“添加对称协程控制转移”由Gor Nishanov于2018年提出，通过提供一种机制允许一个协程对称地挂起并恢复另一个协程而不消耗任何额外的栈空间来解决这个问题。

该论文提出了两个关键变化：

- 允许从`await_suspend()`返回一个`std::coroutine_handle<T>`作为指示执行应转移到由返回句柄标识的协程的一种方式。
- 添加了一个`std::experimental::noop_coroutine()`函数，它返回一个特殊的`std::coroutine_handle`，可以从`await_suspend()`返回以挂起当前协程，并从调用`.resume()`返回而不是将执行转移到另一个协程。

那么，“对称转移”是什么意思呢？

当你通过调用某个协程的`std::coroutine_handle`上的`.resume()`来恢复协程时，调用`.resume()`的调用方在协程执行时保持活跃在栈上。当此协程下次挂起且针对该挂起点的`await_suspend()`调用返回`void`（表示无条件挂起）或`true`（表示有条件挂起）时，对`.resume()`的调用将会返回。

这可以被视作执行到协程的“非对称转移”，其行为就像普通函数调用一样。调用`.resume()`的可以是任何函数（可能是也可能不是一个协程）。当那个协程挂起并从`await_suspend()`返回`true`或`void`时，执行将从调用`.resume()`返回。

每次我们通过调用`.resume()`来恢复协程时，我们为该协程的执行创建一个新的栈帧。

然而，使用“对称转移”时，我们只是简单地挂起一个协程并恢复另一个协程。这两个协程之间没有隐式的调用者/被调用者关系——当一个协程挂起时，它可以将执行转移到任何挂起的协程（包括其自身），并不一定需要在下次挂起或完成时将执行权转移回之前的协程。

让我们看看当awaiter利用对称转移时编译器如何降低`co_await`表达式的过程：

---

```c++
{
  decltype(auto) value = <expr>;
  decltype(auto) awaitable =
      get_awaitable(promise, static_cast<decltype(value)&&>(value));
  decltype(auto) awaiter =
      get_awaiter(static_cast<decltype(awaitable)&&>(awaitable));
  if (!awaiter.await_ready())
  {
    using handle_t = std::coroutine_handle<P>;

    //<suspend-coroutine>

    auto h = awaiter.await_suspend(handle_t::from_promise(p));
    h.resume();
    //<return-to-caller-or-resumer>
    
    //<resume-point>
  }

  return awaiter.await_resume();
}
```

Let’s zoom in on the key part that differs from other `co_await` forms:

---
让我们聚焦于与其他`co_await`形式不同的关键部分：

---

```c++
auto h = awaiter.await_suspend(handle_t::from_promise(p));
h.resume();
//<return-to-caller-or-resumer>
```

Once the coroutine state-machine is lowered (a topic for another post), the `<return-to-caller-or-resumer>` part basically becomes a `return;` statement which causes the call to `.resume()` that last resumed the coroutine to return to its caller.

This means that we have the situation where we have a call to another function with the same signature, `std::coroutine_handle::resume()`, followed by a `return;` from the current function which is itself the body of a `std::coroutine_handle::resume()` call.

Some compilers, when optimisations are enabled, are able to apply an optimisation that turns calls to other functions the tail-position (ie. just before returning) into tail-calls as long as some conditions are met.

It just so happens that this kind of tail-call optimisation is exactly the kind of thing we want to be able to do to avoid the stack-overflow problem we were encountering before. But instead of being at the mercy of the optimiser as to whether or not the tail-call transformation is perfromed, we want to be able to guarantee that the tail-call transformation occurs, even when optimisations are not enabled.

But first let’s dig into what we mean by tail-calls.

---
一旦协程状态机被降低（这是另一篇文章的主题），`<return-to-caller-or-resumer>`部分基本上变成了一个`return;`语句，这会导致最后一次恢复协程的`.resume()`调用返回到其调用者。

这意味着我们有这样一种情况：有一个对另一个具有相同签名的函数的调用，即`std::coroutine_handle::resume()`，紧接着是从当前函数的返回，而该函数本身是`std::coroutine_handle::resume()`调用体的一部分。

一些编译器在启用优化时，能够应用一种优化，将处于尾部位置（即，在即将返回之前）的其他函数调用转化为尾调用，只要满足某些条件。

这种尾调用优化正是我们想要实现的，以避免之前遇到的栈溢出问题。但是，我们希望能够在即使没有启用优化的情况下也确保尾调用转换的发生，而不是取决于优化器是否执行尾调用转换。

但首先让我们深入了解一下尾调用的含义。

---

### Tail-calls

A tail-call is one where the current stack-frame is popped before the call and the current function’s return address becomes the return-address for the callee. ie. the callee will return directly the the caller of this function.

On X86/X64 architectures this generally means that the compiler will generate code that first pops the current stack-frame and then uses a `jmp` instruction to jump to the called function’s entry-point instead of using a `call` instruction and then popping the current stack-frame after the `call` returns.

This optimisation is generally only possible to do in limited circumstances, however.

In particular, it requires that:

- the calling convention supports tail-calls and is the same for the caller and callee;
- the return-type is the same;
- there are no non-trivial destructors that need to be run after the call before returning to the caller; and
- the call is not inside a try/catch block.

The shape of the symmetric-transfer form of `co_await` has actually been designed specifically to allow coroutines to satisfy all of these requirements. Let’s look at them individually.

**Calling convention** When the compiler lowers a coroutine into machine code it actually splits the coroutine up into two parts: the ramp (which allocates and initialises the coroutine frame) and the body (which contains the state-machine for the user-authored coroutine body).

The function signature of the coroutine (and thus any user-specified calling-convention) affects only the ramp part, whereas the body part is under the control of the compiler and is never directly called by any user-code - only by the ramp function and by `std::coroutine_handle::resume()`.

The calling-convention of the coroutine body part is not user-visible and is entirely up to the compiler and thus it can choose an appropriate calling convention that supports tail-calls and that is used by all coroutine bodies.

**Return type is the same** The return-type for both the source and target coroutine’s `.resume()` method is `void` so this requirement is trivially satisfied.

**No non-trivial destructors** When performing a tail-call we need to be able to free the current stack-frame before calling the target function and this requires the lifetime of all stack-allocated objects to have ended prior to the call.

Normally, this would be problematic as soon as there are any objects with non-trivial destructors in-scope as the lifetime of those objects would not yet have ended and those objects would have been allocated on the stack.

However, when a coroutine suspends it does so without exiting any scopes and the way it achieves this is by placing any objects whose lifetime spans a suspend-point in the coroutine frame rather than allocating them on the stack.

Local variables with lifetimes that do not span a suspend-point may be allocated on the stack, but the lifetime of these objects will have already ended and their destructors will have been called before the coroutine next suspends.

Thus there should be no non-trivial destructors for stack-allocated objects that need to be run after the return of the tail-call.

**Call not inside a try/catch block** This one is a little tricker as within every coroutine there is an implicit try/catch block that encloses the user-authored body of the coroutine.

From the specification, we see that the coroutine is defined as:

---
尾调用是指当前栈帧在调用之前被弹出，并且当前函数的返回地址成为被调用者的返回地址。也就是说，被调用者将直接返回到该函数的调用者。

在X86/X64架构上，这通常意味着编译器会生成首先弹出当前栈帧然后使用`jmp`指令跳转到被调用函数入口点的代码，而不是使用`call`指令并在`call`返回后弹出当前栈帧。

然而，这种优化通常只能在有限的情况下进行：

- 调用约定支持尾调用并且对于调用者和被调用者是一致的；
- 返回类型相同；
- 在调用之后返回给调用者之前不需要运行任何非平凡的析构函数；以及
- 调用不在try/catch块内。

实际上，`co_await`对称传输形式的设计就是为了使协程能够满足所有这些要求。让我们逐一查看这些要求：

**调用约定** 当编译器将协程转换为机器码时，它实际上将协程分为两部分：ramp（分配并初始化协程帧）和body（包含用户编写的协程体的状态机）。

协程的函数签名（因此任何用户指定的调用约定）只影响ramp部分，而body部分由编译器控制且不会被任何用户代码直接调用——仅由ramp函数和`std::coroutine_handle::resume()`调用。

协程body部分的调用约定对用户不可见，并完全取决于编译器，因此它可以选择一个支持尾调用并由所有协程body使用的适当调用约定。

**返回类型相同** 源协程和目标协程的`.resume()`方法的返回类型都是`void`，因此这个要求很容易满足。

**没有非平凡的析构函数** 进行尾调用时，我们需要能够在调用目标函数之前释放当前栈帧，这就要求所有栈分配对象的生命周期在调用前已经结束。

通常情况下，如果有任何具有非平凡析构函数的对象在作用域内，这可能会有问题，因为那些对象的生命周期尚未结束且这些对象会被分配在栈上。

然而，当协程挂起时，它不会退出任何作用域，并通过将跨越暂停点的任何对象放置在协程帧中而不是在栈上分配来实现这一点。

生命周期不跨越暂停点的局部变量可能被分配在栈上，但在协程下次挂起之前，这些对象的生命周期已经结束且它们的析构函数已经被调用。

因此，在尾调用返回后，不应该有需要运行的栈分配对象的非平凡析构函数。

**调用不在try/catch块内** 这一点有点棘手，因为在每个协程中都有一个隐式的try/catch块包围着用户编写的协程体。

根据规范，协程定义为：

---

```c++
{
  promise_type promise;
  co_await promise.initial_suspend();
  try { F; }
  catch (...) { promise.unhandled_exception(); }
final_suspend:
  co_await promise.final_suspend();
}
```

Where `F` is the user-authored part of the coroutine body.

Thus every user-authored `co_await` expression (other than initial/final_suspend) exists within the context of a try/catch block.

However, implementations work around this by actually executing the call to `.resume()` *outside* of the context of the try-block.

I hope to be able to go into this aspect in more detail in another blog post that goes into the details of the lowering of a coroutine to machine-code (this post is already long enough).

> Note, however, that the current wording in the C++ specification is not clear on requiring implementations to do this and it is only a non-normative note that hints that this is something that might be required. Hopefully we’ll be able to fix the specification in the future.

So we see that coroutines performing a symmetric-transfer generally satisfy all of the requirements for being able to perform a tail-call. The compiler guarantees that this will always be a tail-call, regardless of whether optimisations are enabled or not.

This means that by using the `std::coroutine_handle`-returning flavour of `await_suspend()` we can suspend the current coroutine and transfer execution to another coroutine without consuming extra stack-space.

This allows us to write coroutines that mutually and recursively resume each other to an arbitrary depth without fear of overflowing the stack.

This is exactly what we need to fix our `task` implementation.

---
其中 `F` 是协程体中用户编写的部分。

因此，每个用户编写的 `co_await` 表达式（除了初始/最终挂起）都存在于 try/catch 块的上下文中。

然而，实现方式通过实际上在 try 块的上下文之外执行 `.resume()` 调用来绕过这个问题。

我希望能在另一篇详细介绍协程降低为机器码细节的博客文章中进一步探讨这一方面（这篇帖子已经足够长了）。

> 注意，当前 C++ 规范中的措辞并不明确要求实现必须这样做，仅有一个非规范性的注释暗示这可能是必需的。希望我们将来能够修正规范。

由此可见，执行对称传输的协程通常满足进行尾调用的所有要求。编译器保证这将始终是一个尾调用，无论是否启用了优化。

这意味着通过使用返回 `std::coroutine_handle` 的 `await_suspend()` 形式，我们可以挂起当前协程并将执行转移到另一个协程，而不会消耗额外的栈空间。

这使我们能够编写相互递归恢复到任意深度的协程，而不必担心栈溢出问题。

这正是我们需要修复我们的 `task` 实现所需的内容。

---

## `task` revisited

So with the new “symmetric transfer” capability under our belt let’s go back and fix our `task` type implementation.

To do this we need to make changes to the two `await_suspend()` methods in our implementation:

- First so that when we await the task that we perform a symmetric-transfer to resume the task’s coroutine.
- Second so that when the task’s coroutine completes that it performs a symmetric transfer to resume the awaiting coroutine.

To address the await direction we need to change the `task::awaiter` method from this:

---
因此，有了新的“对称转移”功能，让我们回去修复我们的`task`类型实现。

为此，我们需要对我们实现中的两个`await_suspend()`方法进行修改：

- 首先，当我们等待任务时，执行对称转移以恢复任务的协程。
- 其次，当任务的协程完成时，执行对称转移以恢复等待的协程。

为了解决等待方向的问题，我们需要将`task::awaiter`方法从这样改变： 

---

```c++
void task::awaiter::await_suspend(
    std::coroutine_handle<> continuation) noexcept {
  // Store the continuation in the task's promise so that the final_suspend()
  // knows to resume this coroutine when the task completes.
  coro_.promise().continuation = continuation;

  // Then we resume the task's coroutine, which is currently suspended
  // at the initial-suspend-point (ie. at the open curly brace).
  coro_.resume();
}
```

to this:

```c++
std::coroutine_handle<> task::awaiter::await_suspend(
    std::coroutine_handle<> continuation) noexcept {
  // Store the continuation in the task's promise so that the final_suspend()
  // knows to resume this coroutine when the task completes.
  coro_.promise().continuation = continuation;

  // Then we tail-resume the task's coroutine, which is currently suspended
  // at the initial-suspend-point (ie. at the open curly brace), by returning
  // its handle from await_suspend().
  return coro_;
}
```

And to address the return-path we need to update the `task::promise_type::final_awaiter` method from this:

---
为了解决返回路径的问题，我们需要将`task::promise_type::final_awaiter`方法从这样更新：

---

```c++
void task::promise_type::final_awaiter::await_suspend(
    std::coroutine_handle<promise_type> h) noexcept {
  // The coroutine is now suspended at the final-suspend point.
  // Lookup its continuation in the promise and resume it.
  h.promise().continuation.resume();
}
```

to this:

```c++
std::coroutine_handle<> task::promise_type::final_awaiter::await_suspend(
    std::coroutine_handle<promise_type> h) noexcept {
  // The coroutine is now suspended at the final-suspend point.
  // Lookup its continuation in the promise and resume it symmetrically.
  return h.promise().continuation;
}
```

And now we have a `task` implementation that doesn’t suffer from the stack-overflow problem that the `void`-returning `await_suspend` flavour had and that doesn’t have the non-deterministic resumption context problem of the `bool`-returning `await_suspend` flavour had.

---
现在我们有了一个`task`实现，它不会遭受返回`void`的`await_suspend`版本所遇到的栈溢出问题，也不会有返回`bool`的`await_suspend`版本所具有的非确定性恢复上下文问题。

---

### Visualising the stack

Let’s now go back and have a look at our original example:

---
现在让我们回过头来重新看一下我们最初的示例：

---

```c++
task completes_synchronously() {
  co_return;
}

task loop_synchronously(int count) {
  for (int i = 0; i < count; ++i) {
    co_await completes_synchronously();
  }
}
```

When the `loop_synchronously()` coroutine first starts executing it will be because some other coroutine `co_await`ed the `task` returned. This will have been launched by symmetric transfer from some other coroutine, which would have been resumed by a call to `std::coroutine_handle::resume()`.

Thus the stack will look something like this when `loop_synchronously()` starts:

---
当`loop_synchronously()`协程第一次开始执行时，是因为其他某个协程对返回的`task`进行了`co_await`。这将通过来自另一个协程的对称转移启动，该协程是通过调用`std::coroutine_handle::resume()`恢复的。

因此，当`loop_synchronously()`开始执行时，栈看起来会像这样：

---

```
           Stack                                                Heap
+---------------------------+  <-- top of stack   +--------------------------+
| loop_synchronously$resume | active coroutine -> | loop_synchronously frame |
+---------------------------+                     | +----------------------+ |
| coroutine_handle::resume  |                     | | task::promise        | |
+---------------------------+                     | | - continuation --.   | |
|     ...                   |                     | +------------------|---+ |
+---------------------------+                     | ...                |     |
                                                  +--------------------|-----+
                                                                       V
                                                  +--------------------------+
                                                  | awaiting_coroutine frame |
                                                  |                          |
                                                  +--------------------------+
```

Now, when it executes `co_await completes_synchronously()` it will perform a symmetric transfer to `completes_synchronously` coroutine.

It does this by:

- calling the `task::operator co_await()` which then returns the `task::awaiter` object
- then suspends and calls `task::awaiter::await_suspend()` which then returns the `coroutine_handle` of the `completes_synchronously` coroutine.
- then performs a tail-call / jump to `completes_synchronously` coroutine. This pops the `loop_synchronously` frame before activing the `completes_synchronously` frame.

If we now look at the stack just after `completes_synchronously` is resumed it will now look like this:

---
现在，当它执行`co_await completes_synchronously()`时，它将通过对称转移进入`completes_synchronously`协程。

它通过以下步骤实现这一点：

- 调用`task::operator co_await()`，这将返回`task::awaiter`对象
- 然后挂起并调用`task::awaiter::await_suspend()`，这将返回`completes_synchronously`协程的`coroutine_handle`
- 然后执行尾调用/跳转到`completes_synchronously`协程。这会在激活`completes_synchronously`帧之前弹出`loop_synchronously`帧。

如果我们现在查看`completes_synchronously`恢复后的栈，它将看起来像这样：

---

```
              Stack                                          Heap
                                            .-> +--------------------------+ <-.
                                            |   | completes_synchronously  |   |
                                            |   | frame                    |   |
                                            |   | +----------------------+ |   |
                                            |   | | task::promise        | |   |
                                            |   | | - continuation --.   | |   |
                                            |   | +------------------|---+ |   |
                                            `-, +--------------------|-----+   |
                                              |                      V         |
+-------------------------------+ <-- top of  | +--------------------------+   |
| completes_synchronously$resume|     stack   | | loop_synchronously frame |   |
+-------------------------------+ active -----' | +----------------------+ |   |
| coroutine_handle::resume      | coroutine     | | task::promise        | |   |
+-------------------------------+               | | - continuation --.   | |   |
|     ...                       |               | +------------------|---+ |   |
+-------------------------------+               | task temporary     |     |   |
                                                | - coro_       -----|---------`
                                                +--------------------|-----+
                                                                     V
                                                +--------------------------+
                                                | awaiting_coroutine frame |
                                                |                          |
                                                +--------------------------+
```

Note that the number of stack-frames has not grown here.

After the `completes_synchronously` coroutine completes and execution reaches the closing curly brace it will evaluate `co_await promise.final_suspend()`.

This will suspend the coroutine and call `final_awaiter::await_suspend()` which return the continuation’s `std::coroutine_handle` (ie. the handle that points to the `loop_synchronously` coroutine). This will then do a symmetric transfer/tail-call to resume the `loop_synchronously` coroutine.

If we look at the stack just after `loop_synchronously` is resumed then it will look something like this:

---
注意，这里的栈帧数量没有增加。

当`completes_synchronously`协程完成后，在执行到结束大括号时，它将评估`co_await promise.final_suspend()`。

这将挂起协程并调用`final_awaiter::await_suspend()`，该函数返回继续体的`std::coroutine_handle`（即指向`loop_synchronously`协程的句柄）。然后，这将通过对称转移/尾调用来恢复`loop_synchronously`协程。

如果我们查看`loop_synchronously`恢复后的栈，它将看起来像这样： 

---

```
           Stack                                                   Heap
                                                   +--------------------------+ <-.
                                                   | completes_synchronously  |   |
                                                   | frame                    |   |
                                                   | +----------------------+ |   |
                                                   | | task::promise        | |   |
                                                   | | - continuation --.   | |   |
                                                   | +------------------|---+ |   |
                                                   +--------------------|-----+   |
                                                                        V         |
+----------------------------+  <-- top of stack   +--------------------------+   |
| loop_synchronously$resume  | active coroutine -> | loop_synchronously frame |   |
+----------------------------+                     | +----------------------+ |   |
| coroutine_handle::resume() |                     | | task::promise        | |   |
+----------------------------+                     | | - continuation --.   | |   |
|     ...                    |                     | +------------------|---+ |   |
+----------------------------+                     | task temporary     |     |   |
                                                   | - coro_       -----|---------`
                                                   +--------------------|-----+
                                                                        V
                                                   +--------------------------+
                                                   | awaiting_coroutine frame |
                                                   |                          |
                                                   +--------------------------+
```

The first thing the `loop_synchronously` coroutine is going to do once resumed is to call the destructor of the temporary `task` that was returned from the call to `completes_synchronously` when execution reaches the semicolon. This will destroy the coroutine-frame, freeing its memory and leaving us with the following sitution:

---
`loop_synchronously`协程一旦恢复，要做的第一件事就是调用从`completes_synchronously`调用返回的临时`task`的析构函数，当执行到达分号时。这将销毁协程帧，释放其内存，从而留下以下情况：

---

```
           Stack                                                   Heap
+---------------------------+  <-- top of stack   +--------------------------+
| loop_synchronously$resume | active coroutine -> | loop_synchronously frame |
+---------------------------+                     | +----------------------+ |
| coroutine_handle::resume  |                     | | task::promise        | |
+---------------------------+                     | | - continuation --.   | |
|     ...                   |                     | +------------------|---+ |
+---------------------------+                     | ...                |     |
                                                  +--------------------|-----+
                                                                       V
                                                  +--------------------------+
                                                  | awaiting_coroutine frame |
                                                  |                          |
                                                  +--------------------------+
```

We are now back to executing the `loop_synchronously` coroutine and we now have the same number of stack-frames and coroutine-frames as we started, and will do so each time we go around the loop.

Thus we can perform as many iterations of the loop as we want and will only use a constant amount of storage space.

For a full example of the symmetric-transfer version of the `task` type see the following Compiler Explorer link: https://godbolt.org/z/9baieF.

---
我们现在回到执行`loop_synchronously`协程，并且现在我们拥有的栈帧和协程帧数量与开始时相同，每次循环一圈时都会如此。

因此，我们可以执行循环的任意多次迭代，并且只会使用固定数量的存储空间。

关于`task`类型的对称传输版本的完整示例，请参见以下Compiler Explorer链接：[https://godbolt.org/z/9baieF](https://godbolt.org/z/9baieF)。 

---

## Symmetric Transfer as the Universal Form of await_suspend

Now that we see the power and importance of the symmetric-transfer form of the awaitable concept, I want to show you that this form is actually the universal form, which can theoretically replace the `void` and `bool`-returning forms of `await_suspend()`.

But first we need to look at the other piece that the [P0913R0](https://wg21.link/P0913R0) proposal added to the coroutines design: `std::noop_coroutine()`.

---
现在我们看到了对称传输形式的可等待概念的强大功能和重要性，我想向您展示这种形式实际上是通用形式，理论上它可以替代返回`void`和`bool`的`await_suspend()`形式。

但首先，我们需要看一下[P0913R0](https://wg21.link/P0913R0)提案为协程设计添加的另一部分内容：`std::noop_coroutine()`。

---

### Terminating the recursion

With the symmetric-transfer form of coroutines, every time the coroutine suspends it symmetrically resumes another coroutine. This is great as long as you have another coroutine to resume, but sometimes we don’t have another coroutine to execute and just need to suspend and let execution return to the caller of `std::coroutine_handle::resume()`.

Both the `void`-returning and `bool`-returning flavours of `await_suspend()` allow a coroutine to suspend and return from `std::coroutine_handle::resume()`, so how do we do that with the symmetric-transfer flavour?

The answer is by using the special builtin `std::coroutine_handle`, called the “noop coroutine handle” which is produced by the function `std::noop_coroutine()`.

The “noop coroutine handle” is named as such because its `.resume()` implementation is such that it just immediately returns. i.e. resuming the coroutine is a no-op. Typically its implementation contains a single `ret` instruction.

If the `await_suspend()` method returns the `std::noop_coroutine()` handle then instead of transferring execution to the next coroutine, it transfers execution back to the caller of `std::coroutine_handle::resume()`.

---
有了协程的对称转移形式，每次协程挂起时都会对称地恢复另一个协程。只要还有另一个协程可以恢复这很好，但有时候我们没有另一个协程可以执行，只需要挂起并让执行返回到 `std::coroutine_handle::resume()` 的调用者。

`await_suspend()` 的返回 `void` 和返回 `bool` 两种形式都允许协程挂起并从 `std::coroutine_handle::resume()` 返回，那么使用对称转移形式时我们如何做到这一点呢？

答案是使用特殊的内置 `std::coroutine_handle`，称为“空操作协程句柄”，它由函数 `std::noop_coroutine()` 生成。

“空操作协程句柄”之所以这样命名，是因为它的 `.resume()` 实现就是这样立即返回。也就是说，恢复协程是一个空操作。通常其实施包含单一的 `ret` 指令。

如果 `await_suspend()` 方法返回的是 `std::noop_coroutine()` 句柄，则不是将执行转移到下一个协程，而是将执行转移回调用 `std::coroutine_handle::resume()` 的调用者。

---

### Representing the other flavours of `await_suspend()`

With this information in-hand we can now show how to represent the other flavours of `await_suspend()` using the symmetric-transfer form.

The `void`-returning form

---
有了这些信息，我们现在可以展示如何使用对称转移形式表示 `await_suspend()` 的其他形式。

返回 `void` 的形式。

---

```c++
void my_awaiter::await_suspend(std::coroutine_handle<> h) {
  this->coro = h;
  enqueue(this);
}
```

can also be written using both the `bool`-returning form:

---
也可以使用返回 `bool` 的形式来写： 

---

```c++
bool my_awaiter::await_suspend(std::coroutine_handle<> h) {
  this->coro = h;
  enqueue(this);
  return true;
}
```

and can be written using the symmetric-transfer form:

---
并且可以使用对称转移形式来写：

---
```c++
std::noop_coroutine_handle my_awaiter::await_suspend(
    std::coroutine_handle<> h) {
  this->coro = h;
  enqueue(this);
  return std::noop_coroutine();
}
```

The `bool`-returning form:

---
返回 bool 的形式：

---

```c++
bool my_awaiter::await_suspend(std::coroutine_handle<> h) {
  this->coro = h;
  if (try_start(this)) {
    // Operation will complete asynchronously.
    // Return true to transfer execution to caller of
    // coroutine_handle::resume().
    return true;
  }

  // Operation completed synchronously.
  // Return false to immediately resume the current coroutine.
  return false;
}
```

can also be written using the symmetric-transfer form:

---
也可以使用对称转移形式来写：

---

```c++
std::coroutine_handle<> my_awaiter::await_suspend(std::coroutine_handle<> h) {
  this->coro = h;
  if (try_start(this)) {
    // Operation will complete asynchronously.
    // Return std::noop_coroutine() to transfer execution to caller of
    // coroutine_handle::resume().
    return std::noop_coroutine();
  }

  // Operation completed synchronously.
  // Return current coroutine's handle to immediately resume
  // the current coroutine.
  return h;
}
```

### Why have all three flavours?

So why do we still have the `void` and `bool`-returning flavours of `await_suspend()` when we have the symmetric-transfer flavour?

The reason is partly historical, partly pragmatic and partly performance.

The `void`-returning version could be entirely replaced by returning the `std::noop_coroutine_handle` type from `await_suspend()` as this would be an equivalent signal to the compiler that the coroutine is unconditionally transfering execution to the caller of `std::coroutine_handle::resume()`.

That it was kept was, IMO, partly because it was already in-use prior to the introduction of symmetric-transfer and partly because the `void`-form results in less-code/less-typing for the unconditional suspend case.

The `bool`-returning version, however, can have a slight win in terms of optimisability in some cases compared to the symmetric-transfer form.

Consider the case where we have a `bool`-returning `await_suspend()` method that is defined in another translation unit. In this case the compiler can generate code in the awaiting coroutine that will suspend the current coroutine and then conditionally resume it after the call to `await_suspend()` returns by just executing the next piece of code. It knows exactly the piece of code to execute next if `await_suspend()` returns `false`.

With the symmetric-transfer flavour we still need to represent the same outcomes; either return to the caller/resume or resume the current coroutine. Instead of returning `true` or `false` we need to return `std::noop_coroutine()` or the handle to the current coroutine. We can coerce both of these handles into a `std::coroutine_handle<void>` type and return it.

However, now, because the `await_suspend()` method is defined in another translation unit the compiler can’t see what coroutine the returned handle is referring to and so when it resumes the coroutine it now has to perform some more expensive indirect calls and possibly some branches to resume the coroutine, compared to a single branch for the `bool`-returning case.

Now, it’s possible that we might be able to get equivalent performance out of the symmetric transfer version one day. For example, we could write our code in such a way that `await_suspend()` is defined inline but calls a `bool`-returning method that is defined out-of-line and then conditionally returns the appropriate handle.

For example:

---
所以我们为什么在有了对称转移形式的情况下仍然保留了返回 `void` 和 `bool` 的 `await_suspend()` 形式？

原因部分是历史的，部分是实用的，还有部分是性能的。

返回 `void` 的版本可以通过从 `await_suspend()` 返回 `std::noop_coroutine_handle` 类型来完全替代，因为这相当于向编译器发出信号，表明协程无条件地将执行转移到 `std::coroutine_handle::resume()` 的调用者。

之所以保留它，部分是因为在引入对称转移之前它已经被使用，部分是因为对于无条件挂起的情况，返回 `void` 的形式会导致更少的代码/打字。

然而，返回 `bool` 的版本在某些情况下相比对称转移形式，在优化方面可能会有一些优势。

考虑这样一个情况：我们有一个返回 `bool` 的 `await_suspend()` 方法，它定义在另一个翻译单元中。在这种情况下，编译器可以在等待的协程中生成代码，该代码会挂起当前协程，然后通过仅仅执行下一段代码在 `await_suspend()` 返回后有条件地恢复它。如果 `await_suspend()` 返回 `false`，它确切知道接下来要执行哪一段代码。

使用对称转移形式，我们仍然需要表示相同的结果；要么返回给调用者/恢复，要么恢复当前协程。而不是返回 `true` 或 `false`，我们需要返回 `std::noop_coroutine()` 或当前协程的句柄。我们可以将这两个句柄强制转换为 `std::coroutine_handle<void>` 类型并返回它。

然而，现在由于 `await_suspend()` 方法定义在另一个翻译单元中，编译器无法看到返回的句柄指的是哪个协程，因此当它恢复协程时，现在必须执行一些更昂贵的间接调用，并可能进行一些分支以恢复协程，相比之下，返回 `bool` 的情况只需要一个分支。

现在，有可能有一天我们能够让对称转移版本获得等效的性能。例如，我们可以编写代码，使得 `await_suspend()` 被定义为内联但调用一个定义为非内联的返回 `bool` 的方法，然后根据情况返回适当的句柄。

例如：

---

```c++
struct my_awaiter {
  bool await_ready();

  // Compilers should in-theory be able to optimise this to the same
  // as the bool-returning version, but currently don't do this optimisation.
  std::coroutine_handle<> await_suspend(std::coroutine_handle<> h) {
    if (try_start(h)) {
      return std::noop_coroutine();
    } else {
      return h;
    }
  }

  void await_resume();

private:
  // This method is defined out-of-line in a separate translation unit.
  bool try_start(std::coroutine_handle<> h);
}
```

However, current compilers (c. Clang 10) are not currently able to optimise this to as efficient code as the equivalent `bool`-returning version. Having said that, you’re probably not going to notice the difference unless you’re awaiting this in a really tight loop.

So, for now, the general rule is:

- If you need to unconditionally return to `.resume()` caller, use the `void`-returning flavour.
- If you need to conditionally return to `.resume()` caller or resume current coroutine use the `bool`-returning flavour.
- If you need to resume another coroutine use the symmetric-transfer flavour.

---
然而，当前的编译器（例如 Clang 10）还不能将此优化为与返回 `bool` 的版本一样高效的代码。话虽如此，除非你在非常紧凑的循环中等待这个操作，否则你可能不会注意到这种差异。

所以，目前的一般规则是：

- 如果你需要无条件地返回到 `.resume()` 的调用者，使用返回 `void` 的形式。
- 如果你需要有条件地返回到 `.resume()` 的调用者或恢复当前协程，使用返回 `bool` 的形式。
- 如果你需要恢复另一个协程，使用对称转移形式。

---

# Rounding out

The new symmetric transfer capability added to coroutines for C++20 makes it much easier to write coroutines that recursively resume each other without fear of running into stack-overflow. This capability is key to making efficient and safe async coroutine types, such as the `task` one presented here.

This ended up being a much longer than expected post on symmetric transfer. If you made it this far, then thanks for sticking with it! I hope you found it useful.

In the next post, I’ll dive into understanding how the compiler transforms a coroutine function into a state-machine.

---
新增到 C++20 协程中的对称转移能力使得编写递归恢复彼此的协程变得更加容易，而不用担心遇到栈溢出的问题。这种能力是实现高效且安全的异步协程类型（如这里介绍的 `task`）的关键。

这篇文章最终比预期的要长得多，讨论了对称转移。如果你看到了这里，感谢你的坚持！希望你发现它是有用的。

在下一篇文章中，我将深入探讨编译器如何将协程函数转换为状态机。

---

# Thanks

Thanks to Eric Niebler and Corentin Jabot for providing feedback on drafts of this post.

---
感谢 Eric Niebler 和 Corentin Jabot 对本文草稿提供的反馈。

---