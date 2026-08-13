Here is a clean, professional `README.md` for your project. I have included a dedicated "Demo" section at the top with placeholder image tags where you can drop in your screenshots.

# 📸 Live Photo Booth

A lightweight, zero-dependency Live Photo Booth server written in Go. It runs as a single binary on Windows, macOS, or Linux. 

Guests can upload photos from their phones, and the photos instantly appear on a live presentation screen. It includes a built-in admin panel to moderate photos and adjust display settings on the fly.

## ✨ Features
* **Zero Dependencies:** Compiles to a single binary with all HTML/JS embedded.
* **Instant Sync:** New uploads instantly jump to the screen without waiting.
* **Admin Moderation:** One-click approve/deny photos in real-time.
* **Two Modes:** Choose between "Auto-allow" (blacklist mode) or "Require approval" (whitelist mode).
* **Dual-Port Security:** Public uploads run on port `8080`, while the private presentation and admin panel run on port `8081`.

---

## 🚀 Demo

<!-- Add your demo images to a 'docs' or 'assets' folder and update the paths below -->

### Upload Screen (Mobile)
> *Placeholder: Add an image of the mobile upload UI here*
> 
> `![Upload Screen UI](path/to/your/upload-image.png)`

### Admin Panel
> *Placeholder: Add an image of the admin panel with the "Now Showing" badge here*
> 
> `![Admin Panel UI](path/to/your/admin-image.png)`

### Live Display
> *Placeholder: Add an image of the live presentation screen here*
> 
> `![Live Display UI](path/to/your/display-image.png)`

---

## 🛠 Getting Started

### 1. Run the Server
Simply execute the binary for your OS. It will automatically create an `uploads/` folder in the same directory to store the photos.

```bash
# On Linux / macOS
./photobooth

# On Windows
photobooth.exe

```

### 2. Access the Application

The server runs on two separate ports to keep your admin tools secure:

* **Upload Page (Public):** `http://localhost:8080/`
* **Live Display (Private):** `http://localhost:8081/show`
* **Admin Panel (Private):** `http://localhost:8081/admin`

---

## 🌍 Exposing the Upload Page via Tailscale

To let guests upload photos without being on your local Wi-Fi, you can use [Tailscale Funnel](https://tailscale.com/kb/1223/funnel) to expose **only** the upload port (`8080`), keeping your admin panel securely hidden.

Run this script to start the funnel:

```bash
#!/bin/bash
# Route public internet traffic to local port 8080
tailscale serve https / [http://127.0.0.1:8080](http://127.0.0.1:8080)
# Turn on the funnel
tailscale funnel 8080 on
```

Tailscale will generate a public URL (e.g., `https://your-machine.tailnet-name.ts.net`) that you can turn into a QR code for your guests!

---

## 💻 Building from Source

To compile the application yourself, you need [Go](https://go.dev/) installed.

```bash
# Clone the repository and cd into it
# Build for your current OS:
go build -o photobooth main.go

```

**Cross-compiling for macOS from Linux/Windows:**

```bash
# For Apple Silicon (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o photobooth-mac main.go

# For Intel Macs
GOOS=darwin GOARCH=amd64 go build -o photobooth-mac-intel main.go

```

### ⚠️ Note for macOS Users

If you download the pre-compiled Mac binary, macOS Gatekeeper may block it. To run it, open your terminal and run:

```bash
# Make it executable
chmod +x photobooth-mac
# Remove the quarantine flag
xattr -d com.apple.quarantine photobooth-mac

```

Then, execute it normally with `./photobooth-mac`.
