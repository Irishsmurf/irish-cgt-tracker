# AI Context & Rules: Irish CGT Tracker

## Project Goal
A self-hosted Go application to track stock portfolios (RSU/GSU) and calculate Capital Gains Tax specifically for the **Irish Revenue (Revenue.ie)** context.

## Critical Business Logic (The "Irish Rule")
1.  **Dual Conversion:** We do NOT convert the USD profit to EUR.
    * We convert the **Acquisition Cost** to EUR using the exchange rate on the **Vesting Date**.
    * We convert the **Disposal Value** to EUR using the exchange rate on the **Sale Date**.
    * Gain/Loss = (EUR Disposal) - (EUR Cost).
2.  **Exchange Rates:**
    * Source: ECB Reference Rates (via Frankfurter API).
    * Fallback: If a date (e.g., Sunday) has no rate, use the most recent previous trading day (e.g., Friday).
3.  **Money Handling:**
    * NEVER use floats for currency storage.
    * ALWAYS use `int64` representing "cents" (USD) or "cents" (EUR).
    * Floats are ONLY allowed for Exchange Rates.

## Tech Stack
* **Language:** Go (Golang) 1.21+
* **DB:** SQLite (via `modernc.org/sqlite` - pure Go, no CGO).
* **Frontend:** HTMX + Go `html/template`.

## Project Structure
* `internal/models`: Struct definitions.
* `internal/db`: SQLite connectivity and schema.
* `internal/currency`: External API client (Frankfurter).
* `internal/portfolio`: Business logic (CRUD, computations).
