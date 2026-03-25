# PhishSim - Phishing Campaign Simulation

A web-based phishing simulation platform. This tool enables to safely simulate phishing campaigns, track user interactions, and generate comprehensive reports.

## Features

### Campaign Management
- **Create Campaigns**: Define campaign parameters including sender name, subject line, and target recipients
- **List Campaigns**: View all campaigns with summary statistics
- **View Campaign Details**: Access detailed reports with per-target metrics
- **Send Campaigns**: Trigger email distribution to all targets via SMTP

### Tracking & Analytics
- **Email Open Tracking**: Know when recipients open phishing emails
- **Link Click Tracking**: Track which users click malicious links
- **Credential Submission**: Record when users submit credentials
- **Event Timeline**: View chronological record of all interactions
- **IP Logging**: Capture source IP addresses for security analysis

### Reporting
- **Campaign Reports**: Generate HTML reports with comprehensive statistics
- **Risk Scoring**: Automatic risk assessment per user
- **Engagement Metrics**: 
  - Open rate
  - Click rate
  - Submission rate
- **Report Export**: Download reports as standalone HTML files

### Admin Dashboard
- Clean, responsive web interface
- Real-time statistics
- Campaign overview with status badges
- Easy navigation between campaigns

## Architecture

PhishSim is built using:

- **Backend**: Go with Chi router framework
- **Frontend**: HTML templates with embedded CSS
- **Database**: SQLite for persistent storage
- **Email**: SMTP integration (tested with Mailtrap)
- **Tracking**: Pixel-based email open detection and query parameters for link tracking

### Component Overview

```
┌─────────────────────────────────────────┐
│         Admin Dashboard (Web UI)        │
├─────────────────────────────────────────┤
│  ├─ /admin/              (List campaigns)
│  ├─ /admin/new           (Create campaign)
│  └─ /admin/campaign/{id} (View campaign)
├─────────────────────────────────────────┤
│         Tracker Endpoints                │
│  ├─ /t/open              (Log opens)
│  ├─ /t/click             (Log clicks)
│  └─ /t/submit            (Log submissions)
├─────────────────────────────────────────┤
│      Core Services                       │
│  ├─ Admin Handler       (Campaign mgmt)
│  ├─ Mailer              (SMTP sending)
│  ├─ Reporter            (Analytics)
│  ├─ Tracker             (Event logging)
│  └─ Database            (SQLite)
└─────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.18 or higher
- SQLite 3
- SMTP server access (e.g., Mailtrap, Gmail, corporate mail server)

### Setup

1. **Clone the repository**

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the server**
   ```bash
   go run ./cmd/server/main.go
   ```

   The application will start on `http://localhost:8080`

4. **Access the admin dashboard**
   - Navigate to `http://localhost:8080/`

## Configuration

### SMTP Configuration

Edit `cmd/server/main.go` to configure your SMTP server:

```go
m := mailer.New(mailer.Config{
    Host:     "sandbox.smtp.mailtrap.io",  // SMTP server
    Port:     587,                          // SMTP port
    Username: "your-username",              // SMTP username
    Password: "your-password",              // SMTP password
    BaseURL:  "http://localhost:8080",     // Tracking URL base
}, database)
```

### Recommended Providers

- **Mailtrap** (Development/Testing): Sandbox SMTP for safe email testing
- **Gmail**: Requires app-specific passwords
- **Corporate Mail Server**: For internal deployments

## Usage

### Creating a Campaign

1. Navigate to **New Campaign**
2. Fill in campaign details:
   - **Campaign Name**: Identifier for internal reference
   - **Email Subject**: Subject line for phishing email
   - **Sender Name**: Display name in email
   - **Sender Email**: From address
   - **Targets**: Format: `email@example.com,FirstName,LastName` (one per line)
3. Click **Create Campaign**
4. You'll be redirected to the campaign view

### Sending a Campaign

1. From the campaign view, click **📧 Send Campaign**
2. Emails are sent to all targets immediately
3. Tracking pixel and click link are automatically embedded
4. Monitor real-time interactions on the campaign page

### Viewing Results

- **Campaign List**: Quick overview with key metrics
  - Targets count
  - Emails sent
  - Open rate
  - Click rate
  - Submission rate
  
- **Campaign Detail**: Comprehensive analytics
  - Per-user interaction history
  - Chronological event timeline
  - Risk scoring per user
  - IP addresses and timestamps

### Downloading Reports

1. View campaign details
2. Click **Download Report** button
3. HTML report is generated and downloaded

## Project Structure

