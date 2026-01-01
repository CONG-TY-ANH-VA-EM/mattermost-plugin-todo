# Hướng Dẫn Sử Dụng Chi Tiết - Mattermost Todo Plugin v2.0

## Mục Lục
1. [Giới Thiệu](#giới-thiệu)
2. [Cài Đặt Plugin](#cài-đặt-plugin)
3. [Cấu Hình](#cấu-hình)
4. [Sử Dụng Cơ Bản](#sử-dụng-cơ-bản)
5. [Tính Năng AI](#tính-năng-ai)
6. [Tính Năng Nâng Cao](#tính-năng-nâng-cao)
7. [Câu Hỏi Thường Gặp](#câu-hỏi-thường-gặp)

## Giới Thiệu

Mattermost Todo Plugin v2.0 là một công cụ quản lý công việc mạnh mẽ được tích hợp ngay trong Mattermost. Phiên bản 2.0 đã được cải tiến toàn diện với:

- **🤖 Trí Tuệ Nhân Tạo**: Tạo công việc bằng ngôn ngữ tự nhiên
- **🗄️ Cơ Sở Dữ Liệu SQL**: Hiệu suất cao, hỗ trợ hàng nghìn công việc
- **🌍 Tiếng Việt**: Giao diện hoàn toàn tiếng Việt
- **💬 Bình Luận**: Thảo luận trực tiếp trên công việc
- **⏰ Hạn Chót & Ưu Tiên**: Quản lý deadline hiệu quả

## Cài Đặt Plugin

### Bước 1: Tải Plugin
1. Truy cập [GitHub Releases](https://github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo/releases)
2. Tải file `com.mattermost.plugin-todo.tar.gz` phiên bản mới nhất

### Bước 2: Cài Đặt Trên Mattermost
1. Đăng nhập Mattermost với tài khoản **System Admin**
2. Vào **Main Menu → System Console**
3. Chọn **Plugins → Plugin Management**
4. Nhấn **Choose File** và chọn file đã tải
5. Nhấn **Upload**
6. Kích hoạt plugin bằng cách toggle **Enable Plugin**

### Bước 3: Xác Nhận
- Kiểm tra thanh bên phải, bạn sẽ thấy biểu tượng **Todo**
- Gõ `/todo` để kiểm tra plugin hoạt động

## Cấu Hình

### Cấu Hình Cơ Bản

1. Vào **System Console → Plugins → Todo**
2. Các tùy chọn có sẵn:

| Tùy Chọn | Mô Tả | Mặc Định |
|----------|-------|----------|
| Hide Team Sidebar | Ẩn nút Todo trên thanh bên | `false` |
| Enable Smart Todo | Bật tính năng AI | `false` |
| OpenAI API Key | Khóa API để sử dụng AI | (trống) |
| OpenAI Model | Model AI sử dụng | `gpt-4o` |

### Cấu Hình AI (Tùy Chọn)

Để sử dụng tính năng tạo todo bằng giọng nói tự nhiên:

#### Bước 1: Lấy OpenAI API Key
1. Truy cập [OpenAI Platform](https://platform.openai.com/signup)
2. Đăng ký/đăng nhập tài khoản
3. Vào **API Keys** → **Create new secret key**
4. Sao chép key (bắt đầu bằng `sk-...`)

#### Bước 2: Cấu Hình Plugin
1. Vào **System Console → Plugins → Todo**
2. **Enable Smart Todo**: Đặt thành `true`
3. **OpenAI API Key**: Dán key đã sao chép
4. **OpenAI Model**: Để mặc định `gpt-4o` (khuyến nghị)
5. Nhấn **Save**

> **💡 Lưu Ý**: OpenAI API là dịch vụ trả phí. Kiểm tra [bảng giá](https://openai.com/pricing) trước khi sử dụng.

## Sử Dụng Cơ Bản

### Tạo Công Việc Mới

#### Cách 1: Dùng Lệnh
```
/todo add Chuẩn bị báo cáo tuần
```

#### Cách 2: Dùng Giao Diện
1. Nhấn vào biểu tượng **Todo** trên thanh bên phải
2. Nhấn nút **"+"** hoặc **"Thêm Todo"**
3. Điền thông tin:
   - **Tiêu đề**: Mô tả công việc
   - **Mô tả**: Chi tiết thêm (tùy chọn)
   - **Ưu tiên**: Thấp / Trung bình / Cao
   - **Hạn chót**: Chọn ngày giờ deadline
4. Nhấn **Thêm**

### Xem Danh Sách Công Việc

- **Công việc của tôi**: Tab "Việc của tôi"
- **Đang đến**: Việc người khác giao cho bạn
- **Đã gửi**: Việc bạn giao cho người khác

### Hoàn Thành Công Việc

- Tích vào ô checkbox bên cạnh công việc
- Hoặc dùng lệnh: `/todo pop` (hoàn thành việc cũ nhất)

### Chỉnh Sửa Công Việc

1. Nhấp vào công việc để mở chi tiết
2. Nhấn nút **Chỉnh sửa** (biểu tượng bút chì)
3. Thay đổi thông tin
4. Nhấn **Lưu**

### Giao Việc Cho Người Khác

```
/todo send @nguoidung Xem lại tài liệu thiết kế
```

Người nhận sẽ:
- Nhận thông báo từ Todo Bot
- Thấy công việc trong tab **Đang đến**
- Có thể chấp nhận hoặc từ chối

## Tính Năng AI

> **⚠️ Yêu Cầu**: Phải cấu hình OpenAI API Key (xem [Cấu Hình AI](#cấu-hình-ai-tùy-chọn))

### Cách Sử Dụng

Thay vì dùng lệnh `/todo add`, bạn có thể gõ như nói chuyện tự nhiên:

#### Ví Dụ Cơ Bản
```
/todo Cần gọi cho Hào lúc 9h30 sáng mai
```

**AI sẽ tự động nhận diện:**
- Nội dung: "Gọi cho Hào"
- Hạn chót: ngày mai, 09:30
- Ưu tiên: Trung bình (mặc định)

#### Ví Dụ Với Ưu Tiên
```
/todo urgent Fix lỗi server trước 5pm hôm nay
```

**AI nhận diện:**
- Nội dung: "Fix lỗi server"
- Hạn chót: hôm nay, 17:00
- Ưu tiên: Cao (từ từ "urgent")

#### Ví Dụ Phức Tạp
```
/todo Hoàn thành slide thuyết trình cho meeting thứ 6 tuần sau ưu tiên cao
```

**AI nhận diện:**
- Nội dung: "Hoàn thành slide thuyết trình"  
- Hạn chót: Thứ 6 tuần sau
- Ưu tiên: Cao

### Từ Khóa AI Nhận Diện

| Loại | Từ Khóa |
|------|---------|
| **Thời gian** | ngày mai, hôm nay, tuần sau, thứ hai, 9h, 3pm |
| **Ưu tiên cao** | urgent, khẩn cấp, gấp, quan trọng, ưu tiên cao |
| **Ưu tiên thấp** | không gấp, ưu tiên thấp, có thể làm sau |

## Tính Năng Nâng Cao

### Bình Luận & Thảo Luận

1. **Mở công việc**: Nhấp vào công việc trong danh sách
2. **Xem bình luận**: Cuộn xuống phần bình luận
3. **Thêm bình luận**: 
   - Gõ nội dung vào ô "Thêm bình luận…"
   - Nhấn **Gửi**
4. **Mention**: Dùng `@tên` để tag đồng nghiệp

**Ứng dụng:**
- Báo cáo tiến độ
- Yêu cầu hỗ trợ
- Lưu lại ghi chú quan trọng

### Nhắc Nhở Hàng Ngày

1. Gõ `/todo settings`
2. Kích hoạt **Daily Reminders**: `on`
3. Mỗi sáng bạn sẽ nhận danh sách việc chưa hoàn thành

### Xử Lý Yêu Cầu Đến

Khi ai đó giao việc cho bạn:

1. Nhận thông báo từ **Todo Bot**
2. Vào tab **Đang đến**
3. Chọn công việc:
   - **Chấp nhận**: Chuyển sang tab "Việc của tôi"
   - **Từ chối**: Gửi lại người giao

### Tìm Kiếm & Lọc

- **Lọc theo ưu tiên**: Nhấp vào badge ưu tiên
- **Lọc theo hạn chót**: Sắp xếp theo ngày
- **Tìm kiếm**: Dùng ô tìm kiếm (nếu có)

## Câu Hỏi Thường Gặp

### 1. AI không hoạt động?

**Kiểm tra:**
- ✅ Enable Smart Todo = `true`
- ✅ API Key đúng và còn credits
- ✅ Xem log Mattermost: `/var/log/mattermost/mattermost.log`

**Thông báo lỗi phổ biến:**
- `Failed to create smart todo`: Kiểm tra API key
- `API rate limit`: Bạn đã hết quota miễn phí

### 2. Không thấy biểu tượng Todo?

- Plugin chưa được kích hoạt
- Xóa cache trình duyệt (Ctrl+Shift+R)
- Kiểm tra **System Console → Plugins**

### 3. Công việc bị mất?

Todo sử dụng cơ sở dữ liệu SQL của Mattermost:
- Dữ liệu được lưu vĩnh viễn
- Backup theo lịch Mattermost
- Kiểm tra tab **Đã hoàn thành** (nếu có)

### 4. Làm sao xóa công việc?

Hiện tại plugin dùng **soft-delete**:
- Công việc hoàn thành sẽ bị ẩn
- Không có nút xóa vĩnh viễn (để audit)

### 5. Có giới hạn số công việc?

Không. Plugin được tối ưu cho:
- Hàng nghìn công việc mỗi user
- Hàng chục nghìn trên toàn hệ thống

### 6. Có thể import/export không?

Hiện chưa hỗ trợ. Roadmap tương lai:
- Export to CSV
- Import from Trello/Asana
- Backup/Restore riêng

## Hỗ Trợ

- **Bug Report**: [GitHub Issues](https://github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo/issues)
- **Feature Request**: Tạo issue với tag `enhancement`
- **Email**: support@ane.vn (nếu có)

---

**Phiên bản tài liệu**: v2.0  
**Cập nhật lần cuối**: 2026-01-01
