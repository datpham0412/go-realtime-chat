import router from './router'

router.beforeEach((to, from, next) => {
  const isAuthenticated = document.cookie.includes('user_id') || process.env.NODE_ENV === 'development'
  
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!isAuthenticated) {
      next({
        path: '/login'
      })
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router 