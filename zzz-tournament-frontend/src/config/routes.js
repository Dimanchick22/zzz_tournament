// Routes Configuration

// Route paths
export const ROUTES = {
    // Public routes
    HOME: '/',
    LOGIN: '/login',
    REGISTER: '/register',
    
    // Protected routes
    DASHBOARD: '/dashboard',
    
    // Rooms
    ROOMS: '/rooms',
    ROOM_DETAILS: '/rooms/:id',
    ROOM_CREATE: '/rooms/create',
    
    // Tournaments
    TOURNAMENT: '/tournament/:id',
    
    // Heroes
    HEROES: '/heroes',
    HERO_DETAILS: '/heroes/:id',
    
    // User
    PROFILE: '/profile',
    LEADERBOARD: '/leaderboard',
    
    // Admin (если понадобится)
    ADMIN: '/admin',
    ADMIN_USERS: '/admin/users',
    ADMIN_HEROES: '/admin/heroes',
    ADMIN_TOURNAMENTS: '/admin/tournaments',
    
    // Error pages
    NOT_FOUND: '/404',
    UNAUTHORIZED: '/401',
    FORBIDDEN: '/403'
  }
  
  // Route builders - функции для создания путей с параметрами
  export const buildRoute = {
    roomDetails: (id) => `/rooms/${id}`,
    tournament: (id) => `/tournament/${id}`,
    heroDetails: (id) => `/heroes/${id}`,
    adminUsers: () => '/admin/users',
    adminHeroes: () => '/admin/heroes',
    adminTournaments: () => '/admin/tournaments'
  }
  
  // Route metadata для навигации
  export const ROUTE_META = {
    [ROUTES.HOME]: {
      title: 'Главная',
      description: 'Главная страница ZZZ Tournament',
      icon: 'home',
      showInNav: false,
      requiresAuth: false,
      layout: 'public'
    },
    
    [ROUTES.LOGIN]: {
      title: 'Вход',
      description: 'Вход в систему',
      icon: 'login',
      showInNav: false,
      requiresAuth: false,
      layout: 'auth'
    },
    
    [ROUTES.REGISTER]: {
      title: 'Регистрация',
      description: 'Создание нового аккаунта',
      icon: 'user-plus',
      showInNav: false,
      requiresAuth: false,
      layout: 'auth'
    },
    
    [ROUTES.DASHBOARD]: {
      title: 'Дашборд',
      description: 'Панель управления',
      icon: 'dashboard',
      showInNav: true,
      requiresAuth: true,
      layout: 'main',
      order: 1
    },
    
    [ROUTES.ROOMS]: {
      title: 'Комнаты',
      description: 'Список игровых комнат',
      icon: 'users',
      showInNav: true,
      requiresAuth: true,
      layout: 'main',
      order: 2
    },
    
    [ROUTES.HEROES]: {
      title: 'Герои',
      description: 'Библиотека героев ZZZ',
      icon: 'sword',
      showInNav: true,
      requiresAuth: true,
      layout: 'main',
      order: 3
    },
    
    [ROUTES.LEADERBOARD]: {
      title: 'Рейтинг',
      description: 'Таблица лидеров',
      icon: 'trophy',
      showInNav: true,
      requiresAuth: true,
      layout: 'main',
      order: 4
    },
    
    [ROUTES.PROFILE]: {
      title: 'Профиль',
      description: 'Настройки профиля',
      icon: 'user',
      showInNav: false,
      requiresAuth: true,
      layout: 'main'
    },
    
    [ROUTES.TOURNAMENT]: {
      title: 'Турнир',
      description: 'Турнирная сетка',
      icon: 'tournament',
      showInNav: false,
      requiresAuth: true,
      layout: 'main'
    },
    
    [ROUTES.ROOM_DETAILS]: {
      title: 'Комната',
      description: 'Детали комнаты',
      icon: 'door-open',
      showInNav: false,
      requiresAuth: true,
      layout: 'main'
    },
    
    [ROUTES.ADMIN]: {
      title: 'Админ',
      description: 'Панель администратора',
      icon: 'shield',
      showInNav: false,
      requiresAuth: true,
      requiresAdmin: true,
      layout: 'admin'
    }
  }
  
  // Navigation configuration
  export const NAVIGATION = {
    // Главная навигация (sidebar)
    main: [
      {
        path: ROUTES.DASHBOARD,
        label: 'Дашборд',
        icon: 'dashboard',
        exact: true
      },
      {
        path: ROUTES.ROOMS,
        label: 'Комнаты',
        icon: 'users',
        badge: 'new' // Можно добавлять бейджи
      },
      {
        path: ROUTES.HEROES,
        label: 'Герои',
        icon: 'sword'
      },
      {
        path: ROUTES.LEADERBOARD,
        label: 'Рейтинг',
        icon: 'trophy'
      }
    ],
    
    // Пользовательское меню (header)
    user: [
      {
        path: ROUTES.PROFILE,
        label: 'Профиль',
        icon: 'user'
      },
      {
        action: 'logout',
        label: 'Выход',
        icon: 'logout',
        variant: 'danger'
      }
    ],
    
    // Админское меню
    admin: [
      {
        path: ROUTES.ADMIN_USERS,
        label: 'Пользователи',
        icon: 'users'
      },
      {
        path: ROUTES.ADMIN_HEROES,
        label: 'Герои',
        icon: 'sword'
      },
      {
        path: ROUTES.ADMIN_TOURNAMENTS,
        label: 'Турниры',
        icon: 'tournament'
      }
    ]
  }
  
  // Breadcrumb configuration
  export const BREADCRUMBS = {
    [ROUTES.DASHBOARD]: [],
    [ROUTES.ROOMS]: [
      { label: 'Главная', path: ROUTES.DASHBOARD }
    ],
    [ROUTES.ROOM_DETAILS]: [
      { label: 'Главная', path: ROUTES.DASHBOARD },
      { label: 'Комнаты', path: ROUTES.ROOMS },
      { label: 'Детали', path: null } // null = current page
    ],
    [ROUTES.HEROES]: [
      { label: 'Главная', path: ROUTES.DASHBOARD }
    ],
    [ROUTES.LEADERBOARD]: [
      { label: 'Главная', path: ROUTES.DASHBOARD }
    ],
    [ROUTES.PROFILE]: [
      { label: 'Главная', path: ROUTES.DASHBOARD }
    ],
    [ROUTES.TOURNAMENT]: [
      { label: 'Главная', path: ROUTES.DASHBOARD },
      { label: 'Комнаты', path: ROUTES.ROOMS },
      { label: 'Турнир', path: null }
    ]
  }
  
  // Route permissions
  export const ROUTE_PERMISSIONS = {
    // Публичные роуты
    public: [
      ROUTES.HOME,
      ROUTES.LOGIN,
      ROUTES.REGISTER,
      ROUTES.NOT_FOUND,
      ROUTES.UNAUTHORIZED,
      ROUTES.FORBIDDEN
    ],
    
    // Роуты для аутентифицированных пользователей
    authenticated: [
      ROUTES.DASHBOARD,
      ROUTES.ROOMS,
      ROUTES.ROOM_DETAILS,
      ROUTES.HEROES,
      ROUTES.HERO_DETAILS,
      ROUTES.LEADERBOARD,
      ROUTES.PROFILE,
      ROUTES.TOURNAMENT
    ],
    
    // Роуты только для админов
    admin: [
      ROUTES.ADMIN,
      ROUTES.ADMIN_USERS,
      ROUTES.ADMIN_HEROES,
      ROUTES.ADMIN_TOURNAMENTS
    ]
  }
  
  // Helper functions
  export const isPublicRoute = (path) => {
    return ROUTE_PERMISSIONS.public.includes(path)
  }
  
  export const isAuthenticatedRoute = (path) => {
    return ROUTE_PERMISSIONS.authenticated.includes(path)
  }
  
  export const isAdminRoute = (path) => {
    return ROUTE_PERMISSIONS.admin.includes(path)
  }
  
  export const requiresAuth = (path) => {
    return isAuthenticatedRoute(path) || isAdminRoute(path)
  }
  
  export const requiresAdmin = (path) => {
    return isAdminRoute(path)
  }
  
  // Get route meta by path
  export const getRouteMeta = (path) => {
    return ROUTE_META[path] || {
      title: 'Страница не найдена',
      description: '',
      icon: 'question',
      showInNav: false,
      requiresAuth: false
    }
  }
  
  // Get navigation items filtered by permissions
  export const getNavigationItems = (userRole = 'user') => {
    const items = [...NAVIGATION.main]
    
    if (userRole === 'admin') {
      items.push({
        path: ROUTES.ADMIN,
        label: 'Админ',
        icon: 'shield',
        children: NAVIGATION.admin
      })
    }
    
    return items
  }
  
  // Get breadcrumbs for route
  export const getBreadcrumbs = (path, params = {}) => {
    const breadcrumbs = BREADCRUMBS[path] || []
    
    // Заменяем параметры в breadcrumbs если нужно
    return breadcrumbs.map(crumb => ({
      ...crumb,
      path: crumb.path && params && Object.keys(params).length > 0
        ? crumb.path.replace(/:(\w+)/g, (match, param) => params[param] || match)
        : crumb.path
    }))
  }
  
  // URL helpers
  export const matchRoute = (pattern, path) => {
    const regex = new RegExp(
      '^' + pattern.replace(/:\w+/g, '([^/]+)').replace(/\*/g, '.*') + '$'
    )
    return regex.test(path)
  }
  
  export const extractParams = (pattern, path) => {
    const regex = new RegExp(
      '^' + pattern.replace(/:\w+/g, '([^/]+)') + '$'
    )
    const match = path.match(regex)
    
    if (!match) return {}
    
    const paramNames = (pattern.match(/:(\w+)/g) || []).map(p => p.slice(1))
    const params = {}
    
    paramNames.forEach((name, index) => {
      params[name] = match[index + 1]
    })
    
    return params
  }
  
  // Page titles for document.title
  export const getPageTitle = (path, params = {}) => {
    const meta = getRouteMeta(path)
    const baseTitle = 'ZZZ Tournament'
    
    if (meta.title) {
      return `${meta.title} | ${baseTitle}`
    }
    
    return baseTitle
  }
  
  export default {
    ROUTES,
    ROUTE_META,
    NAVIGATION,
    BREADCRUMBS,
    ROUTE_PERMISSIONS,
    buildRoute,
    isPublicRoute,
    isAuthenticatedRoute,
    isAdminRoute,
    requiresAuth,
    requiresAdmin,
    getRouteMeta,
    getNavigationItems,
    getBreadcrumbs,
    getPageTitle
  }