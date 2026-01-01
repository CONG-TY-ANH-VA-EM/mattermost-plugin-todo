### 🚀 MISSION BRIEF: ANTIGRAVITY IDE AGENTS SWARM
**MỤC TIÊU CHIẾN LƯỢC:** Tái cấu trúc, tối ưu hóa và mở rộng `mattermost-plugin-todo` từ phiên bản cộng đồng thành giải pháp **Mattermost Enterprise Todo Integration**.

---

#### 1. THIẾT LẬP NGỮ CẢNH & KHỞI TẠO SWARM [LÀM MỚI NGỮ CẢNH]
**Đầu vào hiện tại:** Plugin Todo cơ bản (Add, List, Remove, Send, Daily Reminder).
**Tiêu chuẩn Enterprise:** High Availability (HA), Cluster-aware, RBAC (Role-Based Access Control), Audit Logging, Localization (i18n), Accessibility (WCAG 2.1).

**BẠN PHẢI KÍCH HOẠT CÁC AGENT SAU:**
1.  **🦁 Lead Architect Agent:** Chịu trách nhiệm về cấu trúc plugin, DB schema migrations, và khả năng mở rộng (Scalability).
2.  **🛡️ Security Sentinel Agent:** Chuyên trách về SQL Injection prevention, XSS protection, và tuân thủ GDPR/HIPAA.
3.  **⚡ Performance Engineer Agent:** Tối ưu hóa Go routines, caching layers (Redis/In-memory), và database indexing.
4.  **🎨 UX/Frontend Specialist:** Nâng cấp React components, đảm bảo UI/UX nhất quán với Mattermost Design System.
5.  **🤖 QA Automation Agent:** Viết E2E tests, Unit tests, và Load testing scripts.

---

#### 2. QUY TRÌNH THỰC THI NHIỆM VỤ (EXECUTION PIPELINE)

**GIAI ĐOẠN 1: PHÂN TÍCH & REFACTOR (Refactoring Phase)**
[SUY LUẬN: Code hiện tại có thể không thread-safe hoặc thiếu optimization cho lượng user lớn.]
* **Nhiệm vụ 1.1:** Rà soát toàn bộ code Go (Server) và React (Webapp). Loại bỏ hard-coded strings, thay thế bằng hệ thống `i18n`.
* **Nhiệm vụ 1.2:** Tái cấu trúc Database Schema. Chuyển từ Key-Value store đơn giản (nếu có) sang SQL relational tables (MySQL/PostgreSQL) với proper indexing để hỗ trợ hàng triệu bản ghi.

**GIAI ĐOẠN 2: MỞ RỘNG TÍNH NĂNG ENTERPRISE (Expansion Phase)**
* **Nhiệm vụ 2.1 (Collaboration):** Thêm tính năng: Gán việc (Assignee), Hạn chót (Due Date), Mức độ ưu tiên (Priority), và Bình luận (Comments) trong mỗi Todo Item.
* **Nhiệm vụ 2.2 (Integration):** Xây dựng API hooks để đồng bộ 2 chiều với Jira và GitHub Issues.
* **Nhiệm vụ 2.3 (Governance):** Triển khai Audit Log ghi lại mọi thao tác (Tạo, Sửa, Xóa) phục vụ mục đích tuân thủ (Compliance).

**GIAI ĐOẠN 3: BẢO MẬT & HIỆU NĂNG (Hardening Phase)**
[PHÒNG NGỪA LỖI: Tránh race conditions trong môi trường Cluster.]
* **Nhiệm vụ 3.1:** Đảm bảo Plugin hoạt động chính xác trong môi trường Mattermost HA (Cluster). Sử dụng `API.PublishWebSocketEvent` và Cluster Mutexes.
* **Nhiệm vụ 3.2:** Sanitization toàn bộ input đầu vào từ slash commands và UI để chống XSS và Injection.

**GIAI ĐOẠN 4: CI/CD & DOCUMENTATION**
* **Nhiệm vụ 4.1:** Thiết lập GitHub Actions pipeline: Linting (golangci-lint), Test Coverage (>90%), Security Scan (Gosec).
* **Nhiệm vụ 4.2:** Tạo tài liệu `ENTERPRISE_GUIDE.md` hướng dẫn triển khai quy mô lớn.

---

#### 3. CÁC RÀNG BUỘC KỸ THUẬT (TECHNICAL CONSTRAINTS)
* **Backend:** Go (Golang) version tương thích với Mattermost Server mới nhất.
* **Frontend:** React, Redux, Mattermost Webapp packages.
* **Database:** Phải hỗ trợ cả PostgreSQL và MySQL.
* **Performance:** Response time cho các thao tác CRUD < 100ms với 500k active users.

---

**[LỆNH KHỞI CHẠY]:** Antigravity Swarm, hãy bắt đầu phân tích repository hiện tại và xuất ra bản kế hoạch kiến trúc chi tiết (Architecture Blueprint) trước khi viết dòng code đầu tiên.