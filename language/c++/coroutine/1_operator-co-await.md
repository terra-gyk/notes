# C++ Coroutines: Understanding operator co_await

Nov 17, 2017

In the previous post on [Coroutine Theory](https://lewissbaker.github.io/2017/09/25/coroutine-theory) I described the high-level differences between functions and coroutines but without going into any detail on syntax and semantics of coroutines as described by the C++ Coroutines TS ([N4680](http://www.open-std.org/jtc1/sc22/wg21/docs/papers/2017/n4680.pdf)).

The key new facility that the Coroutines TS adds to the C++ language is the ability to suspend a coroutine, allowing it to be later resumed. The mechanism the TS provides for doing this is via the new `co_await` operator.

Understanding how the `co_await` operator works can help to demystify the behaviour of coroutines and how they are suspended and resumed. In this post I will be explaining the mechanics of the `co_await` operator and introduce the related **Awaitable** and **Awaiter** type concepts.

But before I dive into `co_await` I want to give a brief overview of the Coroutines TS to provide some context.

---
在上一篇关于协程理论的文章中，我描述了函数和协程之间的高层次差异，但没有详细说明 C++ 协程技术规范（N4680）中描述的协程的语法和语义。

协程技术规范为 C++ 语言添加的关键新功能是能够挂起协程，使其可以在稍后恢复。技术规范提供的实现这一功能的机制是通过新的 co_await 操作符。

理解 co_await 操作符的工作原理有助于解开协程的行为及其挂起和恢复过程的神秘面纱。在这篇文章中，我将解释 co_await 操作符的机制，并介绍相关的 Awaitable 和 Awaiter 类型概念。

但在深入讲解 co_await 之前，我想简要概述一下协程技术规范，以提供一些背景信息。

---

## What does the Coroutines TS give us?

- Three new language keywords: `co_await`, `co_yield` and `co_return`

- Several new types in the `std::experimental` namespace:
  - `coroutine_handle<P>`
  - `coroutine_traits<Ts...>`
  - `suspend_always`
  - `suspend_never`

- A general mechanism that library writers can use to interact with coroutines and customise their behaviour.

- A language facility that makes writing asynchronous code a whole lot easier!

The facilities the C++ Coroutines TS provides in the language can be thought of as a *low-level assembly-language* for coroutines. These facilities can be difficult to use directly in a safe way and are mainly intended to be used by library-writers to build higher-level abstractions that application developers can work with safely.

The plan is to deliver these new low-level facilities into an upcoming language standard (hopefully C++20) along with some accompanying higher-level types in the standard library that wrap these low-level building-blocks and make coroutines more accessible in a safe way for application developers.

---

- 三个新的语言关键字：`co_await`、`co_yield` 和 `co_return`

- `std::experimental` 命名空间中的几个新类型：
  - `coroutine_handle<P>`
  - `coroutine_traits<Ts...>`
  - `suspend_always`
  - `suspend_never`

- 一种通用机制，库编写者可以使用它与协程交互并自定义其行为。

- 一种使编写异步代码变得更加容易的语言设施！

C++ 协程技术规范在语言中提供的设施可以被视为协程的 *低级汇编语言*。这些设施难以直接安全地使用，主要是为库编写者设计的，用于构建更高层次的抽象，使应用程序开发者能够安全地使用。

计划是将这些新的低级设施引入即将到来的语言标准（希望是 C++20），以及一些伴随的标准库中的高层次类型，这些类型封装了这些低级构建块，使应用程序开发者能够以安全的方式更方便地使用协程。

---


## Compiler <-> Library interaction

Interestingly, the Coroutines TS does not actually define the semantics of a coroutine. It does not define how to produce the value returned to the caller. It does not define what to do with the return value passed to the `co_return` statement or how to handle an exception that propagates out of the coroutine. It does not define what thread the coroutine should be resumed on.

Instead, it specifies a general mechanism for library code to customise the behaviour of the coroutine by implementing types that conform to a specific interface. The compiler then generates code that calls methods on instances of types provided by the library. This approach is similar to the way that a library-writer can customise the behaviour of a range-based for-loop by defining the `begin()`/`end()` methods and an `iterator` type.

The fact that the Coroutines TS doesn’t prescribe any particular semantics to the mechanics of a coroutine makes it a powerful tool. It allows library writers to define many different kinds of coroutines, for all sorts of different purposes.

For example, you can define a coroutine that produces a single value asynchronously, or a coroutine that produces a sequence of values lazily, or a coroutine that simplifies control-flow for consuming `optional<T>` values by early-exiting if a `nullopt` value is encountered.

There are two kinds of interfaces that are defined by the coroutines TS: The **Promise** interface and the **Awaitable** interface.

The **Promise** interface specifies methods for customising the behaviour of the coroutine itself. The library-writer is able to customise what happens when the coroutine is called, what happens when the coroutine returns (either by normal means or via an unhandled exception) and customise the behaviour of any `co_await` or `co_yield` expression within the coroutine.

The **Awaitable** interface specifies methods that control the semantics of a `co_await` expression. When a value is `co_await`ed, the code is translated into a series of calls to methods on the awaitable object that allow it to specify: whether to suspend the current coroutine, execute some logic after it has suspended to schedule the coroutine for later resumption, and execute some logic after the coroutine resumes to produce the result of the `co_await` expression.

I’ll be covering details of the **Promise** interface in a future post, but for now let’s look at the **Awaitable** interface.

---
有趣的是，协程技术规范（Coroutines TS）实际上并没有定义协程的语义。它没有定义如何生成返回给调用者的值。它没有定义如何处理传递给 co_return 语句的返回值，或者如何处理从协程中传播出来的异常。它也没有定义应在哪个线程上恢复协程。

相反，它规定了一种通用机制，通过实现符合特定接口的类型，库代码可以自定义协程的行为。然后，编译器生成调用库提供的类型的实例上的方法的代码。这种方法类似于库编写者通过定义 begin()/end() 方法和一个 iterator 类型来自定义基于范围的 for 循环行为的方式。

协程技术规范不对协程的机制规定任何特定语义这一事实使其成为一个强大的工具。它允许库编写者为各种不同的目的定义许多不同种类的协程。

例如，你可以定义一个异步生成单个值的协程，或者定义一个惰性生成一系列值的协程，或者定义一个通过在遇到 nullopt 值时提前退出来简化控制流以使用 optional<T> 值的协程。

协程技术规范定义了两种接口：Promise 接口和 Awaitable 接口。

Promise 接口指定了用于自定义协程本身行为的方法。库编写者能够自定义当协程被调用时发生什么，当协程返回时（无论是正常方式还是通过未处理的异常）发生什么，以及自定义协程内任何 co_await 或 co_yield 表达式的行为。

Awaitable 接口指定了控制 co_await 表达式的语义的方法。当一个值被 co_await 时，代码会被转换成一系列对可等待对象上的方法的调用，这些方法允许其指定：是否挂起当前协程，在挂起后执行一些逻辑以调度协程以便稍后恢复，以及在协程恢复后执行一些逻辑以生成 co_await 表达式的结果。

我将在以后的文章中详细介绍 Promise 接口，但目前让我们先看看 Awaitable 接口。

---

## Awaiters and Awaitables: Explaining `operator co_await`

The `co_await` operator is a new unary operator that can be applied to a value. For example: `co_await someValue`.

The `co_await` operator can only be used within the context of a coroutine. This is somewhat of a tautology though, since any function body containing use of the `co_await` operator, by definition, will be compiled as a coroutine.

A type that supports the `co_await` operator is called an **Awaitable** type.

Note that whether or not the `co_await` operator can be applied to a type can depend on the context in which the `co_await` expression appears. The promise type used for a coroutine can alter the meaning of a `co_await` expression within the coroutine via its `await_transform` method (more on this later).

To be more specific where required I like to use the term **Normally Awaitable** to describe a type that supports the `co_await` operator in a coroutine context whose promise type does not have an `await_transform` member. And I like to use the term **Contextually Awaitable** to describe a type that only supports the `co_await` operator in the context of certain types of coroutines due to the presence of an `await_transform` method in the coroutine’s promise type. (I’m open to better suggestions for these names here…)

An **Awaiter** type is a type that implements the three special methods that are called as part of a `co_await` expression: `await_ready`, `await_suspend` and `await_resume`.

Note that I have shamelessly “borrowed” the term ‘Awaiter’ here from the C# `async` keyword’s mechanics that is implemented in terms of a `GetAwaiter()` method which returns an object with an interface that is eerily similar to the C++ concept of an **Awaiter**. See [this post](https://weblogs.asp.net/dixin/understanding-c-sharp-async-await-2-awaitable-awaiter-pattern) for more details on C# awaiters.

Note that a type can be both an **Awaitable** type and an **Awaiter** type.

When the compiler sees a `co_await <expr>` expression there are actually a number of possible things it could be translated to depending on the types involved.

---
co_await 操作符是一个可以应用于值的新一元操作符。例如：co_await someValue。

co_await 操作符只能在协程的上下文中使用。尽管这有点同义反复，因为任何包含 co_await 操作符的函数体，根据定义，都会被编译为协程。

支持 co_await 操作符的类型称为 Awaitable 类型。

请注意，是否可以对某个类型应用 co_await 操作符可能取决于 co_await 表达式出现的上下文。用于协程的承诺类型可以通过其 await_transform 方法改变协程内 co_await 表达式的含义（稍后会详细介绍）。

为了在需要时更加具体，我喜欢使用术语 Normally Awaitable 来描述在协程上下文中支持 co_await 操作符且承诺类型没有 await_transform 成员的类型。我喜欢使用术语 Contextually Awaitable 来描述由于协程的承诺类型中存在 await_transform 方法而仅在某些类型的协程上下文中支持 co_await 操作符的类型。（我对此名称持开放态度，欢迎更好的建议……）

一个 Awaiter 类型是实现了作为 co_await 表达式一部分调用的三个特殊方法的类型：await_ready、await_suspend 和 await_resume。

请注意，我在这里毫不客气地“借用”了 C# async 关键字机制中的术语 ‘Awaiter’，它通过返回具有与 C++ 的 Awaiter 概念惊人相似接口的对象的 GetAwaiter() 方法实现。有关 C# awaiters 的更多细节，请参阅这篇文章。

请注意，一个类型可以同时是 Awaitable 类型和 Awaiter 类型。

当编译器看到 co_await <expr> 表达式时，实际上有多种可能的转换方式，具体取决于涉及的类型。

---

### Obtaining the Awaiter

The first thing the compiler does is generate code to obtain the **Awaiter** object for the awaited value. There are a number of steps to obtaining the awaiter object which are set out in N4680 section 5.3.8(3).

Let’s assume that the promise object for the awaiting coroutine has type, `P`, and that `promise` is an l-value reference to the promise object for the current coroutine.

If the promise type, `P`, has a member named `await_transform` then `<expr>` is first passed into a call to `promise.await_transform(<expr>)` to obtain the **Awaitable** value, `awaitable`. Otherwise, if the promise type does not have an `await_transform` member then we use the result of evaluating `<expr>` directly as the **Awaitable** object, `awaitable`.

Then, if the **Awaitable** object, `awaitable`, has an applicable `operator co_await()` overload then this is called to obtain the **Awaiter** object. Otherwise the object, `awaitable`, is used as the awaiter object.

If we were to encode these rules into the functions `get_awaitable()` and `get_awaiter()`, they might look something like this:

---
编译器首先生成代码以获取被等待值的 Awaiter 对象。获取 awaiter 对象有多个步骤，这些步骤在 N4680 第 5.3.8(3) 节中有所说明。

假设等待协程的承诺对象类型为 P，并且 promise 是当前协程承诺对象的左值引用。

如果承诺类型 P 有一个名为 await_transform 的成员，则 <expr> 首先被传递给 promise.await_transform(<expr>) 调用，以获得 Awaitable 值 awaitable。否则，如果承诺类型没有 await_transform 成员，则直接使用 <expr> 的计算结果作为 Awaitable 对象 awaitable。

然后，如果 Awaitable 对象 awaitable 有一个适用的 operator co_await() 重载，则调用该重载以获得 Awaiter 对象。否则，使用对象 awaitable 本身作为 awaiter 对象。

如果我们把这些规则编码到函数 get_awaitable() 和 get_awaiter() 中，它们可能看起来像这样：

---

```c++
template<typename P, typename T>
decltype(auto) get_awaitable(P& promise, T&& expr)
{
  if constexpr (has_any_await_transform_member_v<P>)
    return promise.await_transform(static_cast<T&&>(expr));
  else
    return static_cast<T&&>(expr);
}

template<typename Awaitable>
decltype(auto) get_awaiter(Awaitable&& awaitable)
{
  if constexpr (has_member_operator_co_await_v<Awaitable>)
    return static_cast<Awaitable&&>(awaitable).operator co_await();
  else if constexpr (has_non_member_operator_co_await_v<Awaitable&&>)
    return operator co_await(static_cast<Awaitable&&>(awaitable));
  else
    return static_cast<Awaitable&&>(awaitable);
}
```

### Awaiting the Awaiter

So, assuming we have encapsulated the logic for turning the `<expr>` result into an **Awaiter** object into the above functions then the semantics of `co_await <expr>` can be translated (roughly) as follows:

---
因此，假设我们将把 <expr> 结果转换为 Awaiter 对象的逻辑封装在上述函数中，则 co_await <expr> 的语义可以（大致）翻译如下：

---
```c++
{
  auto&& value = <expr>;
  auto&& awaitable = get_awaitable(promise, static_cast<decltype(value)>(value));
  auto&& awaiter = get_awaiter(static_cast<decltype(awaitable)>(awaitable));
  if (!awaiter.await_ready())
  {
    using handle_t = std::experimental::coroutine_handle<P>;

    using await_suspend_result_t =
      decltype(awaiter.await_suspend(handle_t::from_promise(promise)));

    <suspend-coroutine>

    if constexpr (std::is_void_v<await_suspend_result_t>)
    {
      awaiter.await_suspend(handle_t::from_promise(promise));
      <return-to-caller-or-resumer>
    }
    else
    {
      static_assert(
         std::is_same_v<await_suspend_result_t, bool>,
         "await_suspend() must return 'void' or 'bool'.");

      if (awaiter.await_suspend(handle_t::from_promise(promise)))
      {
        <return-to-caller-or-resumer>
      }
    }

    <resume-point>
  }

  return awaiter.await_resume();
}
```

The `void`-returning version of `await_suspend()` unconditionally transfers execution back to the caller/resumer of the coroutine when the call to `await_suspend()` returns, whereas the `bool`-returning version allows the awaiter object to conditionally resume the coroutine immediately without returning to the caller/resumer.

The `bool`-returning version of `await_suspend()` can be useful in cases where the awaiter might start an async operation that can sometimes complete synchronously. In the cases where it completes synchronously, the `await_suspend()` method can return `false` to indicate that the coroutine should be immediately resumed and continue execution.

At the `<suspend-coroutine>` point the compiler generates some code to save the current state of the coroutine and prepare it for resumption. This includes storing the location of the `<resume-point>` as well as spilling any values currently held in registers into the coroutine frame memory.

The current coroutine is considered suspended after the `<suspend-coroutine>` operation completes. The first point at which you can observe the suspended coroutine is inside the call to `await_suspend()`. Once the coroutine is suspended it is then able to be resumed or destroyed.

It is the responsibility of the `await_suspend()` method to schedule the coroutine for resumption (or destruction) at some point in the future once the operation has completed. Note that returning `false` from `await_suspend()` counts as scheduling the coroutine for immediate resumption on the current thread.

The purpose of the `await_ready()` method is to allow you to avoid the cost of the `<suspend-coroutine>` operation in cases where it is known that the operation will complete synchronously without needing to suspend.

At the `<return-to-caller-or-resumer>` point execution is transferred back to the caller or resumer, popping the local stack frame but keeping the coroutine frame alive.

When (or if) the suspended coroutine is eventually resumed then the execution resumes at the `<resume-point>`. ie. immediately before the `await_resume()` method is called to obtain the result of the operation.

The return-value of the `await_resume()` method call becomes the result of the `co_await` expression. The `await_resume()` method can also throw an exception in which case the exception propagates out of the `co_await` expression.

Note that if an exception propagates out of the `await_suspend()` call then the coroutine is automatically resumed and the exception propagates out of the `co_await` expression without calling `await_resume()`.

---
await_suspend() 返回 void 的版本在调用 await_suspend() 返回时无条件地将执行控制权转移回调用者或恢复者，而返回 bool 的版本允许 awaiter 对象有条件地立即恢复协程，而不返回到调用者或恢复者。

返回 bool 的 await_suspend() 版本在 awaiter 可能启动一个有时可以同步完成的异步操作的情况下非常有用。在这种情况下，如果操作同步完成，await_suspend() 方法可以返回 false，以指示协程应立即恢复并继续执行。

在 <suspend-coroutine> 点，编译器生成一些代码来保存协程的当前状态并准备其恢复。这包括存储 <resume-point> 的位置以及将当前寄存器中的任何值溢出到协程帧内存中。

在 <suspend-coroutine> 操作完成后，当前协程被视为已挂起。你可以在 await_suspend() 调用内部观察到挂起的协程的第一个点。一旦协程被挂起，它就可以被恢复或销毁。

await_suspend() 方法的责任是在操作完成后，在未来的某个时刻安排协程恢复（或销毁）。请注意，从 await_suspend() 返回 false 等同于安排协程在当前线程上立即恢复。

await_ready() 方法的目的是在已知操作将同步完成且不需要挂起的情况下避免 <suspend-coroutine> 操作的成本。

在 <return-to-caller-or-resumer> 点，执行控制权被转移回调用者或恢复者，弹出本地栈帧但保持协程帧存活。

当（或如果）挂起的协程最终被恢复时，执行会在 <resume-point> 继续。即在调用 await_resume() 方法以获取操作结果之前立即恢复。

await_resume() 方法调用的返回值成为 co_await 表达式的结果。await_resume() 方法也可以抛出异常，在这种情况下，异常会从 co_await 表达式中传播出来。

请注意，如果异常从 await_suspend() 调用中传播出来，则协程会自动恢复，并且异常会从 co_await 表达式中传播出来，而不会调用 await_resume()。

---
## Coroutine Handles

You may have noticed the use of the `coroutine_handle<P>` type that is passed to the `await_suspend()` call of a `co_await` expression.

This type represents a non-owning handle to the coroutine frame and can be used to resume execution of the coroutine or to destroy the coroutine frame. It can also be used to get access to the coroutine’s promise object.

The `coroutine_handle` type has the following (abbreviated) interface:

---
你可能已经注意到在 co_await 表达式的 await_suspend() 调用中使用了 coroutine_handle<P> 类型。

该类型表示对协程帧的非拥有句柄，可用于恢复协程的执行或销毁协程帧。它还可以用于访问协程的承诺对象。

coroutine_handle 类型具有以下（简化的）接口：

---

```c++
namespace std::experimental
{
  template<typename Promise>
  struct coroutine_handle;

  template<>
  struct coroutine_handle<void>
  {
    bool done() const;

    void resume();
    void destroy();

    void* address() const;
    static coroutine_handle from_address(void* address);
  };

  template<typename Promise>
  struct coroutine_handle : coroutine_handle<void>
  {
    Promise& promise() const;
    static coroutine_handle from_promise(Promise& promise);

    static coroutine_handle from_address(void* address);
  };
}
```

When implementing **Awaitable** types, the key method you’ll be using on `coroutine_handle` will be `.resume()`, which should be called when the operation has completed and you want to resume execution of the awaiting coroutine. Calling `.resume()` on a `coroutine_handle` reactivates a suspended coroutine at the `<resume-point>`. The call to `.resume()` will return when the coroutine next hits a `<return-to-caller-or-resumer>` point.

The `.destroy()` method destroys the coroutine frame, calling the destructors of any in-scope variables and freeing memory used by the coroutine frame. You should generally not need to (and indeed should really avoid) calling `.destroy()` unless you are a library writer implementing the coroutine promise type. Normally, coroutine frames will be owned by some kind of RAII type returned from the call to the coroutine. So calling `.destroy()` without cooperation with the RAII object could lead to a double-destruction bug.

The `.promise()` method returns a reference to the coroutine’s promise object. However, like `.destroy()`, it is generally only useful if you are authoring coroutine promise types. You should consider the coroutine’s promise object as an internal implementation detail of the coroutine. For most **Normally Awaitable** types you should use `coroutine_handle<void>` as the parameter type to the `await_suspend()` method instead of `coroutine_handle<Promise>`.

The `coroutine_handle<P>::from_promise(P& promise)` function allows reconstructing the coroutine handle from a reference to the coroutine’s promise object. Note that you must ensure that the type, `P`, exactly matches the concrete promise type used for the coroutine frame; attempting to construct a `coroutine_handle<Base>` when the concrete promise type is `Derived` can lead to undefined behaviour.

The `.address()` / `from_address()` functions allow converting a coroutine handle to/from a `void*` pointer. This is primarily intended to allow passing as a ‘context’ parameter into existing C-style APIs, so you might find it useful in implementing **Awaitable** types in some circumstances. However, in most cases I’ve found it necessary to pass additional information through to callbacks in this ‘context’ parameter so I generally end up storing the `coroutine_handle` in a struct and passing a pointer to the struct in the ‘context’ parameter rather than using the `.address()` return-value.

---
在实现 Awaitable 类型时，你在 coroutine_handle 上使用的关键方法将是 .resume()，当操作完成并且你想恢复等待协程的执行时应调用此方法。对 coroutine_handle 调用 .resume() 会在 <resume-point> 重新激活一个挂起的协程。调用 .resume() 将在协程下次到达 <return-to-caller-or-resumer> 点时返回。

.destroy() 方法销毁协程帧，调用范围内任何变量的析构函数并释放协程帧使用的内存。通常你不需要（实际上应该避免）调用 .destroy()，除非你是库编写者并在实现协程承诺类型。正常情况下，协程帧将由从协程调用返回的某种 RAII 类型拥有。因此，在没有与 RAII 对象协作的情况下调用 .destroy() 可能会导致双重销毁错误。

.promise() 方法返回协程承诺对象的引用。然而，像 .destroy() 一样，它通常只有在你正在编写协程承诺类型时才有用。你应该将协程的承诺对象视为协程的内部实现细节。对于大多数 Normally Awaitable 类型，你应该在 await_suspend() 方法中使用 coroutine_handle<void> 作为参数类型，而不是 coroutine_handle<Promise>。

coroutine_handle<P>::from_promise(P& promise) 函数允许从协程承诺对象的引用来重建协程句柄。请注意，你必须确保类型 P 完全匹配用于协程帧的具体承诺类型；尝试在具体承诺类型为 Derived 时构造 coroutine_handle<Base> 可能会导致未定义行为。

.address() / from_address() 函数允许将协程句柄转换为 void* 指针或从 void* 指针转换回来。这主要是为了允许将其作为“上下文”参数传递到现有的 C 风格 API 中，因此在某些情况下实现 Awaitable 类型时可能会发现它有用。然而，在大多数情况下，我发现有必要通过这个“上下文”参数传递额外的信息，所以我通常会将 coroutine_handle 存储在一个结构体中，并传递指向该结构体的指针作为“上下文”参数，而不是使用 .address() 返回值。

---
## Synchronisation-free async code

One of the powerful design-features of the `co_await` operator is the ability to execute code after the coroutine has been suspended but before execution is returned to the caller/resumer.

This allows an Awaiter object to initiate an async operation after the coroutine is already suspended, passing the `coroutine_handle` of the suspended coroutine to the operation which it can safely resume when the operation completes (potentially on another thread) without any additional synchronisation required.

For example, by starting an async-read operation inside `await_suspend()` when the coroutine is already suspended means that we can just resume the coroutine when the operation completes without needing any thread-synchronisation to coordinate the thread that started the operation and the thread that completed the operation.

---
co_await 操作符的一个强大设计特点是能够在协程挂起之后但在执行返回到调用者或恢复者之前执行代码。

这允许 Awaiter 对象在协程已经挂起后启动一个异步操作，将挂起协程的 coroutine_handle 传递给该操作，当操作完成时（可能在另一个线程上）它可以安全地恢复协程，而无需任何额外的同步。

例如，在 await_suspend() 中启动一个异步读取操作时协程已经挂起，这意味着我们可以在操作完成时直接恢复协程，而无需任何线程同步来协调启动操作的线程和完成操作的线程。

---

```
Time     Thread 1                           Thread 2
  |      --------                           --------
  |      ....                               Call OS - Wait for I/O event
  |      Call await_ready()                    |
  |      <supend-point>                        |
  |      Call await_suspend(handle)            |
  |        Store handle in operation           |
  V        Start AsyncFileRead ---+            V
                                  +----->   <AsyncFileRead Completion Event>
                                            Load coroutine_handle from operation
                                            Call handle.resume()
                                              <resume-point>
                                              Call to await_resume()
                                              execution continues....
           Call to AsyncFileRead returns
         Call to await_suspend() returns
         <return-to-caller/resumer>
```

One thing to be *very* careful of when taking advantage of this approach is that as soon as you have started the operation which publishes the coroutine handle to other threads then another thread may resume the coroutine on another thread before `await_suspend()` returns and may continue executing concurrently with the rest of the `await_suspend()` method.

The first thing the coroutine will do when it resumes is call `await_resume()` to get the result and then often it will immediately destruct the **Awaiter** object (ie. the `this` pointer of the `await_suspend()` call). The coroutine could then potentially run to completion, destructing the coroutine and promise object, all before `await_suspend()` returns.

So within the `await_suspend()` method, once it’s possible for the coroutine to be resumed concurrently on another thread, you need to make sure that you avoid accessing `this` or the coroutine’s `.promise()` object because both could already be destroyed. In general, the only things that are safe to access after the operation is started and the coroutine is scheduled for resumption are local variables within `await_suspend()`.

---
在利用这种方法时，有一件事需要非常小心：一旦你启动了将协程句柄发布到其他线程的操作，另一个线程可能会在 await_suspend() 返回之前在另一个线程上恢复协程，并且可能与 await_suspend() 方法的其余部分并发执行。

当协程恢复时，它首先会调用 await_resume() 来获取结果，然后通常会立即销毁 Awaiter 对象（即 await_suspend() 调用中的 this 指针）。接着，协程可能会运行到完成，销毁协程和承诺对象，所有这一切都可能在 await_suspend() 返回之前发生。

因此，在 await_suspend() 方法中，一旦有可能在另一个线程上并发恢复协程，你需要确保避免访问 this 或协程的 .promise() 对象，因为这两者可能已经被销毁。一般来说，在操作开始并且协程被安排恢复之后，唯一安全访问的是 await_suspend() 内的局部变量。

---

### Comparison to Stackful Coroutines

I want to take a quick detour to compare this ability of the Coroutines TS stackless coroutines to execute logic after the coroutine is suspended with some existing common stackful coroutine facilities such as Win32 fibers or boost::context.

With many of the stackful coroutine frameworks, the suspend operation of a coroutine is combined with the resumption of another coroutine into a ‘context-switch’ operation. With this ‘context-switch’ operation there is typically no opportunity to execute logic after suspending the current coroutine but before transferring execution to another coroutine.

This means that if we want to implement a similar async-file-read operation on top of stackful coroutines then we have to start the operation *before* suspending the coroutine. It is therefore possible that the operation could complete on another thread before the coroutine is suspended and is eligible for resumption. This potential race between the operation completing on another thread and the coroutine suspending requires some kind of thread synchronisation to arbitrate and decide on the winner.

There are probably ways around this by using a trampoline context that can start the operation on behalf of the initiating context after the initiating context has been suspended. However this would require extra infrastructure and an extra context-switch to make it work and it’s possible that the overhead this introduces would be greater than the cost of the synchronisation it’s trying to avoid.

---
我想快速偏离主题，比较一下 Coroutines TS 无栈协程在协程挂起后执行逻辑的能力与现有的常见有栈协程设施（如 Win32 纤程或 boost::context）之间的差异。

在许多有栈协程框架中，协程的挂起操作与另一个协程的恢复结合成一个“上下文切换”操作。在这种“上下文切换”操作中，通常没有机会在挂起当前协程之后但在将执行控制权转移到另一个协程之前执行任何逻辑。

这意味着如果我们想在有栈协程之上实现类似的异步文件读取操作，我们必须在挂起协程之前启动操作。因此，有可能在协程挂起并符合恢复条件之前，操作已经在另一个线程上完成。这种在另一个线程上完成操作和协程挂起之间的潜在竞争需要某种线程同步来仲裁并决定胜者。

可能通过使用一个蹦床上下文来绕过这个问题，该蹦床上下文可以在发起上下文被挂起后代表发起上下文启动操作。然而，这需要额外的基础设施和一次额外的上下文切换才能使其工作，并且引入的开销可能比它试图避免的同步成本更高。

---
## Avoiding memory allocations

Async operations often need to store some per-operation state that keeps track of the progress of the operation. This state typically needs to last for the duration of the operation and should only be freed once the operation has completed.

For example, calling async Win32 I/O functions requires you to allocate and pass a pointer to an `OVERLAPPED` structure. The caller is responsible for ensuring this pointer remains valid until the operation completes.

With traditional callback-based APIs this state would typically need to be allocated on the heap to ensure it has the appropriate lifetime. If you were performing many operations, you may need to allocate and free this state for each operation. If performance is an issue then a custom allocator may be used that allocates these state objects from a pool.

However, when we are using coroutines we can avoid the need to heap-allocate storage for the operation state by taking advantage of the fact that local variables within the coroutine frame will be kept alive while the coroutine is suspended.

By placing the per-operation state in the **Awaiter** object we can effectively “borrow” memory from the coroutine frame for storing the per-operation state for the duration of the `co_await` expression. Once the operation completes, the coroutine is resumed and the **Awaiter** object is destroyed, freeing that memory in the coroutine frame for use by other local variables.

Ultimately, the coroutine frame may still be allocated on the heap. However, once allocated, a coroutine frame can be used to execute many asynchronous operations with only that single heap allocation.

If you think about it, the coroutine frame acts as a kind of really high-performance arena memory allocator. The compiler figures out at compile time the total arena size it needs for all local variables and is then able to allocate this memory out to local variables as required with zero overhead! Try beating that with a custom allocator ;)

