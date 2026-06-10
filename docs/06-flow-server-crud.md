# Luồng vận hành: Cập nhật và Quản lý Server (CRUD)

Mặc dù việc thêm/sửa/xóa Server nghe có vẻ là tính năng cơ bản nhất của một hệ thống, nhưng trong kiến trúc Event-Driven của VCS-SMS, mọi thao tác thay đổi dữ liệu đều tạo ra một chuỗi phản ứng dây chuyền (ripple effect) đến các service khác.

## 1. Luồng Tạo mới Server (Create)

1. **Người dùng gửi yêu cầu**: Quản trị viên (Admin) nhập thông tin server (IP, OS, CPU, RAM...) trên giao diện và bấm Lưu.
2. **Gateway Xác thực**: API Gateway kiểm tra JWT xem người này có quyền `server:create` không. Nếu hợp lệ, nó chuyển tiếp dữ liệu xuống `server-service`.
3. **Validate & Lưu trữ**: `server-service` kiểm tra xem IP hoặc Tên server đã tồn tại chưa. Định dạng IP có chuẩn IPv4 không. Sau đó, nó lưu bản ghi vào bảng `servers` trong PostgreSQL (lúc này trạng thái mặc định là `off`).
4. **Phát sóng Sự kiện (Publish Event)**: 
   Ngay khi lưu DB thành công, `server-service` không gọi điện báo trực tiếp cho ai cả. Nó ném một tin nhắn lên Kafka vào kênh (topic) `server.created`. Tin nhắn này chứa toàn bộ thông tin của server mới.
5. **Monitor Service Lắng nghe**: `monitor-service` (luôn luôn túc trực nghe ngóng trên Kafka) nhận được tin nhắn. Nó tự động cập nhật danh sách server trong bộ nhớ của nó, sẵn sàng cho vòng quét Health-Check ở phút tiếp theo mà không cần phải chờ truy vấn lại toàn bộ Database.

## 2. Luồng Cập nhật Server (Update)

1. Quản trị viên thay đổi thông tin (ví dụ: đổi server từ mục đích "Web" sang "Database").
2. `server-service` cập nhật vào Postgres.
3. Đồng thời xóa Cache (bộ nhớ tạm) của server này trong Redis để lần truy xuất sau người dùng lấy được dữ liệu mới nhất.
4. Bắn event `server.updated` lên Kafka. Các dịch vụ khác nếu cần (ví dụ dịch vụ Alerting sau này) có thể biết được thông tin server vừa thay đổi để điều chỉnh logic.

## 3. Luồng Xóa Server (Soft Delete)

Trong các hệ thống quản trị lớn, chúng ta hiếm khi xóa hẳn dữ liệu khỏi ổ cứng (Hard Delete) vì lý do lịch sử và an toàn.
1. Quản trị viên bấm Xóa Server.
2. `server-service` không gọi lệnh `DELETE FROM`. Thay vào đó, nó cập nhật cột `deleted_at` thành thời gian hiện tại (Soft Delete). 
3. Kể từ lúc này, mọi câu query lấy danh sách server đều tự động thêm điều kiện `WHERE deleted_at IS NULL` (bị ẩn khỏi người dùng).
4. Nó bắn event `server.deleted` lên Kafka.
5. `monitor-service` nhận được event này, nó lập tức gạch tên server này khỏi danh sách Health-Check. Ngừng việc tốn tài nguyên đi ping một server đã bị xóa.

## 4. Luồng Xem danh sách (Read / Filter)

Vì hệ thống có 10.000 server, việc lấy toàn bộ dữ liệu 1 lúc là thảm họa.
1. Người dùng gửi request kèm theo bộ lọc (Filter): Ví dụ "Lấy các server đang ON, chạy Ubuntu, sắp xếp theo tên, trang 1, mỗi trang 20 server".
2. `server-service` sẽ dịch bộ lọc này thành câu truy vấn SQL chuẩn xác, tận dụng các Index (chỉ mục) đã tạo sẵn trong PostgreSQL để tìm kiếm với tốc độ mili-giây.
3. Kết quả được lưu tạm vào Redis (Caching). Lần tới nếu có người truy vấn y hệt trang 1 này, hệ thống sẽ móc từ Redis ra trả về ngay lập tức mà không cần chọc xuống DB, giúp giảm tải cực kỳ hiệu quả.
