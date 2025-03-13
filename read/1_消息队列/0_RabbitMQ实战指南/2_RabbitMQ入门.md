## 2.1 相关概念

![image-20250305104817666](image\image-20250305104817666.png)

### 2.1.1 生产者和消费者

- Producer：生产者，投递消息的一方
- Consumer：消费者，接收消息的一方
- Broker：消息中间件的服务节点



### 2.1.2 队列

- Queue：队列，是RabbitMQ的内部对象，用于存储消息。
- 多个消费者可以订阅同一个队列，这时队列中的消息会被平均分摊（Round-Robin，即轮询）给多个消费者处理。
- RabbitMQ不支持队列层面的广播消费



### 2.1.3 交换器、路由键、绑定

- Exchange：交换器
- RouingKey：路由键，用于指定路由规则
- Binding：绑定，RabbitMQ中通过绑定将交换器与队列关联起来，BindingKey 与 RoutingKey 相匹配时，消息会被路由到对应的队列中。这些绑定允许使用相同的BindingKey（可以通过一个RoutingKey将消息发送到多个队列中）。
- 注意RoutingKey是用于路由的关键字，BindingKey 是绑定时的关键字，不是同一个东西

### 2.1.4 交换器类型

- fanout：把所有发送到该交换器的消息路由到所有与该交换器绑定的队列中。
- direct：它会把消息路由到哪些BindingKey和RoutingKey完全匹配的队列中。
- topic：模糊匹配，RoutingKey 和 BindingKey可以使用 ‘.’ 来分隔，如 log.info.time
- BindingKey：中使用 ’*‘ 和 ’#‘ 用于模糊匹配，
  - `*`（星号）：可以匹配一个单词。例如，`*.error` 可以匹配 `info.error`、`warning.error` 等，但不能匹配 `info.warning.error`（可以使用：*.warning.\* 或 info.\*.\*）。
  - `#`（井号）：可以匹配零个或多个单词。例如，`order.#` 可以匹配 `order`、`order.create`、`order.pay.success` 等。

### 2.1.5 RabbitMQ运转流程

- 生产者向 RabbitMQ 发送消息的详细流程：
  1. 生产者与 RabbitMQ Broker 建立连接并开启信道。
  2. 声明交换器，设置其类型、是否持久化等属性。
  3. 声明队列，设置是否排他、持久化、自动删除等属性。
  4. 通过路由键将交换器和队列绑定。
  5. 发送包含路由键、交换器等信息的消息至 RabbitMQ Broker。
  6. 交换器依据路由键查找匹配队列。
  7. 若找到，消息存入队列。
  8. 若未找到，按生产者配置属性决定丢弃或回退消息。
  9. 关闭信道。
  10. 关闭连接。
- 消费者接收 RabbitMQ 消息的流程：
  1. 消费者连接到 RabbitMQ Broker，建立连接并开启信道。
  2. 向 Broker 请求消费队列中的消息，可能设置回调函数并做准备工作。
  3. 等待 Broker 回应并投递消息，然后接收消息。
  4. 确认（ack）接收到的消息。
  5. RabbitMQ 从队列中删除已确认的消息。
  6. 关闭信道 。
  7. 关闭连接。