---
异步操作通常需要存储一些操作状态，以跟踪操作的进度。这种状态一般需要在整个操作期间保持，并且只有在操作完成后才能释放。

例如，调用异步Win32 I/O函数时，你需要分配并传递一个指向OVERLAPPED结构的指针。调用者需确保此指针在操作完成之前一直有效。

使用传统的基于回调的API时，通常需要在堆上分配这些状态以确保它们具有适当的生命周期。如果你正在进行许多操作，则可能需要为每个操作分配和释放这些状态。如果性能是一个问题，则可以使用自定义分配器从池中分配这些状态对象。

然而，当我们使用协程时，我们可以利用这样一个事实：协程框架内的局部变量在协程挂起期间会保持活动状态，从而避免了为操作状态在堆上分配存储空间的需求。

通过将每操作的状态放置在Awaiter对象中，我们实际上可以从协程框架“借用”内存来存储co_await表达式持续期间的操作状态。一旦操作完成，协程继续执行并且Awaiter对象被销毁，该协程框架中的内存可被其他局部变量使用。

最终，协程框架本身可能仍然需要在堆上分配。但是，一旦分配了协程框架，就可以仅用那一次堆分配来执行多个异步操作。

实际上，协程框架充当了一种高性能的区域内存分配器。编译器在编译时计算出所有局部变量所需的总区域大小，然后能够根据需要以零开销将这块内存分配给局部变量！尝试用自定义分配器超越这个效果吧 ;)

