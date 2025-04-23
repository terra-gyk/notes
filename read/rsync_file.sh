rsync -avz --progress --checksum /mnt/d/1_server_down/0_微服务架构设计模式/ ./0_微服务架构设计模式/
rsync -avz --progress --checksum /mnt/d/1_server_down/1_消息队列/ ./1_消息队列/

find ./ -type f \( -name "*.py" -o -name "*.md" -o -name "*.png" \) -exec chmod 644 {} \;