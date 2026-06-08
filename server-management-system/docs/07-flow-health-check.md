# Luồng vận hành: Health-Check 10.000 Server

Đây là trái tim của hệ thống, hoạt động bền bỉ, lặp đi lặp lại không ngừng nghỉ để theo dõi "nhịp tim" của 10.000 máy chủ.

## 1. Chuẩn bị (Trước khi chạy)

Bộ máy `monitor-service` sở hữu một Đồng hồ đếm nhịp (Cron Scheduler) được đặt lịch cứ đúng **60 giây (1 phút)** là reo chuông một lần.

Khi chuông reo:
1. Nó chạy ra hỏi Redis: "Cho tôi xin cái chìa khóa (Lock)". Redis cấp khóa. Nếu có 1 bản sao `monitor-service` thứ 2 cũng chạy ra xin khóa, nó sẽ bị Redis từ chối để đảm bảo chỉ 1 người được làm việc.
2. Nó kết nối qua Postgres (với quyền đọc chéo Schema), lấy về danh sách toàn bộ 10.000 server đang hoạt động (không bị xóa).

## 2. Quá trình "Ping" (Worker Pool)

Thay vì 1 người chạy đi kiểm tra 10.000 nhà, `monitor-service` phái ra **100 công nhân (Workers)** làm việc song song.
Hệ thống lấy thông tin cấu hình Health-Check của từng server để biết phải dùng cách nào:

- **Nếu cấu hình là TCP**: Công nhân sẽ thử tạo kết nối mạng thực sự (mở TCP Socket) tới địa chỉ IP của server đó. Nếu kết nối thành công trong vòng 5 giây, server được tính là **ON**. Nếu quá 5 giây hoặc bị từ chối kết nối, server bị tính là **OFF**.
- **Nếu cấu hình là Simulator (Giả lập cho Demo)**: Vì không có 10.000 máy thật, công nhân sẽ nhìn vào tỷ lệ sống `uptime_rate` (VD: 95%). Công nhân sẽ tung một viên xúc xắc. 95% khả năng viên xúc xắc rơi vào ô **ON**, 5% rơi vào ô **OFF**. Để đồ thị trông tự nhiên, hệ thống còn thêm vào một chút nhiễu sóng (tăng giảm tỷ lệ theo giờ trong ngày).

## 3. Tổng hợp và Ghi nhận Kết quả (Batch Processing)

Khi 100 công nhân đã báo cáo xong toàn bộ 10.000 kết quả, hệ thống bắt đầu xử lý hậu kỳ. Đây là bước phải làm cực kỳ tối ưu để không làm "sập" Database.

**Bước 3.1: So sánh trạng thái**
Hệ thống cầm 10.000 kết quả mới đi so sánh với 10.000 kết quả của phút trước (đang lưu sẵn trên RAM/Redis).
Nó phát hiện ra: "À, 9.990 server trạng thái vẫn không đổi. Chỉ có 10 server vừa từ ON chuyển sang OFF (bị sập)".

**Bước 3.2: Phát sóng sự kiện thay đổi (Alerting Foundation)**
Với 10 server bị thay đổi trạng thái đó, nó lập tức ném 10 tin nhắn lên Kafka (kênh `server.status.changed`). Hệ thống Alert hoặc màn hình Frontend có thể nhận sự kiện này để kêu bíp bíp, chớp đỏ báo động ngay lập tức theo thời gian thực (Real-time).

**Bước 3.3: Cập nhật PostgreSQL cực nhẹ**
Hệ thống tạo ra 1 câu lệnh SQL duy nhất (Batch Update) gửi xuống PostgreSQL để cập nhật chữ "on" thành "off" cho đúng 10 server bị thay đổi kia. Bỏ qua 9.990 server không đổi. Database thở phào nhẹ nhõm. Cập nhật xong, nó lưu lại trạng thái mới nhất vào Redis cho phút sau so sánh.

**Bước 3.4: Đổ Log vào Elasticsearch (Lưu trữ Big Data)**
Mặc dù DB chỉ cập nhật 10 server, nhưng **Lịch sử nhịp tim (Log)** của cả 10.000 server đều phải được ghi lại để vẽ biểu đồ và tính Uptime.
Hệ thống đóng gói 10.000 bản ghi này thành 1 gói hàng khổng lồ (Bulk Request), gửi 1 lần duy nhất sang Elasticsearch. Elasticsearch (công cụ chuyên trị Big Data) sẽ nuốt trọn 10.000 bản ghi này trong vài chục mili-giây và đánh index (chỉ mục) để sau này tìm kiếm siêu tốc.

Vòng lặp kết thúc, các công nhân nghỉ ngơi chờ chuông reo ở phút tiếp theo.
