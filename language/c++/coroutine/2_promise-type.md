# C++ Coroutines: Understanding the promise type

Sep 5, 2018

This post is the third in the series on the C++ Coroutines TS ([N4736](http://wg21.link/N4736)).

The previous articles in this series cover:

- [Coroutine Theory](https://lewissbaker.github.io/2017/09/25/coroutine-theory)
- [Understanding operator co_await](https://lewissbaker.github.io/2017/11/17/understanding-operator-co-await)

In this post I look at the mechanics of how the compiler translates coroutine code that you write into compiled code and how you can customise the behaviour of a coroutine by defining your own **Promise** type.

---
这篇帖子是关于C++协程技术规范（[N4736](http://wg21.link/N4736)）系列的第三篇。

该系列之前的两篇文章涵盖了：

- [协程理论](https://lewissbaker.github.io/2017/09/25/coroutine-theory)
- [理解`operator co_await`](https://lewissbaker.github.io/2017/11/17/understanding-operator-co-await)

在这篇文章中，我将探讨编译器如何将你编写的协程代码转换为编译后的代码，以及如何通过定义自己的**Promise**类型来自定义协程的行为。

---

## Coroutine Concepts

The Coroutines TS adds three new keywords: `co_await`, `co_yield` and `co_return`. Whenever you use one of these coroutine keywords in the body of a function this triggers the compiler to compile this function as a coroutine rather than as a normal function.

The compiler applies some fairly mechanical transformations to the code that you write to turn it into a state-machine that allows it to suspend execution at particular points within the function and then later resume execution.

In the previous post I described the first of two new interfaces that the Coroutines TS introduces: The **Awaitable** interface. The second interface that the TS introduces that is important to this code transformation is the **Promise** interface.

The **Promise** interface specifies methods for customising the behaviour of the coroutine itself. The library-writer is able to customise what happens when the coroutine is called, what happens when the coroutine returns (either by normal means or via an unhandled exception) and customise the behaviour of any `co_await` or `co_yield` expression within the coroutine.

---
协程技术规范（Coroutines TS）添加了三个新关键字：`co_await`、`co_yield`和`co_return`。每当你在函数体中使用这些协程关键字之一时，这会触发编译器将该函数编译为协程而不是普通函数。

编译器对所写的代码应用了一些相当机械的转换，将其转变为一个状态机，从而使函数能够在特定点挂起执行，然后稍后恢复执行。

在上一篇文章中，我描述了协程技术规范引入的两个新接口中的第一个：**Awaitable**接口。TS引入的对这种代码转换至关重要的第二个接口是**Promise**接口。

**Promise**接口指定了用于自定义协程本身行为的方法。库作者能够自定义当协程被调用时发生什么，当协程返回时（无论是通过正常方式还是由于未处理的异常）发生什么，并自定义协程内任何`co_await`或`co_yield`表达式的行为。

---

## Promise objects

The **Promise** object defines and controls the behaviour of the coroutine itself by implementing methods that are called at specific points during execution of the coroutine.

> Before we go on, I want you to try and rid yourself of any preconceived notions of what a “promise” is. While, in some use-cases, the coroutine promise object does indeed act in a similar role to the `std::promise` part of a `std::future` pair, for other use-cases the analogy is somewhat stretched. It may be easier to think about the coroutine’s promise object as being a “coroutine state controller” object that controls the behaviour of the coroutine and can be used to track its state.

An instance of the promise object is constructed within the coroutine frame for each invocation of a coroutine function.

The compiler generates calls to certain methods on the promise object at key points during execution of the coroutine.

In the following examples, assume that the promise object created in the coroutine frame for a particular invocation of the coroutine is `promise`.

When you write a coroutine function that has a body, `<body-statements>`, which contains one of the coroutine keywords (`co_return`, `co_await`, `co_yield`) then the body of the coroutine is transformed to something (roughly) like the following:

---
每次调用协程函数时，都会在协程框架内构造一个promise对象实例。

编译器在协程执行的关键点生成对`promise`对象上某些方法的调用。

在以下示例中，假设为特定协程调用创建的协程框架中的promise对象是`promise`。

当你编写一个包含协程体`<body-statements>`的协程函数时，该协程体包含其中一个协程关键字（`co_return`、`co_await`、`co_yield`），则协程体会被转换为类似以下的内容：

---
```c++
{
  co_await promise.initial_suspend();
  try
  {
    <body-statements>
  }
  catch (...)
  {
    promise.unhandled_exception();
  }
FinalSuspend:
  co_await promise.final_suspend();
}
```

When a coroutine function is called there are a number of steps that are performed prior to executing the code in the source of the coroutine body that are a little different to regular functions.

Here is a summary of the steps (I’ll go into more detail on each of the steps below).

1. Allocate a coroutine frame using `operator new` (optional).
2. Copy any function parameters to the coroutine frame.
3. Call the constructor for the promise object of type, `P`.
4. Call the `promise.get_return_object()` method to obtain the result to return to the caller when the coroutine first suspends. Save the result as a local variable.
5. Call the `promise.initial_suspend()` method and `co_await` the result.
6. When the `co_await promise.initial_suspend()` expression resumes (either immediately or asynchronously), then the coroutine starts executing the coroutine body statements that you wrote.

Some additional steps are executed when execution reaches a `co_return` statement:

1. Call `promise.return_void()` or `promise.return_value(<expr>)`
2. Destroy all variables with automatic storage duration in reverse order they were created.
3. Call `promise.final_suspend()` and `co_await` the result.

If instead, execution leaves `<body-statements>` due to an unhandled exception then:

1. Catch the exception and call `promise.unhandled_exception()` from within the catch-block.
2. Call `promise.final_suspend()` and `co_await` the result.

Once execution propagates outside of the coroutine body then the coroutine frame is destroyed. Destroying the coroutine frame involves a number of steps:

1. Call the destructor of the promise object.
2. Call the destructors of the function parameter copies.
3. Call `operator delete` to free the memory used by the coroutine frame (optional)
4. Transfer execution back to the caller/resumer.

When execution first reaches a `<return-to-caller-or-resumer>` point inside a `co_await` expression, or if the coroutine runs to completion without hitting a `<return-to-caller-or-resumer>` point, then the coroutine is either suspended or destroyed and the return-object previously returned from the call to `promise.get_return_object()` is then returned to the caller of the coroutine.

---
当调用协程函数时，在执行协程体中的代码之前会执行一些步骤，这些步骤与普通函数的执行有所不同。以下是这些步骤的概述（我将在下面详细说明每个步骤）：

1. **分配协程框架**：使用`operator new`分配协程框架（可选）。
2. **复制函数参数**：将任何函数参数复制到协程框架中。
3. **构造Promise对象**：调用类型为`P`的promise对象的构造函数。
4. **获取返回对象**：调用`promise.get_return_object()`方法以获取在协程首次挂起时返回给调用者的对象，并将其保存为局部变量。
5. **初始挂起**：调用`promise.initial_suspend()`方法，并对结果执行`co_await`。
6. **开始执行协程体**：当`co_await promise.initial_suspend()`表达式恢复（可能是立即恢复或异步恢复）时，协程开始执行你编写的协程体语句。

当执行到达`co_return`语句时，会执行一些额外的步骤：

1. **返回值处理**：调用`promise.return_void()`或`promise.return_value(<expr>)`。
2. **销毁自动存储期变量**：按创建顺序的逆序销毁所有具有自动存储期的变量。
3. **最终挂起**：调用`promise.final_suspend()`并对结果执行`co_await`。

如果执行由于未处理的异常离开`<body-statements>`，则会执行以下步骤：

1. **捕获异常并处理**：在catch块中捕获异常并调用`promise.unhandled_exception()`。
2. **最终挂起**：调用`promise.final_suspend()`并对结果执行`co_await`。

一旦执行传播到协程体之外，协程框架就会被销毁。销毁协程框架涉及以下步骤：

1. **销毁Promise对象**：调用promise对象的析构函数。
2. **销毁函数参数副本**：调用函数参数副本的析构函数。
3. **释放内存**：调用`operator delete`释放协程框架使用的内存（可选）。
4. **返回控制权**：将控制权转移回调用者或恢复者。

当执行首次到达`co_await`表达式中的`<return-to-caller-or-resumer>`点，或者如果协程运行至完成而没有遇到`<return-to-caller-or-resumer>`点，则协程要么被挂起要么被销毁，并且之前从`promise.get_return_object()`调用返回的返回对象将返回给协程的调用者。

这些步骤确保了协程能够正确地管理其生命周期、状态和资源，同时允许库作者通过自定义Promise对象来控制协程的具体行为。

---
### Allocating a coroutine frame

First, the compiler generates a call to `operator new` to allocate memory for the coroutine frame.

If the promise type, `P`, defines a custom `operator new` method then that is called, otherwise the global `operator new` is called.

There are a few important things to note here:

The size passed to `operator new` is not `sizeof(P)` but is rather the size of the entire coroutine frame and is determined automatically by the compiler based on the number and sizes of parameters, size of the promise object, number and sizes of local variables and other compiler-specific storage needed for management of coroutine state.

The compiler is free to elide the call to `operator new` as an optimisation if:

- it is able to determine that the lifetime of the coroutine frame is strictly nested within the lifetime of the caller; and
- the compiler can see the size of coroutine frame required at the call-site.

In these cases, the compiler can allocate storage for the coroutine frame in the caller’s activation frame (either in the stack-frame or coroutine-frame part).

The Coroutines TS does not yet specify any situations in which the allocation elision is guaranteed, so you still need to write code as if the allocation of the coroutine frame may fail with `std::bad_alloc`. This also means that you usually shouldn’t declare a coroutine function as `noexcept` unless you are ok with `std::terminate()` being called if the coroutine fails to allocate memory for the coroutine frame.

There is a fallback, however, that can be used in lieu of exceptions for handling failure to allocate the coroutine frame. This can be necessary when operating in environments where exceptions are not allowed, such as embedded environments or high-performance environments where the overhead of exceptions is not tolerated.

If the promise type provides a static `P::get_return_object_on_allocation_failure()` member function then the compiler will generate a call to the `operator new(size_t, nothrow_t)` overload instead. If that call returns `nullptr` then the coroutine will immediately call `P::get_return_object_on_allocation_failure()` and return the result to the caller of the coroutine instead of throwing an exception.

---
首先，编译器生成对`operator new`的调用以分配协程框架所需的内存。

如果Promise类型`P`定义了自定义的`operator new`方法，则调用该方法；否则，调用全局的`operator new`。

这里有几个重要的注意事项：

- 传递给`operator new`的大小不是`sizeof(P)`，而是整个协程框架的大小。这个大小由编译器根据参数的数量和大小、Promise对象的大小、局部变量的数量和大小以及其他编译器特定的管理协程状态所需存储自动确定。
  
- 编译器可以在某些情况下优化掉对`operator new`的调用：
  - 如果编译器能够确定协程框架的生命周期严格嵌套在调用者的生命周期内；
  - 并且编译器能够在调用点看到所需的协程框架大小。

  在这些情况下，编译器可以在调用者的激活帧（无论是栈帧还是协程帧部分）中为协程框架分配存储空间。

协程技术规范（Coroutines TS）目前还没有指定任何保证分配省略的情况，因此你仍然需要编写代码，假设协程框架的分配可能会失败并抛出`std::bad_alloc`异常。这也意味着除非你接受在协程无法分配内存时调用`std::terminate()`，否则通常不应该将协程函数声明为`noexcept`。

然而，有一个备用方案可以用于处理协程框架分配失败的情况，而不需要使用异常。这在不允许使用异常的环境中是必要的，例如嵌入式环境或高性能环境中，其中异常的开销是不可容忍的。

如果Promise类型提供了静态成员函数`P::get_return_object_on_allocation_failure()`，那么编译器会生成对`operator new(size_t, nothrow_t)`重载的调用。如果该调用返回`nullptr`，则协程将立即调用`P::get_return_object_on_allocation_failure()`，并将结果返回给协程的调用者，而不是抛出异常。 

通过这种方式，即使在不允许使用异常的环境中，也可以安全地处理协程框架分配失败的情况。

---

#### Customising coroutine frame memory allocation

Your promise type can define an overload of `operator new()` that will be called instead of global-scope `operator new` if the compiler needs to allocate memory for a coroutine frame that uses your promise type.

For example:

---
你的Promise类型可以定义一个`operator new()`的重载版本，如果编译器需要为使用你的Promise类型的协程框架分配内存时，将会调用这个重载版本而不是全局作用域的`operator new`。

例如：

---

```c++
struct my_promise_type
{
  void* operator new(std::size_t size)
  {
    void* ptr = my_custom_allocate(size);
    if (!ptr) throw std::bad_alloc{};
    return ptr;
  }

  void operator delete(void* ptr, std::size_t size)
  {
    my_custom_free(ptr, size);
  }

  ...
};
```

“But what about custom allocators?”, I hear you asking.

You can also provide an overload of `P::operator new()` that takes additional arguments which will be called with lvalue references to the coroutine function parameters if a suitable overload can be found. This can be used to hook up `operator new` to call an `allocate()` method on an allocator that was passed as an argument to the coroutine function.

You will need to do some extra work to make a copy of the allocator inside the allocated memory so you can reference it in the corresponding call to `operator delete` since the parameters are not passed to the corresponding `operator delete` call. This is because the parameters are stored in the coroutine-frame and so they will have already been destructed by the time that `operator delete` is called.

For example, you can implement `operator new` so that it allocates extra space after the coroutine frame and use that space to stash a copy of the allocator that can be used to free the coroutine frame memory.

For example:

---
“但是自定义分配器呢？”我听到你在问。

你还可以提供一个重载的`P::operator new()`，它接受额外的参数，这些参数将以协程函数参数的左值引用传递，前提是能找到合适的重载。这可以用于将`operator new`挂钩到调用一个在协程函数中作为参数传递的分配器的`allocate()`方法。

你需要做一些额外的工作来在分配的内存中复制分配器，以便在相应的`operator delete`调用中引用它，因为参数不会传递给相应的`operator delete`调用。这是因为参数存储在协程框架中，因此在调用`operator delete`时它们已经被销毁了。

例如，你可以实现`operator new`，使其在协程框架之后分配额外的空间，并使用该空间来存储分配器的一个副本，以便稍后用于释放协程框架的内存。

例如：

---

```c++
template<typename ALLOCATOR>
struct my_promise_type
{
  template<typename... ARGS>
  void* operator new(std::size_t sz, std::allocator_arg_t, ALLOCATOR& allocator, ARGS&... args)
  {
    // Round up sz to next multiple of ALLOCATOR alignment
    std::size_t allocatorOffset =
      (sz + alignof(ALLOCATOR) - 1u) & ~(alignof(ALLOCATOR) - 1u);

    // Call onto allocator to allocate space for coroutine frame.
    void* ptr = allocator.allocate(allocatorOffset + sizeof(ALLOCATOR));

    // Take a copy of the allocator (assuming noexcept copy constructor here)
    new (((char*)ptr) + allocatorOffset) ALLOCATOR(allocator);

    return ptr;
  }

  void operator delete(void* ptr, std::size_t sz)
  {
    std::size_t allocatorOffset =
      (sz + alignof(ALLOCATOR) - 1u) & ~(alignof(ALLOCATOR) - 1u);

    ALLOCATOR& allocator = *reinterpret_cast<ALLOCATOR*>(
      ((char*)ptr) + allocatorOffset);

    // Move allocator to local variable first so it isn't freeing its
    // own memory from underneath itself.
    // Assuming allocator move-constructor is noexcept here.
    ALLOCATOR allocatorCopy = std::move(allocator);

    // But don't forget to destruct allocator object in coroutine frame
    allocator.~ALLOCATOR();

    // Finally, free the memory using the allocator.
    allocatorCopy.deallocate(ptr, allocatorOffset + sizeof(ALLOCATOR));
  }
}
```

To hook up the custom `my_promise_type` to be used for coroutines that pass `std::allocator_arg` as the first parameter, you need to specialise the `coroutine_traits` class (see section on `coroutine_traits` below for more details).

For example:

---
要将自定义的`my_promise_type`连接到以`std::allocator_arg`作为第一个参数传递的协程，你需要特化`coroutine_traits`类（详见下面关于`coroutine_traits`的部分）。

例如：

---

```c++
namespace std::experimental
{
  template<typename ALLOCATOR, typename... ARGS>
  struct coroutine_traits<my_return_type, std::allocator_arg_t, ALLOCATOR, ARGS...>
  {
    using promise_type = my_promise_type<ALLOCATOR>;
  };
}
```

Note that even if you customise the memory allocation strategy for a coroutine, **the compiler is still allowed to elide the call to your memory allocator**.

---
请注意，即使你自定义了协程的内存分配策略，**编译器仍然可以省略对你的内存分配器的调用**。

---

### Copying parameters to the coroutine frame

The coroutine needs to copy any parameters passed to the coroutine function by the original caller into the coroutine frame so that they remain valid after the coroutine is suspended.

If parameters are passed to the coroutine by value, then those parameters are copied to the coroutine frame by calling the type’s move-constructor.

If parameters are passed to the coroutine by reference (either lvalue or rvalue), then only the references are copied into the coroutine frame, not the values they point to.

Note that for types with trivial destructors, the compiler is free to elide the copy of the parameter if the parameter is never referenced after a reachable `<return-to-caller-or-resumer>` point in the coroutine.

There are many gotchas involved when passing parameters by reference into coroutines as you cannot necessarily rely on the reference remaining valid for the lifetime of the coroutine. Many common techniques used with normal functions, such as perfect-forwarding and universal-references, can result in code that has undefined behaviour if used with coroutines. Toby Allsopp has written a [great article](https://toby-allsopp.github.io/2017/04/22/coroutines-reference-params.html) on this topic if you want more details.

If any of the parameter copy/move constructors throws an exception then any parameters already constructed are destructed, the coroutine frame is freed and the exception propagates back out to the caller.

---
协程需要将原始调用者传递给协程函数的参数复制到协程框架中，以确保在协程挂起后这些参数仍然有效。

如果参数是通过值传递给协程的，那么这些参数会通过调用类型的移动构造函数被复制到协程框架中。

如果参数是通过引用（无论是左值引用还是右值引用）传递给协程的，则只有这些引用会被复制到协程框架中，而不是它们指向的值。

请注意，对于具有平凡析构函数的类型，如果参数在协程的可达`<return-to-caller-or-resumer>`点之后不再被引用，编译器可以自由地省略该参数的复制。

当通过引用将参数传递给协程时，存在许多陷阱，因为你不能依赖引用在整个协程生命周期内保持有效。许多与普通函数一起使用的常见技术，如完美转发和通用引用，如果与协程一起使用可能会导致未定义行为。如果你想了解更多细节，Toby Allsopp写了一篇[很好的文章](https://toby-allsopp.github.io/2017/04/22/coroutines-reference-params.html)讨论这个话题。

如果任何参数的复制/移动构造函数抛出异常，则已经构造的任何参数都会被析构，协程框架会被释放，并且异常会传播回调用者。

---

### Constructing the promise object

Once all of the parameters have been copied into the coroutine frame, the coroutine then constructs the promise object.

The reason the parameters are copied prior to the promise object being constructed is to allow the promise object to be given access to the post-copied parameters in its constructor.

First, the compiler checks to see if there is an overload of the promise constructor that can accept lvalue references to each of the copied parameters. If the compiler finds such an overload then the compiler generates a call to that constructor overload. If it does not find such an overload then the compiler falls back to generating a call to the promise type’s default constructor.

Note that the ability for the promise constructor to “peek” at the parameters was a relatively recent change to the Coroutines TS, being adopted in [N4723](http://wg21.link/N4723) at the Jacksonville 2018 meeting. See [P0914R1](http://wg21.link/P0914R1) for the proposal. Thus it may not be supported by some older versions of Clang or MSVC.

If the promise constructor throws an exception then the parameter copies are destructed and the coroutine frame freed during stack unwinding before the exception propagates out to the caller.

---
一旦所有参数都被复制到协程框架中，协程随后会构造Promise对象。

参数在Promise对象构造之前被复制的原因是为了允许Promise对象在其构造函数中访问已复制的参数。

首先，编译器检查是否存在一个可以接受每个已复制参数的左值引用的Promise构造函数重载。如果编译器找到这样的重载，则生成对该构造函数重载的调用。如果没有找到这样的重载，则编译器回退到生成对Promise类型的默认构造函数的调用。

请注意，Promise构造函数能够“窥视”参数的能力是Coroutines TS中相对较新的变化，在2018年杰克逊维尔会议上通过[N4723](http://wg21.link/N4723)被采纳。参见[P0914R1](http://wg21.link/P0914R1)了解提案详情。因此，一些较旧版本的Clang或MSVC可能不支持这一功能。

如果Promise构造函数抛出异常，则在异常传播回调用者之前，参数副本会在栈展开期间被析构，并释放协程框架。

---

### Obtaining the return object

The first thing a coroutine does with the promise object is obtain the `return-object` by calling `promise.get_return_object()`.

The `return-object` is the value that is returned to the caller of the coroutine function when the coroutine first suspends or after it runs to completion and execution returns to the caller.

You can think of the control flow going something (very roughly) like this:

---
协程对Promise对象做的第一件事是通过调用`promise.get_return_object()`获取`return-object`。

`return-object`是当协程第一次挂起或运行完成并返回执行到调用者时，返回给协程函数调用者的值。

你可以将控制流想象成如下（非常粗略地）进行：

---

```c++
// Pretend there's a compiler-generated structure called 'coroutine_frame'
// that holds all of the state needed for the coroutine. Its constructor
// takes a copy of parameters and default-constructs a promise object.
struct coroutine_frame { ... };

T some_coroutine(P param)
{
  auto* f = new coroutine_frame(std::forward<P>(param));

  auto returnObject = f->promise.get_return_object();

  // Start execution of the coroutine body by resuming it.
  // This call will return when the coroutine gets to the first
  // suspend-point or when the coroutine runs to completion.
  coroutine_handle<decltype(f->promise)>::from_promise(f->promise).resume();

  // Then the return object is returned to the caller.
  return returnObject;
}
```

Note that we need to obtain the return-object before starting the coroutine body since the coroutine frame (and thus the promise object) may be destroyed prior to the call to `coroutine_handle::resume()` returning, either on this thread or possibly on another thread, and so it would be unsafe to call `get_return_object()` after starting execution of the coroutine body.

---
请注意，我们需要在启动协程体之前获取`return-object`，因为协程框架（以及因此的Promise对象）可能在`coroutine_handle::resume()`调用返回之前被销毁，无论是在当前线程还是可能在另一个线程上，所以在启动协程体的执行之后再调用`get_return_object()`将是不安全的。

---

### The initial-suspend point

The next thing the coroutine executes once the coroutine frame has been initialised and the return object has been obtained is execute the statement `co_await promise.initial_suspend();`.

This allows the author of the `promise_type` to control whether the coroutine should suspend before executing the coroutine body that appears in the source code or start executing the coroutine body immediately.

If the coroutine suspends at the initial suspend point then it can be later resumed or destroyed at a time of your choosing by calling `resume()` or `destroy()` on the coroutine’s `coroutine_handle`.

The result of the `co_await promise.initial_suspend()` expression is discarded so implementations should generally return `void` from the `await_resume()` method of the awaiter.

It is important to note that this statement exists outside of the `try`/`catch` block that guards the rest of the coroutine (scroll back up to the definition of the coroutine body if you’ve forgotten what it looks like). This means that any exception thrown from the `co_await promise.initial_suspend()` evaluation prior to hitting its `<return-to-caller-or-resumer>` will be thrown back to the caller of the coroutine after destroying the coroutine frame and the return object.

Be aware of this if your `return-object` has RAII semantics that destroy the coroutine frame on destruction. If this is the case then you want to make sure that `co_await promise.initial_suspend()` is `noexcept` to avoid double-free of the coroutine frame.

> Note that there is a proposal to tweak the semantics so that either all or part of the `co_await promise.initial_suspend()` expression lies inside try/catch block of the coroutine-body so the exact semantics here are likely to change before coroutines are finalised.

For many types of coroutine, the `initial_suspend()` method either returns `std::experimental::suspend_always` (if the operation is lazily started) or `std::experimental::suspend_never` (if the operation is eagerly started) which are both `noexcept` awaitables so this is usually not an issue.

---
一旦协程框架初始化完成并获取了返回对象，协程接下来执行的语句是`co_await promise.initial_suspend();`。

这允许`promise_type`的作者控制协程是否应在执行源代码中出现的协程体之前挂起，或者立即开始执行协程体。

如果协程在初始挂起点挂起，则可以通过调用协程的`coroutine_handle`上的`resume()`或`destroy()`在你选择的时间恢复或销毁协程。

`co_await promise.initial_suspend()`表达式的结果被丢弃，因此实现通常应从awaiter的`await_resume()`方法返回`void`。

需要注意的是，此语句位于保护协程其余部分的`try`/`catch`块之外（如果你忘记了它的样子，请回顾协程体的定义）。这意味着，在到达其`<return-to-caller-or-resumer>`点之前，从`co_await promise.initial_suspend()`评估抛出的任何异常将在销毁协程框架和返回对象后抛回调用者。

如果你的`return-object`具有在销毁时销毁协程框架的RAII语义，请注意这一点。如果是这种情况，你需要确保`co_await promise.initial_suspend()`是`noexcept`，以避免协程框架的双重释放。

> 注意，有一个提案建议调整语义，使`co_await promise.initial_suspend()`表达式的全部或部分位于协程体的`try`/`catch`块内，因此在协程最终确定之前，这里的精确语义可能会发生变化。

对于许多类型的协程，`initial_suspend()`方法要么返回`std::experimental::suspend_always`（如果操作是惰性启动的），要么返回`std::experimental::suspend_never`（如果操作是积极启动的），这两者都是`noexcept`的可等待对象，因此这通常不是问题。

---

### Returning to the caller

When the coroutine function reaches its first `<return-to-caller-or-resumer>` point (or if no such point is reached then when execution of the coroutine runs to completion) then the `return-object` returned from the `get_return_object()` call is returned to the caller of the coroutine.

Note that the type of the `return-object` doesn’t need to be the same type as the return-type of the coroutine function. An implicit conversion from the `return-object` to the return-type of the coroutine is performed if necessary.

> Note that Clang’s implementation of coroutines (as of 5.0) defers executing this conversion until the return-object is returned from the coroutine call, whereas MSVC’s implementation as of 2017 Update 3 performs the conversion immediately after calling `get_return_object()`. Although the Coroutines TS is not explicit on the intended behaviour, I believe MSVC has plans to change their implementation to behave more like Clang’s as this enables some [interesting use cases](https://github.com/toby-allsopp/coroutine_monad).

---
当协程函数到达其第一个`<return-to-caller-or-resumer>`点时（或者如果没有这样的点，则在协程执行完成时），从`get_return_object()`调用返回的`return-object`将返回给协程的调用者。

请注意，`return-object`的类型不需要与协程函数的返回类型相同。如果必要，会隐式地将`return-object`转换为协程的返回类型。

> 需要注意的是，截至5.0版本，Clang的协程实现将此转换推迟到`return-object`从协程调用中返回时执行，而截至2017 Update 3，MSVC的实现则在调用`get_return_object()`之后立即执行转换。尽管协程技术规范（Coroutines TS）对预期行为没有明确说明，但我相信MSVC计划更改其实现以使其行为更像Clang，因为这可以启用一些[有趣的用例](https://github.com/toby-allsopp/coroutine_monad)。

---

### Returning from the coroutine using `co_return`

When the coroutine reaches a `co_return` statement, it is translated into either a call to `promise.return_void()` or `promise.return_value(<expr>)` followed by a `goto FinalSuspend;`.

The rules for the translation are as follows:

- `co_return;`
  -> `promise.return_void();`
- `co_return <expr>;`
  -> `<expr>; promise.return_void();` if `<expr>` has type `void`
  -> `promise.return_value(<expr>);` if `<expr>` does not have type `void`

The subsequent `goto FinalSuspend;` causes all local variables with automatic storage duration to be destructed in reverse order of construction before then evaluating `co_await promise.final_suspend();`.

Note that if execution runs off the end of a coroutine without a `co_return` statement then this is equivalent to having a `co_return;` at the end of the function body. In this case, if the `promise_type` does not have a `return_void()` method then the behaviour is undefined.

If either the evaluation of `<expr>` or the call to `promise.return_void()` or `promise.return_value()` throws an exception then the exception still propagates to `promise.unhandled_exception()` (see below).

---
当协程到达`co_return`语句时，它会被翻译成对`promise.return_void()`或`promise.return_value(<expr>)`的调用，随后是一个`goto FinalSuspend;`。

翻译的规则如下：

- `co_return;`
  -> `promise.return_void();`
  
- `co_return <expr>;`
  -> 如果`<expr>`的类型是`void`，则为`<expr>; promise.return_void();`
  -> 如果`<expr>`的类型不是`void`，则为`promise.return_value(<expr>);`

随后的`goto FinalSuspend;`会导致所有具有自动存储期的局部变量按照构造的逆序被销毁，然后才评估`co_await promise.final_suspend();`。

请注意，如果执行在没有`co_return`语句的情况下运行到协程的末尾，则这相当于在函数体末尾有一个`co_return;`。在这种情况下，如果`promise_type`没有`return_void()`方法，则行为是未定义的。

如果`<expr>`的求值或对`promise.return_void()`或`promise.return_value()`的调用抛出异常，则该异常仍然会传播到`promise.unhandled_exception()`（见下文）。

---

### Handling exceptions that propagate out of the coroutine body

If an exception propagates out of the coroutine body then the exception is caught and the `promise.unhandled_exception()` method is called inside the `catch` block.

Implementations of this method typically call `std::current_exception()` to capture a copy of the exception to store it away to be later rethrown in a different context.

Alternatively, the implementation could immediately rethrow the exception by executing a `throw;` statement. For example see [folly::Optional](https://github.com/facebook/folly/blob/4af3040b4c2192818a413bad35f7a6cc5846ed0b/folly/Optional.h#L587) However, doing so will (likely - see below) cause the the coroutine frame to be immediately destroyed and for the exception to propagate out to the caller/resumer. This could cause problems for some abstractions that assume/require the call to `coroutine_handle::resume()` to be `noexcept`, so you should generally only use this approach when you have full control over who/what calls `resume()`.

Note that the current [Coroutines TS](http://wg21.link/N4736) wording is a [little unclear](https://github.com/GorNishanov/CoroutineWording/issues/17) on the intended behaviour if the call to `unhandled_exception()` rethrows the exception (or for that matter if any of the logic outside of the try-block throws an exception).

My current interpretation of the wording is that if control exits the coroutine-body, either via exception propagating out of `co_await promise.initial_suspend()`, `promise.unhandled_exception()` or `co_await promise.final_suspend()` or by the coroutine running to completion by `co_await p.final_suspend()` completing synchronously then the coroutine frame is automatically destroyed before execution returns to the caller/resumer. However, this interpretation has its own issues.

A future version of the Coroutines specification will hopefully clarify the situation. However, until then I’d stay away from throwing exceptions out of `initial_suspend()`, `final_suspend()` or `unhandled_exception()`. Stay tuned!

---
如果异常从协程体中传播出来，则该异常会被捕获，并在`catch`块内调用`promise.unhandled_exception()`方法。

此方法的实现通常会调用`std::current_exception()`来捕获异常的副本并存储起来，以便稍后在不同的上下文中重新抛出。

或者，实现可以立即通过执行`throw;`语句重新抛出异常。例如，参见[folly::Optional](https://github.com/facebook/folly/blob/4af3040b4c2192818a413bad35f7a6cc5846ed0b/folly/Optional.h#L587)。然而，这样做可能会（很可能——详见下文）导致协程框架立即被销毁，并使异常传播回调用者或恢复者。这可能会对某些假设或要求`coroutine_handle::resume()`为`noexcept`的抽象造成问题，因此通常只有在你完全控制谁/什么调用`resume()`时才应使用这种方法。

需要注意的是，当前的[Coroutines TS](http://wg21.link/N4736)关于如果`unhandled_exception()`调用重新抛出异常（或者对于任何在`try`块外部逻辑抛出异常的情况）的预期行为的描述有点不清楚（参见[GitHub issue](https://github.com/GorNishanov/CoroutineWording/issues/17)）。

根据我对当前措辞的理解，如果控制通过异常从`co_await promise.initial_suspend()`、`promise.unhandled_exception()`或`co_await promise.final_suspend()`中传播出来，或者协程通过同步完成`co_await p.final_suspend()`运行到结束而退出协程体，则在执行返回回调用者或恢复者之前，协程框架会自动被销毁。然而，这种解释本身也存在一些问题。

未来的协程规范版本有望澄清这种情况。但在那之前，建议避免从`initial_suspend()`、`final_suspend()`或`unhandled_exception()`中抛出异常。请继续关注更新！

---

### The final-suspend point

Once execution exits the user-defined part of the coroutine body and the result has been captured via a call to `return_void()`, `return_value()` or `unhandled_exception()` and any local variables have been destructed, the coroutine has an opportunity to execute some additional logic before execution is returned back to the caller/resumer.

The coroutine executes the `co_await promise.final_suspend();` statement.

This allows the coroutine to execute some logic, such as publishing a result, signalling completion or resuming a continuation. It also allows the coroutine to optionally suspend immediately before execution of the coroutine runs to completion and the coroutine frame is destroyed.

Note that it is undefined behaviour to `resume()` a coroutine that is suspended at the `final_suspend` point. The only thing you can do with a coroutine suspended here is `destroy()` it.

The rationale for this limitation, according to Gor Nishanov, is that this provides several optimisation opportunities for the compiler due to the reduction in the number of suspend states that need to be represented by the coroutine and a potential reduction in the number of branches required.

Note that while it is allowed to have a coroutine not suspend at the `final_suspend` point, **it is recommended that you structure your coroutines so that they do suspend at `final_suspend`** where possible. This is because this forces you to call `.destroy()` on the coroutine from outside of the coroutine (typically from some RAII object destructor) and this makes it much easier for the compiler to determine when the scope of the lifetime of the coroutine-frame is nested inside the caller. This in turn makes it much more likely that the compiler can elide the memory allocation of the coroutine frame.

---
一旦执行退出协程体的用户定义部分，并通过调用`return_void()`、`return_value()`或`unhandled_exception()`捕获结果，且所有局部变量已被销毁，协程在执行返回回调用者或恢复者之前有机会执行一些额外的逻辑。

协程执行`co_await promise.final_suspend();`语句。

这允许协程执行一些逻辑，例如发布结果、发出完成信号或恢复延续。它还允许协程在执行完成并销毁协程框架之前选择性地立即挂起。

请注意，在`final_suspend`点挂起的协程上调用`resume()`是未定义行为。对于在此处挂起的协程，唯一可以执行的操作是`destroy()`它。

根据Gor Nishanov的说法，这种限制的理由是，由于需要表示的挂起状态数量减少，这为编译器提供了若干优化机会，并可能减少所需的分支数量。

尽管允许协程不在`final_suspend`点挂起，**建议尽可能将协程结构设计为在`final_suspend`点挂起**。这是因为这迫使你从协程外部（通常是从某些RAII对象的析构函数中）调用`.destroy()`，从而使编译器更容易确定协程框架的生命周期范围是否嵌套在调用者的范围内。这反过来使得编译器更有可能省略协程框架的内存分配。

---

### How the compiler chooses the promise type

So lets look now at how the compiler determines what type of promise object to use for a given coroutine.

The type of the promise object is determined from the signature of the coroutine by using the `std::experimental::coroutine_traits` class.

If you have a coroutine function with signature:

---
现在让我们看看编译器如何确定对给定协程使用哪种类型的Promise对象。

Promise对象的类型是通过`std::experimental::coroutine_traits`类从协程的签名中确定的。

如果你有一个具有以下签名的协程函数：

---

```c++
task<float> foo(std::string x, bool flag);
```

Then the compiler will deduce the type of the coroutine’s promise by passing the return-type and parameter types as template arguments to `coroutine_traits`.

---
然后，编译器将通过将返回类型和参数类型作为模板参数传递给`coroutine_traits`来推导协程的Promise类型。

---

```c++
typename coroutine_traits<task<float>, std::string, bool>::promise_type;
```

If the function is a non-static member function then the class type is passed as the second template parameter to `coroutine_traits`. Note that if your method is overloaded for rvalue-references then the second template parameter will be an rvalue reference.

For example, if you have the following methods:

---
如果函数是非静态成员函数，则类类型将作为第二个模板参数传递给`coroutine_traits`。请注意，如果你的方法被重载以处理右值引用，则第二个模板参数将是右值引用。

例如，如果你有以下方法：

---

```c++
task<void> my_class::method1(int x) const;
task<foo> my_class::method2() &&;
```

Then the compiler will use the following promise types:

---
然后，编译器将使用以下Promise类型：

---

```c++
// method1 promise type
typename coroutine_traits<task<void>, const my_class&, int>::promise_type;

// method2 promise type
typename coroutine_traits<task<foo>, my_class&&>::promise_type;
```

The default definition of `coroutine_traits` template defines the `promise_type` by looking for a nested `promise_type` typedef defined on the return-type. ie. Something like this (but with some extra SFINAE magic so that `promise_type` is not defined if `RET::promise_type` is not defined).

---
`coroutine_traits`模板的默认定义通过查找返回类型上定义的嵌套`promise_type`类型别名来定义`promise_type`。即，类似这样（但带有一些额外的SFINAE魔法，以确保在`RET::promise_type`未定义时不定义`promise_type`）。

---

```c++
namespace std::experimental
{
  template<typename RET, typename... ARGS>
  struct coroutine_traits<RET, ARGS...>
  {
    using promise_type = typename RET::promise_type;
  };
}
```

So for coroutine return-types that you have control over, you can just define a nested `promise_type` in your class to have the compiler use that type as the type of the promise object for coroutines that return your class.

For example:

---
因此，对于你可以控制的协程返回类型，只需在你的类中定义一个嵌套的`promise_type`，就可以让编译器使用该类型作为返回你的类的协程的承诺对象的类型。

例如：

---
```c++
template<typename T>
struct task
{
  using promise_type = task_promise<T>;
  ...
};
```

However, for coroutine return-types that you don’t have control over you can specialise the `coroutine_traits` to define the promise type to use without needing to modify the type.

For example, to define the promise-type to use for a coroutine that returns `std::optional<T>`:

---
然而，对于你无法控制的协程返回类型，你可以专门化`coroutine_traits`以定义要使用的承诺类型，而无需修改该类型。

例如，为返回`std::optional<T>`的协程定义要使用的承诺类型：

---

```c++
namespace std::experimental
{
  template<typename T, typename... ARGS>
  struct coroutine_traits<std::optional<T>, ARGS...>
  {
    using promise_type = optional_promise<T>;
  };
}
```

### Identifying a specific coroutine activation frame

When you call a coroutine function, a coroutine frame is created. In order to resume the associated coroutine or destroy the coroutine frame you need some way to identify or refer to that particular coroutine frame.

The mechanism the Coroutines TS provides for this is the `coroutine_handle` type.

The (abbreviated) interface of this type is as follows:

---
当你调用协程函数时，会创建一个协程帧。为了恢复相关的协程或销毁协程帧，你需要某种方式来标识或引用该特定的协程帧。

协程技术规范为此提供的机制是`coroutine_handle`类型。

此类型的（简化）接口如下：

---

```c++
namespace std::experimental
{
  template<typename Promise = void>
  struct coroutine_handle;

  // Type-erased coroutine handle. Can refer to any kind of coroutine.
  // Doesn't allow access to the promise object.
  template<>
  struct coroutine_handle<void>
  {
    // Constructs to the null handle.
    constexpr coroutine_handle();

    // Convert to/from a void* for passing into C-style interop functions.
    constexpr void* address() const noexcept;
    static constexpr coroutine_handle from_address(void* addr);

    // Query if the handle is non-null.
    constexpr explicit operator bool() const noexcept;

    // Query if the coroutine is suspended at the final_suspend point.
    // Undefined behaviour if coroutine is not currently suspended.
    bool done() const;

    // Resume/Destroy the suspended coroutine
    void resume();
    void destroy();
  };

  // Coroutine handle for coroutines with a known promise type.
  // Template argument must exactly match coroutine's promise type.
  template<typename Promise>
  struct coroutine_handle : coroutine_handle<>
  {
    using coroutine_handle<>::coroutine_handle;

    static constexpr coroutine_handle from_address(void* addr);

    // Access to the coroutine's promise object.
    Promise& promise() const;

    // You can reconstruct the coroutine handle from the promise object.
    static coroutine_handle from_promise(Promise& promise);
  };
}
```

You can obtain a `coroutine_handle` for a coroutine in two ways:
1. It is passed to the `await_suspend()` method during a `co_await` expression.
2. If you have a reference to the coroutine’s promise object, you can reconstruct its `coroutine_handle` using `coroutine_handle<Promise>::from_promise()`.

The `coroutine_handle` of the awaiting coroutine will be passed into the `await_suspend()` method of the awaiter after the coroutine has suspended at the `<suspend-point>` of a `co_await` expression. You can think of this `coroutine_handle` as representing the continuation of the coroutine in a [continuation-passing style](https://en.wikipedia.org/wiki/Continuation-passing_style) call.

Note that the `coroutine_handle` is **NOT** and RAII object. You must manually call `.destroy()` to destroy the coroutine frame and free its resources. Think of it as the equivalent of a `void*` used to manage memory. This is for performance reasons: making it an RAII object would add additional overhead to coroutine, such as the need for reference counting.

You should generally try to use higher-level types that provide the RAII semantics for coroutines, such as those provided by [cppcoro](https://github.com/lewissbaker/cppcoro) (shameless plug), or write your own higher-level types that encapsulate the lifetime of the coroutine frame for your coroutine type.

---
你可以通过两种方式获得协程的`coroutine_handle`：
1. 在`co_await`表达式期间，它被传递给`await_suspend()`方法。
2. 如果你有对协程的promise对象的引用，可以使用`coroutine_handle<Promise>::from_promise()`重构其`coroutine_handle`。

在协程于`co_await`表达式的<suspend-point>处暂停后，等待协程的`coroutine_handle`会被传递到awaiter的`await_suspend()`方法中。你可以将这个`coroutine_handle`视为以[延续传递风格](https://en.wikipedia.org/wiki/Continuation-passing_style)调用中代表协程继续执行的部分。

请注意，`coroutine_handle`**不是**RAII对象。你必须手动调用`.destroy()`来销毁协程帧并释放其资源。可以将其视为用于管理内存的`void*`。这是出于性能考虑：将其作为RAII对象会使协程产生额外的开销，例如需要引用计数。

通常你应该尝试使用提供RAII语义的更高级类型来处理协程，如[cppcoro](https://github.com/lewissbaker/cppcoro)提供的那些（不加掩饰的推荐），或者编写自己的更高级类型来封装你的协程类型的生命周期。

---

### Customising the behaviour of `co_await`

The promise type can optionally customise the behaviour of every `co_await` expression that appears in the body of the coroutine.

By simply defining a method named `await_transform()` on the promise type, the compiler will then transform every `co_await <expr>` appearing in the body of the coroutine into `co_await promise.await_transform(<expr>)`.

This has a number of important and powerful uses:

**It lets you enable awaiting types that would not normally be awaitable.**

For example, a promise type for coroutines with a `std::optional<T>` return-type might provide an `await_transform()` overload that takes a `std::optional<U>` and that returns an awaitable type that either returns a value of type `U` or suspends the coroutine if the awaited value contains `nullopt`.

---
承诺类型可以自定义出现在协程体中的每个`co_await`表达式的行为。

只需在承诺类型中定义一个名为`await_transform()`的方法，编译器就会将协程体中出现的每个`co_await <expr>`转换为`co_await promise.await_transform(<expr>)`。

这有几个重要且强大的用途：

**它允许你启用通常不可等待的类型的等待。**

例如，具有`std::optional<T>`返回类型的协程的承诺类型可能提供一个接受`std::optional<U>`并返回一个可等待类型的`await_transform()`重载，该可等待类型要么返回类型为`U`的值，要么在等待的值包含`nullopt`时挂起协程。

---
```c++
template<typename T>
class optional_promise
{
  ...

  template<typename U>
  auto await_transform(std::optional<U>& value)
  {
    class awaiter
    {
      std::optional<U>& value;
    public:
      explicit awaiter(std::optional<U>& x) noexcept : value(x) {}
      bool await_ready() noexcept { return value.has_value(); }
      void await_suspend(std::experimental::coroutine_handle<>) noexcept {}
      U& await_resume() noexcept { return *value; }
    };
    return awaiter{ value };
  }
};
```

**It lets you disallow awaiting on certain types by declaring `await_transform` overloads as deleted.**

For example, a promise type for `std::generator<T>` return-type might declare a deleted `await_transform()` template member function that accepts any type. This basically disables use of `co_await` within the coroutine.

---
**它允许你通过将`await_transform`重载声明为已删除来禁止对某些类型进行等待。**

例如，具有`std::generator<T>`返回类型的承诺类型可能会声明一个已删除的`await_transform()`模板成员函数，该函数接受任何类型。这基本上禁用了协程内`co_await`的使用。

---
```c++
template<typename T>
class generator_promise
{
  ...

  // Disable any use of co_await within this type of coroutine.
  template<typename U>
  std::experimental::suspend_never await_transform(U&&) = delete;

};
```

**It lets you adapt and change the behaviour of normally awaitable values**

For example, you could define a type of coroutine that ensured that the coroutine always resumed from every `co_await` expression on an associated executor by wrapping the awaitable in a `resume_on()` operator (see `cppcoro::resume_on()`).

---
**它允许你调整和改变通常可等待值的行为。**

例如，你可以定义一种协程类型，通过将可等待对象包装在`resume_on()`操作符中（参见`cppcoro::resume_on()`），确保协程总是从每个`co_await`表达式在相关执行器上恢复。

---

```c++
template<typename T, typename Executor>
class executor_task_promise
{
  Executor executor;

public:

  template<typename Awaitable>
  auto await_transform(Awaitable&& awaitable)
  {
    using cppcoro::resume_on;
    return resume_on(this->executor, std::forward<Awaitable>(awaitable));
  }
};
```

As a final word on `await_transform()`, it’s important to note that if the promise type defines *any* `await_transform()` members then this triggers the compiler to transform *all* `co_await` expressions to call `promise.await_transform()`. This means that if you want to customise the behaviour of `co_await` for just some types that you also need to provide a fallback overload of `await_transform()` that just forwards through the argument.

---
关于`await_transform()`的最后一点，重要的是要注意，如果承诺类型定义了*任何* `await_transform()`成员，则这会触发编译器将*所有* `co_await`表达式转换为调用`promise.await_transform()`。这意味着，如果你只想自定义某些类型的`co_await`行为，你还需要提供一个回传参数的`await_transform()`重载作为备用。

---

### Customising the behaviour of `co_yield`

The final thing you can customise through the promise type is the behaviour of the `co_yield` keyword.

If the `co_yield` keyword appears in a coroutine then the compiler translates the expression `co_yield <expr>` into the expression `co_await promise.yield_value(<expr>)`. The promise type can therefore customise the behaviour of the `co_yield` keyword by defining one or more `yield_value()` methods on the promise object.

Note that, unlike `await_transform`, there is no default behaviour of `co_yield` if the promise type does not define the `yield_value()` method. So while a promise type needs to explicitly opt-out of allowing `co_await` by declaring a deleted `await_transform()`, a promise type needs to opt-in to supporting `co_yield`.

The typical example of a promise type with a `yield_value()` method is that of a `generator<T>` type:

---
通过承诺类型可以自定义的最后一件事是`co_yield`关键字的行为。

如果协程中出现了`co_yield`关键字，编译器会将表达式`co_yield <expr>`转换为表达式`co_await promise.yield_value(<expr>)`。因此，承诺类型可以通过在承诺对象上定义一个或多个`yield_value()`方法来自定义`co_yield`关键字的行为。

请注意，与`await_transform`不同，如果承诺类型未定义`yield_value()`方法，则`co_yield`没有默认行为。因此，虽然承诺类型需要通过声明已删除的`await_transform()`显式地选择不允许`co_await`，但承诺类型需要选择支持`co_yield`。

具有`yield_value()`方法的承诺类型的典型示例是`generator<T>`类型：

---

```c++
template<typename T>
class generator_promise
{
  T* valuePtr;
public:
  ...

  std::experimental::suspend_always yield_value(T& value) noexcept
  {
    // Stash the address of the yielded value and then return an awaitable
    // that will cause the coroutine to suspend at the co_yield expression.
    // Execution will then return from the call to coroutine_handle<>::resume()
    // inside either generator<T>::begin() or generator<T>::iterator::operator++().
    valuePtr = std::addressof(value);
    return {};
  }
};
```

# Summary

In this post I’ve covered the individual transformations that the compiler applies to a function when compiling it as a coroutine.

Hopefully this post will help you to understand how you can customise the behaviour of different types of coroutines through defining different your own promise type. There are a lot of moving parts in the coroutine mechanics and so there are lots of different ways that you can customise their behaviour.

However, there is still one more important transformation that the compiler performs which I have not yet covered - the transformation of the coroutine body into a state-machine. However, this post is already too long so I will defer explaining this to the next post. Stay tuned!

---
在这篇文章中，我介绍了编译器在将函数编译为协程时应用的各个转换。

希望这篇文章能帮助你理解如何通过定义自己的承诺类型来定制不同类型协程的行为。协程机制中有许多活动部件，因此你可以通过很多不同的方式来定制它们的行为。

然而，还有一个重要的转换是编译器执行的，我还没有涵盖——即将协程体转换为状态机的转换。不过，这篇文章已经太长了，所以我将把这个解释留到下一篇文章中。敬请期待！

---