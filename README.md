# 🇮🇪 Irish Personal Stock Portfolio & CGT Tracker

A self-hosted, containerized application designed to meticulously track Restricted Stock Unit (RSU) and Global Stock Unit (GSU) portfolios. It automates the calculation of Capital Gains Tax (CGT) liability in strict accordance with the rules set by the Irish Revenue Commissioners.

## 🎯 Core Value Proposition: The "Irish Rule"

A common mistake in portfolio tracking is to calculate the net profit in a foreign currency (like USD) and then convert the final amount to Euro. **This approach is not compliant with Irish tax law.**

Irish Revenue mandates a **Dual Currency Conversion** methodology for calculating capital gains:

1.  **Acquisition Cost**: The value of the shares at the time of vesting must be converted to EUR using the European Central Bank (ECB) reference exchange rate from the **vesting date**.
2.  **Disposal Value**: The proceeds from the sale of the shares must be converted to EUR using the ECB reference rate from the **sale date**.
3.  **Gain/Loss Calculation**: The chargeable gain or loss is the difference between these two EUR figures: `Gain/Loss = (EUR Disposal Value) - (EUR Acquisition Cost)`.

This application correctly implements this critical logic. It automatically sources historical exchange rates from the Frankfurter.app API (which mirrors the ECB) and includes a fallback mechanism to find the most recent trading day's rate if the transaction date falls on a weekend or public holiday.

---

## ✨ Features

- **Correct CGT Calculation**: Implements the "Irish Rule" for accurate tax assessment.
- **Automated Exchange Rates**: Fetches historical EUR/USD rates automatically.
- **FIFO Accounting**: Matches sales to the earliest available vested shares (First-In, First-Out).
- **Secure**: Protected by a simple, configurable username/password login.
- **Containerized**: Easy to deploy and run anywhere using Docker.
- **Lightweight**: Built in Go with a simple HTMX frontend, ensuring minimal resource usage.

## 🚀 Quick Start (Docker)

Running the application with Docker is the recommended and simplest method.

### 1. Configuration
Before starting, open the `docker-compose.yml` file and set a secure username and password under the `environment` section:
```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - APP_USER=admin
      - APP_PASSWORD=YourSecurePasswordHere! # <-- Change this
```

### 2. Run
From your terminal in the project root, execute the following command:
```sh
docker compose up -d --build
```
This command will build the Docker image, create a container, and start the application in detached mode.

### 3. Access
- **URL**: Navigate to `http://localhost:8080` in your web browser.
- **Login**: Use the credentials you configured in `docker-compose.yml`.

The application's SQLite database will be persisted in the `./data` directory on your host machine, ensuring your data is safe across container restarts.

## 🛠️ Local Development

For those who wish to run the application directly without Docker.

### Prerequisites
- Go 1.21 or newer.

### 1. Install Dependencies
Download the required Go modules:
```sh
go mod download
```

### 2. Run the Application
Execute the following command to start the server:
```sh
go run main.go
```
The server will start on `http://localhost:8080`.

### 3. Environment Variables (Optional)
For local development, the application defaults to the credentials `admin` / `secret`. You can override these by setting environment variables:
```sh
export APP_USER="myuser"
export APP_PASSWORD="mypassword"
go run main.go
```

**Note**: The SQLite database file will be created at `./data/portfolio.db`.

## 📖 User Guide

The user interface is designed to be straightforward.

1.  **Recording a Vest (Acquisition)**
    - Navigate to the "Add New Vest" form.
    - **Input**: The date of the vest, the stock symbol (e.g., GOOGL), the quantity of shares, and the market price in USD at the time of vesting.
    - **System Action**: The application automatically fetches the historical ECB exchange rate for the vesting date and saves the record. This establishes the **Cost Basis** in EUR for this lot of shares.

2.  **Recording a Sale (Disposal)**
    - Navigate to the "Add New Sale" form.
    - **Input**: The date of the sale, the quantity of shares sold, and the sale price in USD per share.
    - **System Action**: The application fetches the ECB rate for the sale date and records the sale. Initially, the sale is marked as **"Unsettled"**.

3.  **Calculating Tax (Settlement)**
    - In the "Sales" table, find the unsettled sale you wish to process.
    - Click the **"Settle"** button.
    - **System Action**: The application applies the FIFO method to identify which vested shares were sold. It then performs the dual-currency conversion to calculate the precise chargeable gain or loss for that transaction and marks the sale as **"Settled"**. The results are logged in the console.