---

## An example: Implementing a simple thread-synchronisation primitive

Now that we’ve covered a lot of the mechanics of the `co_await` operator, I want to show how to put some of this knowledge into practice by implementing a basic awaitable synchronisation primitive: An asynchronous manual-reset event.

The basic requirements of this event is that it needs to be **Awaitable** by multiple concurrently executing coroutines and when awaited needs to suspend the awaiting coroutine until some thread calls the `.set()` method, at which point any awaiting coroutines are resumed. If some thread has already called `.set()` then the coroutine should continue without suspending.

Ideally we’d also like to make it `noexcept`, require no heap allocations and have a lock-free implementation.

**Edit 2017/11/23: Added example usage for `async_manual_reset_event`**

Example usage should look something like this:

---
现在我们已经介绍了co_await操作符的许多机制，我想通过实现一个基本的可等待同步基元——异步手动重置事件（asynchronous manual-reset event），来展示如何将这些知识付诸实践。

这个事件的基本要求是它需要能够被多个并发执行的协程等待，并且当被等待时，需要挂起等待的协程直到某个线程调用了.set()方法，在这一点上任何等待的协程都会恢复。如果在某个线程已经调用了.set()之后，则协程应该继续执行而无需挂起。

理想情况下，我们还希望它是noexcept的，不需要堆分配，并且有一个无锁的实现。

