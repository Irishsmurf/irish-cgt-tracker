# System Design: Personal Stock Portfolio & CGT Tracker

## 1. Architecture
* **Monolithic Binary:** Single Go executable containing HTTP server, business logic, and database driver.
* **Containerization:** Distroless Docker image (<25MB).
* **Persistence:** Single SQLite file mounted via Docker volume.

## 2. Data Flow
1.  **User Input:** User inputs Vest or Sale data (USD).
2.  **Rate Fetch:** System *immediately* fetches ECB Rate for that date (async or blocking).
3.  **Persistence:** Data stored in `vests` or `sales` table with the frozen rate.
4.  **Reporting:** When calculating tax, the system uses the stored rate, ensuring historical immutability.

## 3. Mathematical Model (Irish CGT)

$$\text{EUR Cost Basis} = \sum (\text{Vest Qty} \times \text{Vest Price}_{USD} \times \text{Rate}_{\text{vest}})$$

$$\text{EUR Disposal} = (\text{Sale Qty} \times \text{Sale Price}_{USD} \times \text{Rate}_{\text{sale}})$$

$$\text{Chargeable Gain} = \text{EUR Disposal} - \text{EUR Cost Basis}$$

## 4. Key Components
* **Currency Service:** Handles 404 fallbacks for weekends/holidays.
* **Portfolio Service:** Orchestrates DB writes and Rate fetches.
* **Lot Matcher:** (Future Phase) Implements FIFO logic to link Sales to Vests.