```
Phising-Campaign-Simulation/
├── cmd/
│   ├── server/           # Main web server
│   ├── sendcampaign/     # CLI tool for sending campaigns
│   ├── report/           # CLI tool for generating reports
│   └── seed/             # Database seeding utility
├── internal/
│   ├── admin/            # Admin dashboard handlers
│   ├── campaign/         # Campaign logic
│   ├── db/               # Database layer
│   │   ├── db.go         # DB initialization
│   │   ├── queries.go    # SQL queries
│   │   └── schema.sql    # Database schema
│   ├── mailer/           # Email sending service
│   ├── reporter/         # Report generation
│   └── tracker/          # Tracking endpoints
├── web/
│   ├── static/           # Static assets (CSS, JS)
│   └── templates/        # HTML templates
│       ├── base.html     # Base layout
│       ├── campaigns.html # Campaign list
│       ├── campaign.html # Campaign detail
│       └── new.html      # New campaign form
├── doc/
│   └── README.md         # This file
├── config.yaml           # Configuration file
├── go.mod               # Go module definition
└── phishsim.db          # SQLite database (auto-created)
```

## API Endpoints

### Admin Dashboard

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/admin/` | List all campaigns |
| GET | `/admin/new` | Show new campaign form |
| POST | `/admin/new` | Create campaign |
| GET | `/admin/campaign/{id}` | View campaign details |
| POST | `/admin/campaign/{id}/send` | Send campaign emails |
| GET | `/admin/campaign/{id}/report` | Download HTML report |

### Tracking Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/t/open` | Log email open (tracking pixel) |
| GET | `/t/click` | Log link click |
| POST | `/t/submit` | Log credential submission |

### Tracking Parameters

**Email Open Tracking:**
```
GET /t/open?target_id=<target_id>&campaign_id=<campaign_id>
```

**Link Click Tracking:**
```
GET /t/click?target_id=<target_id>&campaign_id=<campaign_id>&url=<url>
```

**Credential Submission:**
```
POST /t/submit
{
  "target_id": "...",
  "campaign_id": "...",
  "username": "...",
  "password": "..."
}
```

## Database Schema

### Campaigns Table
```sql
CREATE TABLE campaigns (
  id TEXT PRIMARY KEY,
  name TEXT,
  subject TEXT,
  sender_name TEXT,
  sender_email TEXT,
  status TEXT,
  created_at TIMESTAMP
);
```

### Targets Table
```sql
CREATE TABLE targets (
  id TEXT PRIMARY KEY,
  campaign_id TEXT FOREIGN KEY,
  email TEXT,
  first_name TEXT,
  last_name TEXT
);
```

### Events Table
```sql
CREATE TABLE events (
  id TEXT PRIMARY KEY,
  campaign_id TEXT FOREIGN KEY,
  target_id TEXT FOREIGN KEY,
  event_type TEXT (email_sent, opened, clicked, submitted),
  ip_address TEXT,
  user_agent TEXT,
  timestamp TIMESTAMP
);
```

### Submissions Table
```sql
CREATE TABLE submissions (
  id TEXT PRIMARY KEY,
  target_id TEXT FOREIGN KEY,
  campaign_id TEXT FOREIGN KEY,
  username TEXT,
  password TEXT,
  created_at TIMESTAMP
);
```

## Development

### Running from Source

```bash
# Start the server
go run ./cmd/server/main.go

# In another terminal, send a campaign via CLI
go run ./cmd/sendcampaign/main.go <campaign_id>

# Generate a report
go run ./cmd/report/main.go <campaign_id>
```

### Building

```bash
# Build all binaries
go build -o bin/ ./cmd/...

# Run the server binary
./bin/server
```

### Database

- SQLite database is automatically created as `phishsim.db`
- Schema is initialized on first run
- Use `cmd/seed/main.go` to populate sample data

### Modifying Email Template

Edit the email template in `internal/mailer/mailer.go`:

```go
var emailTemplate = template.Must(template.New("email").Parse(`
  <!-- Your email HTML here -->
`))
```

Available template variables:
- `{{.FirstName}}` - Recipient first name
- `{{.TrackingPixel}}` - Tracking pixel URL
- `{{.ClickURL}}` - Phishing link with tracking
- `{{.SenderName}}` - Campaign sender name

## Security Considerations

⚠️ **Important**: This tool is designed for authorized security testing only.

- **Authorization**: Only use on systems you have explicit permission to test
- **Notification**: Ensure participants are informed this is a simulation (before/after)
- **Data Protection**: Credentials captured are for security awareness training only
- **HTTPS**: Deploy with proper SSL/TLS certificates in production
- **Access Control**: Restrict admin dashboard access to authorized personnel
- **Logging**: Keep audit logs of campaign activities
- **Data Retention**: Implement policies for secure data deletion

## Troubleshooting

### Emails not sending
- Verify SMTP credentials in `cmd/server/main.go`
- Check firewall allows outbound SMTP connections
- Ensure `BaseURL` is reachable for tracking

### Tracking not working
- Confirm BaseURL is accessible by email clients
- Check recipient email client opens images by default
- Review server logs for tracking endpoint errors

### Database errors
- Delete `phishsim.db` to reset and recreate schema
- Ensure write permissions in application directory

## License

This project is for authorized security testing and educational purposes only.

## Support

For issues or feature requests, contact your security team or project maintainer.

---

**Last Updated**: March 2026