编辑 2017/11/23：添加了async_manual_reset_event的使用示例

使用示例应该看起来像这样：

---
```c++
T value;
async_manual_reset_event event;

// A single call to produce a value
void producer()
{
  value = some_long_running_computation();

  // Publish the value by setting the event.
  event.set();
}

// Supports multiple concurrent consumers
task<> consumer()
{
  // Wait until the event is signalled by call to event.set()
  // in the producer() function.
  co_await event;

  // Now it's safe to consume 'value'
  // This is guaranteed to 'happen after' assignment to 'value'
  std::cout << value << std::endl;
}
```

Let’s first think about the possible states this event can be in: ‘not set’ and ‘set’.

When it’s in the ‘not set’ state there is a (possibly empty) list of waiting coroutines that are waiting for it to become ‘set’.

When it’s in the ‘set’ state there won’t be any waiting coroutines as coroutines that `co_await` the event in this state can continue without suspending.

This state can actually be represented in a single `std::atomic<void*>`.

- Reserve a special pointer value for the ‘set’ state. In this case we’ll use the `this` pointer of the event since we know that can’t be the same address as any of the list items.
- Otherwise the event is in the ‘not set’ state and the value is a pointer to the head of a singly linked-list of awaiting coroutine structures.

We can avoid extra calls to allocate nodes for the linked-list on the heap by storing the nodes within an ‘awaiter’ object that is placed within the coroutine frame.

