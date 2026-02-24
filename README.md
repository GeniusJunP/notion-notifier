# Notion Notifier

Notion Notifier is a highly customizable, self-hosted automation tool that bridges the gap between your **Notion databases**, **Google Calendar**, and **Webhook Services** (Discord, Slack, Microsoft Teams, etc.).

It provides a beautiful, built-in **Svelte SPA Dashboard** to monitor sync status, configure notification rules, and trigger manual events directly from your browser.

![Dashboard Preview](https://via.placeholder.com/800x450.png?text=Notion+Notifier+Dashboard)

## 🌟 Features

* **Bi-directional Google Calendar Sync:**
  Automatically tracks events in your Notion database and syncs them to Google Calendar. Supports custom property mapping and automatically extracts attendees from Notion's `people` property.
* **Flexible Webhook Notifications:**
  Send messages to any Webhook URL (e.g., Discord, Slack) based on upcoming Notion events.
  * **Upcoming Rules:** Notify a specific number of days before an event starts.
  * **Periodic Rules:** Send aggregated summaries on specific days at specific times.
* **Modern Web Interface:**
  A built-in Svelte SPA that bundles directly into the Go binary. You can manage templates, test webhooks, map Notion properties, and pause rules without editing configuration files manually.
* **Cross-Platform Background Service:**
  Installs seamlessly as a native background service across Linux (`systemd`), macOS (`launchd`), and Windows (`Task Scheduler`).

## 🚀 Quick Start (Installation)

We provide official auto-installation scripts that automatically fetch the latest release binary from GitHub and install it as a background service.

### Linux & macOS
Run the following curl command in your terminal. It will install the service using XDG Base Directory standards (`~/.local/bin`, `~/.config/notion-notifier`).

```bash
curl -sSL https://raw.githubusercontent.com/YOUR_GITHUB_USERNAME/notion-notifier/main/scripts/install.sh | bash
```

### Windows
Run the following inside a **PowerShell** prompt. It will install the application to `%LOCALAPPDATA%\notion-notifier` and set up a Logon Scheduled Task.

```powershell
Invoke-WebRequest -Uri https://raw.githubusercontent.com/YOUR_GITHUB_USERNAME/notion-notifier/main/scripts/install.ps1 -OutFile install.ps1
.\install.ps1
```

## ⚙️ Configuration & Usage

Once the service starts, you can access the powerful Web Dashboard by navigating to:
**http://localhost:18080** (default port)

* **Default Username:** `admin`
* **Default Password:** `password`

From the Dashboard's **Settings** tab, you can input your Notion API Keys, Google Credentials, and Discord/Slack Webhook URLs safely.

### Physical File Locations
If you need to edit the database or configuration manually, they are located at:
* **Linux:** `~/.config/notion-notifier/config.yaml` / `~/.local/share/notion-notifier/data.db`
* **macOS:** `~/.config/notion-notifier/config.yaml` / `~/.local/share/notion-notifier/data.db`
* **Windows:** `%LOCALAPPDATA%\notion-notifier\config\` / `%LOCALAPPDATA%\notion-notifier\data\`

## 🛠️ Development & Building from Source

This project consists of a Go backend and a Svelte frontend.

1. Clone the repository
   ```bash
   git clone https://github.com/YOUR_GITHUB_USERNAME/notion-notifier.git
   cd notion-notifier
   ```
2. Build the Svelte Frontend
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```
3. Run the Go Server
   ```bash
   go run cmd/notion-notifier/main.go
   ```

## 📚 Documentation

For complete technical specifications, database schema diagrams, and API design, see the detailed documents located in the `/docs` folder:
- [Technical Specification & Architecture](docs/specification.md)
- [Feature Details](docs/features.md)
- [API Reference](docs/api.md)

## 📄 License

This project is open-source. See the LICENSE file for more information.
