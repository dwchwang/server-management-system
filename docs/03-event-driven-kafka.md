# Kiến trúc Event-Driven với Kafka KRaft

Hệ thống VCS-SMS sử dụng **Apache Kafka (phiên bản 3.9 KRaft)** làm hệ thần kinh trung ương (Message Broker) để liên kết các microservices với nhau theo mô hình Bất đồng bộ (Asynchronous).

## 1. Tại sao cần Message Broker (Kafka) thay vì gọi HTTP API?

Hãy tưởng tượng luồng Import Excel:
Người dùng upload 1 file Excel chứa 5000 servers. `fileio-service` đọc file. Sau đó nó phải báo cho `server-service` để lưu 5000 dòng này vào Postgres.
- Nếu dùng HTTP API: `fileio-service` gửi request, `server-service` mất 10 giây để chèn DB. Kết nối HTTP bị treo 10 giây, người dùng ngồi nhìn màn hình loading xoay tròn, rủi ro timeout rất cao.
- **Dùng Kafka (Event-Driven)**: `fileio-service` chỉ việc bắn một tin nhắn (Event) "Có 5000 server cần import" vào Kafka (mất vài mili-giây) rồi phản hồi ngay cho người dùng: "Hệ thống đang xử lý, vui lòng kiểm tra lại sau". Sau đó, `server-service` từ từ lấy tin nhắn đó từ Kafka ra để xử lý ngầm (Background job).

**Lợi ích của Event-Driven:**
1. **Decoupling (Giải phóng phụ thuộc)**: Service gửi (Producer) không cần biết Service nhận (Consumer) là ai, có đang sống hay không. Nếu `server-service` đang sập bảo trì, thông báo vẫn nằm an toàn trong Kafka. Khi `server-service` bật lên, nó sẽ đọc tiếp mà không mất dữ liệu.
2. **Buffer Spikes (Chống sốc tải)**: Khi có 1 lúc 10 người cùng import 50.000 servers, Kafka sẽ lưu trữ chúng lại như một hàng đợi. `server-service` sẽ rút dần ra xử lý theo sức của nó, không bị sập vì quá tải.
3. **Eventual Consistency (Nhất quán cuối cùng)**: Trạng thái hệ thống được đồng bộ dần dần. Khi Server thay đổi trạng thái (On -> Off), event `server.status.changed` bắn ra. Elasticsearch sẽ bắt event này để log lại. Các thành phần đều nhất quán nhờ theo dõi chung 1 dòng sự kiện.

## 2. Apache Kafka KRaft là gì?

Trước đây (trước bản 3.x), Kafka bắt buộc phải đi kèm với một hệ thống khác tên là **ZooKeeper** để quản lý metadata (thông tin cụm, bầu leader, v.v.). Điều này làm tăng độ phức tạp: phải cài đặt 2 hệ thống, monitor 2 hệ thống.

Từ bản 3.3 trở đi, Kafka giới thiệu **KRaft (Kafka Raft)** — một giao thức nội bộ thay thế hoàn toàn ZooKeeper. 
Trong dự án này, chúng ta sử dụng Apache Kafka 3.9 ở chế độ KRaft:
- Loại bỏ hoàn toàn container ZooKeeper khỏi Docker.
- Hệ thống nhẹ hơn, start/stop nhanh hơn.
- Cấu hình qua các tham số `process.roles=broker,controller`. Node Kafka tự đóng vai trò vừa là broker lưu data, vừa là controller quản lý metadata.

## 3. Khái niệm Topic, Partition và Consumer Group