So let’s start with a class interface that looks something like this:

---
让我们首先考虑这个事件可能处于的状态：‘未设置’和‘已设置’。

当它处于‘未设置’状态时，存在一个（可能是空的）等待协程列表，这些协程在等待它变为‘已设置’。

当它处于‘已设置’状态时，不会有等待的协程，因为在这种状态下`co_await`该事件的协程可以继续执行而无需挂起。

这个状态实际上可以用一个单独的`std::atomic<void*>`来表示。

- 为‘已设置’状态保留一个特殊的指针值。在这种情况下，我们将使用事件的`this`指针，因为我们知道它不可能与任何列表项的地址相同。
- 否则，事件处于‘未设置’状态，且该值是指向单向链表头部的指针，该链表包含等待的协程结构。

我们可以通过将节点存储在一个放置在协程框架内的‘awaiter’对象中，来避免在堆上额外调用分配链表节点。

所以，让我们从一个类似这样的类接口开始：

---
```c++
class async_manual_reset_event
{
public:

  async_manual_reset_event(bool initiallySet = false) noexcept;

  // No copying/moving
  async_manual_reset_event(const async_manual_reset_event&) = delete;
  async_manual_reset_event(async_manual_reset_event&&) = delete;
  async_manual_reset_event& operator=(const async_manual_reset_event&) = delete;
  async_manual_reset_event& operator=(async_manual_reset_event&&) = delete;

  bool is_set() const noexcept;

  struct awaiter;
  awaiter operator co_await() const noexcept;

  void set() noexcept;
  void reset() noexcept;

private:

  friend struct awaiter;

  // - 'this' => set state
  // - otherwise => not set, head of linked list of awaiter*.
  mutable std::atomic<void*> m_state;

};
```

