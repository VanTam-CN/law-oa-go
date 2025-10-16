# Law OA Frontend

This is the frontend application for the Law Office Automation system, built with React, TypeScript, and Bootstrap.

## Features

- User authentication (login/register)
- Dashboard with statistics
- Client management
- Case management
- User management (admin only)
- Profile management

## Tech Stack

- React 18
- TypeScript
- React Router v6
- Bootstrap 5
- React Bootstrap
- Axios

## Prerequisites

- Node.js 16+
- npm or yarn

## Getting Started

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start the development server:
   ```bash
   npm start
   ```

3. Build for production:
   ```bash
   npm run build
   ```

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
REACT_APP_API_BASE_URL=http://localhost:8080/api/v1
```

## Project Structure

```
src/
├── components/     # Reusable components
├── pages/          # Page components
├── services/       # API services
├── contexts/       # React contexts
├── types/          # TypeScript types
├── utils/          # Utility functions
├── assets/         # Static assets
├── styles/         # Global styles
└── hooks/          # Custom hooks
```

## Development

### Available Scripts

- `npm start` - Runs the app in development mode
- `npm test` - Launches the test runner
- `npm run build` - Builds the app for production
- `npm run eject` - Ejects the Create React App configuration

## API Integration

The frontend communicates with the backend API at `/api/v1` endpoints. All API calls are handled through the `apiClient` service which includes:

- Automatic token management
- Request/response interceptors
- Error handling
- Type-safe API responses

## Authentication

The app uses JWT tokens for authentication:

1. Login/register endpoints return access and refresh tokens
2. Tokens are stored in localStorage
3. Requests automatically include the access token
4. Expired tokens are refreshed automatically
5. Users are redirected to login on authentication errors

## Routing

The app uses React Router for navigation:

- `/login` - Login page
- `/register` - Registration page
- `/` - Main app layout
- `/dashboard` - Dashboard with statistics
- `/clients` - Client management
- `/cases` - Case management
- `/users` - User management (admin only)
- `/profile` - User profile management

## Styling

The app uses Bootstrap 5 for styling with React Bootstrap components. Custom styles are defined in:

- `src/index.css` - Global styles
- `src/App.css` - App-specific styles
- Component-specific styles in individual component files

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT