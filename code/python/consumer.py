import pika
import pika

# 设置用户名和密码
credentials = pika.PlainCredentials('terra', 'kkly2021.com@@')
# 连接到 RabbitMQ 服务器
connection = pika.BlockingConnection(pika.ConnectionParameters('8.138.98.54', credentials=credentials))
channel = connection.channel()


# 声明队列
# channel.queue_declare(queue='log_1')
# 定义回调函数，用于处理接收到的消息
def callback(ch, method, properties, body):
    print(f" [x] 接收到消息: {body.decode()}")
# 开始消费消息
channel.basic_consume(queue='log_warn',
                      auto_ack=True,
                      on_message_callback=callback)
print(' [*] 正在等待消息。按 CTRL+C 退出。')
channel.start_consuming()