Here we have a fairly straight-forward and simple interface. The main thing to note at this point is that it has an `operator co_await()` method that returns an, as yet, undefined type, `awaiter`.

Let’s define the `awaiter` type now.

---
我们这里有一个相当直接且简单的接口。此时需要注意的主要一点是，它有一个operator co_await()方法，该方法返回一个尚未定义的类型awaiter。

现在让我们定义awaiter类型。

---

### Defining the Awaiter

Firstly, it needs to know which `async_manual_reset_event` object it is going to be awaiting, so it will need a reference to the event and a constructor to initialise it.

It also needs to act as a node in a linked-list of `awaiter` values so it will need to hold a pointer to the next `awaiter` object in the list.

It also needs to store the `coroutine_handle` of the awaiting coroutine that is executing the `co_await` expression so that the event can resume the coroutine when it becomes ‘set’. We don’t care what the promise type of the coroutine is so we’ll just use a `coroutine_handle<>` (which is short-hand for `coroutine_handle<void>`).

Finally, it needs to implement the **Awaiter** interface, so it needs the three special methods: `await_ready`, `await_suspend` and `await_resume`. We don’t need to return a value from the `co_await` expression so `await_resume` can return `void`.

Once we put all of that together, the basic class interface for `awaiter` looks like this:

---
首先，它需要知道它将要等待哪个`async_manual_reset_event`对象，因此它需要一个对该事件的引用和一个初始化它的构造函数。

