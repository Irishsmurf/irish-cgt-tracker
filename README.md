# 🇮🇪 Irish Personal Stock Portfolio & CGT Tracker

A self-hosted, containerized application to track RSU/GSU vests and calculate Capital Gains Tax (CGT) liability, strictly adhering to **Irish Revenue rules**.

## 🎯 Core Value Proposition (The "Irish Rule")

Most stock portfolio trackers calculate profit in USD ($) and then convert the net result to Euro (€). **This is incorrect for Irish Tax purposes.**

Irish Revenue requires a **Dual Conversion** method:
1.  **Acquisition Cost:** Converted to EUR using the ECB reference rate on the **date of vesting**.
2.  **Disposal Value:** Converted to EUR using the ECB reference rate on the **date of sale**.
3.  **Calculation:** `Gain/Loss = (EUR Disposal Value) - (EUR Cost Basis)`

This application automates this logic, including handling weekend/holiday exchange rate fallbacks (using the most recent previous trading day).

---

## 🚀 Quick Start (Docker)

This is the recommended way to run the application.

### 1. Configuration
Open `docker-compose.yml` and set your preferred credentials:
```yaml
environment:
  - APP_USER=admin
  - APP_PASSWORD=ChangeThisPassword123!
```

### 2. Run
```sh
docker compose up -d --build
```

### 3. Access
- **URL**: http://localhost:8080
- **Login**: Credentials from step 1.

## Local Development

If you want to run without Docker

1. Initialize
```sh
go mod download
```

2. Run: 
```sh
go run main.go
```

3. **Note**: The database will be created at `.data/portfolio.db`.

## User Guide
1. **Recording a Vest**
  - **Input**: Data, Symbol, Quantity, Strike Price (USD).
  - **System Actions**: Fetches historical ECB rate.
  - **Context**: Establishes Cost Basis.

2. **Recording a Sale**
  - **Input**: Date, Symbol, Sale Price (USD)
  - **System Actions**: Fetches the ECB Rate.
  - **Status**: Marked Unsettled.

3. **Calculating Tax (FIFO)**
  - Click **Calculate Tax** on an unsettled sale.
