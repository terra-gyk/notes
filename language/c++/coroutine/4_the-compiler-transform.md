# C++ Coroutines: Understanding the Compiler Transform

Aug 27, 2022

- [Introduction](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#introduction)
- [Setting the Scene](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#setting-the-scene)
- [Defining the `task` type](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#defining-the-task-type)
- [Step 1: Determining the promise type](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-1-determining-the-promise-type)
- [Step 2: Creating the coroutine state](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-2-creating-the-coroutine-state)
- [Step 3: Call `get_return_object()`](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-3-call-get_return_object)
- [Step 4: The initial-suspend point](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-4-the-initial-suspend-point)
- [Step 5: Recording the suspend-point](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-5-recording-the-suspend-point)
- [Step 6: Implementing `coroutine_handle::resume()` and `coroutine_handle::destroy()`](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-6-implementing-coroutine_handleresume-and-coroutine_handledestroy)
- [Step 7: Implementing `coroutine_handle::promise()` and `from_promise()`](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-7-implementing-coroutine_handlepromisepromise-and-from_promise)
- [Step 8: The beginnings of the coroutine body](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-8-the-beginnings-of-the-coroutine-body)
- [Step 9: Lowering the `co_await` expression](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-9-lowering-the-co_await-expression)
- [Step 10: Implementing `unhandled_exception()`](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-10-implementing-unhandled_exception)
- [Step 11: Implementing `co_return`](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-11-implementing-co_return)
- [Step 12: Implementing `final_suspend()`](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-12-implementing-final_suspend)
- [Step 13: Implementing symmetric-transfer and the noop-coroutine](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#step-13-implementing-symmetric-transfer-and-the-noop-coroutine)
- [One last thing](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#one-last-thing)
- [Tying it all together](https://lewissbaker.github.io/2022/08/27/understanding-the-compiler-transform#tying-it-all-together)

# Introduction

Previous blogs in the series on “Understanding C++ Coroutines” talked about the different kinds of transforms the compiler performs on a coroutine and its `co_await`, `co_yield` and `co_return` expressions. These posts described how each expression was lowered by the compiler to calls to various customisation points/methods on user-defined types.

1. [Coroutine Theory](https://lewissbaker.github.io/2017/09/25/coroutine-theory)
2. [C++ Coroutines: Understanding operator co_await](https://lewissbaker.github.io/2017/11/17/understanding-operator-co-await)
3. [C++ Coroutines: Understanding the promise type](https://lewissbaker.github.io/2018/09/05/understanding-the-promise-type)
4. [C++ Coroutines: Understanding Symmetric Transfer](https://lewissbaker.github.io/2020/05/11/understanding_symmetric_transfer)

However, there was one part of these descriptions that may have left you unsatisfied. The all hand-waved over the concept of a “suspend-point” and said something vague like “the coroutine suspends here” and “the coroutine resumes here” but didn’t really go into detail about what that actually means or how it might be implemented by the compiler.

In this post I am going to go a bit deeper to show how all the concepts from the previous posts come together. I’ll show what happens when a coroutine reaches a suspend-point by walking through the lowering of a coroutine into equivalent non-coroutine, imperative C++ code.

Note that I am not going to describe exactly how a particular compiler lowers coroutines into machine code (compilers have extra tricks up their sleeves here), but rather just one possible lowering of coroutines into portable C++ code.

Warning: This is going to be a fairly deep dive!

---
前几篇关于“理解C++协程”系列的文章讨论了编译器对协程及其 `co_await`、`co_yield` 和 `co_return` 表达式执行的不同转换。这些文章描述了每个表达式是如何被编译器降低为对用户定义类型的各个自定义点/方法的调用。

1. [协程理论](https://lewissbaker.github.io/2017/09/25/coroutine-theory)
2. [C++协程：理解 operator co_await](https://lewissbaker.github.io/2017/11/17/understanding-operator-co-await)
3. [C++协程：理解承诺类型](https://lewissbaker.github.io/2018/09/05/understanding-the-promise-type)
4. [C++协程：理解对称转移](https://lewissbaker.github.io/2020/05/11/understanding_symmetric_transfer)

然而，这些描述中有一部分内容可能会让你感到不满足。所有的内容都对“挂起点”的概念一带而过，并模糊地说类似“协程在这里挂起”和“协程在这里恢复”，但并没有真正深入说明这实际上意味着什么，或者编译器可能如何实现它。

在这篇文章中，我将更深入地探讨，展示前几篇文章中的所有概念是如何结合起来的。我将通过将协程降级为等效的非协程、命令式C++代码来展示当协程到达挂起点时发生了什么。

请注意，我不会描述编译器如何将协程降低为机器代码的具体方式（编译器在这方面有一些额外的技巧），而是仅描述一种可能的将协程降级为可移植C++代码的方式。

警告：这将是一次相当深入的探讨！

---

# Setting the Scene

For starters, let’s assume we have a basic `task` type that acts as both an awaitable and a coroutine return-type. For the sake of simplicity, let’s assume that this coroutine type allows producing a result of type `int` asynchronously.

In this post we are going to walk through how to lower the following coroutine function into C++ code that does not contain any of the coroutine keywords `co_await`, `co_return` so that we can better understand what this means.

---
首先，让我们假设我们有一个基本的 `task` 类型，它既作为一个可等待对象，也作为协程的返回类型。为了简单起见，假设这个协程类型允许异步生成一个 `int` 类型的结果。

在这篇文章中，我们将逐步讲解如何将以下协程函数降低为不包含任何协程关键字 `co_await`、`co_return` 的C++代码，以便我们更好地理解这意味着什么。

---

```c++
// Forward declaration of some other function. Its implementation is not relevant.
task f(int x);

// A simple coroutine that we are going to translate to non-C++ code
task g(int x) {
    int fx = co_await f(x);
    co_return fx * fx;
}
```

# Defining the `task` type

To begin, let us first declare the `task` class that we will be working with.

For the purposes of understanding how the coroutine is lowered, we do not need to know the definitions of the methods for this type. The lowering will just be inserting calls to them.

The definitions of these methods are not complicated, and I will leave them as an exercise for the reader as practice for understanding the previous posts.

---
首先，让我们首先声明我们将要使用的 `task` 类。

为了理解协程是如何被降低的，我们不需要知道这个类型的各个方法的具体定义。降低过程只是插入对这些方法的调用。

这些方法的定义并不复杂，我将它们留作练习，供读者练习理解前几篇文章的内容。

---

```c++
class task {
public:
    struct awaiter;

    class promise_type {
    public:
        promise_type() noexcept;
        ~promise_type();

        struct final_awaiter {
            bool await_ready() noexcept;
            std::coroutine_handle<> await_suspend(
                std::coroutine_handle<promise_type> h) noexcept;
            void await_resume() noexcept;
        };

        task get_return_object() noexcept;
        std::suspend_always initial_suspend() noexcept;
        final_awaiter final_suspend() noexcept;
        void unhandled_exception() noexcept;
        void return_value(int result) noexcept;

    private:
        friend task::awaiter;
        std::coroutine_handle<> continuation_;
        std::variant<std::monostate, int, std::exception_ptr> result_;
    };

    task(task&& t) noexcept;
    ~task();
    task& operator=(task&& t) noexcept;

    struct awaiter {
        explicit awaiter(std::coroutine_handle<promise_type> h) noexcept;
        bool await_ready() noexcept;
        std::coroutine_handle<promise_type> await_suspend(
            std::coroutine_handle<> h) noexcept;
        int await_resume();
    private:
        std::coroutine_handle<promise_type> coro_;
    };

    awaiter operator co_await() && noexcept;

private:
    explicit task(std::coroutine_handle<promise_type> h) noexcept;

    std::coroutine_handle<promise_type> coro_;
};
```

The structure of this task type should be familiar to those that have read the [C++ Coroutines: Understanding Symmetric Transfer](https://lewissbaker.github.io/2020/05/11/understanding_symmetric_transfer) post.

---
这个 `task` 类型的结构对于那些阅读过 [C++协程：理解对称转移](https://lewissbaker.github.io/2020/05/11/understanding_symmetric_transfer) 文章的人来说应该很熟悉。

---

# Step 1: Determining the promise type

```c++
task g(int x) {
    int fx = co_await f(x);
    co_return fx * fx;
}
```

When the compiler sees that this function contains one of the three coroutine keywords (`co_await`, `co_yield` or `co_return`) it starts the coroutine transformation process.

The first step here is determining the `promise_type` to use for this coroutine.

This is determined by substituting the return-type and argument-types of the signature as template arguments to the `std::coroutine_traits` type.

e.g. For our function, `g`, which has return type `task` and a single argument of type `int`, the compiler will look this up using `std::coroutine_traits<task, int>::promise_type`.

Let’s define an alias so we can refer to this type later:

---
当编译器看到这个函数包含三个协程关键字（`co_await`、`co_yield` 或 `co_return`）之一时，它就开始了协程转换过程。

第一步是确定此协程要使用的 `promise_type`。

这是通过将签名的返回类型和参数类型作为模板参数替换到 `std::coroutine_traits` 类型中来确定的。

例如，对于我们的函数 `g`，其返回类型为 `task`，并且有一个 `int` 类型的参数，编译器将使用 `std::coroutine_traits<task, int>::promise_type` 来查找。

让我们定义一个别名，以便稍后引用此类型：

---

```c++
using __g_promise_t = std::coroutine_traits<task, int>::promise_type;
```

**Note: I am using leading double-underscore here to indicate symbols internal to the** **compiler that the compiler generates. Such symbols are reserved by the implementation** **and should \*not\* be used in your own code.**

Now, as we have not specialised `std::coroutine_traits` this will instantiate the primary template which just defines the nested `promise_type` as an alias of the nested `promise_type` name of the return-type. i.e. this should resolve to the type `task::promise_type` in our case.

---
**注意：这里使用前导双下划线来表示编译器内部生成的符号。这些符号由实现保留，不应在您自己的代码中使用。**

现在，由于我们没有专门化 `std::coroutine_traits`，这将实例化主模板，该模板仅将嵌套的 `promise_type` 定义为返回类型的嵌套 `promise_type` 名称的别名。也就是说，在我们的情况下，这应该解析为类型 `task::promise_type`。

---

# Step 2: Creating the coroutine state

A coroutine function needs to preserve the state of the coroutine, parameters and local variables when it suspends so that they remain available when the coroutine is later resumed.

This state, in C++ standardese, is called the *coroutine state* and is typically heap allocated.

Let’s start by defining a struct for the coroutine-state for the coroutine, `g`.

We don’t know what the contents of this type are going to be yet, so let’s just leave it empty for now.

---
协程函数在挂起时需要保存协程的状态、参数和局部变量，以便在稍后恢复协程时它们仍然可用。

这种状态，在C++标准术语中称为 *协程状态*，通常是在堆上分配的。

让我们首先为协程 `g` 定义一个用于协程状态的结构体。

我们还不知道这个类型的具体内容是什么，所以现在先让它保持为空。

---

```c++
struct __g_state {
  // to be filled out
};
```

The coroutine state contains a number of different things:

- The promise object
- Copies of any function parameters
- Information about the suspend-point that the coroutine is currently suspended at and how to resume/destroy it
- Storage for any local variables / temporaries whose lifetimes span a suspend-point

Let’s start by adding storage for the promise object and parameter copies.

---
协程状态包含多个不同的内容：

- 承诺对象（promise object）
- 函数参数的副本
- 关于协程当前挂起的挂起点的信息以及如何恢复或销毁它
- 跨越挂起点的任何局部变量或临时变量的存储

让我们首先添加对承诺对象和参数副本的存储。

---

```c++
struct __g_state {
    int x;
    __g_promise_t __promise;

    // to be filled out
};
```

Next we should add a constructor to initialise these data-members.

Recall that the compiler will first attempt to call the promise constructor with lvalue-references to the parameter copies, if that call is valid, otherwise fall back to calling the default constructor of the promise type.

Let’s create a simple helper to assist with this:

---
接下来，我们应该添加一个构造函数来初始化这些数据成员。

回想一下，编译器将首先尝试使用参数副本的左值引用调用承诺构造函数，如果该调用有效，则执行此操作；否则，回退到调用承诺类型的默认构造函数。

让我们创建一个简单的辅助函数来帮助实现这一点：

---

```c++
template<typename Promise, typename... Params>
Promise construct_promise([[maybe_unused]] Params&... params) {
    if constexpr (std::constructible_from<Promise, Params&...>) {
        return Promise(params...);
    } else {
        return Promise();
    }
}
```

Thus the coroutine-state constructor might look something like this:

---
因此，协程状态的构造函数可能看起来像这样：

---

```c++
struct __g_state {
    __g_state(int&& x)
    : x(static_cast<int&&>(x))
    , __promise(construct_promise<__g_promise_t>(this->x))
    {}

    int x;
    __g_promise_t __promise;
    // to be filled out
};
```

Now that we have the beginnings of a type to represent the coroutine-state, let’s also start to stub out the beginnings of the lowered implementation of `g()` by having it heap-allocate an instance of the `__g_state` type, passing the function parameters so they can be copied/ moved into the coroutine-state.

Some terminology - I use the term “ramp function” to refer to the part of the coroutine implementation containing the logic that initialises the coroutine state and gets it ready to start executing the coroutine. i.e. it is like an on-ramp for entering execution of the coroutine body.

---
现在我们已经有了一个表示协程状态的类型的雏形，让我们也开始为 `g()` 的降低实现打桩，通过在堆上分配一个 `__coroutine_state` 类型的实例，并传递函数参数，以便它们可以被复制/移动到协程状态中。

一些术语——我使用“ramp function”（引导函数）这个术语来指代包含初始化协程状态并使其准备好开始执行协程逻辑的那部分协程实现。也就是说，它就像进入协程体执行的一个入口坡道。

---

```c++
task g(int x) {
    auto* state = new __g_state(static_cast<int&&>(x));
    // ... implement rest of the ramp function
}
```

Note that our promise-type does not define its own custom `operator new` overloads, and so we are just calling global `::operator new` here.

If the promise type *did* define a custom `operator new` then we’d call that instead of the global `::operator new`. We would first check whether `operator new` was callable with the argument list `(size, paramLvalues...)` and if so call it with that argument list. Otherwise, we’d call it with just the `(size)` argument list. The ability for the `operator new` to get access to the parameter list of the coroutine function is sometimes called “parameter preview” and is useful in cases where you want to use an allocator passed as a parameter to allocate storage for the coroutine-state.

If the compiler found any definition of `__g_promise_t::operator new` then we’d lower to the following logic instead:

---
注意，我们的承诺类型没有定义自己的自定义 `operator new` 重载，因此我们在这里只是调用全局的 `::operator new`。

如果承诺类型确实定义了自定义的 `operator new`，那么我们将调用该自定义的 `operator new` 而不是全局的 `::operator new`。我们会首先检查 `operator new` 是否可以使用参数列表 `(size, paramLvalues...)` 进行调用，如果是，则使用该参数列表调用它。否则，我们将仅使用 `(size)` 参数列表调用它。`operator new` 能够访问协程函数的参数列表有时被称为“参数预览”，这在你想使用作为参数传递的分配器来为协程状态分配存储时非常有用。

如果编译器找到了任何 `__coroutine_state::operator new` 的定义，我们将降低为以下逻辑：

---
```c++
template<typename Promise, typename... Args>
void* __promise_allocate(std::size_t size, [[maybe_unused]] Args&... args) {
  if constexpr (requires { Promise::operator new(size, args...); }) {
    return Promise::operator new(size, args...);
  } else {
    return Promise::operator new(size);
  }
}

task g(int x) {
    void* state_mem = __promise_allocate<__g_promise_t>(sizeof(__g_state), x);
    __g_state* state;
    try {
        state = ::new (state_mem) __g_state(static_cast<int&&>(x));
    } catch (...) {
        __g_promise_t::operator delete(state_mem);
        throw;
    }
    // ... implement rest of the ramp function
}
```

Also, this promise-type does not define the `get_return_object_on_allocation_failure()` static member function. If this function is defined on the promise-type then the allocation here would instead use the `std::nothrow_t` form of `operator new` and upon returning `nullptr` would then `return __g_promise_t::get_return_object_on_allocation_failure();`.

i.e. it would look something like this instead:

---
此外，这个承诺类型没有定义 `get_return_object_on_allocation_failure()` 静态成员函数。如果在承诺类型中定义了这个函数，那么这里的分配将使用 `operator new` 的 `std::nothrow_t` 形式，并在返回 `nullptr` 时调用 `__coroutine_state::get_return_object_on_allocation_failure()`。

也就是说，它看起来会像这样：

---

```c++
task g(int x) {
    auto* state = ::new (std::nothrow) __g_state(static_cast<int&&>(x));
    if (state == nullptr) {
        return __g_promise_t::get_return_object_on_allocation_failure();
    }
    // ... implement rest of the ramp function
}
```

For simplicity for the rest of the example, we’ll just use the simplest form that calls the global `::operator new` memory allocation function.

---
为了简单起见，在接下来的例子中，我们将仅使用调用全局 `::operator new` 内存分配函数的最简形式。

---

# Step 3: Call `get_return_object()`

The next thing the ramp function does is to call the `get_return_object()` method on the promise object to obtain the return-value of the ramp function.

The return value is stored as a local variable and is returned at the end of the ramp function (after the other steps have been completed).

---
接下来，引导函数所做的就是调用承诺对象上的 `get_return_object()` 方法以获得引导函数的返回值。

返回值存储为局部变量，并在引导函数结束时返回（在其他步骤完成后）。

---

```c++
task g(int x) {
    auto* state = new __g_state(static_cast<int&&>(x));
    decltype(auto) return_value = state->__promise.get_return_object();
    // ... implement rest of ramp function
    return return_value;
}
```

However, now it’s possible that the call to `get_return_object()` might throw, and in which case we want to free the allocated coroutine state. So for good measure, let’s give ownership of the state to a `std::unique_ptr` so that it’s freed in case a subsequent operation throws an exception:

---
然而，现在有可能 `get_return_object()` 的调用会抛出异常，在这种情况下，我们希望释放已分配的协程状态。因此，为了安全起见，让我们将状态的所有权交给 `std::unique_ptr`，以便在后续操作抛出异常时自动释放它：

---

```c++
task g(int x) {
    std::unique_ptr<__g_state> state(new __g_state(static_cast<int&&>(x)));
    decltype(auto) return_value = state->__promise.get_return_object();
    // ... implement rest of ramp function
    return return_value;
}
```

# Step 4: The initial-suspend point

The next thing the ramp function does after calling `get_return_object()` is to start executing the body of the coroutine, and the first thing to execute in the body of the coroutine is the initial suspend-point. i.e. we evaluate `co_await promise.initial_suspend()`.

Now, ideally we’d just treat the coroutine as initially suspended and then just implement the launching of the coroutine as a resumption of the initially suspended coroutine. However, the specification of the initial-suspend point has a few quirks with regards to how it handles exceptions and the lifetime of the coroutine state. This was a late tweak to the semantics of the initial-suspend point just before C++20 was released to fix some perceived issues here.

Within the evaluation of the initial-suspend-point, if an exception is thrown either from:

- the call to `initial_suspend()`,
- the call to `operator co_await()` on the returned awaitable (if one is defined),
- the call to `await_ready()` on the awaiter, or
- the call to `await_suspend()` on the awaiter

Then the exception propagates back to the caller of the ramp function and the coroutine state is automatically destroyed.

If an exception is thrown either from:

- the call to `await_resume()`,
- the destructor of the object returned from `operator co_await()` (if applicable), or
- the destructor of the object returned from `initial_suspend()`

Then this exception is caught by the coroutine body and `promise.unhandled_exception()` is called.

This means we need to be a bit careful how we handle transforming this part, as some parts will need to live in the ramp function and other parts in the coroutine body.

Also, since the objects returned from `initial_suspend()` and (optionally) `operator co_await()` will have lifetimes that span a suspend-point (they are created before the point at which the coroutine suspends and are destroyed after it resumes) the storage for those objects will need to be placed in the coroutine state.

In our particular case, the type returned from `initial_suspend()` is `std::suspend_always`, which happens to be an empty, trivially constructible type. However, logically we still need to store an instance of this type in the coroutine state, so we’ll add storage for it anyway just to show how this works.

This object will only be constructed at the point that we call `initial_suspend()`, so we need to add a data-member of a certain type that that allows us to explicitly control its lifetime.

To support this, let’s first define a helper class, `manual_lifetime` that is trivally constructible and trivially destructible but that lets us explicitly construct/destruct the value stored there when we need to.

---
`ramp`函数在调用`get_return_object()`之后要做的下一件事，就是开始执行协程体，而协程体中首先要执行的是初始挂起点。也就是说，我们要计算`co_await promise.initial_suspend()`。

理想情况下，我们可以将协程视为初始挂起状态，然后将协程的启动当作对初始挂起协程的恢复。然而，关于初始挂起点在处理异常以及协程状态生命周期方面的规范，存在一些特殊之处。这是在C++20发布前夕对初始挂起点语义的最后调整，以解决这里察觉到的一些问题。

在初始挂起点的计算过程中，如果从以下任何一处抛出异常：
 - 对`initial_suspend()`的调用，
 - 对返回的可等待对象调用`operator co_await()`（如果定义了的话），
 - 对等待器调用`await_ready()`，或者
 - 对等待器调用`await_suspend()`

那么该异常会传播回`ramp`函数的调用者，并且协程状态会自动销毁。

如果从以下任何一处抛出异常：
 - 对`await_resume()`的调用，
 - 从`operator co_await()`返回的对象的析构函数（如果适用），或者
 - 从`initial_suspend()`返回的对象的析构函数

那么这个异常会被协程体捕获，并且会调用`promise.unhandled_exception()`。

这意味着我们在处理这部分转换时需要格外小心，因为有些部分需要放在`ramp`函数中，而其他部分则要放在协程体中。

此外，由于从`initial_suspend()`以及（可选的）`operator co_await()`返回的对象的生命周期会跨越一个挂起点（它们在协程挂起之前创建，在恢复之后销毁），所以这些对象的存储需要放在协程状态中。

在我们这个特定的例子中，从`initial_suspend()`返回的类型是`std::suspend_always`，它恰好是一个空的、可平凡构造的类型。然而，从逻辑上讲，我们仍然需要在协程状态中存储这个类型的一个实例，所以无论如何我们都要为它添加存储空间，以展示这是如何工作的。

这个对象只会在我们调用`initial_suspend()`的时候构造，所以我们需要添加一个特定类型的数据成员，以便能够显式地控制它的生命周期。

为了实现这一点，我们首先定义一个辅助类`manual_lifetime`，它可以平凡构造和平凡析构，但能让我们在需要的时候显式地构造/析构存储在其中的值。 

---

```c++
template<typename T>
struct manual_lifetime {
    manual_lifetime() noexcept = default;
    ~manual_lifetime() = default;

    // Not copyable/movable
    manual_lifetime(const manual_lifetime&) = delete;
    manual_lifetime(manual_lifetime&&) = delete;
    manual_lifetime& operator=(const manual_lifetime&) = delete;
    manual_lifetime& operator=(manual_lifetime&&) = delete;

    template<typename Factory>
        requires
            std::invocable<Factory&> &&
            std::same_as<std::invoke_result_t<Factory&>, T>
    T& construct_from(Factory factory) noexcept(std::is_nothrow_invocable_v<Factory&>) {
        return *::new (static_cast<void*>(&storage)) T(factory());
    }

    void destroy() noexcept(std::is_nothrow_destructible_v<T>) {
        std::destroy_at(std::launder(reinterpret_cast<T*>(&storage)));
    }

    T& get() & noexcept {
        return *std::launder(reinterpret_cast<T*>(&storage));
    }

private:
    alignas(T) std::byte storage[sizeof(T)];
};
```

Note that the `construct_from()` method is designed to take a lambda here rather than taking the constructor arguments. This allows us to make use of the guaranteed copy-elision when initialising a variable with the result of a function-call to construct the object in-place. If it were instead to take the constructor arguments then we’d end up calling an extra move-constructor unnecessarily.

Now we can declare a data-member for the temporary returned by `promise.initial_suspend()` using this `manual_lifetime` structure.

---
注意，`construct_from()` 方法设计为接受一个 lambda 表达式，而不是直接接受构造函数的参数。这使我们能够在用函数调用的结果初始化变量时利用保证的拷贝省略来就地构造对象。如果它改为接受构造函数的参数，我们将最终不必要地调用额外的移动构造函数。

现在我们可以使用这个 `manual_lifetime` 结构声明一个数据成员来存储 `promise.initial_suspend()` 返回的临时对象。

---

```c++
struct __g_state {
    __g_state(int&& x);

    int x;
    __g_promise_t __promise;
    manual_lifetime<std::suspend_always> __tmp1;
    // to be filled out
};
```

The `std::suspend_always` type does not have an `operator co_await()` so we do not need to reserve storage for an extra temporary for the result of that call here.

Once we’ve constructed this object by calling `intial_suspend()`, we then need to call the trio of methods to implement the `co_await` expression: `await_ready()`, `await_suspend()` and `await_resume()`.

When invoking `await_suspend()` we need to pass it a handle to the current coroutine. For now we can just call `std::coroutine_handle<__g_promise_t>::from_promise()` and pass a reference to that promise. We’ll look at the internals of what this does a little later.

Also, the result of the call to `.await_suspend(handle)` has type `void` and so we do not need to consider whether to resume this coroutine or another coroutine after calling `await_suspend()` like we do for the `bool` and `coroutine_handle`-returning flavours.

Finally, as all of the method invocations on the `std::suspend_always` awaiter are declared `noexcept`, we don’t need to worry about exceptions. If they were potentially throwing then we’d need to add extra code to make sure that the temporary `std::suspend_always` object was destroyed before the exception propagated out of the ramp function.

Once we get to the point where `await_suspend()` has returned successfully or where we are about to start executing the coroutine body we enter the phase where we no longer need to automatically destroy the coroutine-state if an exception is thrown. So we can call `release()` on the `std::unique_ptr` owning the coroutine state to prevent it from being destroyed when we return from the function.

So now we can implement the first part of the initial-suspend expression as follows:

---
`std::suspend_always`类型没有`operator co_await()`，所以我们无需在此为该调用结果预留额外临时变量的存储空间。

通过调用`intial_suspend()`构造好这个对象后，我们接着需要调用三个方法来实现`co_await`表达式，即`await_ready()`、`await_suspend()`和`await_resume()`。

调用`await_suspend()`时，我们需要向其传递当前协程的句柄。目前，我们可以直接调用`std::coroutine_handle<__g_promise_t>::from_promise()`，并传递对该承诺对象的引用。稍后我们会深入研究这一操作的内部原理。

此外，`.await_suspend(handle)`调用的返回类型为`void`，所以与返回`bool`和`coroutine_handle`的情况不同，调用`await_suspend()`后，我们无需考虑是恢复当前协程还是其他协程。

最后，由于`std::suspend_always`等待器上所有方法调用都声明为`noexcept`，我们无需担心异常问题。如果这些方法可能抛出异常，那么我们就需要添加额外代码，以确保在异常从`ramp`函数传播出去之前，临时的`std::suspend_always`对象已被销毁。

一旦`await_suspend()`成功返回，或者我们即将开始执行协程体，此时就进入了一个新阶段，即如果抛出异常，无需自动销毁协程状态。所以我们可以对拥有协程状态的`std::unique_ptr`调用`release()`，防止从函数返回时协程状态被销毁。

因此，我们现在可以按如下方式实现初始挂起表达式的第一部分： 

---
```c++
task g(int x) {
    std::unique_ptr<__g_state> state(new __g_state(static_cast<int&&>(x)));
    decltype(auto) return_value = state->__promise.get_return_object();

    state->__tmp1.construct_from([&]() -> decltype(auto) {
        return state->__promise.initial_suspend();
    });
    if (!state->__tmp1.get().await_ready()) {
        //
        // ... suspend-coroutine here
        //
        state->__tmp1.get().await_suspend(
            std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));

        state.release();

        // fall through to return statement below.
    } else {
        // Coroutine did not suspend.

        state.release();

        //
        // ... start executing the coroutine body
        //
    }
    return __return_val;
}
```

The call to `await_resume()` and the destructor of `__tmp1` will appear in the coroutine body and so they do not appear in the ramp function.

We now have a (mostly) functional evaluation of the initial-suspend point, but we still have a couple of TODO’s in the code for this ramp function. To be able to resolve these we will first need to take a detour to look at the strategy for suspending a coroutine and later resuming it.

---
对`await_resume()`的调用和`__tmp1`的析构函数将出现在协程体中，因此它们不会出现在ramp函数中。

现在我们已经（基本上）完成了对初始挂起点的功能评估，但是这个ramp函数的代码里还有一些待办事项（TODO）。为了能够解决这些问题，我们需要先绕个弯，了解一下挂起协程以及之后恢复它的策略。

---

# Step 5: Recording the suspend-point

When a coroutine suspends, it needs to make sure it resumes at the same point in the control flow that it suspended at.

It also needs to keep track of which objects with automatic-storage duration are alive at each suspend-point so that it knows what needs to be destroyed if the coroutine is destroyed instead of being resumed.

One way to implement this is to assign each suspend-point in the coroutine a unique number and then store this in an integer data-member of the coroutine state.

Then whenever a coroutine suspends, it writes the number of the suspend-point at which it is suspending to the coroutine state, and when it is resumed/destroyed we then inspect this integer to see which suspend point it was suspended at.

Note that this is not the only way of storing the suspend-point in the coroutine state, however all 3 major compilers (MSVC, Clang, GCC) use this approach as the time this post was authored (c. 2022). Another potential solution is to use separate resume/destroy function-pointers for each suspend-point, although we will not be exploring this strategy in this post.

So let’s extend our coroutine-state with an integer data-member to store the suspend-point index and initialise it to zero (we’ll always use this as the value for the initial-suspend point).

---
当协程挂起时，它需要确保在控制流中同一位置恢复。

它还需要跟踪每个挂起点处具有自动存储持续期的对象，以便在协程被销毁而非恢复时知道哪些对象需要被析构。

实现这一点的一种方法是给协程中的每个挂起点分配一个唯一的编号，然后将其存储在协程状态的整型数据成员中。

因此，每当协程挂起时，它会将挂起点的编号写入协程状态。而在其被恢复或销毁时，我们通过检查这个整数来确定协程是在哪个挂起点上被挂起的。

需要注意的是，这并不是在协程状态中存储挂起点的唯一方式，然而，在这篇帖子撰写时（约2022年），所有三大编译器（MSVC、Clang、GCC）都使用这种方法。另一种可能的解决方案是对每个挂起点使用单独的恢复/销毁函数指针，不过在这篇文章中不会探讨这种策略。

那么，让我们扩展我们的协程状态，添加一个整型数据成员以存储挂起点索引，并将其初始化为零（我们将始终把这个值用于初始挂起点）。

---

```c++
struct __g_state {
    __g_state(int&& x);

    int x;
    __g_promise_t __promise;
    int __suspend_point = 0;  // <-- add the suspend-point index
    manual_lifetime<std::suspend_always> __tmp1;
    // to be filled out
};
```

# Step 6: Implementing `coroutine_handle::resume()` and `coroutine_handle::destroy()`

When a coroutine is resumed by calling `coroutine_handle::resume()` we need this to end up invoking some function that implements the rest of the body of the suspended coroutine. The invoked body function can then look up the suspend-point index and jump to the appropriate point in the control-flow.

We also need to implement the `coroutine_handle::destroy()` function so that it invokes the appropriate logic to destroy any in-scope objects at the current suspend-point and we need to implement `coroutine_handle::done()` to query whether the current suspend-point is a final-suspend-point.

The interface of the `coroutine_handle` methods does not know about the concrete coroutine state type - the `coroutine_handle<void>` type can point to *any* coroutine instance. This means we need to implement them in a way that type-erases the coroutine state type.

We can do this by storing function-pointers to the resume/destroy functions for that coroutine type and having `coroutine_handle::resume/destroy()` invoke those function-pointers.

The `coroutine_handle` type also needs to be able to be converted to/from a `void*` using the `coroutine_handle::address()` and `coroutine_handle::from_address()` methods.

Furthermore, the coroutine can be resumed/destroyed from *any* handle to that coroutine - not just the handle that was passed to the most recent `await_suspend()` call.

These requirements lead us to define the `coroutine_handle` type so that it only contains a pointer to the coroutine-state and that we store the resume/destroy function pointers as data-members of the coroutine state, rather than, say, storing the resume/destroy function pointers in the `coroutine_handle`.

Also, since we need the `coroutine_handle` to be able to point to an arbitrary coroutine-state object we need the layout of the function-pointer data-members to be consistent across all coroutine-state types.

One straight forward way of doing this is having each coroutine-state type inherit from some base-class that contains these data-members.

e.g. We can define the following type as the base-class for all coroutine-state types

---
当通过调用 `coroutine_handle::resume()` 来恢复协程时，我们需要这一操作最终调用某个函数来执行挂起协程主体的剩余部分。被调用的主体函数可以查找挂起点索引并跳转到控制流中的适当位置。

我们还需要实现 `coroutine_handle::destroy()` 函数，以便在当前挂起点销毁任何范围内对象时调用适当的逻辑，并且需要实现 `coroutine_handle::done()` 来查询当前挂起点是否是最终挂起点。

`coroutine_handle` 方法的接口并不了解具体的协程状态类型——`coroutine_handle<void>` 类型可以指向*任何*协程实例。这意味着我们需要以一种类型擦除协程状态类型的方式来实现它们。

我们可以通过存储指向该协程类型的恢复/销毁函数的函数指针，并让 `coroutine_handle::resume/destroy()` 调用这些函数指针来实现这一点。

`coroutine_handle` 类型还需要能够使用 `coroutine_handle::address()` 和 `coroutine_handle::from_address()` 方法与 `void*` 进行转换。

此外，可以从*任何*对该协程的句柄恢复或销毁协程——而不仅仅是传递给最近的 `await_suspend()` 调用的那个句柄。

这些要求引导我们将 `coroutine_handle` 类型定义为只包含指向协程状态的指针，并将恢复/销毁函数指针作为协程状态的数据成员存储，而不是将恢复/销毁函数指针存储在 `coroutine_handle` 中。

而且，由于我们需要 `coroutine_handle` 能够指向任意的协程状态对象，我们需要确保所有协程状态类型的函数指针数据成员布局一致。

一种直接的方法是让每个协程状态类型都继承自包含这些数据成员的基类。

例如，我们可以定义以下类型作为所有协程状态类型的基类。

---
```c++
struct __coroutine_state {
    using __resume_fn = void(__coroutine_state*);
    using __destroy_fn = void(__coroutine_state*);

    __resume_fn* __resume;
    __destroy_fn* __destroy;
};
```

Then the `coroutine_handle::resume()` method can simply call `__resume()`, passing a pointer to the `__coroutine_state` object. Similarly, we can do this for the `coroutine_handle::destroy()` method and the `__destroy` function-pointer.

For the `coroutine_handle::done()` method, we choose to treat a null `__resume` function pointer as an indication that we are at a final-suspend-point. This is convenient since the final suspend point does not support `resume()`, only `destroy()`. If someone tries to call `resume()` on a coroutine suspended at the final-suspend-point (which has undefined-behaviour) then they end up calling a null function pointer which should fail pretty quickly and point out their error.

Given this, we can implement the `coroutine_handle<void>` type as follows:

---
然后，`coroutine_handle::resume()` 方法可以简单地调用 `__resume()`，传递一个指向 `__coroutine_state` 对象的指针。同样地，对于 `coroutine_handle::destroy()` 方法和 `__destroy` 函数指针，我们也可以这样做。

对于 `coroutine_handle::done()` 方法，我们选择将空的 `__resume` 函数指针视为处于最终挂起点的标志。这很方便，因为最终挂起点不支持 `resume()`，只支持 `destroy()`。如果有人尝试对处于最终挂起点的协程调用 `resume()`（这有未定义行为），那么他们最终会调用一个空的函数指针，这应该很快就会失败并指出他们的错误。

基于此，我们可以按如下方式实现 `coroutine_handle<void>` 类型：

---

```c++
namespace std
{
    template<typename Promise = void>
    class coroutine_handle;

    template<>
    class coroutine_handle<void> {
    public:
        coroutine_handle() noexcept = default;
        coroutine_handle(const coroutine_handle&) noexcept = default;
        coroutine_handle& operator=(const coroutine_handle&) noexcept = default;

        void* address() const {
            return static_cast<void*>(state_);
        }

        static coroutine_handle from_address(void* ptr) {
            coroutine_handle h;
            h.state_ = static_cast<__coroutine_state*>(ptr);
            return h;
        }

        explicit operator bool() noexcept {
            return state_ != nullptr;
        }
        
        friend bool operator==(coroutine_handle a, coroutine_handle b) noexcept {
            return a.state_ == b.state_;
        }

        void resume() const {
            state_->__resume(state_);
        }
        void destroy() const {
            state_->__destroy(state_);
        }

        bool done() const {
            return state_->__resume == nullptr;
        }

    private:
        __coroutine_state* state_ = nullptr;
    };
}
```

# Step 7: Implementing `coroutine_handle<Promise>::promise()` and `from_promise()`

For the more general `coroutine_handle<Promise>` specialisation, most of the implementations can just reuse the `coroutine_handle<void>` implementations. However, we also need to be able to get access to the promise object of the coroutine-state, returned from the `promise()` method, and also construct a `coroutine_handle` from a reference to the promise-object.

However, again we cannot simply point to the concrete coroutine state type since the `coroutine_handle<Promise>` type must be able to refer to any coroutine-state whose promise-type is `Promise`.

We need to define a new coroutine-state base-class that inherits from `__coroutine_state` and which contains the promise object so we can then define all coroutine-state types that use a particular promise-type to inherit from this base-class.

---
对于更通用的 `coroutine_handle<Promise>` 特化实现，大部分实现可以重用 `coroutine_handle<void>` 的实现。然而，我们还需要能够访问协程状态中返回的承诺对象，即通过 `promise()` 方法获取承诺对象，并且也能够从承诺对象的引用构造一个 `coroutine_handle`。

然而，我们再次不能简单地指向具体的协程状态类型，因为 `coroutine_handle<Promise>` 类型必须能够引用任何承诺类型为 `Promise` 的协程状态。

我们需要定义一个新的协程状态基类，该基类继承自 `__coroutine_state` 并包含承诺对象。这样，我们可以定义所有使用特定承诺类型的协程状态类型都继承自这个基类。

---

```c++
template<typename Promise>
struct __coroutine_state_with_promise : __coroutine_state {
    __coroutine_state_with_promise() noexcept {}
    ~__coroutine_state_with_promise() {}

    union {
        Promise __promise;
    };
};
```

You might be wondering why we declare the `__promise` member inside an anonymous union here…

The reason for this is that the derived class created for a particular coroutine function contains the definition for the argument-copy data-members. Data members from derived classes are by default initialised after data-members of any base-classes, so declaring the promise object as a normal data-member would mean that the promise object was constructed before the argument-copy data-members.

However, we need the constructor of the promise to be called *after* the constructor of the argument-copies - references to the argument-copies might need to be passed to the promise constructor.

So we reserve storage for the promise object in this base-class so that it has a consistent offset from the start of the coroutine-state, but leave the derived class responsible for calling the constructor/destructor at the appropriate point after the argument-copies have been initialised. Declaring the `__promise` as a union-member provides this control.

Let’s update the `__g_state` class to now inherit from this new base-class.

---
你可能想知道为什么我们在这里声明了一个匿名联合体中的 `__promise` 成员……

这样做的原因是，为特定协程函数创建的派生类包含了参数复制数据成员的定义。默认情况下，派生类的数据成员是在任何基类的数据成员之后初始化的，因此如果将承诺对象声明为普通数据成员，则意味着承诺对象将在参数复制数据成员之前构造。

然而，我们需要的是在参数复制的构造函数被调用之后再调用承诺对象的构造函数——可能需要将对参数复制的引用传递给承诺对象的构造函数。

因此，我们在基类中保留承诺对象的存储空间，以便它与协程状态开始处有一个一致的偏移量，但让派生类负责在参数复制初始化后的适当时候调用构造函数/析构函数。将 `__promise` 声明为联合体成员提供了这种控制。

让我们更新 `__g_state` 类以继承这个新的基类。

---

```c++
struct __g_state : __coroutine_state_with_promise<__g_promise_t> {
    __g_state(int&& __x)
    : x(static_cast<int&&>(__x)) {
        // Use placement-new to initialise the promise object in the base-class
        ::new ((void*)std::addressof(this->__promise))
            __g_promise_t(construct_promise<__g_promise_t>(x));
    }

    ~__g_state() {
        // Also need to manually call the promise destructor before the
        // argument objects are destroyed.
        this->__promise.~__g_promise_t();
    }

    int __suspend_point = 0;
    int x;
    manual_lifetime<std::suspend_always> __tmp1;
    // to be filled out
};
```

Now that we have defined the promise-base-class we can now implement the `std::coroutine_handle<Promise>` class template.

Most of the implementation should be largely identical to the equivalent methods in `coroutine_handle<void>` except with a `__coroutine_state_with_promise<Promise>` pointer instead of `__coroutine_state` pointer.

The only new part is the addition of the `promise()` and `from_promise()` functions.

- The `promise()` method is straight-forward - it just returns a reference to the `__promise` member of the coroutine-state.
- The `from_promise()` method requires us to calculate the address of the coroutine-state from the address of the promise object. We can do this by just subtracting the offset of the `__promise` member from the address of the promise object.

Implementation of `coroutine_handle<Promise>`:

---
现在我们已经定义了承诺基类，我们可以实现 `std::coroutine_handle<Promise>` 类模板。

大多数实现应该与 `coroutine_handle<void>` 中的等效方法大体相同，只不过使用的是指向 `__coroutine_state_with_promise<Promise>` 的指针而不是指向 `__coroutine_state` 的指针。

唯一的新部分是添加了 `promise()` 和 `from_promise()` 函数。

- `promise()` 方法很简单——它只是返回协程状态的 `__promise` 成员的引用。
- `from_promise()` 方法要求我们根据承诺对象的地址计算协程状态的地址。我们可以通过从承诺对象的地址中减去 `__promise` 成员的偏移量来做到这一点。

`coroutine_handle<Promise>` 的实现：

---

```c++
namespace std
{
    template<typename Promise>
    class coroutine_handle {
        using state_t = __coroutine_state_with_promise<Promise>;
    public:
        coroutine_handle() noexcept = default;
        coroutine_handle(const coroutine_handle&) noexcept = default;
        coroutine_handle& operator=(const coroutine_handle&) noexcept = default;

        operator coroutine_handle<void>() const noexcept {
            return coroutine_handle<void>::from_address(address());
        }

        explicit operator bool() const noexcept {
            return state_ != nullptr;
        }

        friend bool operator==(coroutine_handle a, coroutine_handle b) noexcept {
            return a.state_ == b.state_;
        }

        void* address() const {
            return static_cast<void*>(static_cast<__coroutine_state*>(state_));
        }

        static coroutine_handle from_address(void* ptr) {
            coroutine_handle h;
            h.state_ = static_cast<state_t*>(static_cast<__coroutine_state*>(ptr));
            return h;
        }

        Promise& promise() const {
            return state_->__promise;
        }

        static coroutine_handle from_promise(Promise& promise) {
            coroutine_handle h;

            // We know the address of the __promise member, so calculate the
            // address of the coroutine-state by subtracting the offset of
            // the __promise field from this address.
            h.state_ = reinterpret_cast<state_t*>(
                reinterpret_cast<unsigned char*>(std::addressof(promise)) -
                offsetof(state_t, __promise));

            return h;
        }

        // Define these in terms of their `coroutine_handle<void>` implementations

        void resume() const {
            static_cast<coroutine_handle<void>>(*this).resume();
        }

        void destroy() const {
            static_cast<coroutine_handle<void>>(*this).destroy();
        }

        bool done() const {
            return static_cast<coroutine_handle<void>>(*this).done();
        }

    private:
        state_t* state_;
    };
}
```

Now that we have defined the mechanism by which coroutines are resumed, we can now return to our “ramp” function and update it to initialise the new function-pointer data-members we’ve added to the coroutine-state.

---
现在我们已经定义了恢复协程的机制，现在我们可以回到我们的“ramp”函数，并更新它以初始化我们添加到协程状态的新函数指针数据成员。

---

# Step 8: The beginnings of the coroutine body

Let’s now forward-declare resume/destroy functions of the right signature and update the `__g_state` constructor to initialise the coroutine-state so that the resume/destroy function-pointers point at them:

---
我们现在先向前声明具有正确签名的resume/destroy函数，并更新`__g_state`构造函数以初始化协程状态，从而使resume/destroy函数指针指向它们：

---

```c++
void __g_resume(__coroutine_state* s);
void __g_destroy(__coroutine_state* s);

struct __g_state : __coroutine_state_with_promise<__g_promise_t> {
    __g_state(int&& __x)
    : x(static_cast<int&&>(__x)) {
        // Initialise the function-pointers used by coroutine_handle methods.
        this->__resume = &__g_resume;
        this->__destroy = &__g_destroy;

        // Use placement-new to initialise the promise object in the base-class
        ::new ((void*)std::addressof(this->__promise))
            __g_promise_t(construct_promise<__g_promise_t>(x));
    }

    // ... rest omitted for brevity
};


task g(int x) {
    std::unique_ptr<__g_state> state(new __g_state(static_cast<int&&>(x)));
    decltype(auto) return_value = state->__promise.get_return_object();

    state->__tmp1.construct_from([&]() -> decltype(auto) {
        return state->__promise.initial_suspend();
    });
    if (!state->__tmp1.get().await_ready()) {
        state->__tmp1.get().await_suspend(
            std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));
        state.release();
        // fall through to return statement below.
    } else {
        // Coroutine did not suspend. Start executing the body immediately.
        __g_resume(state.release());
    }
    return return_value;
}
```

This now completes the ramp function and we can now focus on the resume/destroy functions for `g()`.

Let’s start by completing the lowering of the initial-suspend expression.

When `__g_resume()` is called and the `__suspend_point` index is 0 then we need it to resume by calling `await_resume()` on `__tmp1` and then calling the destructor of `__tmp1`.

---
现在这完成了斜坡函数，我们可以把注意力转向`g()`的恢复/销毁功能。

让我们首先完成初始挂起表达式的降低。

当调用`__g_resume()`且`__suspend_point`索引为0时，我们需要通过在`__tmp1`上调用`await_resume()`然后调用`__tmp1`的析构函数来恢复。 

---
```c++
void __g_resume(__coroutine_state* s) {
    // We know that 's' points to a __g_state.
    auto* state = static_cast<__g_state*>(s);

    // Generate a jump-table to jump to the correct place in the code based
    // on the value of the suspend-point index.
    switch (state->__suspend_point) {
    case 0: goto suspend_point_0;
    default: std::unreachable();
    }

suspend_point_0:
    state->__tmp1.get().await_resume();
    state->__tmp1.destroy();

    // TODO: Implement rest of coroutine body.
    //
    //  int fx = co_await f(x);
    //  co_return fx * fx;
}
```

And when `__g_destroy()` is called and the `__suspend_point` index is 0 then we need it to just destroy `__tmp1` before then destroying and freeing the coroutine-state.

---
当调用`__g_destroy()`且`__suspend_point`索引为0时，我们需要它在销毁并释放协程状态之前，仅销毁`__tmp1`。

---
```c++
void __g_destroy(__coroutine_state* s) {
    auto* state = static_cast<__g_state*>(s);

    switch (state->__suspend_point) {
    case 0: goto suspend_point_0;
    default: std::unreachable();
    }

suspend_point_0:
    state->__tmp1.destroy();
    goto destroy_state;

    // TODO: Add extra logic for other suspend-points here.

destroy_state:
    delete state;
}
```

# Step 9: Lowering the `co_await` expression

Next, let’s take a look at lowering the `co_await f(x)` expression.

First we need to evaluate `f(x)` which returns a temporary `task` object.

As the temporary `task` is not destroyed until the semicolon at the end of the statement and the statement contains a `co_await` expression, the lifetime of the `task` therefore spans a suspend-point and so it must be stored in the coroutine-state.

When the `co_await` expression is then evaluated on this temporary `task`, we need to call the `operator co_await()` method which returns a temporary `awaiter` object. The lifetime of this object also spans the suspend-point and so must be stored in the coroutine-state.

Let’s add the necessary members to the `__g_state` type:

---
接下来，让我们看一下降低`co_await f(x)`表达式的情况。

首先，我们需要评估`f(x)`，它返回一个临时的`task`对象。

由于这个临时的`task`不会在语句末尾的分号之前被销毁，且该语句包含了一个`co_await`表达式，因此`task`的生命周期跨越了一个挂起点，所以它必须被存储在协程状态中。

当在这个临时`task`上评估`co_await`表达式时，我们需要调用`operator co_await()`方法，该方法返回一个临时的`awaiter`对象。此对象的生命周期同样跨越了挂起点，因此也必须存储在协程状态中。

让我们给`__g_state`类型添加必要的成员：

---

```c++
struct __g_state : __coroutine_state_with_promise<__g_promise_t> {
    __g_state(int&& __x);
    ~__g_state();

    int __suspend_point = 0;
    int x;
    manual_lifetime<std::suspend_always> __tmp1;
    manual_lifetime<task> __tmp2;
    manual_lifetime<task::awaiter> __tmp3;
};
```

Then we can update the `__g_resume()` function to initialise these temporaries and then evaluate the 3 `await_ready`, `await_suspend` and `await_resume` calls that comprise the rest of the `co_await` expression.

Note that the `task::awaiter::await_suspend()` method returns a coroutine-handle so we need to generate code that resumes the returned handle.

We also need to update the suspend-point index before calling `await_suspend()` (we’ll use the index 1 for this suspend-point) and then add an extra entry to the jump-table to ensure that we resume back at the right spot.

---
然后，我们可以更新`__g_resume()`函数来初始化这些临时对象，然后评估构成`co_await`表达式其余部分的3个`await_ready`、`await_suspend`和`await_resume`调用。

请注意，`task::awaiter::await_suspend()`方法返回一个协程句柄，因此我们需要生成代码来恢复返回的句柄。

我们还需要在调用`await_suspend()`之前更新挂起点索引（对于这个挂起点，我们将使用索引1），然后在跳转表中添加一个额外的条目，以确保我们在正确的位置恢复执行。

---

```c++
void __g_resume(__coroutine_state* s) {
    // We know that 's' points to a __g_state.
    auto* state = static_cast<__g_state*>(s);

    // Generate a jump-table to jump to the correct place in the code based
    // on the value of the suspend-point index.
    switch (state->__suspend_point) {
    case 0: goto suspend_point_0;
    case 1: goto suspend_point_1; // <-- add new jump-table entry
    default: std::unreachable();
    }

suspend_point_0:
    state->__tmp1.get().await_resume();
    state->__tmp1.destroy();

    //  int fx = co_await f(x);
    state->__tmp2.construct_from([&] {
        return f(state->x);
    });
    state->__tmp3.construct_from([&] {
        return static_cast<task&&>(state->__tmp2.get()).operator co_await();
    });
    if (!state->__tmp3.get().await_ready()) {
        // mark the suspend-point
        state->__suspend_point = 1;

        auto h = state->__tmp3.get().await_suspend(
            std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));
        
        // Resume the returned coroutine-handle before returning.
        h.resume();
        return;
    }

suspend_point_1:
    int fx = state->__tmp3.get().await_resume();
    state->__tmp3.destroy();
    state->__tmp2.destroy();

    // TODO: Implement
    //  co_return fx * fx;
}
```

Note that the `int fx` local variable has a lifetime that does not span a suspend-point and so it does not need to be stored in the coroutine-state. We can just store it as a normal local variable in the `__g_resume` function.

We also need to add the necessary entry to the `__g_destroy()` function to handle when the coroutine is destroyed at this suspend-point.

---
注意，局部变量`int fx`的生命周期不会跨越一个挂起点，因此它不需要存储在协程状态中。我们只需将它作为一个普通的局部变量存储在`__g_resume`函数中即可。

我们还需要在`__g_destroy()`函数中添加必要的条目，以处理当协程在此挂起点被销毁时的情况。

---

```c++
void __g_destroy(__coroutine_state* s) {
    auto* state = static_cast<__g_state*>(s);

    switch (state->__suspend_point) {
    case 0: goto suspend_point_0;
    case 1: goto suspend_point_1; // <-- add new jump-table entry
    default: std::unreachable();
    }

suspend_point_0:
    state->__tmp1.destroy();
    goto destroy_state;

suspend_point_1:
    state->__tmp3.destroy();
    state->__tmp2.destroy();
    goto destroy_state;

    // TODO: Add extra logic for other suspend-points here.

destroy_state:
    delete state;
}
```

So now we have finished implementing the statement:

---
所以我们现在已经完成了实现该语句的工作：

---

```c++
int fx = co_await f(x);
```

However, the function `f(x)` is not marked `noexcept` and so it can potentially throw an exception. Also, the `awaiter::await_resume()` method is also not marked `noexcept` and can also potentially throw an exception.

When an exception is thrown from a coroutine-body the compiler generates code to catch the exception and then invoke `promise.unhandled_exception()` to give the promise an opportunity to do something with the exception. Let’s look at implementing this aspect next.

---
然而，函数 `f(x)` 没有标记为 `noexcept`，因此它有可能抛出异常。同样，`awaiter::await_resume()` 方法也没有标记为 `noexcept`，这也可能抛出异常。

当从协程体中抛出异常时，编译器会生成代码来捕获该异常，然后调用 `promise.unhandled_exception()`，以便给 promise 一个处理这个异常的机会。接下来，让我们看看如何实现这一方面。

---

# Step 10: Implementing `unhandled_exception()`

The specification for coroutine definitions [`[dcl.fct.def.coroutine\]`](https://eel.is/c++draft/dcl.fct.def.coroutine) says that the coroutine behaves as if its function-body were replaced by:

---
协程定义的规范 [`[dcl.fct.def.coroutine]`](https://eel.is/c++draft/dcl.fct.def.coroutine) 指出，协程的行为就如同其函数体被替换为：

---

```c++
{
    promise-type promise promise-constructor-arguments ;
    try {
        co_await promise.initial_suspend() ;
        function-body
    } catch ( ... ) {
        if (!initial-await-resume-called)
            throw ;
        promise.unhandled_exception() ;
    }
final-suspend :
    co_await promise.final_suspend() ;
}
```

We have already handled the `initial-await_resume-called` branch separately in the ramp function, so we don’t need to worry about that here.

Let’s adjust the `__g_resume()` function to insert the try/catch block around the body.

Note that we need to be careful to put the `switch` that jumps to the right place inside the try-block as we are not allowed to enter a try-block using a `goto`.

Also, we need to be careful to call `.resume()` on the coroutine handle returned from `await_suspend()` outside of the try/catch block. If an exception is thrown from the call `.resume()` on the returned coroutine then it should not be caught by the current coroutine, but should instead propagate out of the call to `resume()` that resumed this coroutine. So we stash the coroutine-handle in a variable declared at the top of the function and then `goto` a point outside of the try/catch and execute the call to `.resume()` there.

---
我们已经在斜坡函数中单独处理了 `initial-await_resume-called` 分支，因此这里不需要担心这个问题。

让我们调整 `__g_resume()` 函数，在其主体周围插入 try/catch 块。

请注意，我们需要小心地将跳转到正确位置的 `switch` 语句放在 try 块内部，因为我们不允许使用 `goto` 进入 try 块。

此外，我们需要注意，在 `await_suspend()` 返回的协程句柄上调用 `.resume()` 需要在 try/catch 块外部执行。如果在调用返回的协程上的 `.resume()` 抛出了异常，则该异常不应被当前协程捕获，而应从恢复此协程的 `resume()` 调用处传播出去。因此，我们将协程句柄存储在一个在函数顶部声明的变量中，然后 `goto` 到 try/catch 块外部的一个点，并在那里执行对 `.resume()` 的调用。

---

```c++
void __g_resume(__coroutine_state* s) {
    auto* state = static_cast<__g_state*>(s);

    std::coroutine_handle<void> coro_to_resume;

    try {
        switch (state->__suspend_point) {
        case 0: goto suspend_point_0;
        case 1: goto suspend_point_1; // <-- add new jump-table entry
        default: std::unreachable();
        }

suspend_point_0:
        state->__tmp1.get().await_resume();
        state->__tmp1.destroy();

        //  int fx = co_await f(x);
        state->__tmp2.construct_from([&] {
            return f(state->x);
        });
        state->__tmp3.construct_from([&] {
            return static_cast<task&&>(state->__tmp2.get()).operator co_await();
        });
        
        if (!state->__tmp3.get().await_ready()) {
            state->__suspend_point = 1;
            coro_to_resume = state->__tmp3.get().await_suspend(
                std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));
            goto resume_coro;
        }

suspend_point_1:
        int fx = state->__tmp3.get().await_resume();
        state->__tmp3.destroy();
        state->__tmp2.destroy();

        // TODO: Implement
        //  co_return fx * fx;
    } catch (...) {
        state->__promise.unhandled_exception();
        goto final_suspend;
    }

final_suspend:
    // TODO: Implement
    // co_await promise.final_suspend();

resume_coro:
    coro_to_resume.resume();
    return;
}
```

There is a bug in the above code, however. In the case that the `__tmp3.get().await_resume()` call exits with an exception, we would fail to call the destructors of `__tmp3` and `__tmp2` before catching the exception.

Note that we cannot simply catch the exception, call the destructors and rethrow the exception here as this would change the behaviour of those destructors if they were to call `std::unhandled_exceptions()` since the exception would be “handled”. However if the destructor calls this during exception unwind, then call to `std:::unhandled_exceptions()` should return non-zero.

We can instead define an RAII helper class to ensure that the destructors get called on scope exit in the case an exception is thrown.

---
上述代码中存在一个错误，然而。在`__tmp3.get().await_resume()`调用以异常退出的情况下，我们未能在捕获异常之前调用`__tmp3`和`__tmp2`的析构函数。

请注意，我们不能简单地捕获异常，调用析构函数并重新抛出异常，因为如果析构函数调用了`std::unhandled_exceptions()`，这将改变这些析构函数的行为，由于异常会被视为“已处理”。然而，如果析构函数在异常展开过程中调用了`std::unhandled_exceptions()`，那么调用`std::unhandled_exceptions()`应该返回非零值。

我们可以定义一个RAII辅助类，以确保在抛出异常的情况下，析构函数能在作用域退出时被调用。 

---

```c++
template<typename T>
struct destructor_guard {
    explicit destructor_guard(manual_lifetime<T>& obj) noexcept
    : ptr_(std::addressof(obj))
    {}

    // non-movable
    destructor_guard(destructor_guard&&) = delete;
    destructor_guard& operator=(destructor_guard&&) = delete;

    ~destructor_guard() noexcept(std::is_nothrow_destructible_v<T>) {
        if (ptr_ != nullptr) {
            ptr_->destroy();
        }
    }

    void cancel() noexcept { ptr_ = nullptr; }

private:
    manual_lifetime<T>* ptr_;
};

// Partial specialisation for types that don't need their destructors called.
template<typename T>
    requires std::is_trivially_destructible_v<T>
struct destructor_guard<T> {
    explicit destructor_guard(manual_lifetime<T>&) noexcept {}
    void cancel() noexcept {}
};

// Class-template argument deduction to simplify usage
template<typename T>
destructor_guard(manual_lifetime<T>& obj) -> destructor_guard<T>;
```

Using this utility, we can now use this type to ensure that variables stored in the coroutine-state are destroyed when an exception is thrown.

Let’s also use this class to call the destructors of the existing varibles so that it also calls their destructors when they naturally go out of scope.

---
使用这个工具，我们现在可以使用这种类型来确保当抛出异常时，存储在协程状态中的变量会被销毁。

我们同样可以使用这个类去调用已有变量的析构函数，以便在这些变量自然地超出作用域时，也能调用它们的析构函数。 

---

```c++
void __g_resume(__coroutine_state* s) {
    auto* state = static_cast<__g_state*>(s);

    std::coroutine_handle<void> coro_to_resume;

    try {
        switch (state->__suspend_point) {
        case 0: goto suspend_point_0;
        case 1: goto suspend_point_1; // <-- add new jump-table entry
        default: std::unreachable();
        }

suspend_point_0:
        {
            destructor_guard tmp1_dtor{state->__tmp1};
            state->__tmp1.get().await_resume();
        }

        //  int fx = co_await f(x);
        {
            state->__tmp2.construct_from([&] {
                return f(state->x);
            });
            destructor_guard tmp2_dtor{state->__tmp2};

            state->__tmp3.construct_from([&] {
                return static_cast<task&&>(state->__tmp2.get()).operator co_await();
            });
            destructor_guard tmp3_dtor{state->__tmp3};

            if (!state->__tmp3.get().await_ready()) {
                state->__suspend_point = 1;

                coro_to_resume = state->__tmp3.get().await_suspend(
                    std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));

                // A coroutine suspends without exiting scopes.
                // So cancel the destructor-guards.
                tmp3_dtor.cancel();
                tmp2_dtor.cancel();

                goto resume_coro;
            }

            // Don't exit the scope here.
            //
            // We can't 'goto' a label that enters the scope of a variable with a
            // non-trivial destructor. So we have to exit the scope of the destructor
            // guards here without calling the destructors and then recreate them after
            // the `suspend_point_1` label.
            tmp3_dtor.cancel();
            tmp2_dtor.cancel();
        }

suspend_point_1:
        int fx = [&]() -> decltype(auto) {
            destructor_guard tmp2_dtor{state->__tmp2};
            destructor_guard tmp3_dtor{state->__tmp3};
            return state->__tmp3.get().await_resume();
        }();

        // TODO: Implement
        //  co_return fx * fx;
    } catch (...) {
        state->__promise.unhandled_exception();
        goto final_suspend;
    }

final_suspend:
    // TODO: Implement
    // co_await promise.final_suspend();

resume_coro:
    coro_to_resume.resume();
    return;
}
```

Now our coroutine body will now destroy local variables correctly in the presence of any exceptions and will correctly call `promise.unhandled_exception()` if those exceptions propagate out of the coroutine body.

It’s worth noting here that there can also be special handling needed for the case where the `promise.unhandled_exception()` method itself exits with an exception (e.g. if it rethrows the current exception).

In this case, the coroutine would need to catch the exception, mark the coroutine as suspended at a final-suspend-point, and then rethrow the exception.

For example: The `__g_resume()` function’s catch-block would need to look like this:

---
现在我们的协程主体将在出现任何异常的情况下正确销毁局部变量，并且如果这些异常从协程主体中传播出去，会正确调用`promise.unhandled_exception()`。

这里值得注意的是，当`promise.unhandled_exception()`方法本身因异常退出时（例如，如果它重新抛出了当前的异常），可能需要特殊的处理。

在这种情况下，协程需要捕获该异常，将协程标记为在一个最终挂起点处挂起，然后重新抛出该异常。

例如：`__g_resume()`函数的catch块需要像这样处理：

---
```c++
try {
  // ...
} catch (...) {
    try {
        state->__promise.unhandled_exception();
    } catch (...) {
        state->__suspend_point = 2;
        state->__resume = nullptr; // mark as final-suspend-point
        throw;
    }
}
```

and we’d need to add an extra entry to the `__g_destroy` function’s jump table:

---
我们还需要在`__g_destroy`函数的跳转表中添加一个额外的条目：

---

```c++
switch (state->__suspend_point) {
case 0: goto suspend_point_0;
case 1: goto suspend_point_1;
case 2: goto destroy_state; // no variables in scope that need to be destroyed
                            // just destroy the coroutine-state object.
}
```

Note that in this case, the final-suspend-point is not necessarily the same suspend-point as the final-suspend-point as the `co_await promise.final_suspend()` suspend-point.

This is because the `promise.final_suspend()` suspend-point will often have some extra temporary objects related to the `co_await` expression which need to be destroyed when `coroutine_handle::destroy()` is called. Whereas, if `promise.unhandled_exception()` exits with an exception then those temporary objects will not exist and so won’t need to be destroyed by `coroutine_handle::destroy()`.

---
注意，在这种情况下，最终挂起点不一定与`co_await promise.final_suspend()`挂起点相同。

这是因为`promise.final_suspend()`挂起点通常会有一些与`co_await`表达式相关的额外临时对象，在调用`coroutine_handle::destroy()`时需要销毁这些临时对象。而如果`promise.unhandled_exception()`因异常退出，则这些临时对象将不存在，因此不需要由`coroutine_handle::destroy()`来销毁。

---

# Step 11: Implementing `co_return`

The next step is to implement the `co_return fx * fx;` statement.

This is relatively straight-forward compared to some of the previous steps.

The `co_return <expr>` statement gets mapped to:

---
下一步是实现`co_return fx * fx;`语句。

与之前的某些步骤相比，这相对直接。

`co_return <expr>`语句被映射为：

---

```c++
promise.return_value(<expr>);
goto final-suspend-point;
```

So we can simply replace the TODO comment with:

---
所以我们只需将TODO注释替换为：

---

```c++
state->__promise.return_value(fx * fx);
goto final_suspend;
```

Easy.

# Step 12: Implementing `final_suspend()`

The final TODO in the code is now to implement the `co_await promise.final_suspend()` statement.

The `final_suspend()` method returns a temporary `task::promise_type::final_awaiter` type, which will need to be stored in the coroutine-state and destroyed in `__g_destroy`.

This type does not have its own `operator co_await()`, so we don’t need an additional temporary object for the result of that call.

Like the `task::awaiter` type, this also uses the coroutine-handle-returning form of `await_suspend()`. So we need to ensure that we call `resume()` on the returned handle.

If the coroutine does not suspend at the final-suspend-point then the coroutine-state is implicitly destroyed. So we need to delete the state object if execution reaches the end of the coroutine.

Also, as all of the final-suspend logic is required to be noexcept, we don’t need to worry about exceptions being thrown from any of the sub-expressions here.

Let’s first add the data-member to the `__g_state` type.

---
最后代码中的TODO是实现`co_await promise.final_suspend()`语句。

`final_suspend()`方法返回一个临时的`task::promise_type::final_awaiter`类型，这个类型需要被存储在协程状态中并在`__g_destroy`中销毁。

这个类型没有自己的`operator co_await()`，所以我们不需要为那次调用的结果额外准备一个临时对象。

像`task::awaiter`类型一样，这也使用了返回协程句柄形式的`await_suspend()`。因此我们需要确保在返回的句柄上调用`resume()`。

如果协程在最终挂起点不挂起，则协程状态会被隐式销毁。因此，如果执行到达协程末尾，我们需要删除状态对象。

此外，由于所有最终挂起逻辑都要求是noexcept，我们不需要担心这里任何子表达式会抛出异常。

让我们首先向`__g_state`类型添加数据成员。

---

```c++
struct __g_state : __coroutine_state_with_promise<__g_promise_t> {
    __g_state(int&& __x);
    ~__g_state();

    int __suspend_point = 0;
    int x;
    manual_lifetime<std::suspend_always> __tmp1;
    manual_lifetime<task> __tmp2;
    manual_lifetime<task::awaiter> __tmp3;
    manual_lifetime<task::promise_type::final_awaiter> __tmp4; // <---
};
```

Then we can implement the body of the final-suspend expression as follows:

---
接着，我们可以按如下方式实现`final_suspend`表达式的主体：

---

```c++
final_suspend:
    // co_await promise.final_suspend
    {
        state->__tmp4.construct_from([&]() noexcept {
            return state->__promise.final_suspend();
        });
        destructor_guard tmp4_dtor{state->__tmp4};

        if (!state->__tmp4.get().await_ready()) {
            state->__suspend_point = 2;
            state->__resume = nullptr; // mark as final suspend-point

            coro_to_resume = state->__tmp4.get().await_suspend(
                std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));

            tmp4_dtor.cancel();
            goto resume_coro;
        }

        state->__tmp4.get().await_resume();
    }

    //  Destroy coroutine-state if execution flows off end of coroutine
    delete state;
    return;
```

And now we also need to update the `__g_destroy` function to handle this new suspend-point.

---
现在我们还需要更新`__g_destroy`函数以处理这个新的挂起点。

---

```c++
void __g_destroy(__coroutine_state* state) {
    auto* state = static_cast<__g_state*>(s);

    switch (state->__suspend_point) {
    case 0: goto suspend_point_0;
    case 1: goto suspend_point_1;
    case 2: goto suspend_point_2;
    default: std::unreachable();
    }

suspend_point_0:
    state->__tmp1.destroy();
    goto destroy_state;

suspend_point_1:
    state->__tmp3.destroy();
    state->__tmp2.destroy();
    goto destroy_state;

suspend_point_2:
    state->__tmp4.destroy();
    goto destroy_state;

destroy_state:
    delete state;
}
```

We now have a fully functional lowering of the `g()` coroutine function.

We’re done! That’s it!

Or is it….

---
我们现在有了一个功能完全的`g()`协程函数的降级版本。

我们完成了！就是这样！

或者，真的是这样吗….

---

# Step 13: Implementing symmetric-transfer and the noop-coroutine

It turns out there is actually a problem with the way we have implemented our `__g_resume()` function above.

The problems with this were discussed in more detail in the previous blog post so if you want to understand the problem more deeply please take a look at the post [C++ Coroutines: Understanding Symmetric Transfer](https://lewissbaker.github.io/2020/05/11/understanding_symmetric_transfer).

The specification for [[expr.await\]](https://eel.is/c++draft/expr.await) gives a little hint about how we should be handling the coroutine-handle-returning flavour of `await_suspend`:

> If the type of *await-suspend* is `std::coroutine_­handle<Z>`, *await-suspend*`.resume()` is evaluated.
>
> [*Note* 1: This resumes the coroutine referred to by the result of *await-suspend*. Any number of coroutines can be successively resumed in this fashion, eventually returning control flow to the current coroutine caller or resumer ([[dcl.fct.def.coroutine\]](https://eel.is/c++draft/dcl.fct.def.coroutine)). —- *end note*]

The note there, while non-normative and thus non-binding, is strongly encouraging compilers to implement this in such a way that it performs a tail-call to resume the next coroutine rather than resuming the next coroutine recursively. This is because resuming the next coroutine recursively can easily lead to unbounded stack growth if coroutines resume each other in a loop.

The problem is that we are calling `.resume()` on the next coroutine from within the body of the `__g_resume()` function and then returning, so the stack space used by the `__g_resume()` frame is not freed until after the next coroutine suspends and returns.

Compilers are able to do this by implementing the resumption of the next coroutine as a tail-call. In this way, the compiler generates code that first pops the the current stack frame, preserving the return-address, and then executes a `jmp` to the next coroutine’s resume-function.

As we don’t have a mechanism in C++ to specify that a function-call in the tail-position should be a tail-call we will need to instead actually return from the resume-function so that its stack-space can be freed, and then have the caller resume the next coroutine.

As the next coroutine may also need to resume another coroutine when it suspends, and this may happen indefinitely, the caller will need to resume the coroutines in a loop.

Such a loop is typically called a “trampoline loop” as we return back to the loop from one coroutine and then “bounce” off the loop back into the next coroutine.

If we modify the signature of the resume-function to return a pointer to the next coroutine’s coroutine-state instead of returning void, then the `coroutine_handle::resume()` function can then just immediately call the `__resume()` function-pointer for the next coroutine to resume it.

Let’s change the signature of the `__resume_fn` for a `__coroutine_state`:

---
原来我们上面实现的`__g_resume()`函数确实存在问题。

这些问题在之前的博客文章中有更详细的讨论，如果你想更深入地理解这个问题，请参阅文章[C++协程：理解对称转移](https://lewissbaker.github.io/2020/05/11/understanding_symmetric_transfer)。

[[expr.await]](https://eel.is/c++draft/expr.await)规范对于如何处理返回协程句柄类型的`await_suspend`提供了一点提示：

> 如果*await-suspend*的类型是`std::coroutine_­handle<Z>`，则会计算*await-suspend*`.resume()`。
>
> [*注* 1：这将恢复由*await-suspend*的结果所指向的协程。通过这种方式，任意数量的协程可以依次被恢复，最终将控制流返回给当前协程的调用者或恢复者([[dcl.fct.def.coroutine]](https://eel.is/c++draft/dcl.fct.def.coroutine))。—— *结束注释*]

这里的注释虽然是非规范性的，因此没有约束力，但它强烈建议编译器以这样的方式实现：执行尾调用来恢复下一个协程，而不是递归地恢复下一个协程。这是因为如果协程之间在一个循环中相互恢复，递归恢复下一个协程很容易导致栈空间无界增长。

问题在于我们在`__g_resume()`函数体内对下一个协程调用了`.resume()`然后返回，这样，在下一个协程挂起并返回之前，`__g_resume()`帧使用的栈空间不会被释放。

编译器可以通过将下一个协程的恢复实现为尾调用来解决这个问题。这样，编译器生成的代码首先弹出当前的栈帧，保留返回地址，然后执行一个`jmp`到下一个协程的恢复函数。

由于C++中没有机制指定尾位置的函数调用应作为尾调用，我们需要实际从恢复函数返回，以便释放其栈空间，然后让调用者恢复下一个协程。

由于下一个协程在暂停时可能也需要恢复另一个协程，并且这种情况可能会无限期发生，调用者需要在一个循环中恢复协程。

这种循环通常被称为“蹦床循环”，因为我们从一个协程返回到这个循环，然后从循环“弹”回下一个协程。

如果我们修改恢复函数的签名，使其返回下一个协程的协程状态指针而不是返回void，则`coroutine_handle::resume()`函数可以直接调用下一个要恢复的协程的`__resume()`函数指针。

让我们改变`__coroutine_state`的`__resume_fn`的签名。

---

```c++
struct __coroutine_state {
    using __resume_fn = __coroutine_state* (__coroutine_state*);
    using __destroy_fn = void (__coroutine_state*);

    __resume_fn* __resume;
    __destroy_fn* __destroy;
};
```

Then we can write the `coroutine_handle::resume()` function something like this:

---
然后，我们可以将`coroutine_handle::resume()`函数写成这样：

---

```c++
void std::coroutine_handle<void>::resume() const {
    __coroutine_state* s = state_;
    do {
        s = s->__resume(s);
    } while (/* some condition */);
}
```

The next question then becomes: “What should the condition be?”

This is where the `std::noop_coroutine()` helper comes into the picture.

The `std::noop_coroutine()` is a factory function that returns a special coroutine handle that has a no-op `resume()` and `destroy()` method. If a coroutine suspends and returns the noop-coroutine-handle from the `await_suspend()` method then this indicates that there is no more coroutine to resume and that the invocation of `coroutien_handle::resume()` that resumed this coroutine should return to its caller.

So we need to implement `std::noop_coroutine()` and the condition in `coroutine_handle::resume()` so that the condition returns false and the loop exits when the `__coroutine_state` pointer points to the noop-coroutine-state.

One strategy we can use here is to define a static instance of `__coroutine_state` that is designated as the noop-coroutine-state. The `std::noop_coroutine()` function can return a coroutine-handle that points to this object, and we can compare the `__coroutine_state` pointer to the address of that object to see if a particular coroutine handle is the noop-coroutine.

First let’s define this special noop-coroutine-state object:

---
下一个问题就变成了：“条件应该是什么？”

这里就是`std::noop_coroutine()`助手函数发挥作用的地方。

`std::noop_coroutine()`是一个工厂函数，它返回一个特殊的协程句柄，该句柄的`resume()`和`destroy()`方法都是空操作（no-op）。如果一个协程暂停并从`await_suspend()`方法返回了空操作协程句柄，这表明没有更多的协程需要恢复，并且恢复此协程的`coroutien_handle::resume()`调用应该返回到其调用者。

因此，我们需要实现`std::noop_coroutine()`以及在`coroutine_handle::resume()`中的条件，以便当`__coroutine_state`指针指向空操作协程状态时，条件返回false且循环退出。

我们可以在这里使用的一种策略是定义一个被指定为空操作协程状态的`__coroutine_state`静态实例。`std::noop_coroutine()`函数可以返回指向这个对象的协程句柄，我们可以通过将`__coroutine_state`指针与该对象的地址进行比较来查看特定的协程句柄是否为空操作协程。

首先，让我们定义这个特殊的空操作协程状态对象：

---

```c++
struct __coroutine_state {
    using __resume_fn = __coroutine_state* (__coroutine_state*);
    using __destroy_fn = void (__coroutine_state*);

    __resume_fn* __resume;
    __destroy_fn* __destroy;

    static __coroutine_state* __noop_resume(__coroutine_state* state) noexcept {
        return state;
    }

    static void __noop_destroy(__coroutine_state*) noexcept {}

    static const __coroutine_state __noop_coroutine;
};

inline const __coroutine_state __coroutine_state::__noop_coroutine{
    &__coroutine_state::__noop_resume,
    &__coroutine_state::__noop_destroy
};
```

Then we can implement the `std::coroutine_handle<noop_coroutine_promise>` specialisation.

---
然后，我们可以实现`std::coroutine_handle<noop_coroutine_promise>`的特化版本。

---

```c++
namespace std
{
    struct noop_coroutine_promise {};

    using noop_coroutine_handle = coroutine_handle<noop_coroutine_promise>;

    noop_coroutine_handle noop_coroutine() noexcept;

    template<>
    class coroutine_handle<noop_coroutine_promise> {
    public:
        constexpr coroutine_handle(const coroutine_handle&) noexcept = default;
        constexpr coroutine_handle& operator=(const coroutine_handle&) noexcept = default;

        constexpr explicit operator bool() noexcept { return true; }

        constexpr friend bool operator==(coroutine_handle, coroutine_handle) noexcept {
            return true;
        }

        operator coroutine_handle<void>() const noexcept {
            return coroutine_handle<void>::from_address(address());
        }

        noop_coroutine_promise& promise() const noexcept {
            static noop_coroutine_promise promise;
            return promise;
        }

        constexpr void resume() const noexcept {}
        constexpr void destroy() const noexcept {}
        constexpr bool done() const noexcept { return false; }

        constexpr void* address() const noexcept {
            return const_cast<__coroutine_state*>(&__coroutine_state::__noop_coroutine);
        }
    private:
        constexpr coroutine_handle() noexcept = default;

        friend noop_coroutine_handle noop_coroutine() noexcept {
            return {};
        }
    };
}
```

And we can update `coroutine_handle::resume()` to exit when the noop-coroutine-state is returned.

---
我们可以更新`coroutine_handle::resume()`，以便在返回空操作协程状态时退出。

---

```c++
void std::coroutine_handle<void>::resume() const {
    __coroutine_state* s = state_;
    do {
        s = s->__resume(s);
    } while (s != &__coroutine_state::__noop_coroutine);
}
```

And finally, we can update our `__g_resume()` function to now return the `__coroutine_state*`.

This just involves updating the signature and replacing:

---
最后，我们可以更新我们的`__g_resume()`函数，使其现在返回`__coroutine_state*`。

这仅仅涉及到更新函数签名并替换相关内容。 

---

```c++
coro_to_resume = ...;
goto resume_coro;
```

with

```c++
auto h = ...;
return static_cast<__coroutine_state*>(h.address());
```

and then at the very end of the function (after the `delete state;` statement) adding

---
并在函数的最后（在`delete state;`语句之后）添加

---

```c++
return static_cast<__coroutine_state*>(std::noop_coroutine().address());
```

# One last thing

Those with a keen eye may have noticed that the coroutine-state type `__g_state` is actually larger than it needs to be.

The data-members for the 4 temporary values each reserve storage for their respective values. However, the lifetimes of some of the temporary values do not overlap and so in theory we can save space in the coroutine-state by reusing the storage of an object for the next object after its lifetime has ended.

To be able to take advantage of this we can instead define the data-members in an anonymous union where appropriate.

Looking at the lifetimes of the temporary varaibles we have:

- `__tmp1` - exists only within `co_await promise.initial_suspend();` statement
- `__tmp2` - exists only within `int fx = co_await f(x);` statement
- `__tmp3` - exists only within `int fx = co_await f(x);` statement - nested inside lifetime of `__tmp2`
- `__tmp4` - exists only within `co_await promise.final_suspend();` statement

Since lifetimes of `__tmp2` and `__tmp3` overlap we must place them in a struct together as they both need to exist at the same time.

However, the `__tmp1` and `__tmp4` members do not have lifetimes that overlap and so they can be placed together in an anonymous `union`.

Thus we can change our data-member definition to:

---
那些眼光敏锐的人可能已经注意到，协程状态类型`__g_state`实际上比它需要的要大。

四个临时值的每个数据成员都分别为其各自的值保留了存储空间。然而，某些临时值的生命周期不会重叠，因此理论上我们可以通过在前一个对象的生命期结束后重复使用它的存储空间来节省协程状态中的空间。

为了能够利用这一点，我们可以适当地将数据成员定义在一个匿名联合体内。

查看临时变量的生命期，我们有：

- `__tmp1` —— 仅存在于`co_await promise.initial_suspend();`语句内
- `__tmp2` —— 仅存在于`int fx = co_await f(x);`语句内
- `__tmp3` —— 仅存在于`int fx = co_await f(x);`语句内 —— 嵌套在`__tmp2`的生命期内
- `__tmp4` —— 仅存在于`co_await promise.final_suspend();`语句内

由于`__tmp2`和`__tmp3`的生命期重叠，我们必须将它们一起放在一个结构体中，因为它们都需要同时存在。

然而，`__tmp1`和`__tmp4`成员的生命期不重叠，因此它们可以被放置在一个匿名的`union`中。

因此，我们可以将数据成员定义更改为：

---

```c++
struct __g_state : __coroutine_state_with_promise<__g_promise_t> {
    __g_state(int&& x);
    ~__g_state();

    int __suspend_point = 0;
    int x;

    struct __scope1 {
        manual_lifetime<task> __tmp2;
        manual_lifetime<task::awaiter> __tmp3;
    };

    union {
        manual_lifetime<std::suspend_always> __tmp1;
        __scope1 __s1;
        manual_lifetime<task::promise_type::final_awaiter> __tmp4;
    };
};
```

Then, because the `__tmp2` and `__tmp3` variables are now nested inside the `__s1` object, we need to update references to them to now be e.g. `state->__s1.tmp2`. But otherwise the rest of the code stays the same.

This should save an additional 16 bytes of the coroutine-state size as we no longer need extra storage + padding for the `__tmp1` and `__tmp4` data-members - which would otherwise be padded to the size of a pointer, despite being empty types.

---
然后，因为`__tmp2`和`__tmp3`变量现在嵌套在`__s1`对象内，我们需要更新对它们的引用，使其变为例如`state->__s1.tmp2`。除此之外，其余代码保持不变。

这应该能额外节省协程状态大小的16字节，因为我们不再需要为`__tmp1`和`__tmp4`数据成员提供额外的存储空间和填充——否则的话，即使它们是空类型，也会被填充到指针的大小。

---

# Tying it all together

Ok, so the final code we have generated for the coroutine function:

---
好的，所以我们为协程函数生成的最终代码为：

---

```c++
task g(int x) {
    int fx = co_await f(x);
    co_return fx * fx;
}
```

is the following:

```c++
/////
// The coroutine promise-type

using __g_promise_t = std::coroutine_traits<task, int>::promise_type;

__coroutine_state* __g_resume(__coroutine_state* s);
void __g_destroy(__coroutine_state* s);

/////
// The coroutine-state definition

struct __g_state : __coroutine_state_with_promise<__g_promise_t> {
    __g_state(int&& x)
    : x(static_cast<int&&>(x)) {
        // Initialise the function-pointers used by coroutine_handle methods.
        this->__resume = &__g_resume;
        this->__destroy = &__g_destroy;

        // Use placement-new to initialise the promise object in the base-class
        // after we've initialised the argument copies.
        ::new ((void*)std::addressof(this->__promise))
            __g_promise_t(construct_promise<__g_promise_t>(this->x));
    }

    ~__g_state() {
        this->__promise.~__g_promise_t();
    }

    int __suspend_point = 0;

    // Argument copies
    int x;

    // Local variables/temporaries
    struct __scope1 {
        manual_lifetime<task> __tmp2;
        manual_lifetime<task::awaiter> __tmp3;
    };

    union {
        manual_lifetime<std::suspend_always> __tmp1;
        __scope1 __s1;
        manual_lifetime<task::promise_type::final_awaiter> __tmp4;
    };
};

/////
// The "ramp" function

task g(int x) {
    std::unique_ptr<__g_state> state(new __g_state(static_cast<int&&>(x)));
    decltype(auto) return_value = state->__promise.get_return_object();

    state->__tmp1.construct_from([&]() -> decltype(auto) {
        return state->__promise.initial_suspend();
    });
    if (!state->__tmp1.get().await_ready()) {
        state->__tmp1.get().await_suspend(
            std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));
        state.release();
        // fall through to return statement below.
    } else {
        // Coroutine did not suspend. Start executing the body immediately.
        __g_resume(state.release());
    }
    return return_value;
}

/////
//  The "resume" function

__coroutine_state* __g_resume(__coroutine_state* s) {
    auto* state = static_cast<__g_state*>(s);

    try {
        switch (state->__suspend_point) {
        case 0: goto suspend_point_0;
        case 1: goto suspend_point_1; // <-- add new jump-table entry
        default: std::unreachable();
        }

suspend_point_0:
        {
            destructor_guard tmp1_dtor{state->__tmp1};
            state->__tmp1.get().await_resume();
        }

        //  int fx = co_await f(x);
        {
            state->__s1.__tmp2.construct_from([&] {
                return f(state->x);
            });
            destructor_guard tmp2_dtor{state->__s1.__tmp2};

            state->__s1.__tmp3.construct_from([&] {
                return static_cast<task&&>(state->__s1.__tmp2.get()).operator co_await();
            });
            destructor_guard tmp3_dtor{state->__s1.__tmp3};

            if (!state->__s1.__tmp3.get().await_ready()) {
                state->__suspend_point = 1;

                auto h = state->__s1.__tmp3.get().await_suspend(
                    std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));

                // A coroutine suspends without exiting scopes.
                // So cancel the destructor-guards.
                tmp3_dtor.cancel();
                tmp2_dtor.cancel();

                return static_cast<__coroutine_state*>(h.address());
            }

            // Don't exit the scope here.
            // We can't 'goto' a label that enters the scope of a variable with a
            // non-trivial destructor. So we have to exit the scope of the destructor
            // guards here without calling the destructors and then recreate them after
            // the `suspend_point_1` label.
            tmp3_dtor.cancel();
            tmp2_dtor.cancel();
        }

suspend_point_1:
        int fx = [&]() -> decltype(auto) {
            destructor_guard tmp2_dtor{state->__s1.__tmp2};
            destructor_guard tmp3_dtor{state->__s1.__tmp3};
            return state->__s1.__tmp3.get().await_resume();
        }();

        //  co_return fx * fx;
        state->__promise.return_value(fx * fx);
        goto final_suspend;
    } catch (...) {
        state->__promise.unhandled_exception();
        goto final_suspend;
    }

final_suspend:
    // co_await promise.final_suspend
    {
        state->__tmp4.construct_from([&]() noexcept {
            return state->__promise.final_suspend();
        });
        destructor_guard tmp4_dtor{state->__tmp4};

        if (!state->__tmp4.get().await_ready()) {
            state->__suspend_point = 2;
            state->__resume = nullptr; // mark as final suspend-point

            auto h = state->__tmp4.get().await_suspend(
                std::coroutine_handle<__g_promise_t>::from_promise(state->__promise));

            tmp4_dtor.cancel();
            return static_cast<__coroutine_state*>(h.address());
        }

        state->__tmp4.get().await_resume();
    }

    //  Destroy coroutine-state if execution flows off end of coroutine
    delete state;

    return static_cast<__coroutine_state*>(std::noop_coroutine().address());
}

/////
// The "destroy" function

void __g_destroy(__coroutine_state* s) {
    auto* state = static_cast<__g_state*>(s);

    switch (state->__suspend_point) {
    case 0: goto suspend_point_0;
    case 1: goto suspend_point_1;
    case 2: goto suspend_point_2;
    default: std::unreachable();
    }

suspend_point_0:
    state->__tmp1.destroy();
    goto destroy_state;

suspend_point_1:
    state->__s1.__tmp3.destroy();
    state->__s1.__tmp2.destroy();
    goto destroy_state;

suspend_point_2:
    state->__tmp4.destroy();
    goto destroy_state;

destroy_state:
    delete state;
}
```

For a fully compilable version of the final code, see: https://godbolt.org/z/xaj3Yxabn

This concludes the 5-part series on understanding the mechanics of C++ coroutines.

This is probably more information than you ever wanted to know about coroutines, but hopefully it helps you to understand what’s going on under the hood and demystifies them just a bit.

Thanks for making it through to the end!

Until next time, Lewis.

---
如需获取最终代码的完整可编译版本，请访问：https://godbolt.org/z/xaj3Yxabn 。

这就结束了关于理解C++ 协程机制的五部分系列内容。

这可能包含了比你期望了解的更多关于协程的信息，但希望它能帮助你理解其底层原理，揭开协程的些许神秘面纱。

感谢你一直看到最后！

期待下次再见，Lewis。 

---