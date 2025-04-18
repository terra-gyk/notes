# Coroutine Theory

Sep 25, 2017

This is the first of a series of posts on the [C++ Coroutines TS](http://www.open-std.org/jtc1/sc22/wg21/docs/papers/2017/n4680.pdf), a new language feature that is currently on track for inclusion into the C++20 language standard.

In this series I will cover how the underlying mechanics of C++ Coroutines work as well as show how they can be used to build useful higher-level abstractions such as those provided by the [cppcoro](https://github.com/lewissbaker/cppcoro) library.

In this post I will describe the differences between functions and coroutines and provide a bit of theory about the operations they support. The aim of this post is introduce some foundational concepts that will help frame the way you think about C++ Coroutines.

---

这是关于[C++协程技术规范](http://www.open-std.org/jtc1/sc22/wg21/docs/papers/2017/n4680.pdf)系列文章的第一篇，这是一个新的语言特性，目前正按计划纳入C++20语言标准。

在这个系列中，我将介绍C++协程的基本机制，以及如何使用它们构建有用的高层次抽象，例如由[cppcoro](https://github.com/lewissbaker/cppcoro)库提供的那些抽象。

在这篇文章中，我将描述函数与协程之间的差异，并提供一些关于它们支持的操作的理论知识。本文的目标是介绍一些基础概念，帮助你更好地理解C++协程的工作原理。

---

## Coroutines are Functions are Coroutines

A coroutine is a generalisation of a function that allows the function to be suspended and then later resumed.

I will explain what this means in a bit more detail, but before I do I want to first review how a “normal” C++ function works.

---

协程是函数的一种泛化，它允许函数被挂起然后在之后恢复执行。

我将会更详细地解释这意味着什么，但在此之前，我想先回顾一下‘普通’C++函数的工作原理。

---

## “Normal” Functions

A normal function can be thought of as having two operations: **Call** and **Return** (Note that I’m lumping “throwing an exception” here broadly under the **Return** operation).

The **Call** operation creates an activation frame, suspends execution of the calling function and transfers execution to the start of the function being called.

The **Return** operation passes the return-value to the caller, destroys the activation frame and then resumes execution of the caller just after the point at which it called the function.

Let’s analyse these semantics a little more…

---

一个普通的函数可以被认为具有两种操作：调用和返回（注意，这里我将‘抛出异常’大致归类在返回操作之下）。

调用操作创建一个激活帧，暂停调用函数的执行，并将执行转移到被调用函数的起始位置。

返回操作将返回值传递给调用者，销毁激活帧，然后在调用函数调用该函数的位置之后恢复其执行。

让我们更详细地分析一下这些语义……

---

### Activation Frames

So what is this ‘activation frame’ thing?

You can think of the activation frame as the block of memory that holds the current state of a particular invocation of a function. This state includes the values of any parameters that were passed to it and the values of any local variables.

For “normal” functions, the activation frame also includes the return-address - the address of the instruction to transfer execution to upon returning from the function - and the address of the activation frame for the invocation of the calling function. You can think of these pieces of information together as describing the ‘continuation’ of the function-call. ie. they describe which invocation of which function should continue executing at which point when this function completes.

With “normal” functions, all activation frames have strictly nested lifetimes. This strict nesting allows use of a highly efficient memory allocation data-structure for allocating and freeing the activation frames for each of the function calls. This data-structure is commonly referred to as “the stack”.

When an activation frame is allocated on this stack data structure it is often called a “stack frame”.

This stack data-structure is so common that most (all?) CPU architectures have a dedicated register for holding a pointer to the top of the stack (eg. in X64 it is the `rsp` register).

To allocate space for a new activation frame, you just increment this register by the frame-size. To free space for an activation frame, you just decrement this register by the frame-size.

---

那么这个‘激活帧’究竟是什么呢？

你可以将激活帧视为保存某个函数调用当前状态的内存块。这个状态包括传递给它的任何参数的值以及任何局部变量的值。

对于‘普通’函数，激活帧还包括返回地址——即从函数返回时要转移执行的指令地址——以及调用函数调用的激活帧地址。你可以把这些信息看作描述了函数调用的‘延续’。也就是说，它们描述了当这个函数完成后，哪个函数的哪次调用应该在哪个点继续执行。

对于‘普通’函数，所有激活帧都有严格的嵌套生命周期。这种严格的嵌套允许使用一种高效的内存分配数据结构来分配和释放每个函数调用的激活帧。这种数据结构通常被称为‘栈’。

当一个激活帧在这个栈数据结构上分配时，它通常被称为‘栈帧’。

这种栈数据结构非常普遍，以至于大多数（所有？）CPU架构都有一个专门的寄存器用于保存指向栈顶的指针（例如，在X64架构中是rsp寄存器）。

要为新的激活帧分配空间，只需将这个寄存器增加帧大小即可。要释放激活帧的空间，只需将这个寄存器减去帧大小即可。

---

### The ‘Call’ Operation

When a function calls another function, the caller must first prepare itself for suspension.

This ‘suspend’ step typically involves saving to memory any values that are currently held in CPU registers so that those values can later be restored if required when the function resumes execution. Depending on the calling convention of the function, the caller and callee may coordinate on who saves these register values, but you can still think of them as being performed as part of the **Call** operation.

The caller also stores the values of any parameters passed to the called function into the new activation frame where they can be accessed by the function.

Finally, the caller writes the address of the resumption-point of the caller to the new activation frame and transfers execution to the start of the called function.

In the X86/X64 architecture this final operation has its own instruction, the `call` instruction, that writes the address of the next instruction onto the stack, increments the stack register by the size of the address and then jumps to the address specified in the instruction’s operand.

---

当一个函数调用另一个函数时，调用者必须首先为挂起做准备。

这个‘挂起’步骤通常涉及将当前保存在CPU寄存器中的任何值保存到内存中，以便在函数恢复执行时可以重新加载这些值。根据函数的调用约定，调用者和被调用者可能会协调谁来保存这些寄存器值，但你可以将其视为调用操作的一部分。

调用者还将传递给被调用函数的任何参数的值存储到新的激活帧中，这样函数就可以访问这些值。

最后，调用者将调用者的恢复点地址写入新的激活帧，并将执行转移到被调用函数的起始位置。

在X86/X64架构中，这一最终操作有其自己的指令，即call指令，它将下一条指令的地址写入栈中，通过增加栈寄存器的值（增加量为地址大小）然后跳转到指令操作数指定的地址。

---

### The ‘Return’ Operation

When a function returns via a `return`-statement, the function first stores the return value (if any) where the caller can access it. This could either be in the caller’s activation frame or the function’s activation frame (the distinction can get a bit blurry for parameters and return values that cross the boundary between two activation frames).

Then the function destroys the activation frame by:

- Destroying any local variables in-scope at the return-point.
- Destroying any parameter objects
- Freeing memory used by the activation-frame

And finally, it resumes execution of the caller by:

- Restoring the activation frame of the caller by setting the stack register to point to the activation frame of the caller and restoring any registers that might have been clobbered by the function.
- Jumping to the resume-point of the caller that was stored during the ‘Call’ operation.

Note that as with the ‘Call’ operation, some calling conventions may split the responsibilities of the ‘Return’ operation across both the caller and callee function’s instructions.

---

当一个函数通过`return`语句返回时，函数首先将返回值（如果有）存储在调用者可以访问的地方。这可能是在调用者的激活帧中，也可能是在函数的激活帧中（对于跨越两个激活帧边界的参数和返回值，这种区别可能会变得模糊）。

然后，函数通过以下步骤销毁激活帧：
- 销毁在返回点范围内任何局部变量。
- 销毁任何参数对象。
- 释放激活帧使用的内存。

最后，它通过以下步骤恢复调用者的执行：
- 通过将栈寄存器设置为指向调用者的激活帧，并恢复任何可能被该函数覆盖的寄存器，来恢复调用者的激活帧。
- 跳转到在‘调用’操作期间存储的调用者的恢复点。

需要注意的是，与‘调用’操作一样，某些调用约定可能会将‘返回’操作的责任分配给调用者和被调用者的指令。

---

## Coroutines

Coroutines generalise the operations of a function by separating out some of the steps performed in the **Call** and **Return** operations into three extra operations: **Suspend**, **Resume** and **Destroy**.

The **Suspend** operation suspends execution of the coroutine at the current point within the function and transfers execution back to the caller or resumer without destroying the activation frame. Any objects in-scope at the point of suspension remain alive after the coroutine execution is suspended.

Note that, like the **Return** operation of a function, a coroutine can only be suspended from within the coroutine itself at well-defined suspend-points.

The **Resume** operation resumes execution of a suspended coroutine at the point at which it was suspended. This reactivates the coroutine’s activation frame.

The **Destroy** operation destroys the activation frame without resuming execution of the coroutine. Any objects that were in-scope at the suspend point will be destroyed. Memory used to store the activation frame is freed.

---

协程通过将‘调用’和‘返回’操作中执行的一些步骤分离成三个额外的操作：**挂起**、**恢复**和**销毁**，从而泛化了函数的操作。

**挂起**操作在函数当前执行点挂起协程的执行，并将执行控制权转移回调用者或恢复者，而不会销毁激活帧。在挂起点范围内任何对象在协程执行挂起后仍然保持存活。

需要注意的是，与函数的‘返回’操作类似，协程只能在其内部定义良好的挂起点处进行挂起。

**恢复**操作在协程被挂起的点恢复其执行。这重新激活了协程的激活帧。

**销毁**操作在不恢复协程执行的情况下销毁激活帧。在挂起点范围内任何对象都将被销毁，用于存储激活帧的内存也将被释放。

---

### Coroutine activation frames

Since coroutines can be suspended without destroying the activation frame, we can no longer guarantee that activation frame lifetimes will be strictly nested. This means that activation frames cannot in general be allocated using a stack data-structure and so may need to be stored on the heap instead.

There are some provisions in the C++ Coroutines TS to allow the memory for the coroutine frame to be allocated from the activation frame of the caller if the compiler can prove that the lifetime of the coroutine is indeed strictly nested within the lifetime of the caller. This can avoid heap allocations in many cases provided you have a sufficiently smart compiler.

With coroutines there are some parts of the activation frame that need to be preserved across coroutine suspension and there are some parts that only need to be kept around while the coroutine is executing. For example, the lifetime of a variable with a scope that does not span any coroutine suspend-points can potentially be stored on the stack.

You can logically think of the activation frame of a coroutine as being comprised of two parts: the ‘coroutine frame’ and the ‘stack frame’.

The ‘coroutine frame’ holds part of the coroutine’s activation frame that persists while the coroutine is suspended and the ‘stack frame’ part only exists while the coroutine is executing and is freed when the coroutine suspends and transfers execution back to the caller/resumer.

---

由于协程可以在不销毁激活帧的情况下被挂起，我们不能再保证激活帧的生命周期会严格嵌套。这意味着激活帧通常不能使用栈数据结构进行分配，因此可能需要存储在堆上。

C++协程技术规范（Coroutines TS）中有一些规定，允许编译器从调用者的激活帧中分配协程帧所需的内存，前提是编译器能够证明协程的生命周期确实严格嵌套在调用者的生命周期内。如果编译器足够智能，这可以在许多情况下避免堆分配。

对于协程，激活帧中有些部分需要在协程挂起期间保留，而有些部分只需要在协程执行期间保留。例如，作用域不跨越任何协程挂起点的变量的生命周期可以潜在地存储在栈上。

你可以逻辑上将协程的激活帧视为由两部分组成：‘协程帧’和‘栈帧’。

‘协程帧’保存协程激活帧的一部分，这部分在协程挂起期间依然存在；而‘栈帧’部分仅在协程执行期间存在，并在协程挂起并将执行控制权转移回调用者或恢复者时释放。

---

### The ‘Suspend’ operation

The **Suspend** operation of a coroutine allows the coroutine to suspend execution in the middle of the function and transfer execution back to the caller or resumer of the coroutine.

There are certain points within the body of a coroutine that are designated as suspend-points. In the C++ Coroutines TS, these suspend-points are identified by usages of the `co_await` or `co_yield` keywords.

When a coroutine hits one of these suspend-points it first prepares the coroutine for resumption by:

- Ensuring any values held in registers are written to the coroutine frame
- Writing a value to the coroutine frame that indicates which suspend-point the coroutine is being suspended at. This allows a subsequent **Resume** operation to know where to resume execution of the coroutine or so a subsequent **Destroy** to know what values were in-scope and need to be destroyed.

Once the coroutine has been prepared for resumption, the coroutine is considered ‘suspended’.

The coroutine then has the opportunity to execute some additional logic before execution is transferred back to the caller/resumer. This additional logic is given access to a handle to the coroutine-frame that can be used to later resume or destroy it.

This ability to execute logic after the coroutine enters the ‘suspended’ state allows the coroutine to be scheduled for resumption without the need for synchronisation that would otherwise be required if the coroutine was scheduled for resumption prior to entering the ‘suspended’ state due to the potential for suspension and resumption of the coroutine to race. I’ll go into this in more detail in future posts.

The coroutine can then choose to either immediately resume/continue execution of the coroutine or can choose to transfer execution back to the caller/resumer.

If execution is transferred to the caller/resumer the stack-frame part of the coroutine’s activation frame is freed and popped off the stack.

----

协程的**挂起**操作允许协程在函数执行过程中挂起，并将执行控制权转移回调用者或恢复者。

在协程的函数体内，某些点被指定为挂起点。在C++协程技术规范（Coroutines TS）中，这些挂起点通过使用`co_await`或`co_yield`关键字来标识。

当协程到达其中一个挂起点时，它首先通过以下步骤为恢复做准备：

- 确保任何保存在寄存器中的值都被写入协程帧。
- 将一个值写入协程帧，指示协程是在哪个挂起点被挂起的。这使得后续的**恢复**操作能够知道从何处恢复协程的执行，或者使后续的**销毁**操作能够知道哪些值在作用域内并需要被销毁。

一旦协程准备好恢复，协程就被认为是‘挂起’状态。

然后，协程有机会在执行控制权转移回调用者或恢复者之前执行一些额外的逻辑。这个额外的逻辑可以获得对协程帧的句柄，用于稍后恢复或销毁协程。

这种在进入‘挂起’状态后执行逻辑的能力，允许协程在不需要同步的情况下被调度恢复。否则，如果协程在进入‘挂起’状态之前被调度恢复，可能会由于挂起和恢复的竞争条件而需要同步。我将在以后的文章中详细介绍这一点。

协程可以选择立即恢复/继续执行，也可以选择将执行控制权转移回调用者或恢复者。

如果执行控制权转移给调用者或恢复者，协程的激活帧的栈帧部分将被释放并从栈中弹出。

---

### The ‘Resume’ operation

The **Resume** operation can be performed on a coroutine that is currently in the ‘suspended’ state.

When a function wants to resume a coroutine it needs to effectively ‘call’ into the middle of a particular invocation of the function. The way the resumer identifies the particular invocation to resume is by calling the `void resume()` method on the coroutine-frame handle provided to the corresponding **Suspend** operation.

Just like a normal function call, this call to `resume()` will allocate a new stack-frame and store the return-address of the caller in the stack-frame before transferring execution to the function.

However, instead of transferring execution to the start of the function it will transfer execution to the point in the function at which it was last suspended. It does this by loading the resume-point from the coroutine-frame and jumping to that point.

When the coroutine next suspends or runs to completion this call to `resume()` will return and resume execution of the calling function.

---

**恢复**操作可以对当前处于‘挂起’状态的协程执行。

当一个函数想要恢复一个协程时，它需要有效地‘调用’到该函数特定调用的中间位置。恢复者通过在对应**挂起**操作提供的协程帧句柄上调用`void resume()`方法来识别要恢复的具体调用。

就像普通的函数调用一样，对`resume()`的调用会分配一个新的栈帧，并在将执行控制权转移到函数之前将调用者的返回地址存储在栈帧中。

然而，与转移到函数的起始位置不同，它会将执行控制权转移到函数上次挂起的位置。这是通过从协程帧加载恢复点并跳转到该点实现的。

当下一次协程挂起或运行完成时，这次对`resume()`的调用将返回并继续执行调用函数。

---

### The ‘Destroy’ operation

The **Destroy** operation destroys the coroutine frame without resuming execution of the coroutine.

This operation can only be performed on a suspended coroutine.

The **Destroy** operation acts much like the **Resume** operation in that it re-activates the coroutine’s activation frame, including allocating a new stack-frame and storing the return-address of the caller of the **Destroy** operation.

However, instead of transferring execution to the coroutine body at the last suspend-point it instead transfers execution to an alternative code-path that calls the destructors of all local variables in-scope at the suspend-point before then freeing the memory used by the coroutine frame.

Similar to the **Resume** operation, the **Destroy** operation identifies the particular activation-frame to destroy by calling the `void destroy()` method on the coroutine-frame handle provided during the corresponding **Suspend** operation.

---

**销毁**操作在不恢复协程执行的情况下销毁协程帧。

此操作只能对处于挂起状态的协程执行。

**销毁**操作的作用类似于**恢复**操作，因为它会重新激活协程的激活帧，包括分配一个新的栈帧，并存储调用**销毁**操作的调用者的返回地址。

然而，与将执行控制权转移到协程体的最后一个挂起点不同，它会将执行控制权转移到一个替代代码路径，该路径会在释放协程帧使用的内存之前调用挂起点范围内所有局部变量的析构函数。

类似于**恢复**操作，**销毁**操作通过在对应**挂起**操作提供的协程帧句柄上调用`void destroy()`方法来识别要销毁的具体激活帧。

---

### The ‘Call’ operation of a coroutine

The **Call** operation of a coroutine is much the same as the call operation of a normal function. In fact, from the perspective of the caller there is no difference.

However, rather than execution only returning to the caller when the function has run to completion, with a coroutine the call operation will instead resume execution of the caller when the coroutine reaches its first suspend-point.

When performing the **Call** operation on a coroutine, the caller allocates a new stack-frame, writes the parameters to the stack-frame, writes the return-address to the stack-frame and transfers execution to the coroutine. This is exactly the same as calling a normal function.

The first thing the coroutine does is then allocate a coroutine-frame on the heap and copy/move the parameters from the stack-frame into the coroutine-frame so that the lifetime of the parameters extends beyond the first suspend-point.

---

协程的**调用**操作与普通函数的调用操作非常相似。实际上，从调用者的角度来看，两者之间没有区别。

然而，与普通函数在执行完成时才返回调用者不同，对于协程，调用操作会在协程到达第一个挂起点时恢复调用者的执行。

在对协程执行**调用**操作时，调用者会分配一个新的栈帧，将参数写入栈帧，将返回地址写入栈帧，并将执行控制权转移到协程。这与调用普通函数完全相同。

协程首先要做的是在堆上分配一个协程帧，并将参数从栈帧复制/移动到协程帧中，从而使参数的生命周期扩展到第一个挂起点之后。

---

### The ‘Return’ operation of a coroutine

The **Return** operation of a coroutine is a little different from that of a normal function.

When a coroutine executes a `return`-statement (`co_return` according to the TS) operation it stores the return-value somewhere (exactly where this is stored can be customised by the coroutine) and then destructs any in-scope local variables (but not parameters).

The coroutine then has the opportunity to execute some additional logic before transferring execution back to the caller/resumer.

This additional logic might perform some operation to publish the return value, or it might resume another coroutine that was waiting for the result. It’s completely customisable.

The coroutine then performs either a **Suspend** operation (keeping the coroutine-frame alive) or a **Destroy** operation (destroying the coroutine-frame).

Execution is then transferred back to the caller/resumer as per the **Suspend**/**Destroy** operation semantics, popping the stack-frame component of the activation-frame off the stack.

It is important to note that the return-value passed to the **Return** operation is not the same as the return-value returned from a **Call** operation as the return operation may be executed long after the caller resumed from the initial **Call** operation.

---

协程的**返回**操作与普通函数的返回操作略有不同。

当协程执行一个`return`语句（根据技术规范应为`co_return`）时，它会将返回值存储在某个位置（具体存储位置可以由协程自定义），然后销毁范围内任何局部变量（但不包括参数）。

接着，协程有机会在将执行控制权转移回调用者或恢复者之前执行一些额外的逻辑。

这些额外的逻辑可能用于发布返回值，或者恢复另一个等待结果的协程。这完全是可定制的。

然后，协程执行一个**挂起**操作（保持协程帧存活）或一个**销毁**操作（销毁协程帧）。

根据**挂起**/**销毁**操作的语义，执行控制权随后被转移回调用者或恢复者，并从栈中弹出激活帧的栈帧部分。

需要注意的是，传递给**返回**操作的返回值与从**调用**操作返回的返回值并不相同，因为返回操作可能在调用者从初始**调用**操作恢复很久之后才执行。

---

## An illustration

To help put these concepts into pictures, I want to walk through a simple example of what happens when a coroutine is called, suspends and is later resumed.

So let’s say we have a function (or coroutine), `f()` that calls a coroutine, `x(int a)`.

Before the call we have a situation that looks a bit like this:

---

为了帮助将这些概念可视化，我们可以通过一个简单的例子来展示当协程被调用、挂起并稍后恢复时发生了什么。

假设我们有一个函数（或协程）`f()`，它调用另一个协程 `x(int a)`。

在调用之前，情况看起来像这样：

---

```
STACK                     REGISTERS               HEAP

                          +------+
+---------------+ <------ | rsp  |
|  f()          |         +------+
+---------------+
| ...           |
|               |
```

Then when `x(42)` is called, it first creates a stack frame for `x()`, as with normal functions.

---

当我们调用 `x(42)` 时，它首先为 `x()` 创建一个栈帧，就像普通函数一样。

---

```
STACK                     REGISTERS               HEAP
+----------------+ <-+
|  x()           |   |
| a  = 42        |   |
| ret= f()+0x123 |   |    +------+
+----------------+   +--- | rsp  |
|  f()           |        +------+
+----------------+
| ...            |
|                |
```

Then, once the coroutine `x()` has allocated memory for the coroutine frame on the heap and copied/moved parameter values into the coroutine frame we’ll end up with something that looks like the next diagram. Note that the compiler will typically hold the address of the coroutine frame in a separate register to the stack pointer (eg. MSVC stores this in the `rbp` register).

---

然后，一旦协程 `x()` 在堆上分配了协程帧的内存，并将参数值复制/移动到协程帧中，我们将得到如下图所示的状态。请注意，编译器通常会将协程帧的地址存储在一个与栈指针分开的寄存器中（例如，MSVC 使用 `rbp` 寄存器来存储协程帧的地址）。

---

```
STACK                     REGISTERS               HEAP
+----------------+ <-+
|  x()           |   |
| a  = 42        |   |                   +-->  +-----------+
| ret= f()+0x123 |   |    +------+       |     |  x()      |
+----------------+   +--- | rsp  |       |     | a =  42   |
|  f()           |        +------+       |     +-----------+
+----------------+        | rbp  | ------+
| ...            |        +------+
|                |
```

If the coroutine `x()` then calls another normal function `g()` it will look something like this.

---

如果协程 `x()` 接着调用另一个普通函数 `g()`，它将看起来像这样。

---

```
STACK                     REGISTERS               HEAP
+----------------+ <-+
|  g()           |   |
| ret= x()+0x45  |   |
+----------------+   |
|  x()           |   |
| coroframe      | --|-------------------+
| a  = 42        |   |                   +-->  +-----------+
| ret= f()+0x123 |   |    +------+             |  x()      |
+----------------+   +--- | rsp  |             | a =  42   |
|  f()           |        +------+             +-----------+
+----------------+        | rbp  |
| ...            |        +------+
|                |
```

When `g()` returns it will destroy its activation frame and restore `x()`’s activation frame. Let’s say we save `g()`’s return value in a local variable `b` which is stored in the coroutine frame.

---

当 `g()` 返回时，它会销毁其激活帧并恢复 `x()` 的激活帧。假设我们将 `g()` 的返回值保存在一个局部变量 `b` 中，该变量存储在协程帧中。

---

```
STACK                     REGISTERS               HEAP
+----------------+ <-+
|  x()           |   |
| a  = 42        |   |                   +-->  +-----------+
| ret= f()+0x123 |   |    +------+       |     |  x()      |
+----------------+   +--- | rsp  |       |     | a =  42   |
|  f()           |        +------+       |     | b = 789   |
+----------------+        | rbp  | ------+     +-----------+
| ...            |        +------+
|                |
```

If `x()` now hits a suspend-point and suspends execution without destroying its activation frame then execution returns to `f()`.

This results in the stack-frame part of `x()` being popped off the stack while leaving the coroutine-frame on the heap. When the coroutine suspends for the first time, a return-value is returned to the caller. This return value often holds a handle to the coroutine-frame that suspended that can be used to later resume it. When `x()` suspends it also stores the address of the resumption-point of `x()` in the coroutine frame (call it `RP` for resume-point).

---

如果 `x()` 现在到达一个挂起点并挂起执行而不销毁其激活帧，则执行控制权返回到 `f()`。

这导致 `x()` 的栈帧部分从栈中弹出，而协程帧仍保留在堆上。当协程首次挂起时，返回值会被返回给调用者。这个返回值通常包含一个可以用于稍后恢复该协程的句柄。当 `x()` 挂起时，它还会将 `x()` 的恢复点地址存储在协程帧中（称为 `RP` 以表示恢复点）。

---

```
STACK                     REGISTERS               HEAP
                                        +----> +-----------+
                          +------+      |      |  x()      |
+----------------+ <----- | rsp  |      |      | a =  42   |
|  f()           |        +------+      |      | b = 789   |
| handle     ----|---+    | rbp  |      |      | RP=x()+99 |
| ...            |   |    +------+      |      +-----------+
|                |   |                  |
|                |   +------------------+
```

This handle may now be passed around as a normal value between functions. At some point later, potentially from a different call-stack or even on a different thread, something (say, `h()`) will decide to resume execution of that coroutine. For example, when an async I/O operation completes.

The function that resumes the coroutine calls a `void resume(handle)` function to resume execution of the coroutine. To the caller, this looks just like any other normal call to a `void`-returning function with a single argument.

This creates a new stack-frame that records the return-address of the caller to `resume()`, activates the coroutine-frame by loading its address into a register and resumes execution of `x()` at the resume-point stored in the coroutine-frame.

---

这个句柄现在可以像普通值一样在函数之间传递。在稍后的某个时刻，可能从不同的调用栈甚至不同的线程中，某个函数（假设为 `h()`）会决定恢复该协程的执行。例如，当一个异步 I/O 操作完成时。

恢复协程的函数调用一个 `void resume(handle)` 函数来恢复协程的执行。对于调用者来说，这看起来就像对一个返回 `void` 且带有一个参数的普通函数的调用。

这会创建一个新的栈帧，记录调用 `resume()` 的返回地址，通过将协程帧的地址加载到寄存器中激活协程帧，并在协程帧中存储的恢复点处恢复 `x()` 的执行。

---

```
STACK                     REGISTERS               HEAP
+----------------+ <-+
|  x()           |   |                   +-->  +-----------+
| ret= h()+0x87  |   |    +------+       |     |  x()      |
+----------------+   +--- | rsp  |       |     | a =  42   |
|  h()           |        +------+       |     | b = 789   |
| handle         |        | rbp  | ------+     +-----------+
+----------------+        +------+
| ...            |
|                |
```

## In summary

I have described coroutines as being a generalisation of a function that has three additional operations - ‘Suspend’, ‘Resume’ and ‘Destroy’ - in addition to the ‘Call’ and ‘Return’ operations provided by “normal” functions.

I hope that this provides some useful mental framing for how to think of coroutines and their control-flow.

In the next post I will go through the mechanics of the C++ Coroutines TS language extensions and explain how the compiler translates code that you write into coroutines.

---

我已经描述了协程是函数的一种泛化，它除了具有‘普通’函数提供的‘调用’和‘返回’操作之外，还有三个额外的操作——‘挂起’、‘恢复’和‘销毁’。

我希望这能为如何思考协程及其控制流提供一些有用的思维框架。

在下一篇文章中，我将详细介绍 C++ 协程技术规范（Coroutines TS）的语言扩展，并解释编译器如何将你编写的代码转换为协程。

---

