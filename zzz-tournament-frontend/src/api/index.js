// API exports
export { default as apiClient, apiRequest, setAuthToken, clearAuthToken, checkApiHealth } from './client'
export { authAPI, default as auth } from './auth'
export { usersAPI, default as users } from './users'
export { roomsAPI, default as rooms } from './rooms'
export { tournamentsAPI, default as tournaments } from './tournaments'
export { heroesAPI, default as heroes } from './heroes'
export { wsClient, WS_CONFIG, default as websocket } from './websocket'

// Удобный объект со всеми API
export const api = {
  auth: authAPI,
  users: usersAPI,
  rooms: roomsAPI,
  tournaments: tournamentsAPI,
  heroes: heroesAPI,
  ws: wsClient
}

export default api