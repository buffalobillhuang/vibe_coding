# World Museums Frontend

Next.js application displaying an interactive China map with museum information.

## Setup

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## Environment Variables

Create `.env.local`:
```
NEXT_PUBLIC_API_URL=http://localhost:8000/api
```

## Tech Stack

- Next.js 14 (React framework)
- TypeScript
- Tailwind CSS
- Leaflet + React-Leaflet (map library)
- Axios (HTTP client)

## Project Structure

```
src/
├── components/      # Reusable UI components
│   ├── ChinaMap.tsx
│   ├── MuseumPanel.tsx
│   └── MuseumCard.tsx
├── services/        # API clients and types
│   ├── api.ts
│   └── types.ts
├── styles/          # Global CSS
└── pages/           # Next.js pages
```

## Development

The frontend communicates with the backend API at `NEXT_PUBLIC_API_URL`.
Make sure your backend is running before starting the frontend.
