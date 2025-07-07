import chardet

def convert_bytes_to_utf8(raw_bytes: bytes) -> bytes:
    """
    将任意编码的字节流转换为 UTF-8 编码的字节流。
    
    参数:
        raw_bytes (bytes): 原始字节数据
        
    返回:
        bytes: 使用 UTF-8 编码的字节流
    """
    # 检测原始编码
    result = chardet.detect(raw_bytes)
    encoding = result['encoding']
    confidence = result['confidence']

    print(f"Detected encoding: {encoding} (confidence: {confidence:.2f})")

    # 如果置信度太低，尝试备选解码方式
    if confidence < 0.5:
        try:
            text = raw_bytes.decode('utf-8')
        except UnicodeDecodeError:
            try:
                text = raw_bytes.decode('gbk')
            except UnicodeDecodeError:
                text = raw_bytes.decode('latin1')
    else:
        text = raw_bytes.decode(encoding)

    # 统一转为 UTF-8 编码的 bytes 输出
    return text.encode('utf-8')
