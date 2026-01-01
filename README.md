# Mattermost Todo Plugin v2.0

[![Release](https://img.shields.io/github/v/release/CONG-TY-ANH-VA-EM/mattermost-plugin-todo)](https://github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo/releases/latest)

[English](#english) | [Tiếng Việt](#tiếng-việt)

---

## English

A powerful enterprise-grade Todo plugin for Mattermost with AI-powered natural language processing, SQL backend, and comprehensive task management features.

### ✨ Key Features

#### 🚀 v2.0 Highlights
- **🤖 AI Integration**: Create todos using natural language (powered by OpenAI)
- **🗄️ SQL Backend**: Scalable PostgreSQL/MySQL storage with migrations
- **🌍 Internationalization**: Full support for English and Vietnamese
- **💬 Task Comments**: Threaded discussions on todo items
- **⏰ Due Dates & Priorities**: Track deadlines and urgency levels
- **📊 Audit Logs**: Complete traceability for compliance
- **🔒 Security**: RBAC, input sanitization, XSS protection

### 📦 Installation

1. Download the latest release from [GitHub Releases](https://github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo/releases)
2. Navigate to **System Console → Plugin Management**
3. Upload `com.mattermost.plugin-todo.tar.gz`
4. Enable the plugin

### ⚙️ Configuration

#### Basic Setup
1. Go to **System Console → Plugins → Todo**
2. Configure basic settings:
   - **Hide Team Sidebar**: Toggle sidebar buttons visibility

#### 🤖 AI Features (Optional)
To enable natural language todo creation:

1. **Enable Smart Todo**: Set to `true`
2. **OpenAI API Key**: Enter your API key (starts with `sk-...`)
3. **OpenAI Model**: Choose model (default: `gpt-4o`)

**Get an API Key**: Visit [OpenAI Platform](https://platform.openai.com/api-keys)

### 📖 Usage

#### Creating Todos

**Traditional Method:**
```
/todo add Review pull request #123
```

**🤖 AI Method** (if enabled):
```
/todo Call John tomorrow at 3pm urgent
/todo Fix server crash by Friday high priority
/todo Review documentation
```

The AI automatically extracts:
- **Task description**: Main content
- **Due date**: Absolute or relative times
- **Priority**: High, Medium, Low (inferred from context)

#### Managing Tasks

| Command | Description |
|---------|-------------|
| `/todo` | Open your todo list |
| `/todo add <message>` | Create a new todo |
| `/todo list` | View all your todos |
| `/todo pop` | Complete oldest todo |
| `/todo send @username <message>` | Assign todo to someone |
| `/todo settings` | Configure reminders |

#### Using the Sidebar

1. Click the **Todo** icon in the right sidebar
2. **Add Todo**: Click the "+" button
   - Set priority (Low/Medium/High)
   - Set due date
   - Add description
3. **Comments**: Click on a todo to view/add comments
4. **Complete**: Check the box to mark as done

### 🔧 Advanced Features

#### Collaboration
- **Send Tasks**: Delegate to team members with `/todo send @user Task description`
- **Incoming Requests**: Accept or decline tasks sent to you
- **Notifications**: Receive updates via the Todo bot

#### Daily Reminders
Enable daily reminders in settings to get a summary of pending tasks each morning.

#### Comments & Discussion
- Click any todo item to open the comment thread
- Add context, updates, or ask questions
- All comments are tracked in audit logs

### 🛠️ Development

#### Prerequisites
- Go 1.22+
- Node.js 18+
- Mattermost Server 6.5+

#### Building from Source

```bash
# Build server binaries
cd server
go build -o dist/plugin-linux-amd64

# Build webapp
cd webapp
npm install
npm run build

# Package plugin
make dist
```

#### Running Tests

```bash
# Server tests
cd server
go test ./...

# Webapp tests
cd webapp
npm test
```

### 📝 License

This project is licensed under the MIT License.

### 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## Tiếng Việt

Plugin Todo cấp doanh nghiệp mạnh mẽ cho Mattermost với xử lý ngôn ngữ tự nhiên AI, backend SQL và các tính năng quản lý tác vụ toàn diện.

### ✨ Tính Năng Chính

#### 🚀 Nâng Cấp v2.0
- **🤖 Tích Hợp AI**: Tạo todo bằng ngôn ngữ tự nhiên (OpenAI)
- **🗄️ Backend SQL**: Lưu trữ PostgreSQL/MySQL có khả năng mở rộng
- **🌍 Đa Ngôn Ngữ**: Hỗ trợ đầy đủ Tiếng Anh và Tiếng Việt
- **💬 Bình Luận**: Thảo luận trực tiếp trên todo
- **⏰ Hạn Chót & Ưu Tiên**: Theo dõi deadline và mức độ khẩn cấp
- **📊 Nhật Ký Kiểm Toán**: Truy xuất đầy đủ để tuân thủ
- **🔒 Bảo Mật**: RBAC, lọc đầu vào, bảo vệ XSS

### 📦 Cài Đặt

1. Tải phiên bản mới nhất từ [GitHub Releases](https://github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo/releases)
2. Vào **System Console → Plugin Management**
3. Tải lên file `com.mattermost.plugin-todo.tar.gz`
4. Kích hoạt plugin

### ⚙️ Cấu Hình

#### Thiết Lập Cơ Bản
1. Vào **System Console → Plugins → Todo**
2. Cấu hình các thiết lập:
   - **Hide Team Sidebar**: Ẩn/hiện nút trên thanh bên

#### 🤖 Tính Năng AI (Tùy Chọn)
Để kích hoạt tạo todo bằng ngôn ngữ tự nhiên:

1. **Enable Smart Todo**: Đặt thành `true`
2. **OpenAI API Key**: Nhập API key của bạn (bắt đầu bằng `sk-...`)
3. **OpenAI Model**: Chọn model (mặc định: `gpt-4o`)

**Lấy API Key**: Truy cập [OpenAI Platform](https://platform.openai.com/api-keys)

### 📖 Hướng Dẫn Sử Dụng

#### Tạo Todo

**Phương Pháp Truyền Thống:**
```
/todo add Xem lại pull request #123
```

**🤖 Phương Pháp AI** (nếu đã bật):
```
/todo Gọi cho John lúc 3 giờ chiều ngày mai khẩn cấp
/todo Sửa lỗi server trước thứ 6 ưu tiên cao
/todo Xem lại tài liệu
```

AI tự động nhận diện:
- **Mô tả công việc**: Nội dung chính
- **Hạn chót**: Thời gian tuyệt đối hoặc tương đối
- **Ưu tiên**: Cao, Trung bình, Thấp (từ ngữ cảnh)

#### Quản Lý Tác Vụ

| Lệnh | Mô Tả |
|------|-------|
| `/todo` | Mở danh sách todo |
| `/todo add <nội dung>` | Tạo todo mới |
| `/todo list` | Xem tất cả todo |
| `/todo pop` | Hoàn thành todo cũ nhất |
| `/todo send @user <nội dung>` | Giao việc cho ai đó |
| `/todo settings` | Cấu hình nhắc nhở |

#### Sử Dụng Thanh Bên

1. Nhấp vào biểu tượng **Todo** ở thanh bên phải
2. **Thêm Todo**: Nhấp nút "+"
   - Đặt ưu tiên (Thấp/Trung bình/Cao)
   - Đặt hạn chót
   - Thêm mô tả
3. **Bình Luận**: Nhấp vào todo để xem/thêm bình luận
4. **Hoàn Thành**: Tích vào ô để đánh dấu hoàn thành

### 🔧 Tính Năng Nâng Cao

#### Cộng Tác
- **Giao Việc**: Ủy quyền cho thành viên với `/todo send @user Mô tả công việc`
- **Yêu Cầu Đến**: Chấp nhận hoặc từ chối việc được giao
- **Thông Báo**: Nhận cập nhật qua Todo bot

#### Nhắc Nhở Hàng Ngày
Bật nhắc nhở hàng ngày trong cài đặt để nhận tóm tắt các việc chưa hoàn thành mỗi sáng.

#### Bình Luận & Thảo Luận
- Nhấp vào bất kỳ todo nào để mở chuỗi bình luận
- Thêm ngữ cảnh, cập nhật hoặc đặt câu hỏi
- Tất cả bình luận được theo dõi trong nhật ký kiểm toán

### 🛠️ Phát Triển

#### Yêu Cầu
- Go 1.22+
- Node.js 18+
- Mattermost Server 6.5+

#### Build Từ Source

```bash
# Build server binaries
cd server
go build -o dist/plugin-linux-amd64

# Build webapp
cd webapp
npm install
npm run build

# Đóng gói plugin
make dist
```

#### Chạy Tests

```bash
# Server tests
cd server
go test ./...

# Webapp tests
cd webapp
npm test
```

### 📝 Giấy Phép

Dự án này được cấp phép theo MIT License.

### 🤝 Đóng Góp

Rất hoan nghênh các đóng góp! Vui lòng tạo Pull Request.

---

## Troubleshooting / Khắc Phục Sự Cố

### AI Features Not Working / Tính Năng AI Không Hoạt Động

**English:**
- Verify your OpenAI API key is correct
- Check that "Enable Smart Todo" is set to `true`
- Ensure your API key has sufficient credits
- Check Mattermost logs for detailed error messages

**Tiếng Việt:**
- Xác minh API key OpenAI của bạn đúng
- Kiểm tra "Enable Smart Todo" đã đặt thành `true`
- Đảm bảo API key có đủ credits
- Kiểm tra log Mattermost để xem thông báo lỗi chi tiết

### Database Connection Issues / Vấn Đề Kết Nối Database

**English:**
- Plugin automatically uses your Mattermost database configuration
- Check Mattermost logs for SQL connection errors
- Verify PostgreSQL/MySQL is running and accessible

**Tiếng Việt:**
- Plugin tự động sử dụng cấu hình database của Mattermost
- Kiểm tra log Mattermost để tìm lỗi kết nối SQL
- Xác minh PostgreSQL/MySQL đang chạy và có thể truy cập

---

**Repository**: [github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo](https://github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo)

**Support**: [GitHub Issues](https://github.com/CONG-TY-ANH-VA-EM/mattermost-plugin-todo/issues)
