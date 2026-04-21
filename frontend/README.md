# Moedah POS Frontend

This is the modern, responsive web interface for the Moedah POS system. Built with Next.js 16 and Tailwind CSS 4, it provides a premium user experience for retail and restaurant operations.

## ✨ Key Features

- **Responsive Dashboard**: Beautifully designed overview of store performance.
- **Dynamic POS Interface**: Fast and intuitive checkout process for both retail and restaurant modes.
- **Real-time Reports**: Visual data representation using Recharts for sales, expenses, and cash flow.
- **Inventory Management**: Comprehensive tools for tracking stock, batches, and movements.
- **Store Settings**: Easy configuration for tax, currencies, and multi-store roles.
- **Theme Support**: Premium look and feel with full responsiveness across all devices.

## 🛠 Tech Stack

- **Framework**: [Next.js 16 (App Router)](https://nextjs.org/)
- **Runtime**: [React 19](https://react.dev/)
- **Styling**: [Tailwind CSS 4](https://tailwindcss.com/)
- **Icons**: [Lucide React](https://lucide.dev/)
- **Charts**: [Recharts](https://recharts.org/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)
- **Linting & Formatting**: ESLint & Prettier

## 🚀 Getting Started

### Prerequisites

- Node.js 18.x or later
- npm, yarn, or pnpm

### 1. Installation

```bash
cd moedah-pos/frontend
npm install
```

### 2. Environment Configuration

Create a `.env.local` file in the root of the `frontend` directory:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

### 3. Development Server

Run the development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the application.

## 🏛 Architecture & Project Structure

The frontend strictly adheres to a modular, decoupled architecture leveraging the Next.js App Router for optimal performance, layout compounding, and client/server logic separation.

### 1. 📂 Core Directory Layout
- `src/app/`: Handles all routing logic. Built using Next.js layout patterns and cleanly divided into public routes (e.g., login) and `(protected)` authenticated application modules.
- `src/components/ui/`: Contains stateless, highly reusable primitive visualization components (Buttons, Modals, Form Elements, Sync Widgets).
- `src/lib/api/`: The dedicated API Network Layer. Houses the centralized HTTP client wrapper (`client.ts`) and specific domain route abstractions (e.g., `transactionsApi`, `productsApi`).
- `src/hooks/`: Custom React hooks (e.g., `useOfflineTransaction`) encapsulating complex behaviors or lifecycle state independent of visual logic.
- `src/types/`: Centralized TypeScript schema definitions perfectly mirroring backend entity structs.

### 2. 🔌 External Data & Rendering
- Our infrastructure dynamically shifts between static optimization boundaries and powerful Interactive Client Components (`"use client"`) exactly where dashboard grids or the POS registers demand stateful interactivity.
- Communication with the backend executes securely via strict typed class wrappers in `src/lib/api`, unifying error handling and JWT injections.

### 3. 🌐 Offline-First Sync Engine
- **IndexedDB**: Powered by `Dexie.js` (`src/lib/dexie.ts`) to locally cache POS transactions and preserve data resiliency when Wi-Fi unexpectedly drops.
- **Optimistic UI**: Transactions utilize the `useOfflineTransaction` orchestrator to instantly process operations locally—ensuring a 0-latency checkout experience—while asynchronously syncing and resolving issues through the `SyncStatusWidget` background polling queues.

### 4. 🎨 Design System
- Standardized via **Tailwind CSS 4** for lightning-fast responsive grids.
- All styles enforce comprehensive thematic support relying structurally on `dark:` variants, prioritizing precise micro-animations and "glass" aesthetic standards across high-density elements.

## 🧹 Quality Control

Before submitting any changes, ensure your code passes the quality checks:

```bash
# Run all checks (type-check, lint, format)
npm run analyze

# Individual commands
npm run type-check   # TypeScript validation
npm run lint         # ESLint checking
npm run format:check # Prettier formatting check
```

---

## 📜 License

This project is proprietary and confidential. Unauthorized copying is strictly prohibited.