Để làm việc với Kafka, bạn cần nắm rõ:
- **Topic**: Giống như "kênh radio". Ví dụ `server.created`. Producer phát sóng lên kênh này, ai quan tâm thì đăng ký nghe.
- **Partition**: Để tăng tốc độ, một Topic có thể chia thành nhiều luồng song song gọi là Partition. Ví dụ chia làm 3 partitions. Dữ liệu sẽ rải đều vào 3 ống này.
- **Consumer Group**: Một nhóm các instances của 1 Service. Ví dụ bạn chạy 3 bản `monitor-service`. Khi Kafka thấy group này có 3 người, nó sẽ tự động chia: mỗi người đọc 1 partition. Nhờ đó, 3 bản `monitor-service` sẽ hợp sức xử lý log mà không bị trùng lặp dữ liệu (Load Balancing).

## 4. Các luồng sự kiện chính trong VCS-SMS

1. **Luồng Cập nhật Server**: `server-service` CRUD DB xong -> Bắn event `server.created/updated/deleted`. `monitor-service` nghe để cập nhật lại danh sách mục tiêu cần ping.
2. **Luồng Trạng thái (Status)**: `monitor-service` check thấy server sập -> Bắn event `server.status.changed`. Bất cứ service nào quan tâm (VD hệ thống Alert sau này) có thể nghe.
3. **Luồng Batch Log**: Mỗi 1 phút có 10.000 kết quả ping. `monitor-service` gom lại thành 1 mảng lớn (Batch) -> Bắn event `server.health.batch`. `report-service` (hoặc logstash) nghe để đẩy vào Elasticsearch tối ưu hóa network.

## 5. Go Kafka Client: segmentio/kafka-go

### Tại sao chọn segmentio/kafka-go thay vì IBM/sarama?

| Tiêu chí | segmentio/kafka-go | IBM/sarama |
|----------|:------------------:|:----------:|
| **Pure Go** | ✅ Có (không CGo) | ✅ Có |
| **Context support** | ✅ Native (`context.Context` trong mọi API) | ⚠️ Partial |
| **API simplicity** | ⭐ `WriteMessages()`/`ReadMessage()` | ⭐⭐ Cần ConsumerGroupHandler |
| **Connection management** | Tự động reconnect, health-check | Manual qua config |
| **Performance** | Tốt cho high-throughput | Tốt, nhưng nặng hơn |
| **Dependencies** | Nhẹ, ít transitive deps | Nhiều transitive deps (gokrb5, etc.) |
| **Community** | 14K+ stars, Segment (Twilio) maintain | 11K+ stars, IBM maintain |

**Lý do chọn `segmentio/kafka-go`:**
1. **API đơn giản hơn**: `Writer` cho producer, `Reader` cho consumer — không cần ConsumerGroupHandler phức tạp.
2. **Context-native**: Mọi method đều nhận `context.Context`, phù hợp với pattern Go chuẩn.
3. **Ít dependencies hơn**: Không kéo theo `gokrb5` (Kerberos), `gofork`, giảm attack surface.
4. **Tự động reconnect**: Writer/Reader tự reconnect khi mất kết nối — không cần code retry.

### Producer Pattern (Writer)

```go
w := &kafka.Writer{
    Addr:         kafka.TCP("localhost:9092"),
    Balancer:     &kafka.LeastBytes{},
    BatchSize:    100,
    BatchTimeout: 100 * time.Millisecond,
    RequiredAcks: kafka.RequireAll,
    Compression:  kafka.Snappy,
}
defer w.Close()

err := w.WriteMessages(ctx, kafka.Message{
    Topic: "server.health.batch",
    Key:   []byte(serverID),
    Value: jsonBytes,
})
```

### Consumer Pattern (Reader)

```go
r := kafka.NewReader(kafka.ReaderConfig{
    Brokers:        []string{"localhost:9092"},
    GroupID:        "monitor-group",
    Topic:          "server.created",
    MinBytes:       10e3,
    MaxBytes:       10e6,
    CommitInterval: 5 * time.Second,
    StartOffset:    kafka.LastOffset,
})
defer r.Close()

for {
    msg, err := r.ReadMessage(ctx)
    if err != nil { break }
    // xử lý msg.Value
}
```
