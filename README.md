# 🧪 Interactive Code Execution Engine (Sandboxed)

A **production-grade, interactive code execution engine** built in **Go**, designed to safely run untrusted user code inside **isolated Docker containers** with **real-time stdin/stdout streaming over WebSockets**.

This system is inspired by how **online IDEs, coding interview platforms, and cloud sandboxes** work internally.

---

## ✨ Features

* ⚡ **Interactive execution**

  * Real-time `stdin`, `stdout`, `stderr`
  * Send input after execution starts
* 🔌 **WebSocket-based streaming**

  * Live output chunks (not buffered until exit)
* 🧠 **Session lifecycle management**

  * Created → Running → Finished / Terminated
* 🐳 **Strong Docker isolation**
* 🔐 **Sandboxed execution**

  * CPU, memory, process, disk, and network limits
* 🧼 **Automatic cleanup**

  * Containers and sessions are always removed
* 🌍 **Frontend-agnostic**

  * Works with browser, CLI, or any WS client
* 🧱 **Extensible architecture**

  * Multi-language ready (Python implemented)

---

## 🛡️ Security & Sandboxing

The engine is hardened against **common sandbox attacks**:

| Threat               | Protection               |
| -------------------- | ------------------------ |
| Fork bombs           | `PidsLimit`              |
| CPU exhaustion       | `NanoCPUs`               |
| Memory bombs         | cgroup memory limit      |
| Disk filling         | `tmpfs` workspace        |
| Network abuse        | Network disabled         |
| Privilege escalation | Dropped capabilities     |
| Infinite execution   | Idle + execution timeout |
| Output flooding      | Output size cap          |

### Docker-level protections used

* `ReadonlyRootfs`
* `CapDrop: ALL`
* `no-new-privileges`
* `NetworkMode: none`
* `PidsLimit`
* `NanoCPUs`
* `tmpfs` for `/workspace` and `/tmp`

This is **real-world sandboxing**, not a demo.

---

## 🧠 Architecture Overview

```
Browser / Client
        │
        │ WebSocket (stdin / stdout)
        ▼
   HTTP + WS API (Gin)
        │
        ▼
     Engine
   (session manager)
        │
        ▼
     Session
 (lifecycle, timers)
        │
        ▼
   Executor
 (Docker runtime)
        │
        ▼
  Isolated Container
```

### Responsibility split

* **Engine**

  * Session registry
  * Lifecycle orchestration
* **Session**

  * Timeouts
  * Activity tracking
  * Output limits
  * Cancellation
* **Executor**

  * Docker container creation
  * Attach stdin/stdout
  * Enforce OS-level isolation

---

## 📁 Project Structure

```
.
├── cmd/
│   └── server/           # main.go (HTTP + WS server)
│
├── internal/
│   ├── engine/           # Engine interface + implementation
│   ├── executor/         # Docker execution logic
│   ├── session/          # Session lifecycle, timers, limits
│   ├── language/         # Language specs (Python)
│   └── modules/          # Request/response models
│
├── frontend/
│   └── index.html        # Simple browser-based terminal UI
│
├── README.md
└── go.mod
```

---

## 🚀 Getting Started

### Prerequisites

* Go **1.21+**
* Docker (running)
* Linux / macOS (Docker Desktop)

---

### 1️⃣ Clone the repository

```bash
git clone https://github.com/your-username/interactive-execution-engine.git
cd interactive-execution-engine
```

---

### 2️⃣ Run the server

```bash
go run ./cmd/server
```

Server starts on:

```
http://localhost:8080
```

---

### 3️⃣ Open the frontend

Open `frontend/index.html` directly in the browser.

You can:

* Write Python code
* Run it
* Send interactive input
* See real-time output

---

## 🧪 Example: Interactive Python Code

```python
import time

print("Program started")
name = input("Enter your name: ")
print("Hello", name)

for i in range(5):
    time.sleep(1)
    print(i)
```

### Output (streamed live)

```
Program started
Enter your name: Nikesh
Hello Nikesh
0
1
2
3
4
```

---

## 🔄 API Overview

### Create Session (HTTP)

```
POST /session
```

```json
{
  "language": "python",
  "code": "print('hello')"
}
```

Response:

```json
{
  "sessionId": "uuid"
}
```

---

### Attach WebSocket

```
GET /ws/session/{sessionId}
```

#### Client → Server

```json
{
  "type": "input",
  "data": "hello\n"
}
```

#### Server → Client

```json
{
  "type": "stdout",
  "data": "hello\n"
}
```

```json
{
  "type": "state",
  "state": "FINISHED"
}
```

---

## ⏱️ Timeouts & Limits

| Limit        | Default       |
| ------------ | ------------- |
| Idle timeout | 30 seconds    |
| Max output   | 1 MB          |
| Memory       | 256 MB        |
| CPU          | 0.5 core      |
| Processes    | 32            |
| Disk         | 64 MB (tmpfs) |

---

## 🧯 Failure Handling

* WebSocket disconnect → session terminated
* Infinite loops → killed by timeout
* Fork bombs → blocked by PIDs limit
* Output flood → terminated early
* No zombie containers
