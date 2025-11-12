# RCC Web UI - Project Structure

## 📁 Directory Organization

```
RadioControlContainer/
├── rcc/                          # RCC Container (Go backend)
│   ├── cmd/                      # Main application
│   ├── internal/                  # Internal packages
│   ├── docs/                     # Architecture documentation
│   └── ...
├── web-ui/                       # RCC Web UI (This project)
│   ├── static/                   # Web assets
│   │   ├── index.html           # Main HTML page
│   │   ├── style.css            # CSS styles
│   │   └── app.js               # JavaScript application
│   ├── main.go                  # Go HTTP server
│   ├── config.json              # CB-TIMING v0.3 configuration
│   ├── go.mod                   # Go module
│   ├── rcc-webui               # Compiled binary
│   ├── README.md                # Documentation
│   ├── CHANGES.md               # Change log
│   ├── run-tests.sh            # Test runner
│   ├── open-firewall.sh        # Firewall helper
│   └── audit.log               # Audit log file
└── docs/                        # Shared documentation
```

## 🎯 **Web UI Components**

### **Frontend (Static Assets)**
- `static/index.html` - Desktop-first single page UI
- `static/style.css` - Accessible, responsive styles  
- `static/app.js` - OpenAPI v1 + SSE v1 client

### **Backend (Go Server)**
- `main.go` - HTTP server with reverse proxy
- `config.json` - CB-TIMING v0.3 configuration
- `rcc-webui` - Compiled binary

### **Documentation & Testing**
- `README.md` - Setup and usage guide
- `CHANGES.md` - Change history
- `run-tests.sh` - Automated test runner
- `open-firewall.sh` - Firewall configuration helper

## 🚀 **Quick Start**

```bash
cd RadioControlContainer/web-ui
./run-tests.sh
```

## 🌐 **Access URLs**

- **Local**: http://127.0.0.1:3000
- **Network**: http://192.168.1.120:3000
- **Alternative**: http://10.200.200.41:3000

## 📡 **Integration**

The Web UI connects to RCC container at `http://localhost:8080` and provides:
- Radio selection and control
- Power management (0-39 dBm)
- Channel selection (abstract 1,2,3...)
- Live telemetry monitoring
- Audit logging
