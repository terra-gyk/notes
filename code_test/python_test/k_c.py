from kafka import KafkaConsumer

# 确保这里的地址和端口与 Kafka 实际监听的一致
consumer = KafkaConsumer(
    'test_topic',
    bootstrap_servers='localhost:9092',
    group_id='my_consumer_group'
)

for message in consumer:
    print(f"Received message: {message.value.decode('utf-8')}")