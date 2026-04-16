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

## 📁 Project Structure

- `src/app/`: Next.js App Router pages and layouts.
- `src/components/`: Reusable UI components.
- `src/hooks/`: Custom React hooks for state and logic.
- `src/lib/`: Utility functions and shared constants.
- `src/types/`: TypeScript interfaces and types.

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