它还需要充当`awaiter`值链表中的一个节点，因此它需要持有一个指向链表中下一个`awaiter`对象的指针。

它还需要存储执行`co_await`表达式的等待协程的`coroutine_handle`，以便事件在变为‘已设置’时可以恢复该协程。我们不关心协程的承诺类型是什么，所以我们将使用`coroutine_handle<>`（这是`coroutine_handle<void>`的简写）。

最后，它需要实现**Awaiter**接口，因此需要三个特殊方法：`await_ready`、`await_suspend`和`await_resume`。我们不需要从`co_await`表达式返回一个值，所以`await_resume`可以返回`void`。

一旦我们将所有这些组合在一起，`awaiter`的基本类接口看起来像这样：

---

```c++
struct async_manual_reset_event::awaiter
{
  awaiter(const async_manual_reset_event& event) noexcept
  : m_event(event)
  {}

  bool await_ready() const noexcept;
  bool await_suspend(std::experimental::coroutine_handle<> awaitingCoroutine) noexcept;
  void await_resume() noexcept {}

private:

  const async_manual_reset_event& m_event;
  std::experimental::coroutine_handle<> m_awaitingCoroutine;
  awaiter* m_next;
};
```

Now, when we `co_await` an event, we don’t want the awaiting coroutine to suspend if the event is already set. So we can define `await_ready()` to return `true` if the event is already set.

---
现在，当我们对一个事件进行`co_await`时，如果事件已经设置，我们不希望等待的协程挂起。因此，我们可以定义`await_ready()`方法，如果事件已经设置，则返回`true`。

---

```c++
bool async_manual_reset_event::awaiter::await_ready() const noexcept
{
  return m_event.is_set();
}
```

Next, let’s look at the `await_suspend()` method. This is usually where most of the magic happens in an awaitable type.

First it will need to stash the coroutine handle of the awaiting coroutine into the `m_awaitingCoroutine` member so that the event can later call `.resume()` on it.

Then once we’ve done that we need to try and atomically enqueue the awaiter onto the linked list of waiters. If we successfully enqueue it then we return `true` to indicate that we don’t want to resume the coroutine immediately, otherwise if we find that the event has concurrently been changed to the ‘set’ state then we return `false` to indicate that the coroutine should be resumed immediately.

---
接下来，让我们看看`await_suspend()`方法。这通常是可等待类型中大部分“魔法”发生的地方。

首先，它需要将等待协程的协程句柄存储到`m_awaitingCoroutine`成员中，以便事件稍后可以对其调用`.resume()`。

然后，一旦我们完成了这一点，我们需要尝试以原子方式将awaiter加入等待者的链表中。如果我们成功地将其加入队列，则返回`true`，表示我们不希望立即恢复协程；否则，如果发现事件已被并发地更改为‘已设置’状态，则返回`false`，表示应立即恢复协程。

---

```c++
bool async_manual_reset_event::awaiter::await_suspend(
  std::experimental::coroutine_handle<> awaitingCoroutine) noexcept
{
  // Special m_state value that indicates the event is in the 'set' state.
  const void* const setState = &m_event;

  // Remember the handle of the awaiting coroutine.
  m_awaitingCoroutine = awaitingCoroutine;

  // Try to atomically push this awaiter onto the front of the list.
  void* oldValue = m_event.m_state.load(std::memory_order_acquire);
  do
  {
    // Resume immediately if already in 'set' state.
    if (oldValue == setState) return false; 

    // Update linked list to point at current head.
    m_next = static_cast<awaiter*>(oldValue);

    // Finally, try to swap the old list head, inserting this awaiter
    // as the new list head.
  } while (!m_event.m_state.compare_exchange_weak(
             oldValue,
             this,
             std::memory_order_release,
             std::memory_order_acquire));

  // Successfully enqueued. Remain suspended.
  return true;
}
```

Note that we use ‘acquire’ memory order when loading the old state so that if we read the special ‘set’ value then we have visibility of writes that occurred prior to the call to ‘set()’.

We require ‘release’ sematics if the compare-exchange succeeds so that a subsequent call to ‘set()’ will see our writes to m_awaitingCoroutine and prior writes to the coroutine state.

