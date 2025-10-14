# Praxis Explorer UI

A modern, beautiful dark-themed UI for the Praxis Explorer service, built with Next.js, TypeScript, and Tailwind CSS.

## Features

- 🎨 **Beautiful Dark Theme** - Inspired by prxs.ai official design
- 🔍 **Advanced Search** - Search agents by name, domain, skills
- 🎯 **Smart Filtering** - Filter by network, trust model, capabilities, skills, and tags
- ⚡ **Real-time Status** - See which agents are online
- 📊 **Agent Details** - Comprehensive view of agent information
- 🌈 **Smooth Animations** - Elegant transitions and hover effects
- 📱 **Responsive Design** - Works perfectly on all devices

## Tech Stack

- **Next.js 15** - React framework with App Router
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first styling
- **SWR** - Data fetching with caching
- **Lato Font** - Clean, modern typography

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn
- Praxis Explorer service running on `http://localhost:8080`

### Installation

```bash
# Install dependencies
npm install

# Run development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to see the application.

### Configuration

Configure the API endpoint in `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Project Structure

```
ui-new/
├── app/                  # Next.js app directory
│   ├── page.tsx         # Main explorer page
│   ├── layout.tsx       # Root layout with global styles
│   └── agent/           # Agent detail pages
├── components/          # React components
│   ├── Header.tsx       # Navigation header
│   ├── SearchBar.tsx    # Advanced search interface
│   ├── AgentCard.tsx    # Agent card component
│   └── LoadingSpinner.tsx
├── lib/                 # Utilities and API
│   ├── api.ts          # API client
│   └── utils.ts        # Helper functions
├── hooks/              # Custom React hooks
│   └── useAgents.ts    # SWR data fetching hook
├── styles/             # Global styles
│   └── globals.css     # Tailwind and custom CSS
└── types/              # TypeScript types
    └── agent.ts        # Agent data types
```

## Design System

### Colors
- **Background**: Pure black (#000000)
- **Primary**: Orange (#FF8562)
- **Secondary**: Cyan (#96EEEA)
- **Accent**: Blue (#9393FF)
- **Gray Scale**: Multiple shades for hierarchy

### Components
- Rounded corners (30px border radius)
- Gradient overlays for depth
- Glow effects on hover
- Smooth transitions (300ms)

## Available Scripts

```bash
npm run dev      # Start development server
npm run build    # Build for production
npm run start    # Start production server
npm run lint     # Run ESLint
```

## API Integration

The UI connects to the Praxis Explorer service API:

- `GET /agents` - Search and filter agents
- `GET /agents/:chainId/:agentId` - Get agent details
- `POST /admin/refresh` - Refresh agent data

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

MIT
