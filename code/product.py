import pika
import pika

# 设置用户名和密码
credentials = pika.PlainCredentials('terra', 'kkly2021.com@@')
# 连接到 RabbitMQ 服务器
connection = pika.BlockingConnection(pika.ConnectionParameters('8.138.98.54', credentials=credentials))
channel = connection.channel()


# 声明一个名为 'test' 的队列
channel.queue_declare(queue='test')
# 要发送的消息
message = "这是一条测试消息"
# 向队列发送消息
channel.basic_publish(exchange='',
                      routing_key='test',
                      body=message)
print(f" [x] 已发送消息: '{message}'")
# 关闭连接
connection.close()