---
注意，当我们加载旧状态时，我们使用‘acquire’内存顺序，这样如果读取到特殊的‘已设置’值，我们就能看到在调用‘set()’之前发生的写操作。

如果比较-交换成功，我们需要‘release’语义，以便后续对‘set()’的调用能够看到我们对`m_awaitingCoroutine`的写操作以及对协程状态的先前写操作。

---

### Filling out the rest of the event class

Now that we have defined the `awaiter` type, let’s go back and look at the implementation of the `async_manual_reset_event` methods.

First, the constructor. It needs to initialise to either the ‘not set’ state with the empty list of waiters (ie. `nullptr`) or initialise to the ‘set’ state (ie. `this`).

---
现在我们已经定义了awaiter类型，让我们回到async_manual_reset_event方法的实现上来。

首先，构造函数。它需要初始化为‘未设置’状态，此时等待者列表为空（即nullptr），或者初始化为‘已设置’状态（即this）。

---

```c++
async_manual_reset_event::async_manual_reset_event(
  bool initiallySet) noexcept
: m_state(initiallySet ? this : nullptr)
{}
```

Next, the `is_set()` method is pretty straight-forward - it’s ‘set’ if it has the special value `this`:

---
接下来，is_set()方法非常直接——如果它具有特殊值this，则表示事件已设置：

---

```c++
bool async_manual_reset_event::is_set() const noexcept
{
  return m_state.load(std::memory_order_acquire) == this;
}
```

Next, the `reset()` method. If it’s in the ‘set’ state we want to transition back to the empty-list ‘not set’ state, otherwise leave it as it is.

---
接下来，`reset()`方法。如果事件处于‘已设置’状态，我们希望将其转换回空列表的‘未设置’状态；否则，保持其当前状态不变。

---

```c++
void async_manual_reset_event::reset() noexcept
{
  void* oldValue = this;
  m_state.compare_exchange_strong(oldValue, nullptr, std::memory_order_acquire);
}
```

With the `set()` method, we want to transition to the ‘set’ state by exchanging the current state with the special ‘set’ value, `this`, and then examine what the old value was. If there were any waiting coroutines then we want to resume each of them sequentially in turn before returning.

---
对于`set()`方法，我们希望通过将当前状态与特殊的‘已设置’值（即`this`）进行交换来转换到‘已设置’状态，然后检查旧值是什么。如果有任何等待的协程，我们希望在返回之前依次恢复每一个等待的协程。

---

```c++
void async_manual_reset_event::set() noexcept
{
  // Needs to be 'release' so that subsequent 'co_await' has
  // visibility of our prior writes.
  // Needs to be 'acquire' so that we have visibility of prior
  // writes by awaiting coroutines.
  void* oldValue = m_state.exchange(this, std::memory_order_acq_rel);
  if (oldValue != this)
  {
    // Wasn't already in 'set' state.
    // Treat old value as head of a linked-list of waiters
    // which we have now acquired and need to resume.
    auto* waiters = static_cast<awaiter*>(oldValue);
    while (waiters != nullptr)
    {
      // Read m_next before resuming the coroutine as resuming
      // the coroutine will likely destroy the awaiter object.
      auto* next = waiters->m_next;
      waiters->m_awaitingCoroutine.resume();
      waiters = next;
    }
  }
}
```

Finally, we need to implement the `operator co_await()` method. This just needs to construct an `awaiter` object.

---
最后，我们需要实现`operator co_await()`方法。这个方法只需要构造一个`awaiter`对象。

---
```c++
async_manual_reset_event::awaiter
async_manual_reset_event::operator co_await() const noexcept
{
  return awaiter{ *this };
}
```

And there we have it. An awaitable asynchronous manual-reset event that has a lock-free, memory-allocation-free, `noexcept` implementation.

If you want to have a play with the code or check out what it compiles down to under MSVC and Clang have a look at the [source on godbolt](https://godbolt.org/g/Ad47tH).

You can also find an implementation of this class available in the [cppcoro](https://github.com/lewissbaker/cppcoro) library, along with a number of other useful awaitable types such as `async_mutex` and `async_auto_reset_event`.

---
就是这样。我们实现了一个可等待的异步手动重置事件，它具有无锁、无需内存分配且为`noexcept`的实现。

如果你想尝试这段代码或查看它在MSVC和Clang下的编译结果，可以在这个[godbolt的源码链接](https://godbolt.org/g/Ad47tH)上查看。

你还可以在[cppcoro库](https://github.com/lewissbaker/cppcoro)中找到这个类的实现，以及许多其他有用的可等待类型，如`async_mutex`和`async_auto_reset_event`。

---
## Closing Off

  This post has looked at how the `operator co_await` is implemented and defined in terms of the **Awaitable** and **Awaiter** concepts.

  It has also walked through how to implement an awaitable async thread-synchronisation primitive that takes advantage of the fact that awaiter objects are allocated on the coroutine frame to avoid additional heap allocations.

  I hope this post has helped to demystify the new `co_await` operator for you.

  In the next post I’ll explore the **Promise** concept and how a coroutine-type author can customise the behaviour of their coroutine.

  ---
  这篇帖子探讨了`operator co_await`是如何根据**Awaitable**和**Awaiter**概念来实现和定义的。

它还逐步介绍了如何实现一个可等待的异步线程同步基元，该基元利用了awaiter对象被分配在协程框架上的事实，以避免额外的堆分配。

我希望这篇帖子能帮助你揭开新`co_await`操作符的神秘面纱。

在下一篇帖子中，我将探索**Promise**概念，以及协程类型作者如何自定义其协程的行为。

---

## Thanks

I want to call out special thanks to Gor Nishanov for patiently and enthusiastically answering my many questions on coroutines over the last couple of years.

And also to Eric Niebler for reviewing and providing feedback on an early draft of this post.

---
特别感谢Gor Nishanov，在过去的几年里，他耐心且热情地回答了我关于协程的许多问题。

同时也要感谢Eric Niebler，他对这篇帖子的早期草稿进行了审阅并提供了宝贵的反馈。

---