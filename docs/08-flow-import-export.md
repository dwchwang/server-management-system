# Luồng vận hành: Import & Export Excel (Bất đồng bộ)

Bài toán Import 5000 server bằng file Excel chứa đựng rủi ro cực lớn về hiệu năng (Timeout). VCS-SMS giải quyết bằng cơ chế Xử lý nền (Background Job) kết hợp với Kafka.

## 1. Luồng Import Excel (Asynchronous)

Hãy tưởng tượng bạn ra ngân hàng làm thủ tục, thay vì đứng chờ nhân viên làm xong mới được đi về, bạn bốc số, đưa hồ sơ rồi đi uống cà phê. Lát sau có tin nhắn báo hồ sơ đã duyệt xong. Đó chính là Asynchronous (Bất đồng bộ).

1. **Upload File**: Người dùng chọn file `.xlsx` chứa hàng ngàn server và bấm Import. File bay thẳng vào API Gateway rồi truyền tới `fileio-service`.
2. **Khởi tạo Job**: `fileio-service` lập tức tạo một bản ghi "Hồ sơ" trong bảng `import_jobs` với trạng thái là `PENDING` (Đang chờ xử lý). Nó sinh ra một mã `Job_ID` (ví dụ: JOB-999).
3. **Phản hồi ngay lập tức**: Service trả ngay mã `JOB-999` về cho màn hình của người dùng. Màn hình người dùng tắt vòng xoay loading, hiện thanh tiến trình (Progress bar) và chuyển sang trạng thái chờ. Người dùng có thể đi thao tác màn hình khác.
4. **Đẩy việc vào Kafka**: Cùng lúc đó, `fileio-service` ném một sự kiện `import.job.created` lên Kafka (gửi kèm file hoặc đường dẫn file).
5. **Xử lý nền**: Một "Công nhân ngầm" (Background Worker) của chính `fileio-service` (hoặc `server-service`) rảnh rỗi sẽ nhặt sự kiện từ Kafka xuống. Nó bắt đầu mở file Excel, đọc từng dòng (parse), kiểm tra tính hợp lệ của IP, cấu hình.
6. **Lưu trữ hàng loạt**: Nó gom dữ liệu thành từng lô (Batch), ví dụ 500 dòng 1 lần, rồi nhét thẳng (Insert) vào Database (bảng `servers`).
7. **Cập nhật tiến độ**: Làm xong lô nào, công nhân cập nhật % tiến độ vào Database.
8. **Hoàn tất**: Khi đọc hết file, công nhân đổi trạng thái Job thành `COMPLETED` (hoặc `FAILED` nếu file lỗi). 
9. **Tracking**: Trong lúc hệ thống đang cật lực xử lý ngầm, giao diện Web của người dùng thỉnh thoảng (vài giây 1 lần) gửi mã `JOB-999` lên để hỏi tiến độ, từ đó làm đầy thanh Progress bar trên màn hình cho tới khi báo thành công 100%.

## 2. Luồng Export Excel (Synchronous)

Ngược lại với Import, việc Export (Tải về file Excel) thường đòi hỏi phải trả file ngay lập tức về trình duyệt (Download).

1. Người dùng bấm nút "Export", có thể đính kèm theo các bộ lọc (ví dụ: Chỉ xuất các Server đang OFF).
2. Yêu cầu truyền tới `fileio-service`.
3. Service này sử dụng quyền Đọc chéo (Cross-schema SELECT) để móc danh sách server từ Database lên.
4. Nó sử dụng thư viện `excelize` để tự động tạo một file Excel ảo trong bộ nhớ RAM, vẽ các cột, đổ dữ liệu từng dòng vào.
5. Sau khi file Excel thành hình trong RAM, hệ thống cấu hình HTTP Headers dạng `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` và `Content-Disposition: attachment; filename="servers.xlsx"`.
6. Nó phun luồng byte (stream) của file thẳng về trình duyệt của người dùng. Một popup Tải về sẽ hiện lên trên máy của Admin.

Bằng cách giới hạn bộ lọc hoặc làm phân trang (Pagination) ngầm, việc Export được kiểm soát bộ nhớ chặt chẽ để tránh làm sập service khi người dùng lỡ tay bấm Export toàn bộ dữ liệu hệ thống cùng lúc.
